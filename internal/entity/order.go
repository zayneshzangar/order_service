package entity

import "time"

type Order struct {
	ID         int64
	UserID     int64
	Items      []OrderItem
	TotalPrice float64
	Status     string // "pending", "paid", "shipped"
	CreatedAt  time.Time
}

type OrderItem struct {
	ProductID int64
	Quantity  int
	Price     float64 // Цена на момент заказа
}
