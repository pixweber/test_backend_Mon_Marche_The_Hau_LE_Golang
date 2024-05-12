package main

import (
	"github.com/joho/godotenv"
	"log"
	"net/http"
	"test_backend_Mon_Marche_The_Hau_LE_Golang/consumers"
	"test_backend_Mon_Marche_The_Hau_LE_Golang/handlers"
	"test_backend_Mon_Marche_The_Hau_LE_Golang/persistence"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	// Initialize DataPersistence
	db, err := persistence.NewDataPersistence()
	if err != nil {
		log.Fatalf("Failed to initialize DataPersistence: %v", err)
	}

	// Initialize HTTP handlers
	http.HandleFunc("/ticket", func(w http.ResponseWriter, r *http.Request) {
		handlers.HandleTicketHTTP(w, r, db)
	})

	// Initialize RabbitMQ consumer
	go consumers.ConsumeTickets(db)

	// Start HTTP server
	log.Println("HTTP server started on :8080. RabbitMQ consumer started.")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
