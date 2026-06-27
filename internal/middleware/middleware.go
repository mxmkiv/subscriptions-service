package middleware

import (
	"log/slog"
	"net/http"
)

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (s *statusRecorder) WriteHeader(code int) {
	s.status = code
	s.ResponseWriter.WriteHeader(code)
}

func LoggingRequests(logger *slog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rec := &statusRecorder{
			ResponseWriter: w,
			status:         http.StatusOK,
		}

		next.ServeHTTP(rec, r)
		logger.Info("request", "method", r.Method, "path", r.URL.Path, "remote_addr", r.RemoteAddr, "status", rec.status)
	})
}
