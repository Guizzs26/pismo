package middleware

import (
	"log/slog"
	"net/http"
	"runtime/debug"

	"github.com/Guizzs26/pismo/pkg/httpx"
)

func Recovery(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				slog.Error("panic recovered",
					"panic", rec,
					"stack", string(debug.Stack()),
					"request_id", GetRequestID(r.Context()),
				)
				httpx.InternalServerError(w)
			}
		}()
		next.ServeHTTP(w, r)
	})
}
