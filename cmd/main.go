package main

import (
	"log"
	"net/http"
	"order_service/internal/delivery/grpcclient"
	"order_service/internal/delivery/kafka"
	"order_service/internal/delivery/rest"
	"order_service/internal/repository"
	"order_service/internal/service"
	"os"
	"os/signal"
	"syscall"

	"github.com/sirupsen/logrus"
)

const (
	topic         = "payment_events"
	consumerGroup = "my-consumer-group"
)

var address = []string{"localhost:9091", "localhost:9092", "localhost:9093"}

func main() {
	dbType := repository.DatabaseType(os.Getenv("DB_TYPE"))
	repo, err := repository.NewDatabaseConnection(dbType)
	if err != nil {
		log.Fatal("Error creating repository: ", err)
	}

	productClient, err := grpcclient.NewProductServiceClient("localhost:50051")
	if err != nil {
		log.Fatal("Failed to connect to Product Service:", err)
	}

	paymentClient, err := grpcclient.NewPaymentServiceClient("localhost:50052")
	if err != nil {
		log.Fatal("Failed to connect to Payment Service:", err)
	}

	// Создаём service
	orderService := service.NewOrderService(repo, productClient, paymentClient)

	h := kafka.NewHandler(orderService, productClient)
	c1, err := service.NewConsumer(h, address, topic, consumerGroup, 1)
	if err != nil {
		logrus.Fatal(err)
	}

	c2, err := service.NewConsumer(h, address, topic, consumerGroup, 2)
	if err != nil {
		logrus.Fatal(err)
	}

	c3, err := service.NewConsumer(h, address, topic, consumerGroup, 3)
	if err != nil {
		logrus.Fatal(err)
	}

	go c1.Start()
	go c2.Start()
	go c3.Start()

	// Создаём REST handler
	orderHandler := rest.NewOrderHandler(orderService)

	// Создаём роутер
	router := rest.NewRouter(orderHandler)

	// Настроим CORS
	corsHandler := rest.UserCors(router)

	// Запуск сервера с CORS
	port := ":8081"
	log.Println("Starting server on", port)
	log.Fatal(http.ListenAndServe(port, corsHandler))

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	<-sigChan

	if err := c1.Stop(); err != nil {
		logrus.Error("Ошибка при остановке c1:", err)
	}
	if err := c2.Stop(); err != nil {
		logrus.Error("Ошибка при остановке c2:", err)
	}
	if err := c3.Stop(); err != nil {
		logrus.Error("Ошибка при остановке c3:", err)
	}

	logrus.Info("Консюмеры успешно остановлены, завершаем работу.")
	os.Exit(0) // Завершаем программу после остановки всех консюмеров
}
