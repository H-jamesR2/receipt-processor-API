// controller/responseHandler.go

package controller

import (
    "net/http"
    "encoding/json"
    "rcpt-proc-challenge-ans/config"
    "go.uber.org/zap"
)


func NotFoundHandler(w http.ResponseWriter, r *http.Request) {
    config.Log.Info("Not Found",
        zap.String("path", r.URL.Path),
        zap.String("method", r.Method),
    )

    sendJSONResponse(w, http.StatusNotFound, ErrorResponse{
        Error: "The requested resource was not found.",
    })
}

func MethodNotAllowedHandler(w http.ResponseWriter, r *http.Request) {
    config.Log.Info("Method Not Allowed",
        zap.String("path", r.URL.Path),
        zap.String("method", r.Method),
    )

    sendJSONResponse(w, http.StatusMethodNotAllowed, ErrorResponse{
        Error: "The requested method is not allowed for this resource.",
    })
}

// Helper Methods
func sendJSONResponse(w http.ResponseWriter, status int, payload interface{}) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(status)
    json.NewEncoder(w).Encode(payload)
}