package repository

import "order_service/internal/entity"

type OrderRepository interface {
	Create(order *entity.Order) error
	GetOrderByID(orderID int64) (*entity.Order, error)
}
