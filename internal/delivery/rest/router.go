package rest

import (
	"github.com/gorilla/mux"
)

func NewRouter(orderHandler *OrderHandler) *mux.Router {
	r := mux.NewRouter()
	r.HandleFunc("/orders", orderHandler.CreateOrderHandler).Methods("POST")
	r.HandleFunc("/orders/{id}", orderHandler.GetOrderByIDHandler).Methods("GET")
	return r
}
