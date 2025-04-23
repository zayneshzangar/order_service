package rest

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"order_service/internal/entity"
	"order_service/internal/middleware"
	ServiceMocks "order_service/internal/service/mocks"

	"github.com/gorilla/mux"
	"go.uber.org/mock/gomock"
)

func TestOrderHandler_CreateOrderHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := ServiceMocks.NewMockOrderServiceInterface(ctrl)
	handler := NewOrderHandler(mockService)

	userID := int64(1)

	t.Run("Success", func(t *testing.T) {
		// Подготовка
		reqBody := struct {
			Items      []entity.OrderItem `json:"items"`
			TotalPrice float64            `json:"total_price"`
		}{
			Items: []entity.OrderItem{
				{ProductID: 1, Quantity: 2, Price: 50.0},
			},
			TotalPrice: 100.0,
		}
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/orders", bytes.NewReader(body))
		rr := httptest.NewRecorder()

		ctx := context.WithValue(req.Context(), middleware.UserIDKey, userID)
		req = req.WithContext(ctx)

		expectedResponse := &entity.PaymentResponse{PaymentURL: "http://payment.com/link"}
		mockService.EXPECT().CreateOrder(userID, reqBody.Items, reqBody.TotalPrice).Return(expectedResponse, nil)

		// Выполнение
		handler.CreateOrderHandler(rr, req)

		// Проверка
		if status := rr.Code; status != http.StatusCreated {
			t.Errorf("expected status %v, got %v", http.StatusCreated, status)
		}
		var response entity.PaymentResponse
		if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
			t.Fatal(err)
		}
		if response.PaymentURL != expectedResponse.PaymentURL {
			t.Errorf("expected payment URL %v, got %v", expectedResponse.PaymentURL, response.PaymentURL)
		}
	})

	t.Run("Unauthorized", func(t *testing.T) {
		// Подготовка
		reqBody := struct {
			Items      []entity.OrderItem `json:"items"`
			TotalPrice float64            `json:"total_price"`
		}{
			Items:      []entity.OrderItem{{ProductID: 1, Quantity: 2, Price: 50.0}},
			TotalPrice: 100.0,
		}
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/orders", bytes.NewReader(body))
		rr := httptest.NewRecorder()

		// Выполнение
		handler.CreateOrderHandler(rr, req)

		// Проверка
		if status := rr.Code; status != http.StatusUnauthorized {
			t.Errorf("expected status %v, got %v", http.StatusUnauthorized, status)
		}
	})

	t.Run("InvalidRequestBody", func(t *testing.T) {
		// Подготовка
		req := httptest.NewRequest(http.MethodPost, "/orders", bytes.NewReader([]byte("invalid json")))
		rr := httptest.NewRecorder()

		ctx := context.WithValue(req.Context(), middleware.UserIDKey, userID)
		req = req.WithContext(ctx)

		// Выполнение
		handler.CreateOrderHandler(rr, req)

		// Проверка
		if status := rr.Code; status != http.StatusBadRequest {
			t.Errorf("expected status %v, got %v", http.StatusBadRequest, status)
		}
	})
}

func TestOrderHandler_GetOrderByIDHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := ServiceMocks.NewMockOrderServiceInterface(ctrl)
	handler := NewOrderHandler(mockService)

	t.Run("Success", func(t *testing.T) {
		// Подготовка
		expectedOrder := &entity.Order{
			ID:         1,
			UserID:     1,
			Items:      []entity.OrderItem{{ProductID: 1, Name: "Product 1", Quantity: 2, Price: 50.0}},
			TotalPrice: 100.0,
			Status:     "pending",
		}
		mockService.EXPECT().GetOrderByID(int64(1)).Return(expectedOrder, nil)

		req := httptest.NewRequest(http.MethodGet, "/orders/1", nil)
		rr := httptest.NewRecorder()
		vars := map[string]string{"id": "1"}
		req = mux.SetURLVars(req, vars)

		// Выполнение
		handler.GetOrderByIDHandler(rr, req)

		// Проверка
		if status := rr.Code; status != http.StatusOK {
			t.Errorf("expected status %v, got %v", http.StatusOK, status)
		}
		var response entity.Order
		if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
			t.Fatal(err)
		}
		if response.ID != expectedOrder.ID || response.TotalPrice != expectedOrder.TotalPrice {
			t.Errorf("expected order %v, got %v", expectedOrder, response)
		}
	})

	t.Run("InvalidID", func(t *testing.T) {
		// Подготовка
		req := httptest.NewRequest(http.MethodGet, "/orders/invalid", nil)
		rr := httptest.NewRecorder()
		vars := map[string]string{"id": "invalid"}
		req = mux.SetURLVars(req, vars)

		// Выполнение
		handler.GetOrderByIDHandler(rr, req)

		// Проверка
		if status := rr.Code; status != http.StatusBadRequest {
			t.Errorf("expected status %v, got %v", http.StatusBadRequest, status)
		}
	})

	t.Run("NotFound", func(t *testing.T) {
		// Подготовка
		mockService.EXPECT().GetOrderByID(int64(999)).Return(nil, errors.New("order not found"))

		req := httptest.NewRequest(http.MethodGet, "/orders/999", nil)
		rr := httptest.NewRecorder()
		vars := map[string]string{"id": "999"}
		req = mux.SetURLVars(req, vars)

		// Выполнение
		handler.GetOrderByIDHandler(rr, req)

		// Проверка
		if status := rr.Code; status != http.StatusNotFound {
			t.Errorf("expected status %v, got %v", http.StatusNotFound, status)
		}
	})
}

func TestOrderHandler_GetMyOrdersHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := ServiceMocks.NewMockOrderServiceInterface(ctrl)
	handler := NewOrderHandler(mockService)

	userID := int64(1)

	t.Run("Success", func(t *testing.T) {
		// Подготовка
		expectedOrders := []entity.Order{
			{ID: 1, UserID: userID, TotalPrice: 100.0, Status: "pending"},
			{ID: 2, UserID: userID, TotalPrice: 200.0, Status: "paid"},
		}
		mockService.EXPECT().GetOrdersByUserID(userID).Return(expectedOrders, nil)

		req := httptest.NewRequest(http.MethodGet, "/my-orders", nil)
		rr := httptest.NewRecorder()
		ctx := context.WithValue(req.Context(), middleware.UserIDKey, userID)
		req = req.WithContext(ctx)

		// Выполнение
		handler.GetMyOrdersHandler(rr, req)

		// Проверка
		if status := rr.Code; status != http.StatusOK {
			t.Errorf("expected status %v, got %v", http.StatusOK, status)
		}
		var response []entity.Order
		if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
			t.Fatal(err)
		}
		if len(response) != 2 {
			t.Errorf("expected 2 orders, got %d", len(response))
		}
	})

	t.Run("Unauthorized", func(t *testing.T) {
		// Подготовка
		req := httptest.NewRequest(http.MethodGet, "/my-orders", nil)
		rr := httptest.NewRecorder()

		// Выполнение
		handler.GetMyOrdersHandler(rr, req)

		// Проверка
		if status := rr.Code; status != http.StatusUnauthorized {
			t.Errorf("expected status %v, got %v", http.StatusUnauthorized, status)
		}
	})
}

func TestOrderHandler_CancelOrderHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := ServiceMocks.NewMockOrderServiceInterface(ctrl)
	handler := NewOrderHandler(mockService)

	userID := int64(1)

	t.Run("Success", func(t *testing.T) {
		// Подготовка
		mockService.EXPECT().CancelOrder(userID, int64(1)).Return(nil)

		req := httptest.NewRequest(http.MethodPost, "/orders/1/cancel", nil)
		rr := httptest.NewRecorder()
		vars := map[string]string{"id": "1"}
		req = mux.SetURLVars(req, vars)
		ctx := context.WithValue(req.Context(), middleware.UserIDKey, userID)
		req = req.WithContext(ctx)

		// Выполнение
		handler.CancelOrderHandler(rr, req)

		// Проверка
		if status := rr.Code; status != http.StatusOK {
			t.Errorf("expected status %v, got %v", http.StatusOK, status)
		}
	})

	t.Run("Unauthorized", func(t *testing.T) {
		// Подготовка
		req := httptest.NewRequest(http.MethodPost, "/orders/1/cancel", nil)
		rr := httptest.NewRecorder()
		vars := map[string]string{"id": "1"}
		req = mux.SetURLVars(req, vars)

		// Выполнение
		handler.CancelOrderHandler(rr, req)

		// Проверка
		if status := rr.Code; status != http.StatusUnauthorized {
			t.Errorf("expected status %v, got %v", http.StatusUnauthorized, status)
		}
	})

	t.Run("InvalidID", func(t *testing.T) {
		// Подготовка
		req := httptest.NewRequest(http.MethodPost, "/orders/invalid/cancel", nil)
		rr := httptest.NewRecorder()
		vars := map[string]string{"id": "invalid"}
		req = mux.SetURLVars(req, vars)
		ctx := context.WithValue(req.Context(), middleware.UserIDKey, userID)
		req = req.WithContext(ctx)

		// Выполнение
		handler.CancelOrderHandler(rr, req)

		// Проверка
		if status := rr.Code; status != http.StatusBadRequest {
			t.Errorf("expected status %v, got %v", http.StatusBadRequest, status)
		}
	})
}
