package usecase

import (
	"order_service/internal/entity"
	"order_service/internal/repository"
	"time"
)

type OrderUseCase interface {
	CreateOrder(UserID int64, Items []entity.OrderItem, TotalPrice float64) (*entity.Order, error)
	GetOrderByID(orderID int64) (*entity.Order, error)
}

type orderUseCase struct {
	orderRepo repository.OrderRepository
}

func NewOrderUseCase(orderRepo repository.OrderRepository) OrderUseCase {
	return &orderUseCase{orderRepo: orderRepo}
}

func (u *orderUseCase) CreateOrder(UserID int64, Items []entity.OrderItem, TotalPrice float64) (*entity.Order, error) {
	order := &entity.Order{
		UserID:     UserID,
		Items:      Items,
		TotalPrice: TotalPrice,
		Status:     "pending",
		CreatedAt:  time.Now(),
	}
	// TODO: Возможно, перед созданием заказа надо проверить, существуют ли товары?
	err := u.orderRepo.Create(order)
	if err != nil {
		return nil, err
	}
	return order, nil
}

func (u *orderUseCase) GetOrderByID(orderID int64) (*entity.Order, error) {
	return u.orderRepo.GetOrderByID(orderID)
}
