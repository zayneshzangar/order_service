package main

import (
	"log"
	"net/http"
	"order_service/internal/delivery/rest"
	"order_service/internal/repository"
	"order_service/internal/usecase"
	"os"
)

func main() {
	dbType := repository.DatabaseType(os.Getenv("DB_TYPE"))
	repo, err := repository.NewDatabaseConnection(dbType)
	if err != nil {
		log.Fatal("Error creating repository: ", err)
	}

	// Создаём usecase
	orderUseCase := usecase.NewOrderUseCase(repo)

	// Создаём REST handler
	orderHandler := rest.NewOrderHandler(orderUseCase)

	// Создаём роутер и запускаем сервер
	router := rest.NewRouter(orderHandler)
	port := ":8081"
	log.Println("Starting server on", port)
	log.Fatal(http.ListenAndServe(port, router))
}
