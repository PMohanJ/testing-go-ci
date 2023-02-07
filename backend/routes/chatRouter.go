package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/pmohanj/web-chat-app/controllers"
	"github.com/pmohanj/web-chat-app/middleware"
)

func AddChatRoutes(r *gin.RouterGroup) {
	chat := r.Group("/chat")
	chat.POST("/", middleware.Authenticate(), controllers.AddChatUser())
	chat.GET("/", middleware.Authenticate(), controllers.GetUserChats())
	chat.DELETE("/:chatId", middleware.Authenticate(), controllers.DeleteUserConversation())
	chat.POST("/group", middleware.Authenticate(), controllers.CreateGroupChat())
	chat.PUT("/grouprename", middleware.Authenticate(), controllers.RenameGroupChatName())
	chat.PUT("/groupadd", middleware.Authenticate(), controllers.AddUserToGroupChat())
	chat.PUT("/groupremove", middleware.Authenticate(), controllers.DeleteUserFromGroupChat())
	chat.PUT("/groupexit", middleware.Authenticate(), controllers.UserExitGroup())
}
