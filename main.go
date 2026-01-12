package main

import (
	"fmt"
	"golang-basic/internal/config"
	"golang-basic/internal/routes"
	"log"
	"net/http"
)

func main() {
	if err := config.InitDatabase(); err != nil {
		log.Printf("Warning: Failed to initialize database: %v", err)
		log.Println("Server will continue running without database connection")
	} else {
		defer config.CloseDatabase()
	}

	routes.SetupRoutes()

	port := ":3003"
	fmt.Printf("Server is running on HTTPS port %s\n", port)
	log.Fatal(http.ListenAndServeTLS(port, "server.crt", "server.key", nil))
}
