package main

import (
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

type Client struct {
	ID       string
	Conn     *websocket.Conn
	Room     string
	Username string
}

type Message struct {
	Type      string      `json:"type"`
	Sender    string      `json:"sender"`
	Recipient string      `json:"recipient,omitempty"`
	Room      string      `json:"room,omitempty"`
	Username  string      `json:"username,omitempty"`
	Data      interface{} `json:"data,omitempty"`
}

var (
	clients    = make(map[string]*Client)
	rooms      = make(map[string]map[string]*Client)
	register   = make(chan *Client)
	unregister = make(chan *Client)
	broadcast  = make(chan Message)
	upgrader   = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
	mutex = &sync.Mutex{}
)

func main() {
	// Start hub to manage websocket connection
	go hub()

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("WebRTC Signaling Server"))
	})

	http.HandleFunc("/ws", handleConnections)

	port := ":8080"

	fmt.Printf("Server starting on port %s\n", port)

	err := http.ListenAndServe(port, nil)

	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}

}

func handleConnections(w http.ResponseWriter, r *http.Request) {
	// Upgrade the http connection to a WebSocket connection
	conn, err := upgrader.Upgrade(w, r, nil)

	if err != nil {
		log.Printf("Error upgrading to WebSocket: %v\n", err)
		return
	}

	// Register new client
	clientID := generateUniqueID()
	client := &Client{
		ID:   clientID,
		Conn: conn,
	}

	initMsg := Message{
		Type: "init",
		Data: map[string]string{
			"clientId": clientID,
		},
	}

	err = conn.WriteJSON(initMsg)

	if err != nil {
		log.Printf("Error sending init message: %v", err)
		conn.Close()
		return
	}

	go handleMessages(client)

}

func handleMessages(client *Client) {
	defer func() {
		unregister <- client
		client.Conn.Close()
	}()

	for {
		var msg Message
		err := client.Conn.ReadJSON(&msg)

		if err != nil {
			log.Printf("Error reading message: %v", err)
			break
		}

		switch msg.Type {
		case "join":
			// Get room and username from the message
			roomID, ok := msg.Data.(map[string]interface{})["room"].(string)

			if !ok {
				continue
			}

			username, ok := msg.Data.(map[string]interface{})["username"].(string)

			if !ok {
				username = "Anonymous"
			}

			// Updating client details
			client.Room = roomID
			client.Username = username

			// Registering the client
			register <- client

			// Notify users about peers in the room
			notifyPeers(client)

		case "offer", "answer", "ice-candidate":
			// Forward signaling messages to the specific recipient
			msg.Sender = client.ID
			broadcast <- msg

		case "leave":
			unregister <- client
			return

		}
	}
}

func notifyPeers(client *Client) {
	mutex.Lock()
	defer mutex.Unlock()

	// Check if room exists
	if _, ok := rooms[client.Room]; !ok {
		return
	}

	// Send list of other participants to new client
	peers := []map[string]string{}

	for id, peer := range rooms[client.Room] {
		if id != client.ID {
			peers = append(peers, map[string]string{
				"id":       id,
				"username": peer.Username,
			})
		}
	}

	// Send peers list to the new user
	joinResponse := Message{
		Type: "room-joined",
		Data: map[string]interface{}{
			"room":  client.Room,
			"peers": peers,
		},
	}

	err := client.Conn.WriteJSON(joinResponse)

	if err != nil {
		log.Printf("Error sending room-joined message: %v\n", err)
		return
	}

	// Notify other peers about new participant
	newPeerMsg := Message{
		Type:   "new-peer",
		Sender: client.ID,
		Data: map[string]string{
			"id":       client.ID,
			"username": client.Username,
		},
	}

	for id, peer := range rooms[client.Room] {
		if id != client.ID {
			err := peer.Conn.WriteJSON(newPeerMsg)
			if err != nil {
				log.Printf("Error sending new-peer message: %v", err)
			}
		}
	}
}

func hub() {
	for {
		select {
		case client := <-register:
			mutex.Lock()

			// Add to clients map
			clients[client.ID] = client

			// Add to room
			if _, ok := rooms[client.Room]; !ok {
				rooms[client.Room] = make(map[string]*Client)
			}
			rooms[client.Room][client.ID] = client
			mutex.Unlock()

			log.Printf("Client %s joined room %s as %s", client.ID, client.Room, client.Username)

		case client := <-unregister:
			mutex.Lock()

			// Remove from client map
			delete(clients, client.ID)

			// Remove from the room
			if _, ok := rooms[client.Room]; ok {
				delete(rooms[client.Room], client.ID)

				// If room is empty, remove it
				if len(rooms[client.Room]) == 0 {
					delete(rooms, client.Room)
				} else {
					// Notify remaining peers about user leaving
					peerLeftMsg := Message{
						Type:   "peer-left",
						Sender: client.ID,
					}

					for _, peer := range rooms[client.Room] {
						err := peer.Conn.WriteJSON(peerLeftMsg)
						if err != nil {
							log.Printf("Error sending peer-left message: %v\n", err)
						}
					}
				}
			}
			mutex.Unlock()

			log.Printf("Client %s left room %s", client.ID, client.Room)

		case message := <-broadcast:
			mutex.Lock()

			// Check if recipient is specified
			if message.Recipient != "" {
				// Direct message to specific recipient
				if client, ok := clients[message.Recipient]; ok {
					err := client.Conn.WriteJSON(message)

					if err != nil {
						log.Printf("Error sending direct message: %v\n", err)
					}
				}

			} else if message.Room != "" {
				// Broadcast to everyone in the room except sender
				if roomClient, ok := rooms[message.Room]; ok {
					for id, client := range roomClient {
						if id != message.Sender {
							err := client.Conn.WriteJSON(message)
							if err != nil {
								log.Printf("Error broadcasting message: %v\n", err)
							}
						}
					}
				}
			}
			mutex.Unlock()
		}
	}
}

func generateUniqueID() string {
	return uuid.New().String()
}
