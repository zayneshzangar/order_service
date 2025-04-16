package rest

import (
	"order_service/internal/middleware"

	"github.com/gorilla/mux"
)

func NewRouter(orderHandler *OrderHandler) *mux.Router {
	r := mux.NewRouter()
	// Группа маршрутов, защищённых JWT
	api := r.PathPrefix("/").Subrouter()
	api.Use(middleware.JWTMiddleware)
	api.HandleFunc("/orders", orderHandler.CreateOrderHandler).Methods("POST")
	api.HandleFunc("/orders/{id}", orderHandler.GetOrderByIDHandler).Methods("GET")
	api.HandleFunc("/my-orders", orderHandler.GetMyOrdersHandler).Methods("GET")
	api.HandleFunc("/orders/{id}/cancel", orderHandler.CancelOrderHandler).Methods("POST")
	return r
}
