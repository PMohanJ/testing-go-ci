package websocket

import (
	"github.com/gorilla/websocket"
)

type Client struct {
	WebSockets *WebSockets
	Conn       *websocket.Conn
}
