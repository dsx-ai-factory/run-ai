//go:build integration
// +build integration

package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

const (
	Host       = "http://localhost"
	Port       = 3000
	ValidToken = "dummy_token_pass_validation"
)

type Response struct {
	Ok    bool   `json:"ok"`
	Error string `json:"error"`
}

var requestsRecords = map[string]interface{}{}

// validateToken implemented according to:
// https://api.slack.com/methods/auth.test
func validateToken(w http.ResponseWriter, r *http.Request) {
	requestsRecords[fmt.Sprintf("%s_%s", r.Method, r.URL.String())] = r.Body
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Error parsing form data", http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
	response := getResponse(r.FormValue("token"))
	json.NewEncoder(w).Encode(response)
}

// handlePostedMessage implemented according to:
// https://api.slack.com/methods/chat.postMessage
func handlePostedMessage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	log.Printf("PostMessage got new message to channel %s", r.FormValue("channel"))
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	res := map[string]interface{}{
		"ok":      true,
		"channel": r.FormValue("channel"),
		"token":   r.FormValue("token"),
		"ts":      string(time.Now().Nanosecond()),
		"blocks":  r.FormValue("blocks"),
	}

	requestsRecords["/chat.postMessage"] = res
	log.Printf("PostMessage return response message %+v", res)
	json.NewEncoder(w).Encode(res)
}

func handleGetMessage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	log.Printf("GetMessage return message %+v", requestsRecords["/chat.postMessage"])
	json.NewEncoder(w).Encode(requestsRecords["/chat.postMessage"])
	defer delete(requestsRecords, "/chat.postMessage")
}

func main() {
	http.HandleFunc("/auth.test", validateToken)
	http.HandleFunc("/chat.postMessage", handlePostedMessage)
	http.HandleFunc("/chat.getMessage", handleGetMessage)
	log.Printf("Mock server is running on port %d", Port)
	err := http.ListenAndServe(fmt.Sprintf(":%d", Port), nil)
	if err != nil {
		log.Fatalf("failed to start the slack-mock-server")
	}
}

func getResponse(token string) *Response {
	log.Printf("token: %s", token)
	if token == ValidToken {
		log.Printf("token %s is valid", token)
		return &Response{
			Ok:    true,
			Error: "Token is valid.",
		}
	}

	log.Printf("token %s is invalid", token)
	return &Response{
		Ok:    false,
		Error: "Token is invalid.",
	}
}
