package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/pmohanj/web-chat-app/controllers"
	"github.com/pmohanj/web-chat-app/middleware"
)

func AddMessageRoutes(router *gin.RouterGroup) {
	messageRouter := router.Group("/message")

	messageRouter.POST("/", middleware.Authenticate(), controllers.SendMessage())
	messageRouter.GET("/:chatId", middleware.Authenticate(), controllers.GetMessages())
	messageRouter.PUT("/", middleware.Authenticate(), controllers.EditUserMessage())
	messageRouter.DELETE("/:messageId", middleware.Authenticate(), controllers.DeleteUserMessage())
}
