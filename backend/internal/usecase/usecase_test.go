package usecase

import (
	"WB/internal/models"
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

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

func TestCreateOrder_Success(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(mockOrderRepo)
	mockCache := new(mockCacheRepo)
	mockProd := new(mockMessageBroker)

	order := models.Order{
		OrderUID:          "b563feb7b2b84b6test",
		TrackNumber:       "WBILMTESTTRACK",
		Entry:             "WBIL",
		Locale:            "en",
		InternalSignature: "",
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
		On("Send", ctx, order.OrderUID, mock.MatchedBy(func(b []byte) bool { return assert.JSONEq(t, string(orderJSON), string(b)) })).
		Return(nil).
		Once()

	uc := NewOrderUseCase(mockRepo, mockCache, mockProd)

	err := uc.CreateOrder(ctx, order)

	assert.NoError(t, err)
	mockProd.AssertExpectations(t)
}

func TestCreateOrder_ValidationError(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(mockOrderRepo)
	mockCache := new(mockCacheRepo)
	mockProd := new(mockMessageBroker)

	order := models.Order{} // Invalid order

	uc := NewOrderUseCase(mockRepo, mockCache, mockProd)

	err := uc.CreateOrder(ctx, order)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "validator")
	mockProd.AssertNotCalled(t, "Send")
}

// func TestCreateOrder_MarshalError(t *testing.T) {
// 	ctx := context.Background()
// 	mockRepo := new(mockOrderRepo)
// 	mockCache := new(mockCacheRepo)
// 	mockProd := new(mockMessageBroker)

// 	order := models.Order{
// 		OrderUID: "valid-uid",
// 		// Force marshal error by including unmarshalable field, but since models are structs, hard to force; instead, we can assume it's covered or skip if rare.
// 	}

// 	t.Skip("Marshal error is rare for structs; consider if needed")
// }

func TestCreateOrder_KafkaError(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(mockOrderRepo)
	mockCache := new(mockCacheRepo)
	mockProd := new(mockMessageBroker)

	order := models.Order{
		OrderUID:          "b563feb7b2b84b6test",
		TrackNumber:       "WBILMTESTTRACK",
		Entry:             "WBIL",
		Locale:            "en",
		InternalSignature: "",
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
		On("Send", ctx, order.OrderUID, mock.MatchedBy(func(b []byte) bool { return assert.JSONEq(t, string(orderJSON), string(b)) })).
		Return(errors.New("kafka timeout")).
		Once()

	uc := NewOrderUseCase(mockRepo, mockCache, mockProd)

	err := uc.CreateOrder(ctx, order)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "kafka producer send err")
	assert.Contains(t, err.Error(), "kafka timeout")
	mockProd.AssertExpectations(t)
}

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

	uc := NewOrderUseCase(mockRepo, mockCache, mockProd)

	result, err := uc.GetOrder(ctx, "cache-hit-777")

	assert.NoError(t, err)
	assert.Equal(t, order.OrderUID, result.OrderUID)
	mockCache.AssertExpectations(t)
	mockRepo.AssertNotCalled(t, "GetOrder")
}

func TestGetOrder_FromCache_InvalidJSON(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(mockOrderRepo)
	mockCache := new(mockCacheRepo)
	mockProd := new(mockMessageBroker)

	order := models.Order{OrderUID: "cache-invalid-888"}

	mockCache.
		On("GetOrder", ctx, "cache-invalid-888").
		Return([]byte("invalid json"), nil).
		Once()

	mockCache.
		On("DeleteOrder", ctx, "cache-invalid-888").
		Return(nil).
		Once()

	mockRepo.
		On("GetOrder", "cache-invalid-888").
		Return(order, nil).
		Once()

	orderJSON, _ := json.Marshal(order)
	mockCache.
		On("SetOrder", ctx, "cache-invalid-888", mock.MatchedBy(func(b []byte) bool { return assert.JSONEq(t, string(orderJSON), string(b)) }), 24*time.Hour).
		Return(nil).
		Once()

	uc := NewOrderUseCase(mockRepo, mockCache, mockProd)

	result, err := uc.GetOrder(ctx, "cache-invalid-888")

	assert.NoError(t, err)
	assert.Equal(t, order.OrderUID, result.OrderUID)
	mockCache.AssertExpectations(t)
	mockRepo.AssertExpectations(t)
}

func TestGetOrder_CacheMiss_DB_Success(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(mockOrderRepo)
	mockCache := new(mockCacheRepo)
	mockProd := new(mockMessageBroker)

	order := models.Order{OrderUID: "cache-miss-999"}

	mockCache.
		On("GetOrder", ctx, "cache-miss-999").
		Return([]byte{}, errors.New("not found")).
		Once()

	mockRepo.
		On("GetOrder", "cache-miss-999").
		Return(order, nil).
		Once()

	orderJSON, _ := json.Marshal(order)
	mockCache.
		On("SetOrder", ctx, "cache-miss-999", mock.MatchedBy(func(b []byte) bool { return assert.JSONEq(t, string(orderJSON), string(b)) }), 24*time.Hour).
		Return(nil).
		Once()

	uc := NewOrderUseCase(mockRepo, mockCache, mockProd)

	result, err := uc.GetOrder(ctx, "cache-miss-999")

	assert.NoError(t, err)
	assert.Equal(t, order.OrderUID, result.OrderUID)
	mockCache.AssertExpectations(t)
	mockRepo.AssertExpectations(t)
}

func TestGetOrder_CacheMiss_DB_Error(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(mockOrderRepo)
	mockCache := new(mockCacheRepo)
	mockProd := new(mockMessageBroker)

	mockCache.
		On("GetOrder", ctx, "cache-miss-err").
		Return([]byte{}, errors.New("not found")).
		Once()

	mockRepo.
		On("GetOrder", "cache-miss-err").
		Return(models.Order{}, errors.New("db error")).
		Once()

	uc := NewOrderUseCase(mockRepo, mockCache, mockProd)

	_, err := uc.GetOrder(ctx, "cache-miss-err")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "db error")
	mockCache.AssertExpectations(t)
	mockRepo.AssertExpectations(t)
	mockCache.AssertNotCalled(t, "SetOrder")
}

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
		On("NewOrder", mock.MatchedBy(func(o models.Order) bool { return o.OrderUID == order.OrderUID })).
		Return(nil).
		Once()

	orderJSON, _ := json.Marshal(order)
	mockCache.
		On("SetOrder", ctx, "new-order-abc", mock.MatchedBy(func(b []byte) bool { return assert.JSONEq(t, string(orderJSON), string(b)) }), 24*time.Hour).
		Return(nil).
		Once()

	uc := NewOrderUseCase(mockRepo, mockCache, mockProd)

	err := uc.HandleMessage(ctx, data)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
	mockCache.AssertExpectations(t)
}

func TestHandleMessage_UnmarshalError(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(mockOrderRepo)
	mockCache := new(mockCacheRepo)
	mockProd := new(mockMessageBroker)

	uc := NewOrderUseCase(mockRepo, mockCache, mockProd)

	err := uc.HandleMessage(ctx, []byte("invalid json"))

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to unmarshal message")
	mockRepo.AssertNotCalled(t, "GetOrder")
	mockRepo.AssertNotCalled(t, "NewOrder")
	mockCache.AssertNotCalled(t, "SetOrder")
}

func TestHandleMessage_NewOrder_RepoError(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(mockOrderRepo)
	mockCache := new(mockCacheRepo)
	mockProd := new(mockMessageBroker)

	order := models.Order{OrderUID: "new-order-err"}
	data, _ := json.Marshal(order)

	mockRepo.
		On("GetOrder", "new-order-err").
		Return(models.Order{}, errors.New("not found")).
		Once()

	mockRepo.
		On("NewOrder", mock.MatchedBy(func(o models.Order) bool { return o.OrderUID == order.OrderUID })).
		Return(errors.New("repo save error")).
		Once()

	uc := NewOrderUseCase(mockRepo, mockCache, mockProd)

	err := uc.HandleMessage(ctx, data)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to save order to repository")
	assert.Contains(t, err.Error(), "repo save error")
	mockRepo.AssertExpectations(t)
	mockCache.AssertNotCalled(t, "SetOrder")
}

func TestHandleMessage_GetOrderError_NotNotFound(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(mockOrderRepo)
	mockCache := new(mockCacheRepo)
	mockProd := new(mockMessageBroker)

	order := models.Order{OrderUID: "get-err-proceed"}
	data, _ := json.Marshal(order)

	mockRepo.
		On("GetOrder", "get-err-proceed").
		Return(models.Order{}, errors.New("transient error")).
		Once()

	mockRepo.
		On("NewOrder", mock.MatchedBy(func(o models.Order) bool { return o.OrderUID == order.OrderUID })).
		Return(nil).
		Once()

	orderJSON, _ := json.Marshal(order)
	mockCache.
		On("SetOrder", ctx, "get-err-proceed", mock.MatchedBy(func(b []byte) bool { return assert.JSONEq(t, string(orderJSON), string(b)) }), 24*time.Hour).
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
		Return(order, nil).
		Once()

	uc := NewOrderUseCase(mockRepo, mockCache, mockProd)

	err := uc.HandleMessage(ctx, data)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
	mockRepo.AssertNotCalled(t, "NewOrder")
	mockCache.AssertNotCalled(t, "SetOrder")
}
