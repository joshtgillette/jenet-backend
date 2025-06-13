package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/aws/aws-lambda-go/lambda"
	adapter "github.com/awslabs/aws-lambda-go-api-proxy/httpadapter"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	openai "github.com/sashabaranov/go-openai"
)

func taglineHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("generated ui for <i>you</i>, coming soon"))
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

	r := mux.NewRouter()
	r.HandleFunc("/tagline", taglineHandler).Methods("GET")
	r.HandleFunc("/model", modelHandler).Methods("POST")
	r.NotFoundHandler = http.HandlerFunc(notFoundHandler)

	if os.Getenv("LOCAL") == "1" {
		// Run as a local HTTP server for development
		port := os.Getenv("PORT")
		if port == "" {
			port = "8080"
		}
		fmt.Printf("\nStarting local server at http://localhost:%s\n\n", port)
		panic(http.ListenAndServe(":"+port, r))
	} else {
		// Run as AWS Lambda
		lambda.Start(adapter.NewV2(r).ProxyWithContext)
	}
}
