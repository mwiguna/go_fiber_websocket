package main

import (
	"log"
	"sync"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
)

// Declare Client Connection
type client struct {
	isClosing bool
	mu        sync.Mutex
}

type specificClient struct {
	connection *websocket.Conn
	mu         sync.Mutex
}

var clients = make(map[*websocket.Conn]*client)
var specificClients = make(map[string]*specificClient)

func main() {
	app := fiber.New()
	app.Static("/", "home.html")

	app.Get("/ws", websocket.New(func(c *websocket.Conn) {
		//  Save New Connection to Clients
		clients[c] = &client{}
		log.Println("New Connection")

		for {
			//  Listener Read
			_, msg, err := c.ReadMessage()
			if err != nil {
				log.Println("error read:", err)
				break
			} else {
				log.Printf("receive: %s", msg)
				specificClients["10"] = &specificClient{
					connection: c,
				}
			}

			// sendSpecific(msg)
			broadcastAll(msg)
		}
	}))

	app.Listen(":3000")
}

func sendSpecific(msg []byte) {
	for ID, c := range specificClients {
		go func(ID string, connection *websocket.Conn, c *specificClient) {
			c.mu.Lock()
			defer c.mu.Unlock()
			if ID != string(msg[:]) {
				log.Printf("Bukan user tujuan")
				return
			}

			err := connection.WriteMessage(websocket.TextMessage, msg)
			if err != nil {
				log.Println("write error:", err)

				connection.WriteMessage(websocket.CloseMessage, []byte{})
				connection.Close()
			} else {
				log.Printf("write : %s", msg)
			}
		}(ID, c.connection, c)
	}
}

func broadcastAll(msg []byte) {
	for connection, c := range clients {
		go func(connection *websocket.Conn, c *client) {
			c.mu.Lock()
			defer c.mu.Unlock()
			if c.isClosing {
				return
			}

			err := connection.WriteMessage(websocket.TextMessage, msg)
			if err != nil {
				c.isClosing = true
				log.Println("write error:", err)

				connection.WriteMessage(websocket.CloseMessage, []byte{})
				connection.Close()
			} else {
				log.Printf("write : %s", msg)
			}
		}(connection, c)
	}
}
