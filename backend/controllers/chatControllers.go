package controllers

import (
	"context"
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/pmohanj/web-chat-app/database"
	"github.com/pmohanj/web-chat-app/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// AddChatUser lets the user to add a user to chat with
func AddChatUser() gin.HandlerFunc {
	return func(c *gin.Context) {

		var ids map[string]interface{}

		if err := c.BindJSON(&ids); err != nil {
			log.Panic("error while parsing data")
		}

		log.Printf("ids: %+v", ids)

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		chatCollection := database.OpenCollection(database.Client, "chat")

		// get the ids and convert them back to primitive.ObjectID format for querying
		id1, exists := c.Get("_id")
		if !exists {
			log.Panic("User details not available")
		}
		addingUser := id1.(primitive.ObjectID)

		id2 := ids["userToBeAdded"].(string)
		userToBeAdded, err := primitive.ObjectIDFromHex(id2)
		if err != nil {
			log.Panic(err)
		}

		filter := bson.D{
			{"isGroupChat", false},
			{"$and",
				bson.A{
					bson.D{{"users", bson.D{{"$elemMatch", bson.D{{"$eq", addingUser}}}}}},
					bson.D{{"users", bson.D{{"$elemMatch", bson.D{{"$eq", userToBeAdded}}}}}},
				},
			}}

		// check if the users have chatted before, if so reaturn their chat
		var existedChat models.Chat
		err = chatCollection.FindOne(ctx, filter).Decode(&existedChat)
		if err == nil {
			// chat exist, perform aggragrate operations to join document of Chat with respective chat Users profile
			matchStage := bson.D{
				{
					"$match", bson.D{{"isGroupChat", false},
						{"$and",
							bson.A{
								bson.D{{"users", bson.D{{"$elemMatch", bson.D{{"$eq", addingUser}}}}}},
								bson.D{{"users", bson.D{{"$elemMatch", bson.D{{"$eq", userToBeAdded}}}}}},
							}}},
				},
			}
			lookupStage := LookUpStage("user", "users", "_id", "users")

			projectStage := ProjectStage("users.password", "created_at",
				"updated_at", "users.created_at", "users.updated_at")

			var res []bson.M
			cur, err := chatCollection.Aggregate(ctx, mongo.Pipeline{matchStage, lookupStage, projectStage})
			if err != nil {
				log.Panic(err)
			}

			if err = cur.All(ctx, &res); err != nil {
				log.Panic(err)
			}
			for _, docu := range res {
				log.Printf("docu: %+v", docu)
			}

			c.JSON(http.StatusOK, res[0])
			return
		} else if errors.Is(err, mongo.ErrNoDocuments) {
			log.Println("Chat does't exist")

		} else if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Erroe while checking db"})
			log.Panic(err)
		}

		// No chat existed, so create a chat for the users
		createChat := models.Chat{
			ChatName:    "sender",
			IsGroupChat: false,
			Users:       []primitive.ObjectID{addingUser, userToBeAdded},
		}

		insId, err := chatCollection.InsertOne(ctx, createChat)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "err while inserting chat"})
			log.Panic(err)
		}
		insertedId := insId.InsertedID.(primitive.ObjectID)
		log.Println(insertedId)

		var createdChat []bson.M

		matchStage := MatchStageBySingleField("_id", insertedId)

		lookupStage := LookUpStage("user", "users", "_id", "users")

		projectStage := ProjectStage("users.password", "created_at",
			"updated_at", "users.created_at", "users.updated_at")

		cursor, err := chatCollection.Aggregate(ctx, mongo.Pipeline{matchStage, lookupStage, projectStage})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "err while retreving created chat"})
			log.Println(err)
		}

		if err := cursor.All(ctx, &createdChat); err != nil {
			log.Panic(err)
		}
		c.JSON(http.StatusOK, createdChat[0])
	}
}

func GetUserChats() gin.HandlerFunc {
	return func(c *gin.Context) {
		id, exists := c.Get("_id")
		if !exists {
			log.Panic("User details not available")
		}
		userId := id.(primitive.ObjectID)

		chatCollection := database.OpenCollection(database.Client, "chat")

		matchStage := bson.D{
			{
				"$match", bson.D{
					{
						"users", bson.D{{"$elemMatch", bson.D{{"$eq", userId}}}},
					},
				},
			},
		}

		lookupStage := LookUpStage("user", "users", "_id", "users")

		lookupStageLatestMessage := bson.D{
			{
				"$lookup", bson.D{
					{"from", "message"},
					{"localField", "latestMessage"},
					{"foreignField", "_id"},
					{"as", "latestMessage"},
				},
			},
		}

		projectStage := ProjectStage("users.password", "created_at",
			"updated_at", "users.created_at", "users.updated_at")

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		cursor, err := chatCollection.Aggregate(ctx, mongo.Pipeline{matchStage, lookupStage, lookupStageLatestMessage, projectStage})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error checking documents"})
			log.Panic(err)
		}

		var results []bson.M
		if err := cursor.All(ctx, &results); err != nil {
			log.Panic(err)
		}
		for _, docu := range results {
			log.Println(docu)
		}

		c.JSON(http.StatusOK, results)
	}
}

func DeleteUserConversation() gin.HandlerFunc {
	return func(c *gin.Context) {
		cId := c.Param("chatId")
		chatId, err := primitive.ObjectIDFromHex(cId)
		if err != nil {
			log.Panic(err)
		}

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		// delete all the messages that refer this chatId
		messageCollection := database.OpenCollection(database.Client, "message")

		filter := bson.D{
			{"chat", chatId},
		}
		_, err = messageCollection.DeleteMany(ctx, filter)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error while deleting chat messages"})
			log.Panic(err)
		}

		// delete the chat document too
		chatCollection := database.OpenCollection(database.Client, "chat")
		_, err = chatCollection.DeleteOne(ctx, bson.D{{"_id", chatId}})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error while deleting chat document"})
			log.Panic(err)
		}

		c.Status(http.StatusOK)
	}
}

func CreateGroupChat() gin.HandlerFunc {
	return func(c *gin.Context) {
		var groupData map[string]interface{}

		if err := c.BindJSON(&groupData); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Error while parsing data"})
			log.Panic(err)
		}

		groupName := groupData["groupName"].(string)

		users := groupData["users"].([]interface{})

		aUser, exists := c.Get("_id")
		if !exists {
			log.Panic("User details not available")
		}
		adminUser := aUser.(primitive.ObjectID)

		var usersIds []primitive.ObjectID
		usersIds = append(usersIds, adminUser)

		for _, uId := range users {
			id := uId.(string)

			temp, err := primitive.ObjectIDFromHex(id)
			if err != nil {
				log.Panic(err)
			}
			usersIds = append(usersIds, temp)
		}

		groupChat := models.Chat{
			IsGroupChat: true,
			ChatName:    groupName,
			Users:       usersIds,
			GroupAdmin:  adminUser,
		}

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		chatCollection := database.OpenCollection(database.Client, "chat")

		insId, err := chatCollection.InsertOne(ctx, groupChat)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "err while inserting document"})
			log.Panic(insId)
		}

		insertedId := insId.InsertedID.(primitive.ObjectID)

		matchStage := MatchStageBySingleField("_id", insertedId)

		lookupStage := LookUpStage("user", "users", "_id", "users")

		projectStage := ProjectStage("users.password", "created_at",
			"updated_at", "users.created_at", "users.updated_at")

		cursor, err := chatCollection.Aggregate(ctx, mongo.Pipeline{matchStage, lookupStage, projectStage})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error checking documents"})
			log.Panic(err)
		}

		var results []bson.M
		if err := cursor.All(ctx, &results); err != nil {
			log.Panic(err)
		}

		c.JSON(http.StatusOK, results[0])
	}
}

func RenameGroupChatName() gin.HandlerFunc {
	return func(c *gin.Context) {
		var reqData map[string]interface{}

		if err := c.BindJSON(&reqData); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Error while parsing data"})
			return
		}

		groupName := reqData["groupName"].(string)
		cId := reqData["chatId"].(string)

		chatId, err := primitive.ObjectIDFromHex(cId)
		if err != nil {
			log.Panic(err)
		}

		chatCollection := database.OpenCollection(database.Client, "chat")

		filter := bson.D{{"_id", chatId}}

		update := bson.D{{"$set", bson.D{{"chatName", groupName}}}}

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		_, err = chatCollection.UpdateOne(ctx, filter, update)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error while updating document"})
			log.Panic(err)
		}

		c.JSON(http.StatusOK, gin.H{"updatedGroupName": groupName})
	}
}

func AddUserToGroupChat() gin.HandlerFunc {
	return func(c *gin.Context) {
		var reqData map[string]interface{}

		if err := c.BindJSON(&reqData); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Error while parsing data"})
			return
		}

		uId := reqData["userId"].(string)
		cId := reqData["chatId"].(string)

		chatId, err := primitive.ObjectIDFromHex(cId)
		if err != nil {
			log.Panic(err)
		}

		userId, err := primitive.ObjectIDFromHex(uId)
		if err != nil {
			log.Panic(err)
		}

		chatCollection := database.OpenCollection(database.Client, "chat")

		filter := bson.D{{"_id", chatId}}

		update := bson.D{{"$push", bson.D{{"users", userId}}}}

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		_, err = chatCollection.UpdateOne(ctx, filter, update)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error while updating document"})
			log.Panic(err)
		}

		// User is added to group, now retrieve that document and send into client
		// so that client can update its data, and perfrom necessary rendering
		matchStage := MatchStageBySingleField("_id", chatId)

		lookupStage := LookUpStage("user", "users", "_id", "users")

		projectStage := ProjectStage("users.password", "created_at",
			"updated_at", "users.created_at", "users.updated_at")

		cursor, err := chatCollection.Aggregate(ctx, mongo.Pipeline{matchStage, lookupStage, projectStage})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error checking documents"})
			log.Panic(err)
		}

		// we can only pass array type to cursor even though we know
		// we're just retrieving single document
		var results []bson.M
		if err := cursor.All(ctx, &results); err != nil {
			log.Panic(err)
		}
		for _, docu := range results {
			log.Println(docu)
		}

		c.JSON(http.StatusOK, results[0])
	}
}

func DeleteUserFromGroupChat() gin.HandlerFunc {
	return func(c *gin.Context) {
		var reqData map[string]interface{}

		if err := c.BindJSON(&reqData); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Error while parsing data"})
			return
		}

		uId := reqData["userId"].(string)
		cId := reqData["chatId"].(string)

		chatId, err := primitive.ObjectIDFromHex(cId)
		if err != nil {
			log.Panic(err)
		}

		userId, err := primitive.ObjectIDFromHex(uId)
		if err != nil {
			log.Panic(err)
		}

		chatCollection := database.OpenCollection(database.Client, "chat")

		filter := bson.D{{"_id", chatId}}

		update := bson.D{{"$pull", bson.D{{"users", userId}}}}

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		res, err := chatCollection.UpdateOne(ctx, filter, update)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error while updating document"})
			log.Panic(err)
		}
		log.Printf("Docu up %v", res.ModifiedCount)
		// User is added to group, now retrieve that document and send into client
		// so that client can update its data, and perfrom necessary rendering
		matchStage := MatchStageBySingleField("_id", chatId)

		lookupStage := LookUpStage("user", "users", "_id", "users")

		projectStage := ProjectStage("users.password", "created_at",
			"updated_at", "users.created_at", "users.updated_at")

		cursor, err := chatCollection.Aggregate(ctx, mongo.Pipeline{matchStage, lookupStage, projectStage})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error checking documents"})
			log.Panic(err)
		}

		var results []bson.M
		if err := cursor.All(ctx, &results); err != nil {
			log.Panic(err)
		}
		for _, docu := range results {
			log.Println(docu)
		}

		c.JSON(http.StatusOK, results[0])
	}

}

// UserExitGroup removes a user from Group chat or deletes the whole
// chat if admin of that group is exiting
func UserExitGroup() gin.HandlerFunc {
	return func(c *gin.Context) {
		var reqData map[string]interface{}

		if err := c.BindJSON(&reqData); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Error while parsing data"})
			return
		}

		cId := reqData["chatId"].(string)
		uId, exists := c.Get("_id")
		if !exists {
			log.Panic("User details not available")
		}

		chatId, err := primitive.ObjectIDFromHex(cId)
		if err != nil {
			log.Panic(err)
		}

		userId := uId.(primitive.ObjectID)

		// get chat collection
		chatCollection := database.OpenCollection(database.Client, "chat")

		filter := bson.D{{"_id", chatId}}

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		var chatDocu bson.M
		err = chatCollection.FindOne(ctx, filter).Decode(&chatDocu)
		if err != nil {
			log.Panic(err)
		}

		groupAdmin := chatDocu["groupAdmin"].(primitive.ObjectID)

		// check if admin is exiting Group chat
		if userId.Hex() == groupAdmin.Hex() {
			// delete the whole chat
			deleteResult, err := chatCollection.DeleteOne(ctx, filter)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Error while querying database"})
				log.Panic(err)
			}
			log.Println("Documents deleted: ", deleteResult.DeletedCount)
			c.JSON(http.StatusOK, gin.H{"message": "Exited from group"})
			return
		}

		// just remove the user from Group chat
		update := bson.D{{"$pull", bson.D{{"users", userId}}}}

		res, err := chatCollection.UpdateOne(ctx, filter, update)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error while updating document"})
			log.Panic(err)
		}
		log.Printf("Documents deleted: %v", res.ModifiedCount)

		c.JSON(http.StatusOK, gin.H{"message": "Exited from group"})
	}
}
