package main

import (
	"encoding/json"
	"flag"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"time"
	"fmt"
	"strconv"
)

var addr = flag.String("addr", ":8080", "The inbound http port")

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

type Client struct {
	id int
	connection *websocket.Conn
	username   string
	room       string
}

type ClientMessage struct {
	MsgType string
	Data     map[string]string
}

type ServerMessage struct {
	MsgType string
	Data map[string]string
}

type ConnectionManager struct {
	// monitor clients and direct writes to the appropriate room(s)
	clients       map[int]*Client
	nextClient    int
	broadcastChan chan []byte
	}

func makeConnectionManager() *ConnectionManager {
	return &ConnectionManager{
		make(map[int]*Client),
		0,
		make(chan []byte, 256),
	}
}

func monitorReads(manager *ConnectionManager, connId int) {
	log.Println("Starting goroutine monitorReads")

	// on receive, broadcast it to all connections
	for {
		_, message, err := manager.clients[connId].connection.ReadMessage()
		if err != nil {
			break
		}
		var msgObj ClientMessage
		err2 := json.Unmarshal([]byte(message), &msgObj)
		if err2 != nil {
			break
		}

		log.Println(msgObj)
		// could implement UnmarshalJSON interface to make this unpack to nicer structs
		if msgObj.MsgType == "setup" {
			// do setup stuff
			manager.clients[connId].username = msgObj.Data["user"]
			manager.clients[connId].room = msgObj.Data["room"]

			messageData := ServerMessage{
				MsgType: "pm",
				Data: map[string]string {
					"id": fmt.Sprint(connId),
					"sender": "SERVER",
					"message": fmt.Sprintf("Welcome to #%s, %s!", msgObj.Data["room"], msgObj.Data["user"]),
				},
			}

			toSend, _ := json.Marshal(messageData)

			manager.broadcastChan <- toSend

		} else if msgObj.MsgType == "message" {
			// broadcast
			log.Println("Message received") 
			
			serve := ServerMessage{
				MsgType: "local",
				Data: map[string]string {
					"room": msgObj.Data["room"],
					"sender": msgObj.Data["sender"],
					"message": msgObj.Data["message"],
				},
			}
			serveStr, _ := json.Marshal(serve)
			manager.broadcastChan <- serveStr

		} else if msgObj.MsgType == "listRooms" {
			rooms := map[string]struct{}{}
			for _, client := range manager.clients {
				rooms[client.room] = struct{}{}
			}
			var roomNames []string
			for rn := range rooms {
				roomNames = append(roomNames,rn)
			}
			roomStr, _ := json.Marshal(roomNames)
			serve := ServerMessage{
				MsgType: "roomInfo",
				Data: map[string]string {
					"id": fmt.Sprint(connId), 
					"rooms": string(roomStr), 
				},
			}
			toSend, _ := json.Marshal(serve)
			manager.broadcastChan <- toSend

		}
		// then await next message
	}
}

func broadcast(manager *ConnectionManager, msg []byte) {
	for _, client := range manager.clients {
		writeToConnection(msg, client.connection)		
	}
}

func local(manager *ConnectionManager, room string, msg []byte) {
	for _, client := range manager.clients {
		if client.room == room {
			writeToConnection(msg, client.connection)
		}
	}
}

func pm(manager *ConnectionManager, id int, msg []byte) {
	conn := manager.clients[id].connection
	writeToConnection(msg, conn)
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

func monitorWrites(manager *ConnectionManager) {
	log.Println("Starting goroutine monitorWrites")
	for {
		message, ok := <-manager.broadcastChan
		if !ok {
			// broadcast "close" to all clients
			return
		}
		var msg ServerMessage
		json.Unmarshal(message, &msg)
		log.Println(msg)
		if msg.MsgType == "pm" || msg.MsgType == "roomInfo" {
			connId, _ := strconv.Atoi(msg.Data["id"])
			pm(manager, connId, message)
		} else if msg.MsgType == "local" {
			local(manager, msg.Data["room"], message)	
		} else if msg.MsgType == "broadcast" {
			broadcast(manager, message)
		}
	}
}

func servePage(page string) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Println(r.URL)
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		http.ServeFile(w, r, page)
	}
}

func serve(manager *ConnectionManager, writer http.ResponseWriter, request *http.Request) {
	conn, err := upgrader.Upgrade(writer, request, nil)
	if err != nil {
		log.Println(err)
		return
	}
	manager.clients[manager.nextClient] = &Client{
		id: manager.nextClient,
		connection: conn,
		username: "",
		room: "",	
	}
	// Reads and writes need to be two separate goroutines as they're both blocking
	go monitorReads(manager, manager.nextClient)
	manager.nextClient += 1
}

func main() {
	flag.Parse()
	// main loop

	// need an upgrader as part of the websocket spec
	// process is:
	// client requests ws via an http route
	// server responds with a 101 (change protocol)
	// - and registers client / connection
	// - the library handles this, but likely relevant for re-implementing in C if needed

	manager := makeConnectionManager()
	log.Println("Starting websockets server...")
	go monitorWrites(manager)

	http.HandleFunc("/", servePage("index.html"))
	http.HandleFunc("/chat", servePage("chat.html"))

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
