package grpcclient

import (
	"context"
	"fmt"
	"log"
	"time"

	"order_service/internal/entity"
	"order_service/internal/productpb"

	"google.golang.org/grpc"
)

type ProductServiceClient struct {
	client productpb.ProductServiceClient
}

func NewProductServiceClient(address string) (*ProductServiceClient, error) {
	conn, err := grpc.Dial(address, grpc.WithInsecure(), grpc.WithBlock(), grpc.WithTimeout(5*time.Second))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Product Service: %w", err)
	}

	client := productpb.NewProductServiceClient(conn)
	return &ProductServiceClient{client: client}, nil
}

func (p *ProductServiceClient) GetProductStock(productID []int64) (map[int64]*productpb.ProductStockInfo, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	req := &productpb.ProductStockRequest{
		ProductIds: productID,
	}

	res, err := p.client.GetProductStock(ctx, req)
	if err != nil {
		log.Printf("Error calling GetProductStock: %v", err)
		return nil, err
	}

	return res.StockMap, nil
}

func (p *ProductServiceClient) UpdateProductStock(order *entity.Order) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	arrReq := []*productpb.UpdateProductStockRequest_StockUpdate{}
	for _, item := range order.Items {
		arrReq = append(arrReq, &productpb.UpdateProductStockRequest_StockUpdate{
			ProductId: item.ProductID,
			Quantity:  item.Quantity,
		})
	}

	req := productpb.UpdateProductStockRequest{
		Updates: arrReq,
	}

	res, err := p.client.UpdateProductStock(ctx, &req)
	if err != nil {
		log.Printf("Error calling UpdateProductStock: %v", err)
		return err
	}

	if res.Error != "" {
		return fmt.Errorf("error updating product stock: %s", res.Error)
	}

	return nil
}
