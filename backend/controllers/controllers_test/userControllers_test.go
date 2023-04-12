package controllers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/assert/v2"
	"github.com/joho/godotenv"
	"github.com/pmohanj/web-chat-app/database"
	"github.com/pmohanj/web-chat-app/models"
	"github.com/pmohanj/web-chat-app/routes"
)

var router *gin.Engine
var user1Token string
var chatId string
var chatIdGroup string
var chatIdDelete string
var user0Id string
var user2Id string
var messageIdDelete string
var messageIdEdit string

func TestMain(m *testing.M) {
	router = gin.Default()

	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal("Error loading env variables ", err)
	}
	log.Println("Got env file")
	MongoDBURL := os.Getenv("MONGODB_URL_TESTING")
	// Initiate Databse
	database.DBinstance(MongoDBURL)

	// setup user routes
	api := router.Group("/api")
	routes.AddUserRoutes(api)
	routes.AddMessageRoutes(api)
	routes.AddChatRoutes(api)

	status := setupPhase()
	if status != 0 {
		os.Exit(1)
	}
	code := m.Run()
	tearDownPhase()
	database.CloseDBinstance()
	os.Exit(code)
}

// setupInitialPhase generates data to perform test operations
func setupPhase() int {
	statusCreateUsers := createUsers()
	if statusCreateUsers != 0 {
		return statusCreateUsers
	}

	statusInitiateChats := initiateChats()
	if statusInitiateChats != 0 {
		return statusInitiateChats
	}

	statusCreateMessages := createMessages()
	if statusCreateMessages != 0 {
		return statusCreateMessages
	}

	statusCreateGroupChat := createGroupChat()
	if statusCreateGroupChat != 0 {
		return statusCreateGroupChat
	}
	return 0
}

func createUsers() int {
	input0 := []byte(`{"name":"User0", "email":"user0@gmail.com", "password":"haha123"}`)
	req0, _ := http.NewRequest("POST", "/api/user/", bytes.NewBuffer(input0))

	response0 := httptest.NewRecorder()
	router.ServeHTTP(response0, req0)
	if response0.Code != 200 {
		return response0.Code
	}

	var resUser0 map[string]string
	_ = json.NewDecoder(response0.Body).Decode(&resUser0)
	user0Id = resUser0["_id"]

	input1 := []byte(`{"name":"User1", "email":"user1@gmail.com", "password":"haha123"}`)
	req1, _ := http.NewRequest("POST", "/api/user/", bytes.NewBuffer(input1))

	response1 := httptest.NewRecorder()
	router.ServeHTTP(response1, req1)
	if response1.Code != 200 {
		return response1.Code
	}

	var resUser1 map[string]string
	_ = json.NewDecoder(response1.Body).Decode(&resUser1)
	user1Token = resUser1["token"]

	input2 := []byte(`{"name":"User2", "email":"user2@gmail.com", "password":"haha123"}`)
	req2, _ := http.NewRequest("POST", "/api/user/", bytes.NewBuffer(input2))

	response2 := httptest.NewRecorder()
	router.ServeHTTP(response2, req2)
	if response2.Code != 200 {
		return response2.Code
	}

	var resUser2 map[string]string
	_ = json.NewDecoder(response2.Body).Decode(&resUser2)
	user2Id = resUser2["_id"]

	return 0
}

func initiateChats() int {

	// initiate chat
	data1 := fmt.Sprintf(`{"userToBeAdded":"%s"}`, user2Id)
	input1 := []byte(data1)
	req1, _ := http.NewRequest("POST", "/api/chat/", bytes.NewBuffer(input1))
	req1.Header.Set("Authorization", "Bearer "+user1Token)

	response1 := httptest.NewRecorder()
	router.ServeHTTP(response1, req1)
	if response1.Code != 200 {
		return response1.Code
	}

	var resChat map[string]interface{}
	_ = json.NewDecoder(response1.Body).Decode(&resChat)
	chatId, _ = resChat["_id"].(string)

	// initiate chat for TestDeleteUserConversation
	data2 := fmt.Sprintf(`{"userToBeAdded":"%s"}`, user0Id)
	input2 := []byte(data2)
	req2, _ := http.NewRequest("POST", "/api/chat/", bytes.NewBuffer(input2))
	req2.Header.Set("Authorization", "Bearer "+user1Token)

	response2 := httptest.NewRecorder()
	router.ServeHTTP(response2, req2)
	if response2.Code != 200 {
		return response2.Code
	}

	var resChatDelete map[string]interface{}
	_ = json.NewDecoder(response2.Body).Decode(&resChatDelete)
	chatIdDelete, _ = resChatDelete["_id"].(string)

	return 0
}

func createMessages() int {
	// message for TestDeleteUserMessage
	data1 := fmt.Sprintf(`{"chatId":"%s", "content":"How are you bro"}`, chatId)
	input1 := []byte(data1)
	req1, _ := http.NewRequest("POST", "/api/message/", bytes.NewBuffer(input1))
	req1.Header.Set("Authorization", "Bearer "+user1Token)

	response1 := httptest.NewRecorder()
	router.ServeHTTP(response1, req1)

	if response1.Code != 200 {
		return response1.Code
	}

	var resMessageDelete map[string]interface{}
	_ = json.NewDecoder(response1.Body).Decode(&resMessageDelete)

	messageIdDelete = resMessageDelete["_id"].(string)

	// message for TestEditUserMessage
	data2 := fmt.Sprintf(`{"chatId":"%s", "content":"Message to be edited"}`, chatId)
	input2 := []byte(data2)
	req2, _ := http.NewRequest("POST", "/api/message/", bytes.NewBuffer(input2))
	req2.Header.Set("Authorization", "Bearer "+user1Token)

	response2 := httptest.NewRecorder()
	router.ServeHTTP(response2, req2)

	if response2.Code != 200 {
		return response2.Code
	}

	var resMessageEdit map[string]interface{}
	_ = json.NewDecoder(response2.Body).Decode(&resMessageEdit)

	messageIdEdit = resMessageEdit["_id"].(string)

	return 0
}

func createGroupChat() int {
	// create a group chat for user1 and user2
	data := fmt.Sprintf(`{"groupName":"group for testing", "users":["%s"]}`, user2Id)
	input := []byte(data)
	request, _ := http.NewRequest("POST", "/api/chat/group", bytes.NewBuffer(input))
	request.Header.Set("Authorization", "Bearer "+user1Token)

	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)

	if response.Code != 200 {
		return response.Code
	}

	var result map[string]interface{}
	_ = json.NewDecoder(response.Body).Decode(&result)
	chatIdGroup, _ = result["_id"].(string)
	return 0
}

func tearDownPhase() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	chatCollection := database.OpenCollection(database.Client, "chat")
	messageCollection := database.OpenCollection(database.Client, "message")
	userCollection := database.OpenCollection(database.Client, "user")

	chatCollection.Drop(ctx)
	messageCollection.Drop(ctx)
	userCollection.Drop(ctx)
}

func TestRegisterUser(t *testing.T) {

	t.Run("returns data decoding error", func(t *testing.T) {
		// improperly structured json input
		input := []byte(`{"name":"Sky, email":"checking@gmail.com "password" "78945626"}`)
		req, _ := http.NewRequest("POST", "/api/user/", bytes.NewBuffer(input))

		response := httptest.NewRecorder()
		router.ServeHTTP(response, req)

		var res map[string]string
		_ = json.NewDecoder(response.Body).Decode(&res)

		assert.Equal(t, http.StatusBadRequest, response.Code)

		if res["error"] != "error while decoding user data" {
			t.Errorf("Unexpected results: should return an error, error while decoding user data")
		}
	})

	t.Run("returns user already resgistered", func(t *testing.T) {
		input := []byte(`{"name":"User1", "email":"user1@gmail.com", "password":"haha123"}`)
		request, _ := http.NewRequest("POST", "/api/user/", bytes.NewBuffer(input))

		response := httptest.NewRecorder()
		router.ServeHTTP(response, request)

		var res map[string]string
		_ = json.NewDecoder(response.Body).Decode(&res)

		assert.Equal(t, http.StatusBadRequest, response.Code)

		if res["error"] != "You've already registered with this email" {
			t.Error("Unexpected result: should return error, user already registered with this email")
		}
	})
}

func TestAuthUser(t *testing.T) {

	t.Run("returns user deatils", func(t *testing.T) {
		input := []byte(`{"email":"user1@gmail.com", "password":"haha123"}`)
		req, _ := http.NewRequest("POST", "/api/user/login", bytes.NewBuffer(input))

		response := httptest.NewRecorder()
		router.ServeHTTP(response, req)

		// expected response
		exp_data := models.User{
			Name:  "User1",
			Email: "user1@gmail.com",
			Pic:   "http://res.cloudinary.com/dkqc4za4f/image/upload/v1670340314/clsfmjxnuzsnidzc59np.jpg",
		}

		// decode the response body
		var res map[string]string
		_ = json.NewDecoder(response.Body).Decode(&res)

		assert.Equal(t, http.StatusOK, response.Code)

		if res["name"] != exp_data.Name {
			t.Errorf("Unexpectes result: got %v, want %v", res["name"], exp_data.Name)
		}

		if res["email"] != exp_data.Email {
			t.Errorf("Unexpectes result: got %v, want %v", res["email"], exp_data.Email)
		}

		if res["pic"] == "" {
			t.Errorf("Unexpected result: pic field can't be empty")
		}

		if res["token"] == "" {
			t.Errorf("Unexpected result: token field can't be empty")
		}
	})

	t.Run("returns password invalid error", func(t *testing.T) {
		input := []byte(`{"email":"user1@gmail.com", "password":"haha"}`)
		req, _ := http.NewRequest("POST", "/api/user/login", bytes.NewBuffer(input))

		response := httptest.NewRecorder()
		router.ServeHTTP(response, req)

		var res map[string]string
		_ = json.NewDecoder(response.Body).Decode(&res)

		assert.Equal(t, http.StatusUnauthorized, response.Code)

		if res["error"] != "Given password is invalid" {
			t.Errorf("Unexpected results: should return an error, invalid password")
		}
	})

	t.Run("returns user not resgistered error", func(t *testing.T) {
		input := []byte(`{"email":"unknown@gmail.com", "password":"12345678"}`)
		req, _ := http.NewRequest("POST", "/api/user/login", bytes.NewBuffer(input))

		response := httptest.NewRecorder()
		router.ServeHTTP(response, req)

		var res map[string]string
		_ = json.NewDecoder(response.Body).Decode(&res)

		assert.Equal(t, http.StatusNotFound, response.Code)

		if res["error"] != "User not registered" {
			t.Errorf("Unexpected results: should return an error, user not registered")
		}
	})

	t.Run("returns data decoding error", func(t *testing.T) {
		// improperly structured json input
		input := []byte(`{"email"user1@gmail.com "password":"haha123"}`)
		req, _ := http.NewRequest("POST", "/api/user/login", bytes.NewBuffer(input))

		response := httptest.NewRecorder()
		router.ServeHTTP(response, req)

		var res map[string]string
		_ = json.NewDecoder(response.Body).Decode(&res)

		assert.Equal(t, http.StatusBadRequest, response.Code)

		if res["error"] != "error while decoding user data" {
			t.Errorf("Unexpected results: should return an error, error while decoding user data")
		}
	})
}

func TestSearchUsers(t *testing.T) {

	t.Run("returns no users found", func(t *testing.T) {
		url := "/api/user/search?search=" + "uoaomaxoasvfa*#20"
		request, _ := http.NewRequest("GET", url, nil)
		request.Header.Set("Authorization", "Bearer "+user1Token)

		response := httptest.NewRecorder()
		router.ServeHTTP(response, request)

		var result []map[string]string
		_ = json.NewDecoder(response.Body).Decode(&result)

		assert.Equal(t, http.StatusOK, response.Code)

		if len(result) > 0 {
			t.Errorf("Unexpected result: got %v, want %v", len(result), "0 documents to be returned")
		}
	})

	t.Run("returns user found", func(t *testing.T) {
		url := "/api/user/search?search=" + "user2"
		request, _ := http.NewRequest("GET", url, nil)
		request.Header.Set("Authorization", "Bearer "+user1Token)

		response := httptest.NewRecorder()
		router.ServeHTTP(response, request)

		var result []map[string]string
		_ = json.NewDecoder(response.Body).Decode(&result)

		assert.Equal(t, http.StatusOK, response.Code)

		if len(result) < 1 {
			t.Errorf("Unexpected result: got %v, want %v", len(result), "at least 1 document")
		}
	})
}
