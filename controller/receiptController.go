// controller/receiptController.go

package controller

import (
	"encoding/json"
	"fmt"
	"net/http"
	"rcpt-proc-challenge-ans/config"
	"rcpt-proc-challenge-ans/model"
	"regexp"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

// GetReceipt godoc
// @Summary Get a receipt by ID
// @Description Get a receipt by ID
// @Tags receipts
// @Produce json
// @Param id path string true "Receipt ID"
// @Success 200 {object} model.Receipt
// @Failure 404 {string} string "Receipt not found"
// @Router /receipts/{id} [get]
func GetReceipt(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	receipt, err := model.GetReceiptByID(config.DB, id)
	if err != nil {
		config.Log.Error("Receipt not found", zap.String("id", id), zap.Error(err))
		http.Error(w, "Receipt not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(receipt)
}

// CreateReceipt godoc
// @Summary Create a receipt
// @Description Create a new receipt
// @Tags receipts
// @Accept json
// @Produce json
// @Param receipt body model.Receipt true "Receipt"
// @Success 200 {object} ProcessReceiptResponse
// @Failure 400 {string} string "Invalid input"
// @Failure 500 {string} string "Failed to create receipt"
// @Router /receipts/process [post]
func ProcessReceipt(w http.ResponseWriter, r *http.Request) {
	var receipt model.Receipt
	if err := json.NewDecoder(r.Body).Decode(&receipt); err != nil {
		config.Log.Error("Invalid input", zap.Error(err))
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	// Validate the order
	if err := model.ValidateReceipt(receipt); err != nil {
		config.Log.Error("Invalid receipt data", zap.Error(err))
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// run trimming operation for itemShortDescriptions...
	cleanItemShortDescriptions(&receipt)
	// reformat Date if needed.
	if formattedDate, err := parseAndFormatDate(receipt.PurchaseDate); err == nil {
		receipt.PurchaseDate = formattedDate
	} else {
		fmt.Println(err)
	}

	// reformat Time if needed.
	if formattedTime, err := parseAndFormatTime(receipt.PurchaseTime); err == nil {
		receipt.PurchaseTime = formattedTime
	} else {
		fmt.Println(err)
	}

	receipt.ID = model.GenerateUniqueID()
	model.CalculatePoints(&receipt)

	// AddReceipt
	if err := model.AddReceipt(config.DB, &receipt); err != nil {
		config.Log.Error("Failed to create receipt", zap.Error(err))
		http.Error(w, "Failed to create receipt", http.StatusInternalServerError)
		return
	}


	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ProcessReceiptResponse{
		ID: receipt.ID,
	})
}



// GetReceiptPoints godoc
// @Summary Get receipt points by ID
// @Description Get receipt points by ID
// @Tags receipts
// @Produce json
// @Param id path string true "Receipt ID"
// @Success 200 {object} GetReceiptPointsResponse
// @Failure 404 {string} string "Receipt not found"
// @Router /receipts/{id}/points [get]
func GetReceiptPoints(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	receipt, err := model.GetReceiptByID(config.DB, id)
	if err != nil {
		config.Log.Error("Receipt not found", zap.String("id", id), zap.Error(err))
		http.Error(w, "Receipt not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(GetReceiptPointsResponse{
		Points: receipt.Points,
	})
}


/*
	Helper Functions
*/
// Date Functions
func isISODateFormat(dateStr string) bool {
	// Regular expression to check if the date string is in YYYY-MM-DD format
	re := regexp.MustCompile(`^([0-9]{4})-(0[1-9]|1[0-2])-(0[1-9]|[12][0-9]|3[01])$`)
	return re.MatchString(dateStr)
}

func parseAndFormatDate(dateStr string) (string, error) {
	// If the date string is already in YYYY-MM-DD format, return it directly
	if isISODateFormat(dateStr) {
		return dateStr, nil
	}

	// Define possible date formats
	formats := []string{
		"2006-01-02", // YYYY-MM-DD
		"02-01-2006", // DD-MM-YYYY
		"01/02/2006", // MM/DD/YYYY
		"2006/01/02", // YYYY/MM/DD
	}

	var parsedDate time.Time
	var err error

	// Try parsing with each format
	for _, format := range formats {
		if parsedDate, err = time.Parse(format, dateStr); err == nil {
			break
		}
	}

	// parsed dateStr not a valid dateString.
	if err != nil {
		return "", fmt.Errorf("error parsing date %s: %v", dateStr, err)
	}

	// Format date to YYYY-MM-DD
	return parsedDate.Format("2006-01-02"), nil
}

// Time Functions
func is24HourFormat(timeStr string) bool {
	// Regular expression to check if the time string is in HH:MM format
	re := regexp.MustCompile(`^([01][0-9]|2[0-3]):[0-5][0-9]$`)
	return re.MatchString(timeStr)
}

func parseAndFormatTime(timeStr string) (string, error) {
	// If the time string is already in 24-hour format, return it directly
	if is24HourFormat(timeStr) {
		return timeStr, nil
	}

	// Define possible time formats
	formats := []string{
		"15:04",       // 24-hour clock with minutes
		"15:04:05",    // 24-hour clock with seconds
		"03:04 PM",    // 12-hour clock with AM/PM
		"03:04:05 PM", // 12-hour clock with seconds and AM/PM
	}

	var parsedTime time.Time
	var err error

	// Try parsing with each format
	for _, format := range formats {
		if parsedTime, err = time.Parse(format, timeStr); err == nil {
			break
		}
	}

	// parsed timeStr not a valid timeString.
	if err != nil {
		return "", fmt.Errorf("error parsing time %s: %v", timeStr, err)
	}

	// Format time to 24-hour clock format
	return parsedTime.Format("15:04"), nil
}

// Clean item descriptions by trimming and reducing multiple spaces
func cleanItemShortDescriptions(receipt *model.Receipt) {
	for i, item := range receipt.Items {
		// Trim leading and trailing spaces
		description := strings.TrimSpace(item.ShortDescription)

		// Replace multiple spaces with a single space
		re := regexp.MustCompile(`\s+`)
		description = re.ReplaceAllString(description, " ")

		receipt.Items[i].ShortDescription = description
	}
}
