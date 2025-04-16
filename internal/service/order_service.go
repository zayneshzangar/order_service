package service

import (
	"fmt"
	"log"
	"order_service/internal/delivery/grpcclient"
	"order_service/internal/entity"
	"order_service/internal/repository"
	"slices"
	"time"
)

type orderService struct {
	orderRepo                repository.OrderRepository
	grpcClientOrderService   *grpcclient.ProductServiceClient
	grpcClientPaymentService *grpcclient.PaymentServiceClient
}

func NewOrderService(orderRepo repository.OrderRepository, grpcClientOrderService *grpcclient.ProductServiceClient, grpcClientPaymentService *grpcclient.PaymentServiceClient) OrderService {
	return &orderService{
		orderRepo:                orderRepo,
		grpcClientOrderService:   grpcClientOrderService,
		grpcClientPaymentService: grpcClientPaymentService,
	}
}

func (u *orderService) CreateOrder(UserID int64, Items []entity.OrderItem, TotalPrice float64) (*entity.PaymentResponse, error) {
	// 1️⃣ Собираем product_id из списка товаров
	var productIDs []int64
	for _, item := range Items {
		if slices.Contains(productIDs, item.ProductID) {
			return nil, fmt.Errorf("duplicate product id: %d", item.ProductID)
		}
		productIDs = append(productIDs, item.ProductID)
	}

	// 2️⃣ Запрашиваем stock через gRPC
	stockMap, err := u.grpcClientOrderService.GetProductStock(productIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to get stock: %v", err)
	}

	// var ErrNoReserve = errors.New("pq: relation \"products\" does not exist")
	// 3️⃣ Проверяем, хватает ли товара
	for index, item := range Items {
		availableStock, exists := stockMap[item.ProductID]
		if !exists {
			return nil, fmt.Errorf("product %d not found", item.ProductID)
		}

		if (availableStock.Stock /*- reservedStock*/) < item.Quantity {
			return nil, fmt.Errorf("not enough stock for product %d", item.ProductID)
		}
		Items[index].Name = availableStock.Name
	}

	order := &entity.Order{
		UserID:     UserID,
		Items:      Items,
		TotalPrice: TotalPrice,
		Status:     "pending",
		CreatedAt:  time.Now(),
	}

	order, errCreate := u.orderRepo.Create(order)
	if errCreate != nil {
		return nil, errCreate
	}
	log.Println("Order successfully created:", order.ID)


	// 3️⃣ Запрашиваем Payment Service для получения ссылки на оплату
	payment, err := u.grpcClientPaymentService.GeneratePaymentLink(order.UserID, order.ID, order.TotalPrice)
	if err != nil {
		return nil, fmt.Errorf("failed to get payment link: %v", err)
	}

	// 4️⃣ Возвращаем ID заказа и ссылку на оплату
	return &entity.PaymentResponse{
		PaymentURL: payment.PaymentUrl,
	}, nil

}

// 	if commitErr != nil {
// 		return commitErr
// 	}
// 	return nil
// }

func (u *orderService) GetOrderByID(orderID int64) (*entity.Order, error) {
	return u.orderRepo.GetOrderByID(orderID)
}

func (u *orderService) DeleteOrder(orderID int64) error {
	return u.orderRepo.Delete(orderID)
}

// UpdateOrderStatus обновляет статус заказа в базе данных
func (u *orderService) UpdateOrderStatus(orderID int64, status string) error {
	order, err := u.orderRepo.GetOrderByID(orderID)
	if err != nil {
		return fmt.Errorf("не удалось найти заказ: %w", err)
	}

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

	// Обновляем статус заказа
	order.Status = status
	if err := u.orderRepo.UpdateOrder(order); err != nil {
		return fmt.Errorf("ошибка при обновлении заказа: %w", err)
	}

	return nil
}

func (s *orderService) GetOrdersByUserID(userID int64) ([]entity.Order, error) {
	return s.orderRepo.GetOrdersByUserID(userID)
}

func (s *orderService) CancelOrder(userID int64, orderID int64) error {
	return s.orderRepo.CancelOrder(userID, orderID)
}
