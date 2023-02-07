package controllers_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-playground/assert/v2"
)

func TestAddChatUser(t *testing.T) {
	t.Run("returns users chat", func(t *testing.T) {
		data := fmt.Sprintf(`{"userToBeAdded":"%s"}`, user2Id)
		input := []byte(data)
		request, _ := http.NewRequest("POST", "/api/chat/", bytes.NewBuffer(input))
		request.Header.Set("Authorization", "Bearer "+user1Token)

		response := httptest.NewRecorder()
		router.ServeHTTP(response, request)

		expectedChatId := chatId

		var result map[string]interface{}
		_ = json.NewDecoder(response.Body).Decode(&result)

		assert.Equal(t, http.StatusOK, response.Code)

		if result["_id"] != expectedChatId {
			t.Errorf("Unexpected result: got %v, want %v", result["_id"], expectedChatId)
		}
	})
}

func TestGetUserChats(t *testing.T) {
	t.Run("returns user chats", func(t *testing.T) {
		request, _ := http.NewRequest("GET", "/api/chat/", nil)
		request.Header.Set("Authorization", "Bearer "+user1Token)

		response := httptest.NewRecorder()
		router.ServeHTTP(response, request)

		var result []map[string]interface{}
		_ = json.NewDecoder(response.Body).Decode(&result)

		assert.Equal(t, http.StatusOK, response.Code)

		// wait 'ill change the actual handler so that I don't need to send id explicitly
		if len(result) < 1 {
			t.Errorf("Unexpected result: got %v, want %v", len(result), "at least 1 chat document")
		}
	})
}

func TestDeleteUserConversation(t *testing.T) {
	t.Run("returns status ok for delete conversation", func(t *testing.T) {
		request, _ := http.NewRequest("DELETE", "/api/chat/"+chatIdDelete, nil)
		request.Header.Set("Authorization", "Bearer "+user1Token)

		response := httptest.NewRecorder()
		router.ServeHTTP(response, request)

		assert.Equal(t, http.StatusOK, response.Code)
	})
}

func TestCreateGroupChat(t *testing.T) {
	t.Run("returns status ok for create group", func(t *testing.T) {
		data := fmt.Sprintf(`{"groupName":"Temporary testing group", "users":["%s"]}`, user2Id)
		input := []byte(data)
		request, _ := http.NewRequest("POST", "/api/chat/group", bytes.NewBuffer(input))
		request.Header.Set("Authorization", "Bearer "+user1Token)

		response := httptest.NewRecorder()
		router.ServeHTTP(response, request)

		expectedChatName := "Temporary testing group"
		expectedChatLabel := true

		var result map[string]interface{}
		_ = json.NewDecoder(response.Body).Decode(&result)

		assert.Equal(t, http.StatusOK, response.Code)

		if result["chatName"] != expectedChatName {
			t.Errorf("Unexpected result: got %v, want %v", result["chatName"], expectedChatName)
		}

		if result["isGroupChat"] != expectedChatLabel {
			t.Errorf("Unexpected result: got %v, want %v", result["isGroupChat"], expectedChatLabel)
		}
	})
}

func TestRenameGroupChatName(t *testing.T) {
	t.Run("returns updated group chat name", func(t *testing.T) {
		data := fmt.Sprintf(`{"groupName":"Group for testing renamed", "chatId":"%s"}`, chatIdGroup)
		input := []byte(data)
		request, _ := http.NewRequest("PUT", "/api/chat/grouprename", bytes.NewBuffer(input))
		request.Header.Set("Authorization", "Bearer "+user1Token)

		response := httptest.NewRecorder()
		router.ServeHTTP(response, request)

		expectedChatName := "Group for testing renamed"

		var result map[string]interface{}
		_ = json.NewDecoder(response.Body).Decode(&result)

		assert.Equal(t, http.StatusOK, response.Code)

		if result["updatedGroupName"] != expectedChatName {
			t.Errorf("Unexpected result: got %v, want %v", result["chatName"], expectedChatName)
		}
	})
}

func TestAddUserToGroupChat(t *testing.T) {
	t.Run("returns status ok for add user to chat", func(t *testing.T) {
		// adding another user, user0Id to the group
		data := fmt.Sprintf(`{"userId":"%s", "chatId":"%s"}`, user0Id, chatIdGroup)
		input := []byte(data)
		request, _ := http.NewRequest("PUT", "/api/chat/groupadd", bytes.NewBuffer(input))
		request.Header.Set("Authorization", "Bearer "+user1Token)

		response := httptest.NewRecorder()
		router.ServeHTTP(response, request)

		expectedGroupUsersLength := 3
		var expectedAddedUserId string

		var result map[string]interface{}
		_ = json.NewDecoder(response.Body).Decode(&result)

		resultUsers, ok := result["users"].([]interface{})
		if !ok {
			log.Panic("Type assertion failed")
		}

		assert.Equal(t, http.StatusOK, response.Code)

		if len(resultUsers) < 3 {
			t.Errorf("Unexpected result: got %v, want %v", len(resultUsers), expectedGroupUsersLength)
		}

		// iterate over user objects to get the user id that just got added
		for i := range resultUsers {
			resultUserObject, ok := resultUsers[i].(map[string]interface{})
			if !ok {
				log.Panic("Type assertion failed")
			}
			resultUserId, ok := resultUserObject["_id"].(string)
			if !ok {
				log.Panic("Type assertion failed")
			}

			if resultUserId == user0Id {
				expectedAddedUserId = user0Id
			}
		}

		if expectedAddedUserId == "" {
			t.Errorf("Unexpected result: got %v, want %v", nil, user0Id)
		}
	})
}

func TestDeleteUserFromGroupChat(t *testing.T) {
	t.Run("returns status ok for delete user from group", func(t *testing.T) {
		// remove a user, user2Id from the group
		data := fmt.Sprintf(`{"userId":"%s", "chatId":"%s"}`, user2Id, chatIdGroup)
		input := []byte(data)
		request, _ := http.NewRequest("PUT", "/api/chat/groupremove", bytes.NewBuffer(input))
		request.Header.Set("Authorization", "Bearer "+user1Token)

		response := httptest.NewRecorder()
		router.ServeHTTP(response, request)

		expectedRemovedUserId := ""

		var result map[string]interface{}
		_ = json.NewDecoder(response.Body).Decode(&result)

		resultUsers, ok := result["users"].([]interface{})
		if !ok {
			log.Panic("Type assertion failed")
		}

		assert.Equal(t, http.StatusOK, response.Code)

		// iterate over user objects to check wheather user2Id exist
		for i := range resultUsers {
			resultUserObject, ok := resultUsers[i].(map[string]interface{})
			if !ok {
				log.Panic("Type assertion failed")
			}
			resultUserId, ok := resultUserObject["_id"].(string)
			if !ok {
				log.Panic("Type assertion failed")
			}

			if resultUserId == user2Id {
				expectedRemovedUserId = user2Id
			}
		}

		if expectedRemovedUserId != "" {
			t.Errorf("Unexpected result: got %v, want %v", expectedRemovedUserId, nil)
		}
	})
}

func TestUserExitGroup(t *testing.T) {
	t.Run("returns exited from group", func(t *testing.T) {
		data := fmt.Sprintf(`{"chatId":"%s"}`, chatId)
		input := []byte(data)
		request, _ := http.NewRequest("PUT", "/api/chat/groupexit", bytes.NewBuffer(input))
		request.Header.Set("Authorization", "Bearer "+user1Token)

		response := httptest.NewRecorder()
		router.ServeHTTP(response, request)

		expectedMessage := "Exited from group"

		var result map[string]string
		_ = json.NewDecoder(response.Body).Decode(&result)

		assert.Equal(t, http.StatusOK, response.Code)

		if result["message"] != expectedMessage {
			t.Errorf("Unexpected result: got %v, want %v", result["message"], expectedMessage)
		}
	})
}
