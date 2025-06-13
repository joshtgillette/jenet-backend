package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/aws/aws-lambda-go/lambda"
	adapter "github.com/awslabs/aws-lambda-go-api-proxy/httpadapter"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
)

func taglineHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Generated ui for <i>you</i>, coming soon"))
}

func notFoundHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
	w.Write([]byte("Page not found"))
}

func main() {
	_ = godotenv.Load() // Loads .env file if present

	r := mux.NewRouter()
	r.HandleFunc("/tagline", taglineHandler).Methods("GET")
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
		lambda.Start(adapter.New(r).ProxyWithContext)
	}
}
