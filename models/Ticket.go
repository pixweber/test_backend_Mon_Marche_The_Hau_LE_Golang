package models

type Ticket struct {
	OrderId string
	VAT     float64
	Total   float64
	Products []Product
}