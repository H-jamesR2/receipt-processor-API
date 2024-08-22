// middleware/stripslash.go

package middleware

import (
    "net/http"
    "strings"
    "regexp"
)

// StripSlash is a middleware that normalizes URLs by removing excessive slashes
func StripSlash(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Trim leading and trailing slashes
        path := strings.Trim(r.URL.Path, "/")
        
        // Replace multiple consecutive slashes with a single slash
        re := regexp.MustCompile(`/+`)
        path = re.ReplaceAllString(path, "/")
        
        // Ensure the path starts with a single slash (unless it's empty)
        if path != "" {
            r.URL.Path = "/" + path
        } else {
            r.URL.Path = "/"
        }
        
        next.ServeHTTP(w, r)
    })
}