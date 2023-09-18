package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	m "github.com/elsif-maj/goChat/pkg/mongo"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/net/websocket"
)

var collection *mongo.Collection

type Server struct {
	conns map[*websocket.Conn]bool
}

func NewServer() *Server {
	return &Server{
		conns: make(map[*websocket.Conn]bool),
	}
}

func (s *Server) handleWS(ws *websocket.Conn) {
	fmt.Println("new incoming connection from client:", ws.RemoteAddr())

	s.conns[ws] = true

	s.readLoop(ws)
}

func (s *Server) readLoop(ws *websocket.Conn) {
	buf := make([]byte, 1024)
	for {
		n, err := ws.Read(buf)
		if err != nil {
			if err == io.EOF {
				break
			}
			fmt.Println("read error:", err)
			continue
		}
		msg := buf[:n]
		rmtAddrBytes := []byte(ws.RemoteAddr().String() + ": ")

		msg = append(rmtAddrBytes, msg...)
		m.WriteMsg(collection, msg)
		s.broadcast(msg)
	}
}

func (s *Server) broadcast(b []byte) {
	for ws := range s.conns {
		go func(ws *websocket.Conn) {
			if _, err := ws.Write(b); err != nil {
				fmt.Println("write error:", err)
				ws.Close() // Trying this on for size.
			}
		}(ws)
	}
}

func CORS(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Access-Control-Allow-Origin", "*")
		w.Header().Add("Access-Control-Allow-Credentials", "true")
		w.Header().Add("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		w.Header().Add("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")

		if r.Method == "OPTIONS" {
			http.Error(w, "No Content", http.StatusNoContent)
			return
		}

		next(w, r)
	}
}

func handleHome(collection *mongo.Collection) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		results := m.ReadMsgs(collection)

		// Marshal the results slice into JSON
		jsonData, err := json.Marshal(results)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Set the Content-Type header to indicate JSON
		w.Header().Set("Content-Type", "application/json")

		// Write the JSON data to the response writer
		_, err = w.Write(jsonData)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

func main() {
	server := NewServer()

	db, ctx := m.ConnectToMongo()
	collection = db.Database("goChat").Collection("goChatCapped")

	defer func() {
		if err := db.Disconnect(ctx); err != nil {
			log.Fatal(err)
		}
	}()

	http.Handle("/ws", websocket.Handler(server.handleWS))
	http.HandleFunc("/api/msgs", CORS(handleHome(collection)))

	log.Fatal(http.ListenAndServe(":3001", nil))
}
