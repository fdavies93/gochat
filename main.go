package main

import (
	"flag"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"time"
)

var addr = flag.String("addr", ":8080", "The inbound http port")

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

type ConnectionManager struct {
	connections   map[int]*websocket.Conn
	nextId        int
	broadcastChan chan []byte
}

func monitorReads(manager ConnectionManager, connId int) {
	log.Println("Starting goroutine monitorReads")

	// on receive, broadcast it to all connections
	for {
		_, message, err := manager.connections[connId].ReadMessage()
		if err != nil {
			break
		}
		manager.broadcastChan <- message
		// then await next message
	}
}

func broadcast(manager ConnectionManager, msg []byte) {
	for _, conn := range manager.connections {
		writeToConnection(msg, conn)
	}
}

func writeToConnection(message []byte, connection *websocket.Conn) {
	w, err := connection.NextWriter(websocket.TextMessage)
	if err != nil {
		// raise error? close channel?
		return
	}
	w.Write(message)
	if err := w.Close(); err != nil {
		return
	}

}

func monitorWrites(manager ConnectionManager) {
	log.Println("Starting goroutine monitorWrites")
	for {
		message, ok := <-manager.broadcastChan
		if !ok {
			// broadcast "close" to all clients
			return
		}
		broadcast(manager, message)
	}
}

func serve(manager ConnectionManager, writer http.ResponseWriter, request *http.Request) {
	conn, err := upgrader.Upgrade(writer, request, nil)
	if err != nil {
		log.Println(err)
		return
	}
	manager.connections[manager.nextId] = conn
	// Reads and writes need to be two separate goroutines as they're both blocking
	go monitorReads(manager, manager.nextId)
	manager.nextId += 1
}

func main() {
	flag.Parse()
	// main loop

	// need an upgrader as part of the websocket spec
	// process is:
	// client requests ws via an http route
	// server responds with a 101 (change protocol)
	// - and registers client / connection
	// - the library handles this, but likely relevant for re-implementing in C

	manager := ConnectionManager{
		make(map[int]*websocket.Conn),
		0,
		make(chan []byte, 256),
	}

	log.Println("Starting websockets server...")
	go monitorWrites(manager)

	http.HandleFunc("/ws", func(writer http.ResponseWriter, request *http.Request) {
		serve(manager, writer, request)
		// setup the websocket
	})

	server := &http.Server{
		Addr:              *addr,
		ReadHeaderTimeout: 3 * time.Second,
	}

	err := server.ListenAndServe()
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}

}
