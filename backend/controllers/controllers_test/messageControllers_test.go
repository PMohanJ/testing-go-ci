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

func TestSendMessage(t *testing.T) {

	t.Run("returns error decoding data", func(t *testing.T) {
		data := fmt.Sprintf(`{"chatId""%s", "content":"Hello there}`, chatId)
		input := []byte(data)
		request, _ := http.NewRequest("POST", "/api/message/", bytes.NewBuffer(input))
		request.Header.Set("Authorization", "Bearer "+user1Token)

		response := httptest.NewRecorder()
		router.ServeHTTP(response, request)

		assert.Equal(t, http.StatusBadRequest, response.Code)
	})

	t.Run("returns message object", func(t *testing.T) {
		data := fmt.Sprintf(`{"chatId":"%s", "content":"How are you bro"}`, chatId)
		input := []byte(data)
		request, _ := http.NewRequest("POST", "/api/message/", bytes.NewBuffer(input))
		request.Header.Set("Authorization", "Bearer "+user1Token)

		response := httptest.NewRecorder()
		router.ServeHTTP(response, request)

		expected_res := map[string]interface{}{
			"sender": []map[string]string{{
				"name":  "User1",
				"email": "user1@gmail.com"},
			},
			"content": "How are you bro",
			"chat":    chatId,
		}

		var res map[string]interface{}
		_ = json.NewDecoder(response.Body).Decode(&res)

		assert.Equal(t, http.StatusOK, response.Code)

		if res["content"] != expected_res["content"] {
			t.Errorf("Unexpected result: got %v, want %v", res["content"], expected_res["content"])
		}

		if res["chat"] != expected_res["chat"] {
			t.Errorf("Unexpected result: got %v, want %v", res["chat"], expected_res["chat"])
		}

		ExpSenderField, ok := expected_res["sender"].([]map[string]string)
		if !ok {
			log.Panic("Err type assertion")
		}
		expected_resSender := ExpSenderField[0]

		resSenderField, ok := res["sender"].([]interface{})
		if !ok {
			log.Panic("Err type assertion")
		}

		sR := resSenderField[0]
		senderRes, ok := sR.(map[string]interface{})
		if !ok {
			log.Panic("Err type assertion")
		}
		srName, ok := senderRes["name"].(string)
		if !ok {
			log.Panic("Err type assertion")
		}
		srEmail, ok := senderRes["email"].(string)
		if !ok {
			log.Panic("Err type assertion")
		}
		if srName != expected_resSender["name"] {
			t.Errorf("Unexpectes result: got %v, want %v", senderRes["name"], expected_resSender["name"])
		}

		if srEmail != expected_resSender["email"] {
			t.Errorf("Unexpectes result: got %v, want %v", senderRes["email"], expected_resSender["email"])
		}
	})
}

func TestGetMessage(t *testing.T) {
	t.Run("returns user messages", func(t *testing.T) {
		request, _ := http.NewRequest("GET", "/api/message/"+chatId, nil)
		request.Header.Set("Authorization", "Bearer "+user1Token)

		response := httptest.NewRecorder()
		router.ServeHTTP(response, request)

		assert.Equal(t, http.StatusOK, response.Code)

		var result []map[string]interface{}
		_ = json.NewDecoder(response.Body).Decode(&result)

		if len(result) < 1 {
			t.Errorf("Unexpected result: got %v, want %v", len(result), "atleast 1 message document")
		}
	})
}

func TestEditUserMessage(t *testing.T) {

	t.Run("returns edited message", func(t *testing.T) {
		data := fmt.Sprintf(`{"content":"Message edited", "messageId":"%s"}`, messageIdEdit)
		input := []byte(data)
		request, _ := http.NewRequest("PUT", "/api/message/", bytes.NewBuffer(input))
		request.Header.Set("Authorization", "Bearer "+user1Token)

		response := httptest.NewRecorder()
		router.ServeHTTP(response, request)

		expectedContent := "Message edited"

		var result map[string]interface{}
		_ = json.NewDecoder(response.Body).Decode(&result)

		assert.Equal(t, http.StatusOK, response.Code)

		if result["content"] != expectedContent {
			t.Errorf("Unexpected result: got %v, want %v", result["content"], expectedContent)
		}
	})
}

func TestDeleteUserMessage(t *testing.T) {

	t.Run("returns status ok", func(t *testing.T) {
		request, _ := http.NewRequest("DELETE", "/api/message/"+messageIdDelete, nil)
		request.Header.Set("Authorization", "Bearer "+user1Token)

		response := httptest.NewRecorder()
		router.ServeHTTP(response, request)

		assert.Equal(t, http.StatusOK, response.Code)
	})
}
