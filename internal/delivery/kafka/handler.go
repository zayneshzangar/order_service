package kafka

import (
	"encoding/json"
	"order_service/internal/delivery/grpcclient"
	"order_service/internal/service"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/sirupsen/logrus"
)

type Handler struct {
	orderService   service.OrderServiceInterface
	productService *grpcclient.ProductServiceClient
}

func NewHandler(orderService service.OrderServiceInterface, productService *grpcclient.ProductServiceClient) *Handler {
	return &Handler{orderService: orderService, productService: productService}
}

func (h *Handler) HandleMessage(message []byte, topic kafka.TopicPartition, cn int64) error {
	logrus.Infof("Consumer #%d, Message from kafka with offset %d '%s, on partition %d", cn, topic.Offset, string(message), topic.Partition)

	// Парсим сообщение (допустим, JSON с полем orderID)
	var event struct {
		OrderID int64  `json:"order_id"`
		Status  string `json:"status"`
	}
	if err := json.Unmarshal(message, &event); err != nil {
		logrus.Error("Ошибка парсинга Kafka-сообщения:", err)
		return err
	}

	// Обновляем заказ в БД
	if err := h.orderService.UpdateOrderStatus(event.OrderID, event.Status); err != nil {
		logrus.Error("Ошибка обновления статуса заказа:", err)
		return err
	}

	order, err := h.orderService.GetOrderByID(event.OrderID)
	if err != nil {
		logrus.Error("Ошибка получения заказа:", err)
		return err
	}

	err = h.productService.UpdateProductStock(order)
	if err != nil {
		logrus.Error("Ошибка обновления остатков товаров:", err)
		return err
	}

	return nil
}
