package main

import (
	"embed"
	"fmt"
	"log"
	"net/http"

	"ServerMonitor/internal/api"
	"ServerMonitor/internal/service"
)

//go:embed index.html
var content embed.FS

func main() {
	// 1. Initialize Docker Client
	if err := service.InitDockerClient(); err != nil {
		log.Fatalf("Error connecting to Docker: %v", err)
	}

	// 2. Setup Routes
	mux := http.NewServeMux()

	// Static Files
	mux.Handle("/", http.FileServer(http.FS(content)))

	// API Routes
	api.RegisterRoutes(mux)

	// 3. Start Server
	fmt.Println("Server running on port :8080...")
	log.Fatal(http.ListenAndServe(":8080", mux))
}
