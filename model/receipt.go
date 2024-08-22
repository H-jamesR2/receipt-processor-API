// model/receipt.go

package model

import (
	"errors"
	"fmt"
	"math"
	"strconv"

	"context"
	"encoding/json"
	"strings"
	"time"
	"unicode"
	"regexp"

	"rcpt-proc-challenge-ans/config"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

type Item struct {
	ID               uint      `json:"id"`
	SKU              SKU       `json:"sku"`
	ShortDescription string    `json:"shortDescription"`
	Quantity         int       `json:"quantity"`
	PricePaid        string    `json:"pricePaid"`
	ReceiptID        uuid.UUID `json:"receiptID"`
}

type Receipt struct {
	ID           uuid.UUID `json:"id"`
	Retailer     string    `json:"retailer"`
	PurchaseDate string    `json:"purchaseDate"`
	PurchaseTime string    `json:"purchaseTime"`
	Items        []Item    `json:"items"`
	Total        string    `json:"total"`
	Points       uint      `json:"points"`
}

/*
Prefix: A store-specific prefix (e.g., "AMZ" for Amazon,

	"WMT" for Walmart).

ProductCategory: Broad category (e.g., "ELEC" for electronics,

	"GROC" for grocery).

Manufacturer: The product manufacturer.
ProductLine: Specific product line or model.
Attributes: A map for various attributes like size, color, weight, etc.

	This allows for flexibility across different product types.

UniqueIdentifier: A unique identifier within the store's system.
*/
type SKU struct {
	Prefix           string            `json:"prefix"`
	ProductCategory  string            `json:"productCategory"`
	Manufacturer     string            `json:"manufacturer"`
	ProductLine      string            `json:"productLine"`
	Attributes       map[string]string `json:"attributes"`
	UniqueIdentifier string            `json:"uniqueIdentifier"`
}

type ReceiptStore interface {
    AddReceipt(receipt *Receipt) error
    GetReceiptByID(id uuid.UUID) (*Receipt, error)
    GetAllReceipts() ([]Receipt, error)
}



// GenerateID generates a new UUID and sets it as the receipt's ID
func (r *Receipt) GenerateID() {
    r.ID = config.GenerateUUID()
}

func (receipt *Receipt) ValidateReceipt() error {
	standardErrorPrefix := "error processing receipt:\n   "

	if receipt.Retailer == "" {
		emptyRetailerErr := errors.New(standardErrorPrefix + "retailer cannot be empty")
		config.Log.Error(
			standardErrorPrefix+"retailer error:",
			zap.Error(emptyRetailerErr))
		return emptyRetailerErr
	}

	if len(receipt.Items) == 0 {
		emptyItemsErr := errors.New(standardErrorPrefix + "items cannot be empty")
		config.Log.Error(
			standardErrorPrefix+"items array error:",
			zap.Error(emptyItemsErr))
		return emptyItemsErr
	}

	testItemsTotal := float64(0)
	for _, item := range receipt.Items {
		if item.ShortDescription == "" {
			emptyDescriptionErr := errors.New(standardErrorPrefix + "item description cannot be empty")
			config.Log.Error(
				standardErrorPrefix+"item description error:",
				zap.Error(emptyDescriptionErr))
			return emptyDescriptionErr
		}

		if item.PricePaid == "" {
			emptyItemPriceErr := errors.New(standardErrorPrefix + "item price cannot be empty")
			config.Log.Error(
				standardErrorPrefix+"item price error:",
				zap.Error(emptyItemPriceErr))
			return emptyItemPriceErr
		}

		// price less than or equal to 0 or an error...
		price, priceErr := strconv.ParseFloat(item.PricePaid, 64)
		if price <= 0 {
			return errors.New(standardErrorPrefix + "item price must be greater than zero")
		} else if priceErr != nil {
			config.Log.Error(
				standardErrorPrefix+"error on item price",
				zap.Error(priceErr))
			return priceErr
		} else {
			// add to testTotal to verify and check
			testItemsTotal += price
		}
	}
	// run check on items before cleaning...

	// convert to float + round...
	receiptTotal, receiptErr := strconv.ParseFloat(receipt.Total, 64)
	receiptTotal, testItemsTotal = config.RoundToNearestCent(receiptTotal), config.RoundToNearestCent(testItemsTotal)

	if receiptErr != nil {
		config.Log.Error(
			standardErrorPrefix+"error on total price",
			zap.Error(receiptErr))
		return errors.New(standardErrorPrefix + "error on total price")
	} else if receiptTotal != testItemsTotal {
		mismatchTotalError := errors.New(standardErrorPrefix + "item calculatedTotal does not match Total price")
		config.Log.Error("mismatched total: "+
			"receiptTotal of "+strconv.FormatFloat(receiptTotal, 'f', -1, 64)+
			"versus"+
			"testItemsTotal of "+strconv.FormatFloat(testItemsTotal, 'f', -1, 64),
			zap.Error(mismatchTotalError))
		return mismatchTotalError
	}

	// Validate purchase date
	if err := validateDate(receipt.PurchaseDate); err != nil {
		config.Log.Error("Invalid Purchase Date", zap.Error(err))
		return err
	}

	// Validate purchase time
	if err := validateTime(receipt.PurchaseTime); err != nil {
		config.Log.Error("Invalid Purchase Time", zap.Error(err))
		return err
	}

	return nil
}

//func (receipt *Receipt) ValidateReceipt() error {
func (receipt *Receipt) CalculatePoints() {
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

func (s *SKU) ParseSKU(skuString string) error {
	// Assuming SKU string format:
	// "WMT-GROC-NESTLE-CHOC-WEIGHT-100G-67890"
	skuParts := strings.Split(skuString, "-")
	if len(skuParts) < 5 { // Minimum: Prefix, Category, Manufacturer, ProductLine, UniqueID
		skuParsingErr := errors.New("invalid SKU format")
		config.Log.Error("Invalid Purchase Date", zap.Error(skuParsingErr))
		return skuParsingErr
	}

	s.Prefix = skuParts[0]
	s.ProductCategory = skuParts[1]
	s.Manufacturer = skuParts[2]
	s.ProductLine = skuParts[3]

	s.Attributes = make(map[string]string)
	for i := 4; i < len(skuParts)-1; i += 2 {
		if i+1 < len(skuParts) {
			s.Attributes[skuParts[i]] = skuParts[i+1]
		}
	}

	s.UniqueIdentifier = skuParts[len(skuParts)-1]

	return nil
}

func (s SKU) CombinePartsToString() string {
	skuParts := []string{s.Prefix, s.ProductCategory, s.Manufacturer, s.ProductLine}

	for key, value := range s.Attributes {
		skuParts = append(skuParts, key, value)
	}

	skuParts = append(skuParts, s.UniqueIdentifier)

	return strings.Join(skuParts, "-")
}

func (s *SKU) UnmarshalJSON(data []byte) error {
	var skuString string
	if err := json.Unmarshal(data, &skuString); err != nil {
		return err
	}

	return s.ParseSKU(skuString)
}


// AddReceipt inserts a new receipt and its associated items into the database
func AddReceipt(db *pgxpool.Pool, receipt *Receipt) error {
	//return db.Create(receipt).Error
	startTime := time.Now()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	tx, err := db.Begin(ctx)
	if err != nil {
		config.Log.Error("Failed to begin transaction", zap.Error(err))
		return err
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx, `
		INSERT INTO receipts (id, retailer, purchase_date, purchase_time, total, points)
		VALUES ($1, $2, $3::date, $4::time, $5, $6)
	`, receipt.ID, receipt.Retailer, receipt.PurchaseDate, receipt.PurchaseTime, receipt.Total, receipt.Points)
	if err != nil {
		config.Log.Error("Failed to insert receipt", zap.Error(err))
		return err
	}

	for _, item := range receipt.Items {
		item.ReceiptID = receipt.ID
		var sku SKU

		// Parse the SKU string into the SKU struct
		if err := sku.ParseSKU(item.SKU.CombinePartsToString()); err != nil {
			config.Log.Error("Failed to parse SKU", zap.Error(err))
			return err
		}

		// Check if SKU already exists
		var skuExists bool
		err = tx.QueryRow(ctx, `
            SELECT EXISTS(SELECT 1 FROM skus WHERE unique_identifier = $1)
        `, sku.UniqueIdentifier).Scan(&skuExists)
		if err != nil {
			config.Log.Error("Failed to check SKU existence", zap.Error(err))
			return err
		}

		// Insert SKU if it doesn't exist
		if !skuExists {
			_, err = tx.Exec(ctx, `
                INSERT INTO skus (unique_identifier, prefix, product_category, manufacturer, product_line, attributes)
                VALUES ($1, $2, $3, $4, $5, $6)
            `, sku.UniqueIdentifier, sku.Prefix, sku.ProductCategory, sku.Manufacturer, sku.ProductLine, sku.Attributes)
			if err != nil {
				config.Log.Error("Failed to insert SKU", zap.Error(err))
				return err
			}
		}

		// Insert the item with the SKU's unique identifier
		_, err = tx.Exec(ctx, `
            INSERT INTO items (short_description, quantity, price_paid, receipt_id, sku_id)
            VALUES ($1, $2, $3, $4, $5)
        `, item.ShortDescription, item.Quantity, item.PricePaid, item.ReceiptID, sku.UniqueIdentifier)
		if err != nil {
			config.Log.Error("Failed to insert item", zap.Error(err))
			return err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		config.Log.Error("Failed to commit transaction", zap.Error(err))
		return err
	}

	executionTime := time.Since(startTime)
	config.Log.Info("AddReceipt executed", zap.Duration("duration", executionTime))

	return nil
}


func GetReceiptByID(db *pgxpool.Pool, id uuid.UUID) (*Receipt, error) {
	startTime := time.Now()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	receipt := &Receipt{ID: id}
	err := db.QueryRow(ctx, `
		SELECT retailer, 
			TO_CHAR(purchase_date, 'YYYY-MM-DD') as purchase_date, 
			TO_CHAR(purchase_time, 'HH24:MI') as purchase_time, 
			total, points
		FROM receipts
		WHERE id = $1
	`, id).Scan(&receipt.Retailer, &receipt.PurchaseDate, &receipt.PurchaseTime, &receipt.Total, &receipt.Points)
	if err != nil {
		config.Log.Error("Failed to retrieve receipt", zap.String("id", id.String()), zap.Error(err))
		return nil, err
	}

	rows, err := db.Query(ctx, `
        SELECT i.id, i.short_description, i.quantity, i.price_paid, s.unique_identifier, s.prefix, s.product_category, s.manufacturer, s.product_line, s.attributes
        FROM items i
        JOIN skus s ON i.sku_id = s.unique_identifier
        WHERE i.receipt_id = $1
    `, id)
	if err != nil {
		config.Log.Error("Failed to retrieve items for receipt", zap.String("receipt_id", id.String()), zap.Error(err))
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var item Item
		var sku SKU
		err := rows.Scan(
			&item.ID, &item.ShortDescription, &item.Quantity, &item.PricePaid,
			&sku.UniqueIdentifier, &sku.Prefix, &sku.ProductCategory, &sku.Manufacturer, &sku.ProductLine, &sku.Attributes)
		if err != nil {
			config.Log.Error("Failed to scan item", zap.Error(err))
			return nil, err
		}

		item.ReceiptID = id // Set the ReceiptID to the current receipt's ID
		item.SKU = sku      // Set
		receipt.Items = append(receipt.Items, item)
	}

	executionTime := time.Since(startTime)
	config.Log.Info("GetReceiptByID executed", zap.Duration("duration", executionTime))

	return receipt, nil
}

func GetAllReceipts(db *pgxpool.Pool) ([]Receipt, error) {
	startTime := time.Now()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	rows, err := db.Query(ctx, `
        SELECT id, retailer, 
               TO_CHAR(purchase_date, 'YYYY-MM-DD') as purchase_date, 
               TO_CHAR(purchase_time, 'HH24:MI') as purchase_time, 
               total, points
        FROM receipts
    `)
	if err != nil {
		config.Log.Error("Failed to retrieve receipts", zap.Error(err))
		return nil, err
	}
	defer rows.Close()

	var receipts []Receipt

	for rows.Next() {
		var receipt Receipt
		err := rows.Scan(&receipt.ID, &receipt.Retailer, &receipt.PurchaseDate, &receipt.PurchaseTime, &receipt.Total, &receipt.Points)
		if err != nil {
			config.Log.Error("Failed to scan receipt", zap.Error(err))
			return nil, err
		}

		// Fetch items for this receipt
		itemRows, err := db.Query(ctx, `
            SELECT i.id, i.short_description, i.quantity, i.price_paid, s.unique_identifier, s.prefix, s.product_category, s.manufacturer, s.product_line, s.attributes
            FROM items i
            JOIN skus s ON i.sku_id = s.unique_identifier
            WHERE i.receipt_id = $1
        `, receipt.ID)
		if err != nil {
			config.Log.Error("Failed to retrieve items for receipt", zap.String("receipt_id", receipt.ID.String()), zap.Error(err))
			return nil, err
		}
		defer itemRows.Close()

		for itemRows.Next() {
			var item Item
			var sku SKU
			err := itemRows.Scan(
				&item.ID, &item.ShortDescription, &item.Quantity, &item.PricePaid,
				&sku.UniqueIdentifier, &sku.Prefix, &sku.ProductCategory, &sku.Manufacturer, &sku.ProductLine, &sku.Attributes)
			if err != nil {
				config.Log.Error("Failed to scan item", zap.Error(err))
				return nil, err
			}

			item.ReceiptID = receipt.ID // Set the ReceiptID to the current receipt's ID
			item.SKU = sku
			receipt.Items = append(receipt.Items, item)
		}
		//itemRows.Close()

		receipts = append(receipts, receipt)
	}

	executionTime := time.Since(startTime)
	config.Log.Info("GetAllReceipts executed", zap.Duration("duration", executionTime))

	return receipts, nil
}

// Helper Functions:
// Getters
func GetItemsCount(db *pgxpool.Pool) (int, error) {
    var count int
    err := db.QueryRow(context.Background(), "SELECT COUNT(*) FROM items").Scan(&count)
    if err != nil {
        config.Log.Error("Failed to retrieve item count", zap.Error(err))
        return 0, err
    }
    return count, nil
}

/*
	Validators:
	- Date
	- Time
*/
// Validation Functions
func validateDate(dateStr string) error {
	if _, err := time.Parse("2006-01-02", dateStr); err != nil {
		config.Log.Error("Invalid purchase date", zap.String("date", dateStr), zap.Error(err))
		return errors.New("error processing receipt, invalid purchase date: date " + dateStr + " is invalid")
	}

	config.Log.Info("Valid purchase date", zap.String("date", dateStr))
	return nil
}

func validateTime(timeStr string) error {
	if _, err := time.Parse("15:04", timeStr); err != nil {
		config.Log.Error("Invalid purchase time", zap.String("time", timeStr), zap.Error(err))
		return errors.New("error processing receipt, invalid purchase time: time " + timeStr + " is invalid")
	}

	config.Log.Info("Valid purchase time", zap.String("time", timeStr))
	return nil
}

// Time Check
func isTimeInRange(timeStr string) (bool, error) {
	layout := "15:04"
	t, err := time.Parse(layout, timeStr)
	if err != nil {
		config.Log.Error("Error parsing time", zap.String("time", timeStr), zap.Error(err))
		return false, fmt.Errorf("error parsing time %s: %v", timeStr, err)
	}

	startTime, _ := time.Parse(layout, "14:00")
	endTime, _ := time.Parse(layout, "16:00")

	inRange := t.After(startTime) && t.Before(endTime)
	config.Log.Info("Time range check", zap.String("time", timeStr), zap.Bool("inRange", inRange))
	return inRange, nil
}

// Calculation Functions
func calculatePointsFromRetailerAlphaNumChar(retailer string) uint {
	points := uint(0)
	for _, c := range retailer {
		if unicode.IsLetter(c) || unicode.IsDigit(c) {
			points++
		}
	}

	config.Log.Info("Calculated points from retailer", zap.String("retailer", retailer), zap.Uint("points", points))
	return points
}

func calculatePointsFromTotal(total string) uint {
	points := uint(0)

	totalFloat, err := strconv.ParseFloat(total, 64)
	if err != nil {
		config.Log.Error("Error parsing total", zap.String("total", total), zap.Error(err))
		return points
	}

	if math.Mod(totalFloat, 0.25) == 0 {
		points += 25
		if math.Mod(totalFloat*100, 100) == 0 {
			points += 50
		}
	}

	config.Log.Info("Calculated points from total", zap.String("total", total), zap.Uint("points", points))
	return points
}

func calculatePointsFromItemPriceAndDesc(items []Item) uint {
	points := uint(0)

	for _, item := range items {
		config.Log.Info("Length of Item's Short Description", zap.String("itemDescription", item.ShortDescription), zap.Int("itemDescriptionLength", len(item.ShortDescription)))
		if len(item.ShortDescription)%3 == 0 {
			itemPricePaid, err := strconv.ParseFloat(item.PricePaid, 64)
			if err != nil {
				config.Log.Error("Error parsing item price", zap.String("itemPricePaid", item.PricePaid), zap.Error(err))
				continue
			}

			itemPoints := uint(math.Ceil((itemPricePaid * 0.2)))
			points += itemPoints
			config.Log.Info("Calculated points from item", zap.String("itemDescription", item.ShortDescription), zap.Uint("points", itemPoints))
		}
	}

	config.Log.Info("Total points from items", zap.Uint("totalPoints", points))
	return points
}

func calculatePointsFromPurchaseDate(purchaseDate string) uint {
	points := uint(0)

	layout := "2006-01-02"
	t, err := time.Parse(layout, purchaseDate)
	if err != nil {
		config.Log.Error("Error parsing purchase date", zap.String("purchaseDate", purchaseDate), zap.Error(err))
		return points
	}

	if t.Day()%2 != 0 {
		points += 6
	}

	config.Log.Info("Calculated points from purchase date", zap.String("purchaseDate", purchaseDate), zap.Uint("points", points))
	return points
}

func calculatePointsFromPurchaseTime(purchaseTime string) uint {
	points := uint(0)

	inRange, err := isTimeInRange(purchaseTime)
	if err != nil {
		config.Log.Error("Error checking time range", zap.String("purchaseTime", purchaseTime), zap.Error(err))
		return points
	}

	if inRange {
		points += 10
	}

	config.Log.Info("Calculated points from purchase time", zap.String("purchaseTime", purchaseTime), zap.Uint("points", points))
	return points
}

func (r *Receipt) CleanItemShortDescriptions() {
    for i, item := range r.Items {
        // Trim leading and trailing spaces
        description := strings.TrimSpace(item.ShortDescription)

        // Replace multiple spaces with a single space
        re := regexp.MustCompile(`\s+`)
        description = re.ReplaceAllString(description, " ")

        r.Items[i].ShortDescription = description
    }
}
