package rest

import (
	"encoding/json"
	"net/http"
	"order_service/internal/entity"
	"order_service/internal/usecase"
	"strconv"

	"github.com/gorilla/mux"
)

// OrderHandler отвечает за обработку REST-запросов
type OrderHandler struct {
	orderUseCase usecase.OrderUseCase
}

// NewOrderHandler создаёт новый обработчик
func NewOrderHandler(orderUseCase usecase.OrderUseCase) *OrderHandler {
	return &OrderHandler{orderUseCase: orderUseCase}
}

// CreateOrderHandler — обработчик для создания продукта
func (h *OrderHandler) CreateOrderHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		UserID     int64              `json:"user_id"`
		Items      []entity.OrderItem `json:"items"`
		TotalPrice float64            `json:"total_price"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	order, err := h.orderUseCase.CreateOrder(req.UserID, req.Items, req.TotalPrice)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(order)
}

// GetOrderByIDHandler — обработчик для получения продукта по ID
func (h *OrderHandler) GetOrderByIDHandler(w http.ResponseWriter, r *http.Request) {
	// Получаем ID из URL
	vars := mux.Vars(r)
	idStr, ok := vars["id"]
	if !ok {
		http.Error(w, "missing order ID", http.StatusBadRequest)
		return
	}

	// Конвертируем строку в int64
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "invalid order ID", http.StatusBadRequest)
		return
	}

	// Получаем продукт из usecase
	order, err := h.orderUseCase.GetOrderByID(id)
	if err != nil {
		http.Error(w, "order not found", http.StatusNotFound)
		return
	}

	// Отправляем ответ в JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(order)
}
