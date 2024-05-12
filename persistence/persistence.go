package persistence

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"sync"
	"test_backend_Mon_Marche_The_Hau_LE_Golang/utils"

	"database/sql"
	_ "github.com/lib/pq"
	"github.com/streadway/amqp"
)

type Persistence interface {
	PublishTicketToQueue(ticket string) error
	ConsumeTicketsFromQueue() error
}

// DataPersistence implements the Persistence interface using SQL database and RabbitMQ
type DataPersistence struct {
	db            *sql.DB
	rabbitMQChannel *amqp.Channel
}

func NewDataPersistence() (Persistence, error) {
	connStr := fmt.Sprintf("host=%s port=%s dbname=%s user=%s password=%s sslmode=disable",
		os.Getenv("DB_HOST"), os.Getenv("DB_PORT"), os.Getenv("DB_NAME"), os.Getenv("DB_USER"), os.Getenv("DB_PASSWORD"))
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("Failed to connect to PostgreSQL: %v", err)
	}

	// Connect to RabbitMQ
	conn, err := amqp.Dial(os.Getenv("RABBITMQ_URL"))
	if err != nil {
		return nil, fmt.Errorf("Failed to connect to RabbitMQ: %v", err)
	}

	// Open a channel
	rabbitMQChannel, err := conn.Channel()
	if err != nil {
		return nil, fmt.Errorf("Failed to open a RabbitMQ channel: %v", err)
	}

	persistence := &DataPersistence{db: db, rabbitMQChannel: rabbitMQChannel}

	// Ensure tables exist
	if err := persistence.ensureTablesExist(); err != nil {
		return nil, fmt.Errorf("Failed to ensure tables exist: %v", err)
	}

	return persistence, nil
}

func (d *DataPersistence) ensureTablesExist() error {
	// SQL statement to create the "tickets" table
	createTicketsTableSQL := `
	CREATE TABLE IF NOT EXISTS tickets (
		id SERIAL PRIMARY KEY,
		order_id VARCHAR(255),
		vat NUMERIC,
		total NUMERIC,
	    valid BOOLEAN,
	    ticket_text TEXT,
	    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);
	`

	// SQL statement to create the "products" table
	createProductsTableSQL := `
	CREATE TABLE IF NOT EXISTS products (
		id SERIAL PRIMARY KEY,
		name VARCHAR(255),
		product_id VARCHAR(255) UNIQUE,
		price NUMERIC,
	    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);
	`

	// Execute the SQL statements to create tables
	_, err := d.db.Exec(createTicketsTableSQL)
	if err != nil {
		return fmt.Errorf("Failed to create tickets table: %v", err)
	}

	_, err = d.db.Exec(createProductsTableSQL)
	if err != nil {
		return fmt.Errorf("Failed to create products table: %v", err)
	}

	return nil
}

func (d *DataPersistence) PublishTicketToQueue(ticket string) error {
	err := d.rabbitMQChannel.Publish(
		"",              // Exchange
		"tickets_queue", // Routing key (queue name)
		false,           // Mandatory
		false,           // Immediate
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        []byte(ticket),
		})
	if err != nil {
		return fmt.Errorf("Failed to publish ticket to RabbitMQ: %v", err)
	}

	return nil
}

func (d *DataPersistence) ConsumeTicketsFromQueue() error {
	var wg sync.WaitGroup

	// Declare a queue
	q, err := d.rabbitMQChannel.QueueDeclare(
		"tickets_queue", // Queue name
		true,           // Durable
		false,  	   // Delete when unused
		false,         // Exclusive
		false,          // No-wait
		nil,              // Arguments
	)
	if err != nil {
		return fmt.Errorf("Failed to declare a queue: %v", err)
	}

	// Get max consumers from environment variable
	maxConsumersStr := os.Getenv("MAX_CONSUMERS")
	maxConsumers, err := strconv.Atoi(maxConsumersStr)
	if err != nil {
		return fmt.Errorf("Failed to parse MAX_CONSUMERS: %v", err)
	}

	// Create a pool to limit concurrent consumers
	pool := make(chan struct{}, maxConsumers)

	// Consume messages from the queue
	msgs, err := d.rabbitMQChannel.Consume(
		q.Name, // Queue
		"",     // Consumer
		false,  // Auto-Ack (set to false)
		false,  // Exclusive
		false,  // No-local
		false,  // No-Wait
		nil,    // Args
	)
	if err != nil {
		return fmt.Errorf("Failed to register a consumer: %v", err)
	}

	// Process incoming messages
	for msg := range msgs {
		// Acquire pool
		pool <- struct{}{}

		// Increment wait group
		wg.Add(1)

		// Process the message in a goroutine
		go func(msg amqp.Delivery) {
			defer func() {
				// Release pool
				<-pool

				// Decrement wait group
				wg.Done()
			}()

			// Process the message (store ticket in DB, etc.)
			body := string(msg.Body)
			err := storeTicket(d.db, body)
			if err != nil {
				log.Printf("Failed to store ticket: %v", err)
			} else {
				log.Printf("Ticket stored successfully: %s", body)
			}

			// Acknowledge the message after processing
			msg.Ack(false)
		}(msg)
	}

	// Wait for all goroutines to finish
	wg.Wait()

	return nil
}

func storeTicket(db *sql.DB, ticketString string) error {
	// Initialize ticket variables
	ticketOrderID := ""
	ticketVAT := 0.0
	ticketTotal := 0.0
	ticketValid := false

	// Parse ticket
	ticket, err := utils.ParseTicketFromString(ticketString)
	if err != nil {
		return err
	}

	// Check if ticket is valid
	if utils.ValidTicket(ticketString) {
		// parse ticket using ParseTicketFromString
		ticket, err := utils.ParseTicketFromString(ticketString)
		if err != nil {
			return err
		}

		ticketOrderID = ticket.OrderId
		ticketVAT = ticket.VAT
		ticketTotal = ticket.Total
		ticketValid = true

		fmt.Println(ticketValid)
	}

	// Insert ticket and products into the database
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Insert ticket even if it is invalid
	_, err = tx.Exec("INSERT INTO tickets (order_id, vat, total, valid, ticket_text) " +
		"VALUES ($1, $2, $3, $4, $5)",
		ticketOrderID, ticketVAT, ticketTotal, ticketValid, ticketString)
	if err != nil {
		return err
	}

	// Insert products only if ticket is valid
	if ticketValid == true {
		fmt.Println("Inserting products")
		fmt.Println("ticket: ", ticket)
		for _, product := range ticket.Products {
			fmt.Println(product.Name, product.ID, product.Price)

			_, err := tx.Exec("INSERT INTO products (name, product_id, price) " +
				"VALUES ($1, $2, $3) ON CONFLICT (product_id) DO NOTHING",
				product.Name, product.ID, product.Price)

			if err != nil {
				return err
			}
		}
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}