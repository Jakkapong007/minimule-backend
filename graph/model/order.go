package model

import "time"

type OrderStatus string

const (
	OrderPending    OrderStatus = "pending"
	OrderPaid       OrderStatus = "paid"
	OrderProcessing OrderStatus = "processing"
	OrderShipped    OrderStatus = "shipped"
	OrderDelivered  OrderStatus = "delivered"
	OrderCancelled  OrderStatus = "cancelled"
	OrderRefunded   OrderStatus = "refunded"
)

type Order struct {
	ID          string
	OrderNumber string
	UserID      string
	AddressID   string

	Subtotal       float64
	DiscountAmount float64
	ShippingFee    float64
	Total          float64

	Status OrderStatus

	PromotionCode    *string
	ShippingMethodID *string

	CreatedAt time.Time
	UpdatedAt *time.Time

	Items []*OrderItem
	User  *User
}

type OrderItem struct {
	ID             string
	OrderID        string
	ProductID      string
	VariantID      *string
	CustomDesignID *string

	Quantity  int
	UnitPrice float64
	Subtotal  float64

	Product *Product
}
