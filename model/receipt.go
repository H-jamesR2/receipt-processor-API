// model/receipt.go

package model

import (
	"sync"
	"fmt"
	"math"
	"strconv"
	"time"
	"unicode"
	"github.com/google/uuid"
)

type Item struct {
	ShortDescription string `json:"shortDescription"`
	Price            string `json:"price"`
}

type Receipt struct {
	Retailer     string `json:"retailer"`
	PurchaseDate string `json:"purchaseDate"`
	PurchaseTime string `json:"purchaseTime"`
	Items        []Item `json:"items"`
	Total        string `json:"total"`
	ID           string `json:"id"`
	Points       uint   `json:"points"`
}

var (
	receipts    = make(map[string]Receipt)
	receiptsMux sync.Mutex
)

func GenerateUniqueID() string {
	return uuid.New().String()
}


func CalculatePoints(receipt Receipt) uint {
	// Points Calculation
	points := uint(0)

	// add 1 pt for every alphaNumeric char in retailer name..
	retailerName := receipt.Retailer
	for _, c := range retailerName {
		if unicode.IsLetter(c) || unicode.IsDigit(c) {
			points++
		}
	}

	// If the total is a multiple of 0.25, add 25 pts.
	if totalFloat, err := strconv.ParseFloat(receipt.Total, 64); err == nil && math.Mod(totalFloat, 0.25) == 0 {
		points += 25

		// Get the decimal part of the float value
		totalDecimal := totalFloat - float64(int(totalFloat))
		// If decimal part of total == .00, add 50 pts.
		if totalDecimal == 0.0 {
			points += 50
		}
	} else if err != nil {
		fmt.Printf("Error parsing total: %v\n", err)
	}

	// add 5 points for every TWO items in the receipt.
	// 3/2 -> 1 (discards .5)
	points += (uint(((len(receipt.Items) / 2) * 5)))

	// go through items w/ pre-trimmed descriptions.
	for _, item := range receipt.Items {
		if len(item.ShortDescription)%3 == 0 {
			itemPrice, err := strconv.ParseFloat(item.Price, 64)
			if err != nil {
				fmt.Printf("Error parsing itemPrice: %v\n", err)
				continue
			} else {
				/*	> multiply itemPrice by 0.2
					> round to nearest integer
					> convert to unsigned int
					> add to points (uint)
				*/
				points += (uint(math.Ceil((itemPrice * 0.2))))
			}
		}
	}

	/*
		Processing date + time.
	*/
	// Parse the purchaseDate and check if the day is odd or even.
	layout := "2006-01-02"
	if t, err := time.Parse(layout, receipt.PurchaseDate); err == nil {
		// if odd
		if t.Day()%2 != 0 {
			points += 6
		}
	} else {
		fmt.Printf("Error parsing date %s: %v\n", receipt.PurchaseDate, err)
	}

	// Parse the purchaseTime and check if between 
	// after startTime && before endTime.
	if inRange, err := isTimeInRange(receipt.PurchaseTime); err == nil {
		if inRange {
			points += 10
		}
	} else {
		fmt.Println(err)
	}

	return points
}

func AddReceipt(receipt Receipt) {
	receiptsMux.Lock()
	receipts[receipt.ID] = receipt
	receiptsMux.Unlock()
}

func GetReceiptById(id string) (Receipt, bool) {
	receiptsMux.Lock()
	receipt, exists := receipts[id]
	receiptsMux.Unlock()
	return receipt, exists
}

func GetAllReceipts() []Receipt {
	receiptsMux.Lock()
	defer receiptsMux.Unlock()

	receiptsList := make([]Receipt, 0, len(receipts))
	for _, receipt := range receipts {
		receiptsList = append(receiptsList, receipt)
	}
	return receiptsList
}

// Helper Functions:
// Time Check
func isTimeInRange(timeStr string) (bool, error) {
	// Define the layout for time parsing
	layout := "15:04" // 24-hour time format

	// Parse the input time string
	t, err := time.Parse(layout, timeStr)
	if err != nil {
		return false, fmt.Errorf("error parsing time %s: %v", timeStr, err)
	}

	// Define the start and end times of the range
	startTime, _ := time.Parse(layout, "14:00")
	endTime, _ := time.Parse(layout, "16:00")

	// return True if time -> after Start AND before End
	return t.After(startTime) && t.Before(endTime), nil
}