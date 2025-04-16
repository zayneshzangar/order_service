package service

import (
	"order_service/internal/entity"
)

type OrderService interface {
	CreateOrder(UserID int64, Items []entity.OrderItem, TotalPrice float64) (*entity.PaymentResponse, error)
	GetOrderByID(orderID int64) (*entity.Order, error)
	GetOrdersByUserID(userID int64) ([]entity.Order, error)
	UpdateOrderStatus(orderID int64, status string) error
	DeleteOrder(orderID int64) error
	CancelOrder(userID int64, orderID int64) error
}
