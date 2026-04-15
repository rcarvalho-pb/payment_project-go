package middleware

import (
	"fmt"
	"net/http"
	"time"
)

func LogginMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Printf("[%s] %s %s %s\n", time.Now().Format("15:04:05"), r.Host, r.Method, r.URL.Path)
		next.ServeHTTP(w, r)
	})
}
