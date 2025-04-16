package repository

import (
	"context"
	"database/sql"
	"order_service/internal/entity"
)

type OrderRepository interface {
	Create(order *entity.Order) (*entity.Order, error)
	GetOrderByID(orderID int64) (*entity.Order, error)
	Delete(orderID int64) error
	ReserveStock(ctx context.Context, tx *sql.Tx, orderID, productID int64, quantity int64) error
	GetAvailableStock(ctx context.Context, productID int64) (int64, error)
	BeginTransaction() (*sql.Tx, error)
	ClearExpiredReservations(ctx context.Context) ([]int64, error)
	UpdateOrder(order *entity.Order) error
	GetOrdersByUserID(userID int64) ([]entity.Order, error)
	CancelOrder(userID int64, orderID int64) error
}
