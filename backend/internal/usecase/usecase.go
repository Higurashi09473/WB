package usecase

import (
	"context"
	"encoding/json"
	"time"

	kafka "WB/internal/lib/kafka"
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

// OrderUseCase — основной юзкейс для работы с заказами
type OrderUseCase struct {
	orderRepo     OrderRepository
	cacheRepo     CacheRepository
	kafkaProducer *kafka.Producer
}

// NewOrderUseCase — конструктор с инъекцией зависимостей
func NewOrderUseCase(orderRepo OrderRepository, cacheRepo CacheRepository, kafkaProducer *kafka.Producer) *OrderUseCase {
	return &OrderUseCase{
		orderRepo:     orderRepo,
		cacheRepo:     cacheRepo,
		kafkaProducer: kafkaProducer,
	}
}

// CreateOrder — принимает заказ от клиента и отправляет его в Kafka для асинхронной обработки
func (uc *OrderUseCase) CreateOrder(ctx context.Context, order models.Order) error {
	if err := validator.ValidateOrder(&order); err != nil {
		return err
	}
	
	orderJSON, err := json.Marshal(order)
	if err != nil {
		return err
	}

	// Отправляем в Kafka с ключом = OrderUID (для идемпотентности и правильного партиционирования)
	if err := uc.kafkaProducer.Send(ctx, order.OrderUID, orderJSON); err != nil {
		return err
	}

	return nil
}

// GetOrder — получает заказ по UID: сначала из Redis, потом из PostgreSQL, с кэшированием
func (uc *OrderUseCase) GetOrder(ctx context.Context, orderUID string) (models.Order, error) {
	cached, err := uc.cacheRepo.GetOrder(ctx, orderUID)
	if err == nil && len(cached) > 0 {
		var order models.Order
		if jsonErr := json.Unmarshal(cached, &order); jsonErr == nil {
			return order, nil
		}
		uc.cacheRepo.DeleteOrder(ctx, orderUID)
	}

	// 2. Если нет в кэше — берём из PostgreSQL
	order, err := uc.orderRepo.GetOrder(orderUID)
	if err != nil {
		// if errors.Is(err, repository.ErrOrderNotFound) {
		// 	return models.Order{}, err
		// }
		return models.Order{}, err
	}

	// 3. Кэшируем в Redis на 24 часа (или другой TTL)
	orderJSON, marshalErr := json.Marshal(order)
	if marshalErr == nil {
		_ = uc.cacheRepo.SetOrder(ctx, orderUID, orderJSON, 24*time.Hour)
	}

	return order, nil
}

func (uc *OrderUseCase) HandleMessage(ctx context.Context, key string, value []byte) error {
	var order models.Order
	if err := json.Unmarshal(value, &order); err != nil {
		// log.Printf("Не удалось распарсить JSON заказа %s: %v", key, err)
		return err // неретрябельная ошибка
	}

	// if order.OrderUID != key {
	//     log.Printf("Предупреждение: key (%s) не совпадает с order_uid (%s)", key, order.OrderUID)
	// }

	if _, err := uc.orderRepo.GetOrder(order.OrderUID); err == nil {
		// log.Printf("Заказ %s уже существует — пропускаем", order.OrderUID)
		return nil
	}
	// } else if !errors.Is(err, repository.ErrOrderNotFound) {
	//     return err // ошибка БД — нужно повторить
	// }

	if err := uc.orderRepo.NewOrder(order); err != nil {
		return err 
	}

	// Кэшируем в Redis
	orderJSON, _ := json.Marshal(order)
	if err := uc.cacheRepo.SetOrder(ctx, order.OrderUID, orderJSON, 24*time.Hour); err != nil {

	}

	// log.Printf("Заказ %s успешно сохранён в БД и Redis", order.OrderUID)
	return nil
}
