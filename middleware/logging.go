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
