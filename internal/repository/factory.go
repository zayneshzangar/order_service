package repository

import (
	"database/sql"
	"fmt"
	"os"

	_ "github.com/lib/pq"
)

// DatabaseType определяет доступные базы данных
type DatabaseType string

const (
	Postgres DatabaseType = "postgres"
	Mongo    DatabaseType = "mongo"
)


// NewDatabaseConnection устанавливает соединение с базой данных
func NewDatabaseConnection(dbType DatabaseType) (OrderRepository, error) {
	switch dbType {
	case Postgres:
		return NewPostgresRepository()
	// case Mongo:
	// 	return NewMongoRepository() // Когда добавишь Mongo
	default:
		return nil, fmt.Errorf("unsupported database type: %s", dbType)
	}
}

// NewPostgresRepository создает подключение к Postgres
func NewPostgresRepository() (OrderRepository, error) {
	connStr := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
		os.Getenv("USER_ORDER_SERVICE"),
		os.Getenv("PASSWORD_ORDER_SERVICE"),
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_ORDER_SERVICE"),
		os.Getenv("DB_SSLMODE"),
	)
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %v", err)
	}

	err = db.Ping()
	if err != nil {
		return nil, fmt.Errorf("failed to ping database: %v", err)
	}

	return &PostgresOrderRepository{db: db}, nil
}
