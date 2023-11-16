package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

var addr = flag.String("addr", "8080", "The inbound http port")
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

type ConnectionManager struct {
	connections map[int]*websocket.Conn
	nextId   int
}

func monitor(manager ConnectionManager, connId int) {
	// on receive, broadcast it to all connections
}

func serve(manager ConnectionManager, writer http.ResponseWriter, request *http.Request) {
	conn, err := upgrader.Upgrade(writer, request, nil)
	if err != nil {
		log.Println(err)
		return
	}
	manager.connections[manager.nextId] = conn
	go monitor(manager, manager.nextId)
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
	// -

	manager := ConnectionManager{
		make(map[int]*websocket.Conn),
		0,
	}

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

	fmt.Printf("Hello world!\n")
}
