package usecase

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"WB/internal/lib/validator"
	"WB/internal/models"
)

type OrderRepository interface {
	NewOrder(order models.Order) error
	GetOrder(orderID string) (models.Order, error)
}

type CacheRepository interface {
	GetOrder(ctx context.Context, orderUID string) ([]byte, error)
	SetOrder(ctx context.Context, orderUID string, data []byte, ttl time.Duration) error
	DeleteOrder(ctx context.Context, orderUID string) error
}

type MessageBroker interface {
	Send(ctx context.Context, key string, value []byte) error
	Close() error
}

type OrderUseCase struct {
	orderRepo     OrderRepository
	cacheRepo     CacheRepository
	messageBroker MessageBroker
}

func NewOrderUseCase(orderRepo OrderRepository, cacheRepo CacheRepository, messageBroker MessageBroker) *OrderUseCase {
	return &OrderUseCase{
		orderRepo:     orderRepo,
		cacheRepo:     cacheRepo,
		messageBroker: messageBroker,
	}
}

func (uc *OrderUseCase) CreateOrder(ctx context.Context, order models.Order) error {
	const op = "usecase.CreateOrder"

	if err := validator.ValidateOrder(&order); err != nil {
		return fmt.Errorf("%s: validator: %w", op, err)
	}

	orderJSON, err := json.Marshal(order)
	if err != nil {
		return fmt.Errorf("%s: json marshal err: %w", op, err)
	}

	if err := uc.messageBroker.Send(ctx, order.OrderUID, orderJSON); err != nil {
		return fmt.Errorf("%s: kafka producer send err: %w", op, err)
	}

	return nil
}

func (uc *OrderUseCase) GetOrder(ctx context.Context, orderUID string) (models.Order, error) {
	const op = "usecase.GetOrder"

	cached, err := uc.cacheRepo.GetOrder(ctx, orderUID)
	if err == nil && len(cached) > 0 {
		var order models.Order
		if jsonErr := json.Unmarshal(cached, &order); jsonErr == nil {
			return order, nil
		}
		uc.cacheRepo.DeleteOrder(ctx, orderUID)
	}

	order, err := uc.orderRepo.GetOrder(orderUID)
	if err != nil {
		return models.Order{}, fmt.Errorf("%s: orderRepo get order: %w", op, err)
	}

	orderJSON, marshalErr := json.Marshal(order)
	if marshalErr == nil {
		uc.cacheRepo.SetOrder(ctx, orderUID, orderJSON, 24*time.Hour)
	}

	return order, nil
}

func (uc *OrderUseCase) HandleMessage(ctx context.Context, key string, value []byte) error {
	const op = "usecase.HandleMessage"

	var order models.Order
	if err := json.Unmarshal(value, &order); err != nil {
		return fmt.Errorf("%s: JSON unmarshal err: %w", op, err)
	}

	if _, err := uc.orderRepo.GetOrder(order.OrderUID); err == nil {
		return nil
	}

	if err := uc.orderRepo.NewOrder(order); err != nil {
		return fmt.Errorf("%s: orderRepo new order: %w", op, err)
	}

	orderJSON, marshalErr := json.Marshal(order)
	if marshalErr == nil {
		uc.cacheRepo.SetOrder(ctx, order.OrderUID, orderJSON, 24*time.Hour)
	}

	return nil
}
