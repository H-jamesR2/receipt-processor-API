// model/receipt_test.go

package model

import (
	"testing"
)

func TestValidateReceipt(t *testing.T) {
	tests := []struct {
		receipt   Receipt
		isValid bool
	}{
		{Receipt{Retailer: "Test", PurchaseDate: "2022-01-01", PurchaseTime: "12:00", Items: []Item{{ShortDescription: "Test Item", Price: "10"}}, Total: "100"}, true},
		{Receipt{Retailer: "", PurchaseDate: "2022-01-01", PurchaseTime: "12:00", Items: []Item{{ShortDescription: "Test Item", Price: "10"}}, Total: "100"}, false},
		{Receipt{Retailer: "Test", PurchaseDate: "2022-01-01", PurchaseTime: "12:00", Items: []Item{}}, false},
		{Receipt{Retailer: "Test", PurchaseDate: "invalid-date", PurchaseTime: "12:00", Items: []Item{{ShortDescription: "Test Item", Price: "10"}}, Total: "100"}, false},
		{Receipt{Retailer: "Test", PurchaseDate: "2022-01-01", PurchaseTime: "invalid-time", Items: []Item{{ShortDescription: "Test Item", Price: "10"}}, Total: "100"}, false},
	}

	for _, test := range tests {
		err := ValidateReceipt(test.receipt)
		if (err == nil) != test.isValid {
			t.Errorf("ValidateOrder(%v) returned %v, expected %v", test.receipt, err == nil, test.isValid)
		}
	}
}

func TestCalculatePoints(t *testing.T) {
	receipt := Receipt{
		Retailer:     "Market",
		PurchaseDate: "2022-01-01",
		PurchaseTime: "12:00",	
		Items: []Item{
			{ShortDescription: "Test Item 1", Price: "10"},
			{ShortDescription: "Test Item 2", Price: "10"},
		},
		Total:        "20.00",
	}

	CalculatePoints(&receipt)

	if receipt.Points != 20 {
		t.Errorf("CalculatePoints(%v) returned %d points, expected 20", receipt, receipt.Points)
	}
}
