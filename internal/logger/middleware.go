package logger

import (
	"go.uber.org/zap"
	"net/http"
	"time"
)

type responseWriter struct {
	http.ResponseWriter
	status int
	size   int64
}

func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		wrappedWriter := responseWriter{
			ResponseWriter: w,
			status:         0,
			size:           0,
		}

		next.ServeHTTP(&wrappedWriter, r)

		Logger.Info("HTTP Request",
			zap.String("method", r.Method),
			zap.String("path", r.URL.Path),
			zap.String("query", r.URL.RawQuery),
			zap.String("ip", r.RemoteAddr),
			zap.String("agent", r.UserAgent()),
			zap.Int("status", wrappedWriter.status),
			zap.Duration("duration", time.Since(start)),
			zap.Int64("response_size", wrappedWriter.size),
			zap.String("request_id", r.Header.Get("X-Request-ID")), // 分布式追踪ID
		)
	})
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.status = code
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	size, err := rw.ResponseWriter.Write(b)
	rw.size += int64(size)
	return size, err
}
