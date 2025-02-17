package repository

import (
	"database/sql"
	"order_service/internal/entity"
)

type PostgresOrderRepository struct {
	db *sql.DB
}

func NewPostgresOrderRepository(db *sql.DB) OrderRepository {
	return &PostgresOrderRepository{db: db}
}

func (r *PostgresOrderRepository) Create(order *entity.Order) error {
	// Создаем заказ и получаем его ID
	err := r.db.QueryRow(
		"INSERT INTO orders (user_id, total_price, status, created_at) VALUES ($1, $2, $3, $4) RETURNING id",
		order.UserID, order.TotalPrice, order.Status, order.CreatedAt,
	).Scan(&order.ID)
	if err != nil {
		return err
	}

	// Вставляем товары в заказ
	stmt, err := r.db.Prepare("INSERT INTO order_items (order_id, product_id, quantity, price) VALUES ($1, $2, $3, $4)")
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, item := range order.Items {
		_, err := stmt.Exec(order.ID, item.ProductID, item.Quantity, item.Price)
		if err != nil {
			return err
		}
	}

	return nil
}

func (r *PostgresOrderRepository) GetOrderByID(orderID int64) (*entity.Order, error) {
	var order entity.Order
	err := r.db.QueryRow("SELECT id, user_id, total_price, status, created_at FROM orders WHERE id=$1", orderID).
		Scan(&order.ID, &order.UserID, &order.TotalPrice, &order.Status, &order.CreatedAt)
	if err != nil {
		return nil, err
	}

	rows, err := r.db.Query("SELECT product_id, quantity, price FROM order_items WHERE order_id=$1", orderID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var item entity.OrderItem
		if err := rows.Scan(&item.ProductID, &item.Quantity, &item.Price); err != nil {
			return nil, err
		}
		order.Items = append(order.Items, item)
	}

	return &order, nil
}
