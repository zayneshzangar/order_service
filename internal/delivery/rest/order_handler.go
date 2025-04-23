package rest

import (
	"encoding/json"
	"fmt"
	"net/http"
	"order_service/internal/entity"
	"order_service/internal/middleware"
	"order_service/internal/service"
	"strconv"

	"github.com/gorilla/mux"
)

// OrderHandler отвечает за обработку REST-запросов
type OrderHandler struct {
	orderService service.OrderServiceInterface
}

// NewOrderHandler создаёт новый обработчик
func NewOrderHandler(orderService service.OrderServiceInterface) *OrderHandler {
	return &OrderHandler{orderService: orderService}
}

// CreateOrderHandler — обработчик для создания продукта
func (h *OrderHandler) CreateOrderHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	fmt.Println("CreateOrderHandler")
	// Парсим тело запроса без user_id — он берётся из токена
	req := struct {
		Items      []entity.OrderItem `json:"items"`
		TotalPrice float64            `json:"total_price"`
	}{}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	payment, err := h.orderService.CreateOrder(userID, req.Items, req.TotalPrice)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(payment)
}

// GetOrderByIDHandler — обработчик для получения продукта по ID
func (h *OrderHandler) GetOrderByIDHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr, ok := vars["id"]
	if !ok {
		http.Error(w, "missing order ID", http.StatusBadRequest)
		return
	}

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "invalid order ID", http.StatusBadRequest)
		return
	}

	order, err := h.orderService.GetOrderByID(id)
	if err != nil {
		http.Error(w, "order not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(order)
}

func (h *OrderHandler) GetMyOrdersHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	orders, err := h.orderService.GetOrdersByUserID(userID)
	if err != nil {
		http.Error(w, "Ошибка при получении заказов", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(orders)
}

func (h *OrderHandler) CancelOrderHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	orderIDStr, ok := vars["id"]
	if !ok {
		http.Error(w, "missing order ID", http.StatusBadRequest)
		return
	}

	orderID, err := strconv.ParseInt(orderIDStr, 10, 64)
	if err != nil {
		http.Error(w, "invalid order ID", http.StatusBadRequest)
		return
	}

	err = h.orderService.CancelOrder(userID, orderID)
	if err != nil {
		http.Error(w, "Ошибка при отмене заказа", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Заказ успешно отменен"))
}
