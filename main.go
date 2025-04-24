package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/joho/godotenv"
)

func main() {
	if godotenv.Load() != nil {
		fmt.Println("ERROR loading .env file")
		return
	}

	http.HandleFunc("/tagline", func(w http.ResponseWriter, r *http.Request) {
		devFrontendOrigin := os.Getenv("FRONTEND_DEV_ORIGIN")
		if devFrontendOrigin != "" {
			w.Header().Set("Access-Control-Allow-Origin", devFrontendOrigin)
		}

		fmt.Fprintf(w, "generative ui for <i><b>you</b></i> - coming soon")
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	http.ListenAndServe(":"+port, nil)
}
