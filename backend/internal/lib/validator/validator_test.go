// Package validator provides validation utilities for domain models.
// It uses go-playground/validator under the hood.
package validator

import (
	"WB/internal/models"
	"testing"
	"time"
)

var correctOrder = &models.Order{
	OrderUID:          "test",
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

var incorrectOrder = &models.Order{
	OrderUID:          "test", 
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
		Email:   "testgmail.com", //incorrect email
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


func TestValidateOrder(t *testing.T) {
	type args struct {
		order *models.Order
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "ok",
			args: args{order: correctOrder},
			wantErr: false,
		},
		{
			name: "incorrect email",
			args: args{order: incorrectOrder},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := ValidateOrder(tt.args.order); (err != nil) != tt.wantErr {
				t.Errorf("ValidateOrder() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
