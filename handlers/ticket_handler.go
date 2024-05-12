package handlers

import (
	"fmt"
	"io/ioutil"
	"net/http"

	"test_backend_Mon_Marche_The_Hau_LE_Golang/persistence"
)

func HandleTicketHTTP(w http.ResponseWriter, r *http.Request, db persistence.Persistence) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusInternalServerError)
		return
	}

	ticket := string(body)
	err = db.PublishTicketToQueue(ticket)
	if err != nil {
		http.Error(w, "Failed to publish ticket to RabbitMQ", http.StatusInternalServerError)
		return
	}

	fmt.Fprintf(w, "Ticket published successfully")
}
