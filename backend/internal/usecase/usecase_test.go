// internal/usecase/usecase_test.go

package usecase

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"WB/internal/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockOrderRepo struct {
	mock.Mock
}

func (m *mockOrderRepo) NewOrder(order models.Order) error {
	args := m.Called(order)
	return args.Error(0)
}

func (m *mockOrderRepo) GetOrder(orderID string) (models.Order, error) {
	args := m.Called(orderID)
	return args.Get(0).(models.Order), args.Error(1)
}

type mockCacheRepo struct {
	mock.Mock
}

func (m *mockCacheRepo) GetOrder(ctx context.Context, orderUID string) ([]byte, error) {
	args := m.Called(ctx, orderUID)
	return args.Get(0).([]byte), args.Error(1)
}

func (m *mockCacheRepo) SetOrder(ctx context.Context, orderUID string, data []byte, ttl time.Duration) error {
	args := m.Called(ctx, orderUID, data, ttl)
	return args.Error(0)
}

func (m *mockCacheRepo) DeleteOrder(ctx context.Context, orderUID string) error {
	args := m.Called(ctx, orderUID)
	return args.Error(0)
}

type mockMessageBroker struct {
	mock.Mock
}

func (m *mockMessageBroker) Send(ctx context.Context, key string, value []byte) error {
	args := m.Called(ctx, key, value)
	return args.Error(0)
}

func (m *mockMessageBroker) Close() error {
	args := m.Called()
	return args.Error(0)
}

// ---------------------------------------------------------
// Тесты CreateOrder
// ---------------------------------------------------------

func TestCreateOrder_Success(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(mockOrderRepo)
	mockCache := new(mockCacheRepo)
	mockProd := new(mockMessageBroker)

	order := models.Order{
		OrderUID:          "b563feb7b2b84b6test", // уникальный UID
		TrackNumber:       "WBILMTESTTRACK",
		Entry:             "WBIL",
		Locale:            "en",
		InternalSignature: "", // опционально
		CustomerID:        "test",
		DeliveryService:   "meest",
		Shardkey:          "9",
		SmID:              99,
		DateCreated:       time.Now().UTC(),
		OofShard:          "1",

		Delivery: models.Delivery{
			Name:    "Test Testov",
			Phone:   "+9720000000",
			Zip:     "2639809",
			City:    "Kiryat Mozkin",
			Address: "Ploshad Mira 15",
			Region:  "Kraiot",
			Email:   "test@gmail.com",
		},

		Payment: models.Payment{
			Transaction:  "b563feb7b2b84b6test",
			RequestID:    "",
			Currency:     "USD",
			Provider:     "wbpay",
			Amount:       1817,
			PaymentDt:    1637907727,
			Bank:         "alpha",
			DeliveryCost: 1500,
			GoodsTotal:   317,
			CustomFee:    0,
		},

		Items: []models.Item{
			{
				ChrtID:      9934930,
				TrackNumber: "WBILMTESTTRACK",
				Price:       453,
				Rid:         "ab4219087a764ae0btest",
				Name:        "Mascaras",
				Sale:        30,
				Size:        "0",
				TotalPrice:  317,
				NmID:        2389212,
				Brand:       "Vivienne Sabo",
				Status:      202,
			},
		},
	}

	// Мокируем валидацию (предполагаем, что она проходит)
	// Если валидатор строгий — можно замокать validator.ValidateOrder

	orderJSON, _ := json.Marshal(order)

	mockProd.
		On("Send", ctx, order.OrderUID, orderJSON).
		Return(nil).
		Once()

	uc := NewOrderUseCase(mockRepo, mockCache, mockProd)

	err := uc.CreateOrder(ctx, order)

	assert.NoError(t, err)
	mockProd.AssertExpectations(t)
}

func TestCreateOrder_KafkaError(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(mockOrderRepo)
	mockCache := new(mockCacheRepo)
	mockProd := new(mockMessageBroker)

	order := models.Order{
		OrderUID:          "test", // уникальный UID
		TrackNumber:       "WBILMTESTTRACK",
		Entry:             "WBIL",
		Locale:            "en",
		InternalSignature: "", // опционально
		CustomerID:        "test",
		DeliveryService:   "meest",
		Shardkey:          "9",
		SmID:              99,
		DateCreated:       time.Now().UTC(),
		OofShard:          "1",

		Delivery: models.Delivery{
			Name:    "Test Testov",
			Phone:   "+9720000000",
			Zip:     "2639809",
			City:    "Kiryat Mozkin",
			Address: "Ploshad Mira 15",
			Region:  "Kraiot",
			Email:   "test@gmail.com",
		},

		Payment: models.Payment{
			Transaction:  "b563feb7b2b84b6test",
			RequestID:    "",
			Currency:     "USD",
			Provider:     "wbpay",
			Amount:       1817,
			PaymentDt:    1637907727,
			Bank:         "alpha",
			DeliveryCost: 1500,
			GoodsTotal:   317,
			CustomFee:    0,
		},

		Items: []models.Item{
			{
				ChrtID:      9934930,
				TrackNumber: "WBILMTESTTRACK",
				Price:       453,
				Rid:         "ab4219087a764ae0btest",
				Name:        "Mascaras",
				Sale:        30,
				Size:        "0",
				TotalPrice:  317,
				NmID:        2389212,
				Brand:       "Vivienne Sabo",
				Status:      202,
			},
		},
	}
	orderJSON, _ := json.Marshal(order)

	mockProd.
		On("Send", ctx, order.OrderUID, orderJSON).
		Return(errors.New("kafka timeout")).
		Once()

	uc := NewOrderUseCase(mockRepo, mockCache, mockProd)

	err := uc.CreateOrder(ctx, order)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "kafka producer send err")
	assert.Contains(t, err.Error(), "kafka timeout")
}

// ---------------------------------------------------------
// Тесты GetOrder
// ---------------------------------------------------------

func TestGetOrder_FromCache_Success(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(mockOrderRepo)
	mockCache := new(mockCacheRepo)
	mockProd := new(mockMessageBroker)

	order := models.Order{OrderUID: "cache-hit-777"}
	cachedJSON, _ := json.Marshal(order)

	mockCache.
		On("GetOrder", ctx, "cache-hit-777").
		Return(cachedJSON, nil).
		Once()

	mockRepo.On("GetOrder", mock.Anything).Return(models.Order{}, nil).Maybe()

	uc := NewOrderUseCase(mockRepo, mockCache, mockProd)

	result, err := uc.GetOrder(ctx, "cache-hit-777")

	assert.NoError(t, err)
	assert.Equal(t, order.OrderUID, result.OrderUID)
	mockCache.AssertExpectations(t)
	mockRepo.AssertNotCalled(t, "GetOrder")
}


// ---------------------------------------------------------
// Тесты HandleMessage (consumer)
// ---------------------------------------------------------

func TestHandleMessage_NewOrder_Success(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(mockOrderRepo)
	mockCache := new(mockCacheRepo)
	mockProd := new(mockMessageBroker)

	order := models.Order{OrderUID: "new-order-abc"}
	data, _ := json.Marshal(order)

	mockRepo.
		On("GetOrder", "new-order-abc").
		Return(models.Order{}, errors.New("not found")).
		Once()

	mockRepo.
		On("NewOrder", order).
		Return(nil).
		Once()

	orderJSON, _ := json.Marshal(order)
	mockCache.
		On("SetOrder", ctx, "new-order-abc", orderJSON, 24*time.Hour).
		Return(nil).
		Once()

	uc := NewOrderUseCase(mockRepo, mockCache, mockProd)

	err := uc.HandleMessage(ctx, data)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
	mockCache.AssertExpectations(t)
}

func TestHandleMessage_DuplicateOrder(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(mockOrderRepo)
	mockCache := new(mockCacheRepo)
	mockProd := new(mockMessageBroker)

	order := models.Order{OrderUID: "already-exist"}
	data, _ := json.Marshal(order)

	mockRepo.
		On("GetOrder", "already-exist").
		Return(order, nil). // уже существует
		Once()

	// НЕ должно вызываться создание
	mockRepo.On("NewOrder", mock.Anything).Maybe()

	uc := NewOrderUseCase(mockRepo, mockCache, mockProd)

	err := uc.HandleMessage(ctx, data)

	assert.NoError(t, err) // дубликат — нормальная ситуация
	mockRepo.AssertNotCalled(t, "NewOrder")
}
