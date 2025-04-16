package grpcclient

import (
	"context"
	"fmt"
	"time"

	"order_service/internal/paymentpb"

	"google.golang.org/grpc"
)

type PaymentServiceClient struct {
	client paymentpb.PaymentServiceClient
}

func NewPaymentServiceClient(address string) (*PaymentServiceClient, error) {
	conn, err := grpc.Dial(address, grpc.WithInsecure(), grpc.WithBlock(), grpc.WithTimeout(5*time.Second))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Product Service: %w", err)
	}

	client := paymentpb.NewPaymentServiceClient(conn)
	return &PaymentServiceClient{client: client}, nil
}

func (p *PaymentServiceClient) GeneratePaymentLink(userID int64, orderID int64, totalPrice float64) (*paymentpb.PaymentResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	req := &paymentpb.PaymentRequest{
		UserId:     userID,
		OrderId:    orderID,
		TotalPrice: totalPrice,
	}

	res, err := p.client.GeneratePaymentLink(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to generate payment link: %w", err)
	}
	return res, nil
}
