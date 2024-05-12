package consumers

import (
	"log"
	"test_backend_Mon_Marche_The_Hau_LE_Golang/persistence"
)

func ConsumeTickets(db persistence.Persistence) {
	err := db.ConsumeTicketsFromQueue()
	if err != nil {
		log.Fatalf("Failed to consume tickets from RabbitMQ: %v", err)
	}
}
