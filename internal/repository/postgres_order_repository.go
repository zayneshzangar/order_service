package repository

import (
	"context"
	"database/sql"
	"fmt"
	"order_service/internal/entity"
)

type PostgresOrderRepository struct {
	db *sql.DB
}

func NewPostgresOrderRepository(db *sql.DB) OrderRepository {
	return &PostgresOrderRepository{db: db}
}

func (r *PostgresOrderRepository) Create(order *entity.Order) (*entity.Order, error) {
	// Создаем заказ и получаем его ID
	err := r.db.QueryRow(
		"INSERT INTO orders (user_id, total_price, status, created_at) VALUES ($1, $2, $3, $4) RETURNING id",
		order.UserID, order.TotalPrice, order.Status, order.CreatedAt,
	).Scan(&order.ID)
	if err != nil {
		return nil, err
	}

	// Вставляем товары в заказ
	stmt, err := r.db.Prepare("INSERT INTO order_items (order_id, product_id, name, quantity, price) VALUES ($1, $2, $3, $4, $5)")
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	for _, item := range order.Items {
		_, err := stmt.Exec(order.ID, item.ProductID, item.Name, item.Quantity, item.Price)
		if err != nil {
			return nil, err
		}
	}

	return order, nil
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

func (r *PostgresOrderRepository) Delete(orderID int64) error {
	_, err := r.db.Exec("DELETE FROM orders WHERE id = $1", orderID)
	return err
}

// STOCK methods
func (r *PostgresOrderRepository) BeginTransaction() (*sql.Tx, error) {
	return r.db.Begin()
}

func (r *PostgresOrderRepository) ReserveStock(ctx context.Context, tx *sql.Tx, orderID, productID int64, quantity int64) error {
	_, err := tx.ExecContext(ctx, "INSERT INTO reserved_stock (order_id, product_id, quantity) VALUES ($1, $2, $3)", orderID, productID, quantity)
	return err
}

func (r *PostgresOrderRepository) GetAvailableStock(ctx context.Context, productID int64) (int64, error) {
	var availableStock int64
	query := `
		SELECT stock - COALESCE((SELECT SUM(quantity) FROM reserved_stock WHERE product_id = $1), 0) 
		FROM products WHERE id = $1
	`
	err := r.db.QueryRowContext(ctx, query, productID).Scan(&availableStock)
	return availableStock, err
}

func (r *PostgresOrderRepository) ClearExpiredReservations(ctx context.Context) ([]int64, error) {
	rows, err := r.db.QueryContext(ctx, `DELETE FROM reserved_stock WHERE created_at < NOW() - INTERVAL '5 minutes' RETURNING order_id`)
	if err != nil {
		return []int64{}, err // Возвращаем ПУСТОЙ массив вместо nil
	}
	defer rows.Close()

	var orderIDs []int64
	for rows.Next() {
		var orderID int64
		if err := rows.Scan(&orderID); err != nil {
			return []int64{}, err // Ошибка чтения → пустой массив
		}
		orderIDs = append(orderIDs, orderID)
	}
	return orderIDs, nil
}

// UpdateOrder обновляет заказ в базе данных
func (r *PostgresOrderRepository) UpdateOrder(order *entity.Order) error {
	_, err := r.db.Exec("UPDATE orders SET status = $1 WHERE id = $2", order.Status, order.ID)
	if err != nil {
		return fmt.Errorf("ошибка при обновлении заказа: %w", err)
	}
	return nil
}

func (r *PostgresOrderRepository) GetOrdersByUserID(userID int64) ([]entity.Order, error) {
	query := `SELECT id, user_id, total_price, status, created_at FROM orders WHERE user_id = $1`
	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orders []entity.Order
	for rows.Next() {
		var order entity.Order
		if err := rows.Scan(&order.ID, &order.UserID, &order.TotalPrice, &order.Status, &order.CreatedAt); err != nil {
			return nil, err
		}
		order.Items, err = r.getProductsByOrderID(order.ID)
		if err != nil {
			return nil, err
		}
		orders = append(orders, order)
	}

	return orders, nil
}

func (r *PostgresOrderRepository) getProductsByOrderID(orderID int64) ([]entity.OrderItem, error) {
	rows, err := r.db.Query(`SELECT product_id, name, quantity FROM order_items WHERE order_id = $1`, orderID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var products []entity.OrderItem
	for rows.Next() {
		var p entity.OrderItem
		if err := rows.Scan(&p.ProductID, &p.Name, &p.Quantity); err != nil {
			return nil, err
		}
		products = append(products, p)
	}

	return products, nil
}

func (r *PostgresOrderRepository) CancelOrder(userID int64, orderID int64) error {
	_, err := r.db.Exec("UPDATE orders SET status = 'cancelled' WHERE id = $1 AND user_id = $2", orderID, userID)
	return err
}
