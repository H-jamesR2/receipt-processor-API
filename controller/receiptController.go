// controller/receiptController.go

package controller

import (
	"encoding/json"
	"fmt"
	"net/http"
	"rcpt-proc-challenge-ans/model"
	"regexp"
	"strings"
	"time"
)

// GET MethodS
func GetReceipt(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimSuffix(strings.TrimPrefix(r.URL.Path, "/receipts/"), "/")
	//path := strings.TrimPrefix(r.URL.Path, "/receipts/")
	if path == "" {
		// Return all receipts
		receipts := model.GetAllReceipts()
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(receipts)
		return
	}

	// Check if the path ends with "/points" to route to the GetReceiptPoints handler
	if strings.HasSuffix(path, "/points") {
		id := strings.TrimSuffix(path, "/points")
		GetReceiptPoints(w, r, id)
		return
	}

	// Return specific receipt
	receipt, exists := model.GetReceiptById(path)
	if !exists {
		http.Error(w, "Receipt not found", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(receipt)
}

// POST Method
func ProcessReceipt(w http.ResponseWriter, r *http.Request) {
	var receipt model.Receipt
	receipt.Items = []model.Item{} // Initialize Items to an empty slice

	err := json.NewDecoder(r.Body).Decode(&receipt)
	if err != nil {
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
	receipt.Points = model.CalculatePoints(receipt)
	model.AddReceipt(receipt)

	w.Header().Set("Content-Type", "application/json")
	response := struct {
		ID string `json:"id"`
	}{
		ID: receipt.ID,
	}
	json.NewEncoder(w).Encode(response)
}

// Updated handler to get points for a specific receipt
func GetReceiptPoints(w http.ResponseWriter, r *http.Request, id string) {
	receipt, exists := model.GetReceiptById(id)
	if !exists {
		http.Error(w, "Receipt not found, As such, 0 Points", http.StatusNotFound)
		return
	}

	//access stored Points for receipt @ ID
	points := receipt.Points
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(struct {
		Points uint `json:"points"`
	}{
		Points: points,
	})
}

// NotFoundHandler handles requests to non-existent endpoints
func NotFoundHandler(w http.ResponseWriter, r *http.Request) {
    // Set the status code to 404
    w.WriteHeader(http.StatusNotFound)
    
    // Write a standard error message
    response := map[string]string{
        "error": "Endpoint not found",
        "message": fmt.Sprintf("The requested URL %s was not found on this server.", r.URL.Path),
    }
    
    // Write the response in JSON format
    if err := json.NewEncoder(w).Encode(response); err != nil {
        http.Error(w, "Internal Server Error", http.StatusInternalServerError)
    }
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