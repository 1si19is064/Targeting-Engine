package main

import (
	"log"
	"net/http"
	"os"

	"Targeting-Engine/internal/controllers"
	"Targeting-Engine/internal/database"
	"Targeting-Engine/internal/services"
	"Targeting-Engine/pkg/config"

	"github.com/gorilla/mux"
)

func main() {
	// Load configuration
	cfg := config.Load()

	// Initialize database
	db, err := database.Connect(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Run migrations
	if err := database.Migrate(db); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	// Initialize services
	targetingService := services.NewTargetingService(db)

	// Initialize controllers
	deliveryController := controllers.NewDeliveryController(targetingService)

	// Setup routes
	router := mux.NewRouter()
	router.HandleFunc("/v1/delivery", deliveryController.GetCampaigns).Methods("GET")
	router.HandleFunc("/health", healthCheck).Methods("GET")

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server starting on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, router))
}

func healthCheck(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}
