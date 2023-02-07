package main

import (
	"log"
	"os"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/pmohanj/web-chat-app/database"
	"github.com/pmohanj/web-chat-app/routes"
	"github.com/pmohanj/web-chat-app/websocket"
)

func main() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal("Error loading env variables ", err)
	}

	r := gin.Default()

	// Initiate Databse
	MongoDBURL := os.Getenv("MONGODB_URL")
	database.DBinstance(MongoDBURL)

	// Allows all origins, not suitable for prod environments
	r.Use(cors.New(cors.Config{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{"PUT", "GET", "POST", "DELETE"},
		AllowHeaders: []string{"Content-Type", "Authorization"},
		MaxAge:       12 * time.Hour,
	}))
	api := r.Group("/api")
	routes.AddUserRoutes(api)
	routes.AddChatRoutes(api)
	routes.AddMessageRoutes(api)

	// create websocketserver
	websocket := websocket.CreateWebSocketsServer()

	go websocket.SendMessage()
	routes.AddWebScoketRouter(api, websocket)

	r.Run(":8000")
}
