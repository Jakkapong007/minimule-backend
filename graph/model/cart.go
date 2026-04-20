package model

import "time"

type CartStatus string

const (
	CartActive     CartStatus = "active"
	CartCheckedOut CartStatus = "checked_out"
	CartAbandoned  CartStatus = "abandoned"
)

type Cart struct {
	ID           string
	UserID       *string
	SessionToken *string
	Status       CartStatus
	UpdatedAt    time.Time

	Items []*CartItem
}

type CartItem struct {
	ID             string
	CartID         string
	ProductID      string
	VariantID      *string
	CustomDesignID *string

	Quantity  int
	UnitPrice float64

	Product *Product
}
