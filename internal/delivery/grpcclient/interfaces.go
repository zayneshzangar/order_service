package grpcclient

import (
	"order_service/internal/paymentpb"
	"order_service/internal/productpb"
)

type ProductServiceClientInterface interface {
	GetProductStock(productIDs []int64) (map[int64]*productpb.ProductStockInfo, error)
}

type PaymentServiceClientInterface interface {
	GeneratePaymentLink(userID, orderID int64, amount float64) (*paymentpb.PaymentResponse, error)
}
