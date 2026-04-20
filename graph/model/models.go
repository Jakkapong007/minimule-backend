// Package model contains domain types used throughout minimule-backend.
// Core types are defined in the sibling files:
//   - user.go    — User, UserProfile, UserAddress, UserPaymentMethod
//   - product.go — Product, Category, ProductImage, ProductVariant
//   - cart.go    — Cart, CartItem
//   - order.go   — Order, OrderItem
package model

// CreateProductInput is the input type for the createProduct mutation.
type CreateProductInput struct {
	Name        string
	Description *string
	BasePrice   float64
	StockQty    int
	CategoryID  string
}
