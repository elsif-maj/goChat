package mongo

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MsgEntry struct {
	ID  primitive.ObjectID `bson:"_id"`
	Msg string             `bson:"msg"`
}

type WriteMsgEntry struct {
	Msg string `bson:"msg"`
}

func ConnectToMongo() (*mongo.Client, context.Context) {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file: ", err)
	}

	// Set up a context with a timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Define MongoDB connection options
	clientOptions := options.Client().ApplyURI(os.Getenv("MONGO_SIGNIN"))

	// Connect to MongoDB
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		log.Fatal(err)
	}

	// Ping the MongoDB server to check if the connection is successful
	err = client.Ping(ctx, nil)
	if err != nil {
		log.Fatal(err)
	}

	// You are now connected to MongoDB!
	fmt.Println("Connected to MongoDB")

	return client, ctx
}

func WriteMsg(c *mongo.Collection, msg []byte) error {
	document := WriteMsgEntry{Msg: string(msg)}
	// Set up a context with a timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Insert the document into the collection
	_, err := c.InsertOne(ctx, document)
	if err != nil {
		fmt.Println("here!")
		return err
	}

	return nil
}

func ReadMsgs(c *mongo.Collection) []MsgEntry {
	// Set up a context with a timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Find all documents (match all)
	cursor, err := c.Find(ctx, bson.D{})
	if err != nil {
		log.Fatal(err)
	}
	defer cursor.Close(ctx)

	results := []MsgEntry{}

	// Iterate through the results
	for cursor.Next(ctx) {
		var result MsgEntry // You can define a struct that matches your document structure here instead of bson.M
		if err := cursor.Decode(&result); err != nil {
			log.Fatal(err)
		}
		results = append(results, result)
	}

	// Check for cursor errors
	if err := cursor.Err(); err != nil {
		log.Fatal(err)
	}

	return results
}
