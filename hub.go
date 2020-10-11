// Copyright 2013 The Gorilla WebSocket Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"fmt"

	"github.com/s00500/goloco/lgb"
)

// Hub maintains the set of active clients and broadcasts messages to the
// clients.
type Hub struct {
	// Registered clients.
	clients map[*Client]bool

	// Register requests from the clients.
	register chan *Client

	// Unregister requests from clients.
	unregister chan *Client

	// channel for lgb data to broadcast
	lgbChannel chan lgb.StateChange
}

func newHub(lgbChan chan lgb.StateChange) *Hub {
	return &Hub{
		register:   make(chan *Client),
		unregister: make(chan *Client),
		clients:    make(map[*Client]bool),
		lgbChannel: lgbChan,
	}
}

func (h *Hub) run() {
	for {
		select {
		case client := <-h.register:
			h.clients[client] = true
		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}
		case message := <-h.lgbChannel:
			// now transform channel to message:
			var sendString []byte
			if message.Loco != nil {
				sendString = []byte(fmt.Sprintf("lsc:%d:%d:%t", message.Number, message.Loco.Speed, message.Loco.Light))
			}
			if message.Acc != nil {
				sendString = []byte(fmt.Sprintf("asc:%d:%t", message.Number, message.Acc.State))
			}
			for client := range h.clients {
				select {
				case client.send <- sendString:
				default:
					close(client.send)
					delete(h.clients, client)
				}
			}
		}
	}
}
