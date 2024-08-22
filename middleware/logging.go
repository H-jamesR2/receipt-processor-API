// middleware/logging.go

package middleware

import (
	"net/http"
	"rcpt-proc-challenge-ans/config"
	"time"

	"go.uber.org/zap"
)

func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		next.ServeHTTP(w, r)

		config.Log.Info("Request",
			zap.String("method", r.Method),
			zap.String("url", r.URL.String()),
			zap.Duration("duration", time.Since(start)),
		)
	})
}

/*
// Add this to your middleware package
func PreProcessLoggingMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        config.Log.Info("Incoming request",
            zap.String("method", r.Method),
            zap.String("path", r.URL.Path),
            zap.String("rawPath", r.URL.RawPath))
        next.ServeHTTP(w, r)
    })
} */
