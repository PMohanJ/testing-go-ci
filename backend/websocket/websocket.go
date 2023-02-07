package websocket

import (
	"errors"
	"log"

	"github.com/gin-gonic/gin"
)

type WebSockets struct {
	Clients   map[string][]*Client
	Broadcast chan map[string]interface{}
}

func CreateWebSocketsServer() *WebSockets {
	return &WebSockets{
		Clients:   make(map[string][]*Client),
		Broadcast: make(chan map[string]interface{}),
	}
}

func (ws *WebSockets) WSEndpoint() gin.HandlerFunc {
	return func(c *gin.Context) {
		conn, err := Upgrade(c.Writer, c.Request)
		if err != nil {
			log.Panic(err)
		}

		client := &Client{
			Conn:       conn,
			WebSockets: ws,
		}
		defer func() {
			client.Conn.Close()
			ws.removeClient(client)
		}()
		for {
			var data map[string]interface{}
			err := client.Conn.ReadJSON(&data)
			if err != nil {
				client.Conn.Close()
				log.Println("client closed")
				break
			}

			err = ws.HandleClientMessage(client, data)
			if err != nil {
				log.Panic(err)
			}
		}

	}
}

// removeClient removes client from websocket pool
func (ws *WebSockets) removeClient(clientObj *Client) {
	for chatId, clients := range ws.Clients {
		var done bool
		for i, client := range clients {
			if client == clientObj {
				clients = append(clients[:i], clients[i+1:]...)
				ws.Clients[chatId] = clients
				done = true
				break
			}
		}
		if done {
			break
		}
	}
}

// HandleClientMessage adds client to the respective chats, and also sends client messages
// to websocket broadcast channel
func (ws *WebSockets) HandleClientMessage(clientObj *Client, data map[string]interface{}) error {

	// check if client is initiating the connection
	if data["messageType"] == "setup" {
		// create that chat and add the client
		chatId, ok := data["chat"].(string)
		if !ok {
			return errors.New("chat id type is not string")
		}

		clients, exists := ws.Clients[chatId]
		if exists {
			clients = append(clients, clientObj)
			ws.Clients[chatId] = clients
		} else {
			var clients []*Client
			clients = append(clients, clientObj)
			ws.Clients[chatId] = clients
		}

		log.Printf("Client added to list %+v", clientObj)
	} else {

		// broadcast the message/data
		clientObj.WebSockets.Broadcast <- data
	}
	return nil
}

// SendMessage receives messages from broadcast channel and sends them to
// respective chat members aka clients
func (ws *WebSockets) SendMessage() {
	for {
		msg := <-ws.Broadcast
		chatId := msg["chat"].(string)

		// get the chat to which the msg should be send
		clientsOfThisChat := ws.Clients[chatId]
		log.Printf("size of chat %v and chatid is: %v", len(clientsOfThisChat), chatId)
		for _, client := range clientsOfThisChat {
			err := client.Conn.WriteJSON(msg)
			if err != nil {
				log.Panic(err)
			}
		}
	}
}
