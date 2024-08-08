// controller/response.go

package controller

// CreateReceiptResponse represents the response for creating a receipt
type ProcessReceiptResponse struct {
    ID string `json:"id"`
}

// GetReceiptPointsResponse represents the response for getting receipt points
type GetReceiptPointsResponse struct {
    Points uint `json:"points"`
}
