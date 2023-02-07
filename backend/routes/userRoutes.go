package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/pmohanj/web-chat-app/controllers"
	"github.com/pmohanj/web-chat-app/middleware"
)

func AddUserRoutes(router *gin.RouterGroup) {
	userRouter := router.Group("/user")

	userRouter.GET("/search", middleware.Authenticate(), controllers.SearchUsers())
	userRouter.POST("/", controllers.RegisterUser())
	userRouter.POST("/login", controllers.AuthUser())
}
