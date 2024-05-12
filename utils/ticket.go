package utils

import (
	"fmt"
	"strconv"
	"strings"

	"test_backend_Mon_Marche_The_Hau_LE_Golang/models"
)

func ValidTicket(input string) bool {
	// Split input by lines
	lines := strings.Split(input, "\n")

	// Check if there are at least three lines (header + at least one product)
	if len(lines) < 3 {
		return false
	}

	// Check if there is an empty line separating header and products
	if strings.TrimSpace(lines[3]) != "" {
		return false
	}

	// Check if the fifth line starts with "product"
	if !strings.HasPrefix(lines[4], "product") {
		return false
	}

	// Check totals
	if !validateTotal(input) {
		return false
	}

	// Input seems valid
	return true
}

func validateTotal(ticketText string) bool {
	// Parse ticket text
	lines := strings.Split(ticketText, "\n")
	var totalFromTicket float64
	var products []models.Product

	totalLine := lines[2]
	totalFields := strings.Split(totalLine, ": ")
	totalFromTicket, err := parsePrice(totalFields[1])
	if err != nil {
		// Return false if total is not a valid number
		return false
	}

	// Calculate total price of products
	for index, line := range lines {
		// Starting from the line with the products
		if index < 5 {
			continue
		}

		productFields := strings.Split(line, ",")
		if len(productFields) != 3 {
			continue
		}

		price, err := parsePrice(productFields[2])
		if err != nil {
			// Skip product if price is not a valid number
			continue
		}

		products = append(products, models.Product{
			Name:  productFields[0],
			ID:    productFields[1],
			Price: price,
		})
	}

	// Calculate total price of products
	var totalPrice float64
	for _, p := range products {
		totalPrice += p.Price
	}

	// Check if order total matches
	return totalPrice == totalFromTicket
}

func ParseTicketFromString(ticketText string) (models.Ticket, error) {
	var ticket models.Ticket

	lines := strings.Split(ticketText, "\n")

	// Read header
	ticket.OrderId = getValueFromLine(lines[0])

	// Read VAT
	vatLine := getValueFromLine(lines[1])
	vat, err := parsePrice(vatLine)
	if err != nil {
		return ticket, fmt.Errorf("failed to parse VAT: %v", err)
	}
	ticket.VAT = vat

	// Read Total
	totalLine := getValueFromLine(lines[2])
	total, err := parsePrice(totalLine)
	if err != nil {
		return ticket, fmt.Errorf("failed to parse Total: %v", err)
	}
	ticket.Total = total

	// Read Products
	for i := 5; i < len(lines); i++ {
		line := lines[i]
		if line == "" {
			continue // Skip empty lines
		}
		parts := strings.Split(line, ",")
		if len(parts) != 3 {
			return ticket, fmt.Errorf("invalid product format: %s", line)
		}
		price, err := parsePrice(parts[2])
		if err != nil {
			return ticket, fmt.Errorf("failed to parse price: %v", err)
		}
		product := models.Product{
			Name:  parts[0],
			ID:    parts[1],
			Price: price,
		}
		ticket.Products = append(ticket.Products, product)
	}

	return ticket, nil
}

func getValueFromLine(line string) string {
	return strings.TrimSpace(strings.Split(line, ":")[1])
}

func parsePrice(s string) (float64, error) {
	return strconv.ParseFloat(strings.TrimSpace(s), 64)
}