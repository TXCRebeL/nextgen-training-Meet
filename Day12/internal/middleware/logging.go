package middleware

import (
	"fmt"
	"net/http"
	"strings"
	"time"
)

type responseRecorder struct {
	http.ResponseWriter
	start       time.Time
	status      int
	wroteHeader bool
}

func (rw *responseRecorder) WriteHeader(code int) {
	if rw.wroteHeader {
		return
	}
	rw.wroteHeader = true
	rw.status = code
	duration := time.Since(rw.start)
	rw.Header().Set("X-Response-Time", duration.String())
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseRecorder) Write(b []byte) (int, error) {
	if !rw.wroteHeader {
		rw.WriteHeader(http.StatusOK)
	}
	return rw.ResponseWriter.Write(b)
}

func Logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		defer func() {
			if err := recover(); err != nil {
				fmt.Printf("PANIC RECOVERED: %v\n", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
		}()

		if r.Method == "POST" || r.Method == "PUT" || r.Method == "PATCH" {
			if !strings.Contains(r.Header.Get("Content-Type"), "application/json") {
				http.Error(w, "Content-Type must be application/json", http.StatusUnsupportedMediaType)
				return
			}
		}

		rw := &responseRecorder{
			ResponseWriter: w,
			start:          start,
			status:         http.StatusOK, // Default to 200
		}

		next.ServeHTTP(rw, r)

		if !rw.wroteHeader {
			rw.WriteHeader(http.StatusOK)
		}

		duration := time.Since(start)
		fmt.Printf("%s %s %s %d\n", r.Method, r.URL.Path, duration, rw.status)
	})
}
