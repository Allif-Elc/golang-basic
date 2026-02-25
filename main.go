package main

import (
	"fmt"
	"golang-basic/api/internal/config"
	"golang-basic/api/internal/routes"
	"log"
	"net/http"

	_ "github.com/joho/godotenv/autoload"
)

func main() {
	if err := config.InitDatabase(); err != nil {
		log.Printf("Warning: Failed to initialize database: %v", err)
		log.Println("Server will continue running without database connection")
	} else {
		defer config.CloseDatabase()
	}

	if err := config.InitMinio(); err != nil {
		log.Printf("Warning: MinIO not available: %v", err)
	}

	r := routes.SetupRoutes()

	port := ":3003"
	fmt.Printf("Server is running on HTTPS port %s\n", port)
	log.Fatal(http.ListenAndServeTLS(port, "server.crt", "server.key", r))
}
