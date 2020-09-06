package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/s00500/goloco/lgb"

	"github.com/gorilla/websocket"
)

var lgbSystem *lgb.System

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func reader(conn *websocket.Conn) {
	for {
		// read in a message
		messageType, p, err := conn.ReadMessage()
		if err != nil {
			log.Println(err)
			return
		}
		if messageType != websocket.TextMessage {
			log.Printf("Invalid command: %s\n", string(p))
			continue
		}

		commandParts := strings.Split(string(p), ":")
		if len(commandParts) < 3 {
			log.Printf("Invalid command: %s\n", string(p))
			continue
		}

		switch commandParts[0] {
		case "sa":
			if len(commandParts) < 3 {
				log.Printf("Invalid command: %s\n", string(p))
				continue
			}
			direction := false
			if commandParts[2] == "1" {
				direction = true
			}
			accessory, err := strconv.ParseUint(commandParts[1], 10, 8)
			if err != nil {
				log.Printf("Invalid accessory: %s\n", commandParts[1])
				continue
			}
			lgbSystem.SwitchFunction(uint8(accessory), direction)
		case "ll":
			if len(commandParts) < 2 {
				log.Printf("Invalid command: %s\n", string(p))
				continue
			}
			loco, err := strconv.ParseUint(commandParts[1], 10, 8)

			if err != nil {
				log.Printf("Invalid loco: %s\n", commandParts[1])
				continue
			}
			lgbSystem.LocoLight(uint8(loco))
		case "ls":
		//loco speed
		default:
			log.Printf("Invalid command: %s\n", string(p))
		}
	}
}

func wsEndpoint(w http.ResponseWriter, r *http.Request) {
	upgrader.CheckOrigin = func(r *http.Request) bool { return true }

	// upgrade this connection to a WebSocket
	// connection
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
	}

	log.Println("Client Connected")
	/*
			err = ws.WriteMessage(1, []byte("Hi Client!"))
			if err != nil {
				log.Println(err)
		  }
	*/
	// listen indefinitely for new messages coming
	// through on our WebSocket connection
	reader(ws)
}

func setupRoutes() {
	staticDir := http.FileServer(http.Dir("./static"))
	http.Handle("/", staticDir)
	http.HandleFunc("/ws", wsEndpoint)
}

func main() {
	fmt.Println("Starting Go-Loco")
	setupRoutes()
	portName := "/dev/tty.usbserial-146340"
	if len(os.Args) > 1 {
		portName = os.Args[1]
	}
	lgbSystem = &lgb.System{PortName: portName}

	if err := lgbSystem.Start(); err != nil {
		log.Fatal(err)
	}
	log.Fatal(http.ListenAndServe(":8080", nil))
}
