// model/receipt.go

package model

import (
	"errors"
	"fmt"
	"math"
	"strconv"
	//"sync"
	"time"
	"unicode"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Item struct {
	ID               uint   `gorm:"primary_key" json:"id"`
	ShortDescription string `json:"shortDescription"`
	Price            string `json:"price"`
	ReceiptID        string `json:"receipt_id"`
}

type Receipt struct {
	Retailer     string `json:"retailer"`
	PurchaseDate string `json:"purchaseDate"`
	PurchaseTime string `json:"purchaseTime"`
	Items        []Item `json:"items"`
	Total        string `json:"total"`
	ID           string `gorm:"primary_key" json:"id"`
	Points       uint   `json:"points"`
}

/*
var (
	receipts    = make(map[string]Receipt)
	receiptsMux sync.Mutex
) */

func GenerateUniqueID() string {
	return uuid.New().String()
}

func ValidateReceipt(receipt Receipt) error {
	standardErrorPrefix := "error processing receipt:\n   "

	if receipt.Retailer == "" {
		return errors.New(standardErrorPrefix + "retailer cannot be empty")
	}

	if len(receipt.Items) == 0 {
		return errors.New(standardErrorPrefix + "items cannot be empty")
	}

	testItemsTotal := float64(0)
	for _, item := range receipt.Items {
		if item.ShortDescription == "" {
			return errors.New(standardErrorPrefix + "item description cannot be empty")
		}

		if item.Price == "" {
			return errors.New(standardErrorPrefix + "item price cannot be empty")
		}

		// price less than or equal to 0 or an error...
		price, priceErr := strconv.ParseFloat(item.Price, 64)
		if price <= 0 {
			return errors.New(standardErrorPrefix + "item price must be greater than zero")
		} else if priceErr != nil  {
			return priceErr
		} else {
			// add to testTotal to verify and check
			testItemsTotal += price
		}
	}

	receiptTotal, receiptErr := strconv.ParseFloat(receipt.Total, 64)
	if receiptErr != nil {
		return errors.New(standardErrorPrefix + "error on total price")
	} else if receiptTotal != testItemsTotal {
		return errors.New(standardErrorPrefix + "item calculatedTotal does not match Total price")
	}

	// Validate purchase date
	if err := validateDate(receipt.PurchaseDate); err != nil {
		return err
	}

	// Validate purchase time
	if err := validateTime(receipt.PurchaseTime); err != nil {
		return err
	}

	return nil
}

func CalculatePoints(receipt *Receipt) {
	// Points Calculation
	points := uint(0)

	// add 1 pt for every alphaNumeric char in retailer name..
	points += calculatePointsFromRetailerAlphaNumChar(receipt.Retailer)

	// If the total is a multiple of 0.25, add 25 pts.
	points += calculatePointsFromTotal(receipt.Total)

	// add 5 points for every TWO items in the receipt.
	// 3/2 -> 1 (discards .5)
	points += (uint(((len(receipt.Items) / 2) * 5)))

	// go through items w/ pre-trimmed descriptions.
	points += calculatePointsFromItemPriceAndDesc(receipt.Items)

	/*
		Processing date + time.
	*/
	// Parse the purchaseDate and check if the day is odd or even.
	points += calculatePointsFromPurchaseDate(receipt.PurchaseDate)

	// Parse the purchaseTime and check if between
	// after startTime && before endTime.
	points += calculatePointsFromPurchaseTime(receipt.PurchaseTime)

	receipt.Points = points
}

/*
func AddReceipt(receipt Receipt) {
	receiptsMux.Lock()
	receipts[receipt.ID] = receipt
	receiptsMux.Unlock()
} */

func AddReceipt(db *gorm.DB, receipt *Receipt) error {
	return db.Create(receipt).Error
}

/*
func GetReceiptById(id string) (Receipt, bool) {
	receiptsMux.Lock()
	receipt, exists := receipts[id]
	receiptsMux.Unlock()
	return receipt, exists
} */

func GetReceiptByID(db *gorm.DB, id string) (*Receipt, error) {
	var receipt Receipt
	if err := db.Preload("Items").First(&receipt, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &receipt, nil
}

/*
func GetAllReceipts() []Receipt {
	receiptsMux.Lock()
	defer receiptsMux.Unlock()

	receiptsList := make([]Receipt, 0, len(receipts))
	for _, receipt := range receipts {
		receiptsList = append(receiptsList, receipt)
	}
	return receiptsList
} */

func GetAllReceipts(db *gorm.DB) ([]Receipt, error) {
	var receipts []Receipt
	if err := db.Preload("Items").Find(&receipts).Error; err != nil {
		return nil, err
	}
	return receipts, nil
}

// Helper Functions:

/* 
	Validators:
	- Date
	- Time
*/
func validateDate(dateStr string) error {
	// No need to match if Valid, just convert
	/*
		if matched, _ := regexp.MatchString(`^\d{4}-\d{2}-\d{2}$`, dateStr); !matched {
			return errors.New("error: date format must be YYYY-MM-DD")
		} */
	if _, err := time.Parse("2006-01-02", dateStr); err != nil {
		return errors.New("error processing receipt,\n   invalid purchase date: date " + dateStr + " is invalid")
	}

	return nil
}

func validateTime(timeStr string) error {
	// No need to match if Valid, just convert
	/*
		if matched, _ := regexp.MatchString(`^\d{2}:\d{2}$`, timeStr); !matched {
			return errors.New("error: time format must be HH:MM")
		} */
	if _, err := time.Parse("15:04", timeStr); err != nil {
		return errors.New("error processing receipt,\n   invalid purchase time: time " + timeStr + " is invalid")
	}
	return nil
}

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

/* 
	Calculation Functions:
*/
func calculatePointsFromRetailerAlphaNumChar(retailer string) uint {
	points := uint(0)
	for _, c := range retailer {
		if unicode.IsLetter(c) || unicode.IsDigit(c) {
			points++
		}
	}

	return points
}

func calculatePointsFromTotal(total string) uint {
	points := uint(0)

	if totalFloat, err := strconv.ParseFloat(total, 64); err == nil && math.Mod(totalFloat, 0.25) == 0 {
		points += 25

		// Get the decimal part of the float value
		// totalDecimal := totalFloat - float64(int(totalFloat))
		// If decimal part of total == .00, add 50 pts.
		if math.Mod(totalFloat*100, 100) == 0 {
			points += 50
		}
	} else if err != nil {
		fmt.Printf("Error parsing total: %v\n", err)
	}

	return points
}

func calculatePointsFromItemPriceAndDesc(items []Item) uint {
	points := uint(0)

	for _, item := range items {
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

	return points
}

func calculatePointsFromPurchaseDate(purchaseDate string) uint {
	points := uint(0)

	layout := "2006-01-02"
	if t, err := time.Parse(layout, purchaseDate); err == nil {
		// if odd
		if t.Day()%2 != 0 {
			points += 6
		}
	} else {
		fmt.Printf("Error parsing date %s: %v\n", purchaseDate, err)
	}

	return points
}

func calculatePointsFromPurchaseTime(purchaseTime string) uint {
	points := uint(0)

	if inRange, err := isTimeInRange(purchaseTime); err == nil {
		if inRange {
			points += 10
		}
	} else {
		fmt.Println(err)
	}

	return points
}