package entity

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Order struct {
	ID         int64
	UserID     int64
	Items      []OrderItem
	TotalPrice float64
	Status     string
	CreatedAt  time.Time
}

type OrderItem struct {
	ID        int64   `json:"id"`
	ProductID int64   `json:"product_id"`
	Name      string  `json:"name"`
	Quantity  int64   `json:"quantity"`
	Price     float64 `json:"price"`
}

type PaymentResponse struct {
	PaymentURL string `json:"payment_url"`
}

type ReduceProductStock struct {
	ProductID int64 `json:"product_id"`
	Quantity  int64 `json:"quantity"`
}

type Claims struct {
	UserID int    `json:"user_id"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}
