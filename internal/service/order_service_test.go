package service

import (
	"errors"
	"testing"
	"time"

	GrpcMocks "order_service/internal/delivery/grpcclient/mocks"
	"order_service/internal/entity"
	"order_service/internal/paymentpb"
	"order_service/internal/productpb"
	RepoMocks "order_service/internal/repository/mocks"

	"go.uber.org/mock/gomock"
)

func TestOrderService_CreateOrder(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := RepoMocks.NewMockOrderRepository(ctrl)
	mockProductClient := GrpcMocks.NewMockProductServiceClientInterface(ctrl)
	mockPaymentClient := GrpcMocks.NewMockPaymentServiceClientInterface(ctrl)
	service := NewOrderService(mockRepo, mockProductClient, mockPaymentClient)

	items := []entity.OrderItem{
		{ProductID: 1, Quantity: 2, Price: 50.0},
		{ProductID: 2, Quantity: 1, Price: 100.0},
	}
	userID := int64(1)
	totalPrice := 200.0

	t.Run("Success", func(t *testing.T) {
		// Подготовка
		stockMap := map[int64]*productpb.ProductStockInfo{
			1: {Name: "Product 1", Stock: 10},
			2: {Name: "Product 2", Stock: 5},
		}
		mockProductClient.EXPECT().GetProductStock([]int64{1, 2}).Return(stockMap, nil)

		expectedOrder := &entity.Order{
			UserID:     userID,
			Items:      []entity.OrderItem{{ProductID: 1, Name: "Product 1", Quantity: 2, Price: 50.0}, {ProductID: 2, Name: "Product 2", Quantity: 1, Price: 100.0}},
			TotalPrice: totalPrice,
			Status:     "pending",
			CreatedAt:  time.Now().Truncate(time.Second),
		}
		mockRepo.EXPECT().Create(gomock.Any()).DoAndReturn(func(o *entity.Order) (*entity.Order, error) {
			o.ID = 1
			if o.UserID != expectedOrder.UserID || o.TotalPrice != expectedOrder.TotalPrice || o.Status != expectedOrder.Status {
				t.Errorf("expected order %+v, got %+v", expectedOrder, o)
			}
			return o, nil
		})

		paymentResponse := &paymentpb.PaymentResponse{PaymentUrl: "http://payment.com/link"}
		mockPaymentClient.EXPECT().GeneratePaymentLink(userID, int64(1), totalPrice).Return(paymentResponse, nil)

		// Выполнение
		result, err := service.CreateOrder(userID, items, totalPrice)

		// Проверка
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
		if result.PaymentURL != paymentResponse.PaymentUrl {
			t.Errorf("expected payment URL %v, got %v", paymentResponse.PaymentUrl, result.PaymentURL)
		}
	})

	t.Run("DuplicateProductID", func(t *testing.T) {
		// Подготовка
		invalidItems := []entity.OrderItem{
			{ProductID: 1, Quantity: 2, Price: 50.0},
			{ProductID: 1, Quantity: 1, Price: 50.0},
		}

		// Выполнение
		result, err := service.CreateOrder(userID, invalidItems, totalPrice)

		// Проверка
		if err == nil || err.Error() != "duplicate product id: 1" {
			t.Errorf("expected error 'duplicate product id: 1', got %v", err)
		}
		if result != nil {
			t.Errorf("expected nil result, got %v", result)
		}
	})

	t.Run("ProductNotFound", func(t *testing.T) {
		// Подготовка
		stockMap := map[int64]*productpb.ProductStockInfo{
			1: {Name: "Product 1", Stock: 10},
			// ProductID 2 отсутствует
		}
		mockProductClient.EXPECT().GetProductStock([]int64{1, 2}).Return(stockMap, nil)

		// Выполнение
		result, err := service.CreateOrder(userID, items, totalPrice)

		// Проверка
		if err == nil || err.Error() != "product 2 not found" {
			t.Errorf("expected error 'product 2 not found', got %v", err)
		}
		if result != nil {
			t.Errorf("expected nil result, got %v", result)
		}
	})

	t.Run("NotEnoughStock", func(t *testing.T) {
		// Подготовка
		stockMap := map[int64]*productpb.ProductStockInfo{
			1: {Name: "Product 1", Stock: 1}, // Недостаточно для 2
			2: {Name: "Product 2", Stock: 5},
		}
		mockProductClient.EXPECT().GetProductStock([]int64{1, 2}).Return(stockMap, nil)

		// Выполнение
		result, err := service.CreateOrder(userID, items, totalPrice)

		// Проверка
		if err == nil || err.Error() != "not enough stock for product 1" {
			t.Errorf("expected error 'not enough stock for product 1', got %v", err)
		}
		if result != nil {
			t.Errorf("expected nil result, got %v", result)
		}
	})

	t.Run("RepositoryError", func(t *testing.T) {
		// Подготовка
		stockMap := map[int64]*productpb.ProductStockInfo{
			1: {Name: "Product 1", Stock: 10},
			2: {Name: "Product 2", Stock: 5},
		}
		mockProductClient.EXPECT().GetProductStock([]int64{1, 2}).Return(stockMap, nil)
		mockRepo.EXPECT().Create(gomock.Any()).Return(nil, errors.New("database error"))

		// Выполнение
		result, err := service.CreateOrder(userID, items, totalPrice)

		// Проверка
		if err == nil || err.Error() != "database error" {
			t.Errorf("expected error 'database error', got %v", err)
		}
		if result != nil {
			t.Errorf("expected nil result, got %v", result)
		}
	})

	t.Run("PaymentServiceError", func(t *testing.T) {
		// Подготовка
		stockMap := map[int64]*productpb.ProductStockInfo{
			1: {Name: "Product 1", Stock: 10},
			2: {Name: "Product 2", Stock: 5},
		}
		mockProductClient.EXPECT().GetProductStock([]int64{1, 2}).Return(stockMap, nil)
		mockRepo.EXPECT().Create(gomock.Any()).DoAndReturn(func(o *entity.Order) (*entity.Order, error) {
			o.ID = 1
			return o, nil
		})
		mockPaymentClient.EXPECT().GeneratePaymentLink(userID, int64(1), totalPrice).Return(nil, errors.New("payment service error"))

		// Выполнение
		result, err := service.CreateOrder(userID, items, totalPrice)

		// Проверка
		if err == nil || err.Error() != "failed to get payment link: payment service error" {
			t.Errorf("expected error 'failed to get payment link: payment service error', got %v", err)
		}
		if result != nil {
			t.Errorf("expected nil result, got %v", result)
		}
	})
}

func TestOrderService_GetOrderByID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := RepoMocks.NewMockOrderRepository(ctrl)
	mockProductClient := GrpcMocks.NewMockProductServiceClientInterface(ctrl)
	mockPaymentClient := GrpcMocks.NewMockPaymentServiceClientInterface(ctrl)
	service := NewOrderService(mockRepo, mockProductClient, mockPaymentClient)

	t.Run("Success", func(t *testing.T) {
		// Подготовка
		expectedOrder := &entity.Order{
			ID:         1,
			UserID:     1,
			Items:      []entity.OrderItem{{ProductID: 1, Name: "Product 1", Quantity: 2, Price: 50.0}},
			TotalPrice: 100.0,
			Status:     "pending",
			CreatedAt:  time.Now().Truncate(time.Second),
		}
		mockRepo.EXPECT().GetOrderByID(int64(1)).Return(expectedOrder, nil)

		// Выполнение
		order, err := service.GetOrderByID(1)

		// Проверка
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
		if order.ID != expectedOrder.ID || order.UserID != expectedOrder.UserID {
			t.Errorf("expected order %v, got %v", expectedOrder, order)
		}
	})

	t.Run("NotFound", func(t *testing.T) {
		// Подготовка
		mockRepo.EXPECT().GetOrderByID(int64(999)).Return(nil, errors.New("order not found"))

		// Выполнение
		order, err := service.GetOrderByID(999)

		// Проверка
		if err == nil || err.Error() != "order not found" {
			t.Errorf("expected error 'order not found', got %v", err)
		}
		if order != nil {
			t.Errorf("expected nil order, got %v", order)
		}
	})
}

func TestOrderService_UpdateOrderStatus(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := RepoMocks.NewMockOrderRepository(ctrl)
	mockProductClient := GrpcMocks.NewMockProductServiceClientInterface(ctrl)
	mockPaymentClient := GrpcMocks.NewMockPaymentServiceClientInterface(ctrl)
	service := NewOrderService(mockRepo, mockProductClient, mockPaymentClient)

	t.Run("Success", func(t *testing.T) {
		// Подготовка
		order := &entity.Order{
			ID:         1,
			UserID:     1,
			Items:      []entity.OrderItem{{ProductID: 1, Name: "Product 1", Quantity: 2, Price: 50.0}},
			TotalPrice: 100.0,
			Status:     "pending",
		}
		mockRepo.EXPECT().GetOrderByID(int64(1)).Return(order, nil)
		mockRepo.EXPECT().UpdateOrder(gomock.Any()).Return(nil)

		// Выполнение
		err := service.UpdateOrderStatus(1, "paid")

		// Проверка
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
	})

	t.Run("InvalidStatus", func(t *testing.T) {
		// Выполнение
		err := service.UpdateOrderStatus(1, "invalid")

		// Проверка
		if err == nil || err.Error() != "недопустимый статус: invalid" {
			t.Errorf("expected error 'недопустимый статус: invalid', got %v", err)
		}
	})

	t.Run("OrderNotFound", func(t *testing.T) {
		// Подготовка
		mockRepo.EXPECT().GetOrderByID(int64(999)).Return(nil, errors.New("order not found"))

		// Выполнение
		err := service.UpdateOrderStatus(999, "paid")

		// Проверка
		if err == nil || err.Error() != "не удалось найти заказ: order not found" {
			t.Errorf("expected error 'не удалось найти заказ: order not found', got %v", err)
		}
	})

	t.Run("RepositoryError", func(t *testing.T) {
		// Подготовка
		order := &entity.Order{
			ID:         1,
			UserID:     1,
			Items:      []entity.OrderItem{{ProductID: 1, Name: "Product 1", Quantity: 2, Price: 50.0}},
			TotalPrice: 100.0,
			Status:     "pending",
		}
		mockRepo.EXPECT().GetOrderByID(int64(1)).Return(order, nil)
		mockRepo.EXPECT().UpdateOrder(gomock.Any()).Return(errors.New("database error"))

		// Выполнение
		err := service.UpdateOrderStatus(1, "paid")

		// Проверка
		if err == nil || err.Error() != "ошибка при обновлении заказа: database error" {
			t.Errorf("expected error 'ошибка при обновлении заказа: database error', got %v", err)
		}
	})
}

func TestOrderService_GetOrdersByUserID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := RepoMocks.NewMockOrderRepository(ctrl)
	mockProductClient := GrpcMocks.NewMockProductServiceClientInterface(ctrl)
	mockPaymentClient := GrpcMocks.NewMockPaymentServiceClientInterface(ctrl)
	service := NewOrderService(mockRepo, mockProductClient, mockPaymentClient)

	t.Run("Success", func(t *testing.T) {
		// Подготовка
		expectedOrders := []entity.Order{
			{ID: 1, UserID: 1, TotalPrice: 100.0, Status: "pending"},
			{ID: 2, UserID: 1, TotalPrice: 200.0, Status: "paid"},
		}
		mockRepo.EXPECT().GetOrdersByUserID(int64(1)).Return(expectedOrders, nil)

		// Выполнение
		orders, err := service.GetOrdersByUserID(1)

		// Проверка
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
		if len(orders) != 2 {
			t.Errorf("expected 2 orders, got %d", len(orders))
		}
	})

	t.Run("RepositoryError", func(t *testing.T) {
		// Подготовка
		mockRepo.EXPECT().GetOrdersByUserID(int64(1)).Return(nil, errors.New("database error"))

		// Выполнение
		orders, err := service.GetOrdersByUserID(1)

		// Проверка
		if err == nil || err.Error() != "database error" {
			t.Errorf("expected error 'database error', got %v", err)
		}
		if orders != nil {
			t.Errorf("expected nil orders, got %v", orders)
		}
	})
}

func TestOrderService_CancelOrder(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := RepoMocks.NewMockOrderRepository(ctrl)
	mockProductClient := GrpcMocks.NewMockProductServiceClientInterface(ctrl)
	mockPaymentClient := GrpcMocks.NewMockPaymentServiceClientInterface(ctrl)
	service := NewOrderService(mockRepo, mockProductClient, mockPaymentClient)

	t.Run("Success", func(t *testing.T) {
		// Подготовка
		mockRepo.EXPECT().CancelOrder(int64(1), int64(1)).Return(nil)

		// Выполнение
		err := service.CancelOrder(1, 1)

		// Проверка
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
	})

	t.Run("RepositoryError", func(t *testing.T) {
		// Подготовка
		mockRepo.EXPECT().CancelOrder(int64(1), int64(1)).Return(errors.New("database error"))

		// Выполнение
		err := service.CancelOrder(1, 1)

		// Проверка
		if err == nil || err.Error() != "database error" {
			t.Errorf("expected error 'database error', got %v", err)
		}
	})
}

func TestOrderService_DeleteOrder(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := RepoMocks.NewMockOrderRepository(ctrl)
	mockProductClient := GrpcMocks.NewMockProductServiceClientInterface(ctrl)
	mockPaymentClient := GrpcMocks.NewMockPaymentServiceClientInterface(ctrl)
	service := NewOrderService(mockRepo, mockProductClient, mockPaymentClient)

	t.Run("Success", func(t *testing.T) {
		// Подготовка
		mockRepo.EXPECT().Delete(int64(1)).Return(nil)

		// Выполнение
		err := service.DeleteOrder(1)

		// Проверка
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
	})

	t.Run("RepositoryError", func(t *testing.T) {
		// Подготовка
		mockRepo.EXPECT().Delete(int64(1)).Return(errors.New("database error"))

		// Выполнение
		err := service.DeleteOrder(1)

		// Проверка
		if err == nil || err.Error() != "database error" {
			t.Errorf("expected error 'database error', got %v", err)
		}
	})
}
