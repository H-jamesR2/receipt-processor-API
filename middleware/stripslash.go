// middleware/stripslash.go

package middleware

import (
	"net/http"
	//"strings"
    "regexp"
	
    "rcpt-proc-challenge-ans/config"
	"go.uber.org/zap"
)


// StripSlash is a middleware that normalizes URLs by removing excessive slashes
func StripSlash(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        originalPath := r.URL.Path
        config.Log.Info("CleanURLPath: Original path", zap.String("path", originalPath))

        // Clean the URL
        
        re := regexp.MustCompile(`([^:]\/)\/+`)
        //re := regexp.MustCompile(`([^:]/)/+`)
        cleanedPath := re.ReplaceAllString(originalPath, "$1")

        // Remove trailing slash if it exists and the path is not just "/"
        if len(cleanedPath) > 1 && cleanedPath[len(cleanedPath)-1] == '/' {
            cleanedPath = cleanedPath[:len(cleanedPath)-1]
        }

        config.Log.Info("CleanURLPath",
            zap.String("originalPath", originalPath),
            zap.String("cleanedPath", cleanedPath))

        // Update the request's URL path
        r.URL.Path = cleanedPath

        next.ServeHTTP(w, r)
    })
}
