package service

import (
	"fmt"
	"order_service/internal/delivery/grpcclient"
	"order_service/internal/entity"
	"order_service/internal/repository"
)

type OrderService struct {
	repo          repository.OrderRepository
	productClient grpcclient.ProductServiceClientInterface
	paymentClient grpcclient.PaymentServiceClientInterface
}

func NewOrderService(repo repository.OrderRepository, productClient grpcclient.ProductServiceClientInterface, paymentClient grpcclient.PaymentServiceClientInterface) OrderServiceInterface {
	return &OrderService{
		repo:          repo,
		productClient: productClient,
		paymentClient: paymentClient,
	}
}

func (s *OrderService) CreateOrder(userID int64, items []entity.OrderItem, totalPrice float64) (*entity.PaymentResponse, error) {
	// Проверка на дубликаты продуктов
	seen := make(map[int64]bool)
	var productIDs []int64
	for _, item := range items {
		if seen[item.ProductID] {
			return nil, fmt.Errorf("duplicate product id: %d", item.ProductID)
		}
		seen[item.ProductID] = true
		productIDs = append(productIDs, item.ProductID)
	}

	// Проверка наличия продуктов и их стока
	stockMap, err := s.productClient.GetProductStock(productIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to get product stock: %w", err)
	}

	for index, item := range items {
		availableStock, exists := stockMap[item.ProductID]
		if !exists {
			return nil, fmt.Errorf("product %d not found", item.ProductID)
		}

		if (availableStock.Stock /*- reservedStock*/) < item.Quantity {
			return nil, fmt.Errorf("not enough stock for product %d", item.ProductID)
		}
		items[index].Name = availableStock.Name
	}

	// Создание заказа
	order := &entity.Order{
		UserID:     userID,
		Items:      items,
		TotalPrice: totalPrice,
		Status:     "pending",
	}
	order, err = s.repo.Create(order)
	if err != nil {
		return nil, err
	}

	// Генерация ссылки на оплату
	payment, err := s.paymentClient.GeneratePaymentLink(userID, order.ID, totalPrice)
	if err != nil {
		return nil, fmt.Errorf("failed to get payment link: %w", err)
	}

	return &entity.PaymentResponse{PaymentURL: payment.PaymentUrl}, nil
}

func (u *OrderService) GetOrderByID(orderID int64) (*entity.Order, error) {
	return u.repo.GetOrderByID(orderID)
}

func (u *OrderService) DeleteOrder(orderID int64) error {
	return u.repo.Delete(orderID)
}

func (u *OrderService) UpdateOrderStatus(orderID int64, status string) error {
	// Проверяем, допустим ли такой статус
	validStatuses := map[string]bool{
		"paid":      true,
		"shipped":   true,
		"delivered": true,
		"canceled":  true,
	}

	if !validStatuses[status] {
		return fmt.Errorf("недопустимый статус: %s", status)
	}

	// Получение заказа
	order, err := u.repo.GetOrderByID(orderID)
	if err != nil {
		return fmt.Errorf("не удалось найти заказ: %w", err)
	}

	// Обновляем статус заказа
	order.Status = status
	if err := u.repo.UpdateOrder(order); err != nil {
		return fmt.Errorf("ошибка при обновлении заказа: %w", err)
	}

	return nil
}

func (s *OrderService) GetOrdersByUserID(userID int64) ([]entity.Order, error) {
	return s.repo.GetOrdersByUserID(userID)
}

func (s *OrderService) CancelOrder(userID int64, orderID int64) error {
	return s.repo.CancelOrder(userID, orderID)
}
