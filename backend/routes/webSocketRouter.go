package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/pmohanj/web-chat-app/websocket"
)

func AddWebScoketRouter(router *gin.RouterGroup, ws *websocket.WebSockets) {
	router.GET("/ws", ws.WSEndpoint())
}
