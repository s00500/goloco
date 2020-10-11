package main

import (
	"flag"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/s00500/goloco/lgb"

	"github.com/gorilla/websocket"
	log "github.com/s00500/env_logger"
)

var lgbSystem *lgb.System

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

var resetAll bool

func init() {
	log.ConfigureDefaultLogger()
}

func reader(conn *websocket.Conn) {
	for {
		// read in a message
		messageType, p, err := conn.ReadMessage()
		if err != nil {
			log.Info(err)
			return
		}
		if messageType != websocket.TextMessage {
			log.Info("Invalid command: %s\n", string(p))
			continue
		}

		commandParts := strings.Split(string(p), ":")

		switch commandParts[0] {
		case "sa":
			if len(commandParts) < 3 {
				log.Info("Invalid command: %s\n", string(p))
				continue
			}
			direction := false
			if commandParts[2] == "1" {
				direction = true
			}
			accessory, err := strconv.ParseUint(commandParts[1], 10, 8)
			if err != nil {
				log.Info("Invalid accessory: ", commandParts[1])
				continue
			}
			lgbSystem.SwitchFunction(uint8(accessory), direction)
		case "ll":
			if len(commandParts) < 2 {
				log.Info("Invalid command: ", string(p))
				continue
			}
			loco, err := strconv.ParseUint(commandParts[1], 10, 8)

			if err != nil {
				log.Info("Invalid loco: ", commandParts[1])
				continue
			}
			lgbSystem.LocoLight(uint8(loco))
		case "ls":
			if len(commandParts) < 2 {
				log.Info("Invalid command: ", string(p))
				continue
			}
			loco, err := strconv.ParseUint(commandParts[1], 10, 8)

			if err != nil {
				log.Info("Invalid loco:", commandParts[1])
				continue
			}
			lgbSystem.LocoStop(uint8(loco))
		case "lf":
			if len(commandParts) < 2 {
				log.Info("Invalid command: ", string(p))
				continue
			}
			loco, err := strconv.ParseUint(commandParts[1], 10, 8)

			if err != nil {
				log.Info("Invalid loco: ", commandParts[1])
				continue
			}
			lgbSystem.LocoForward(uint8(loco))
		case "lb":
			if len(commandParts) < 2 {
				log.Info("Invalid command: ", string(p))
				continue
			}
			loco, err := strconv.ParseUint(commandParts[1], 10, 8)

			if err != nil {
				log.Info("Invalid loco: ", commandParts[1])
				continue
			}
			lgbSystem.LocoBackward(uint8(loco))
		case "lfun":
			if len(commandParts) < 3 {
				log.Info("Invalid command: ", string(p))
				continue
			}
			loco, err := strconv.ParseUint(commandParts[1], 10, 8)

			if err != nil {
				log.Info("Invalid loco: ", commandParts[1])
				continue
			}
			fun, err := strconv.ParseUint(commandParts[2], 10, 8)
			if err != nil {
				log.Info("Invalid loco: ", commandParts[1])
				continue
			}
			lgbSystem.LocoFunction(uint8(fun), uint8(loco))
		default:
			log.Info("Invalid command: ", string(p))
		}
	}
}

func wsEndpoint(hub *Hub, w http.ResponseWriter, r *http.Request) {
	upgrader.CheckOrigin = func(r *http.Request) bool { return true }

	// upgrade this connection to a WebSocket
	// connection
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Info(err)
	}

	client := &Client{hub: hub, conn: ws, send: make(chan []byte, 256)}
	client.hub.register <- client
	log.Info("Client Connected")
	/*
			err = ws.WriteMessage(1, []byte("Hi Client!"))
			if err != nil {
				log.Info(err)
		  }
	*/

	for _, locoState := range lgbSystem.GetLocoStates() {
		ws.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf("lsc:%d:%d:%t", uint8(locoState.Number), locoState.Loco.Speed, locoState.Loco.Light)))
	}

	accStateString := "as"
	for _, accState := range lgbSystem.GetAccessoryStates() {
		if accState {
			accStateString = accStateString + ":1"
		} else {
			accStateString = accStateString + ":0"
		}
	}
	ws.WriteMessage(websocket.TextMessage, []byte(accStateString))

	// make sure we get all broadcast messages from the hub!
	go client.writePump()

	// listen indefinitely for new messages coming
	// through on our WebSocket connection
	reader(ws)
}

func setupRoutes(hub *Hub) {
	staticDir := http.FileServer(http.Dir("./static"))
	http.Handle("/", staticDir)
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		wsEndpoint(hub, w, r)
	})
}

func main() {
	flag.BoolVar(&resetAll, "r", false, "Reset all accessories to false (right for switches)")

	flag.Parse()
	args := flag.Args()

	log.Info("Starting Go-Loco")
	portName := "/dev/tty.usbserial-146340"
	if len(args) > 0 {
		portName = args[0]
	}
	lgbSystem = &lgb.System{PortName: portName}

	if err := lgbSystem.Start(resetAll); err != nil {
		log.Fatal(err)
	}

	hub := newHub(lgbSystem.OutChannel)
	go hub.run()
	setupRoutes(hub)
	iface := ":8080"
	log.Info("Listening on ", iface)
	log.Fatal(http.ListenAndServe(iface, nil))
}
