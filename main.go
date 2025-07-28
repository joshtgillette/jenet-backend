package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"slices"

	"github.com/aws/aws-lambda-go/lambda"
	adapter "github.com/awslabs/aws-lambda-go-api-proxy/httpadapter"
	"github.com/joho/godotenv"
	openai "github.com/sashabaranov/go-openai"
)

func isAllowedOrigin(origin string) bool {
	return slices.Contains(
		[]string{
			"https://jenet.ai",
			"https://dev.jenet.ai",
			"http://localhost:3000",
		}, origin)
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if isAllowedOrigin(origin) {
			w.Header().Set("Access-Control-Allow-Origin", origin)
		}
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func taglineHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("generated ui for <i>you</i>, coming soon"))
}

func messageHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		getMessages(w, r)
	case http.MethodPost:
		setMessages(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func setMessages(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	type requestBody struct {
		Text string `json:"text"`
	}
	var body requestBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	if body.Text == "" {
		http.Error(w, "Missing 'text' in request body", http.StatusBadRequest)
		return
	}

	// fmt.Printf("received message: %v\n", body.Text)
}

func getMessages(w http.ResponseWriter, r *http.Request) {
	testData := map[string]interface{}{
		"text":  []string{"Jody Gillette", "Hey how are you doing?", "6/17/2025", "10:50"},
		"event": []string{"Jody Gillette", "Free for a call later?", "6/17/2025", "10:50"},
	}

	jsonData, err := json.Marshal(testData)
	if err != nil {
		http.Error(w, "Failed to encode JSON", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonData)
}

func modelHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	type requestBody struct {
		Text    string `json:"text"`
		Context string `json:"context"`
	}
	var body requestBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	if body.Text == "" {
		http.Error(w, "Missing 'text' in request body", http.StatusBadRequest)
		return
	}

	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		http.Error(w, "OpenAI API key not set", http.StatusInternalServerError)
		return
	}

	client := openai.NewClient(apiKey)
	resp, err := client.CreateChatCompletion(r.Context(), openai.ChatCompletionRequest{
		Model: openai.GPT4Dot1Mini,
		Messages: []openai.ChatCompletionMessage{
			{Role: openai.ChatMessageRoleSystem, Content: body.Context},
			{Role: openai.ChatMessageRoleUser, Content: body.Text},
		},
	})
	if err != nil {
		http.Error(w, "OpenAI API error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	if len(resp.Choices) == 0 {
		http.Error(w, "No response from OpenAI", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(resp.Choices[0].Message.Content))
}

func notFoundHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
	w.Write([]byte("Page not found"))
}

func main() {
	_ = godotenv.Load() // Loads .env file if present

	router := http.NewServeMux()
	middleware_wrapper := corsMiddleware(router)

	router.HandleFunc("/tagline", taglineHandler)
	router.HandleFunc("/model", modelHandler)
	router.HandleFunc("/message", messageHandler)

	if os.Getenv("LOCAL") == "1" {
		// Run as a local HTTP server for development
		port := os.Getenv("PORT")
		if port == "" {
			port = "8080"
		}
		fmt.Printf("\nStarting local server at http://localhost:%s\n\n", port)
		http.ListenAndServe(":"+port, middleware_wrapper)
	} else {
		// Run as AWS Lambda
		lambda.Start(adapter.NewV2(middleware_wrapper).ProxyWithContext)
	}
}
