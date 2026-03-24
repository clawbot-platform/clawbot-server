package middleware

import (
	"context"
	"net/http"

	chimiddleware "github.com/go-chi/chi/v5/middleware"
)

type requestIDKey struct{}

func CaptureRequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := chimiddleware.GetReqID(r.Context())
		ctx := context.WithValue(r.Context(), requestIDKey{}, requestID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func requestIDFromContext(ctx context.Context) string {
	requestID, _ := ctx.Value(requestIDKey{}).(string)
	return requestID
}
