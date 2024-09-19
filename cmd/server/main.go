package main

import (
	"encoding/json"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	httpRequests = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"path", "method", "status"},
	)

	responseTime = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_response_time_seconds",
			Help:    "Response time in seconds",
			Buckets: prometheus.LinearBuckets(0.1, 0.1, 10),
		},
		[]string{"path"},
	)

	activeUsers = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "active_users",
			Help: "Number of active users",
		},
	)
)

func init() {
	prometheus.MustRegister(httpRequests)
	prometheus.MustRegister(responseTime)
	prometheus.MustRegister(activeUsers)
}

func metricsMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Wrap ResponseWriter to capture status code
		wrappedWriter := &responseWriter{ResponseWriter: w, statusCode: 200}

		next.ServeHTTP(wrappedWriter, r)

		duration := time.Since(start).Seconds()

		httpRequests.WithLabelValues(r.URL.Path, r.Method, strconv.Itoa(wrappedWriter.statusCode)).Inc()
		responseTime.WithLabelValues(r.URL.Path).Observe(duration)
	}
}

type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	time.Sleep(time.Duration(rand.Intn(1000)) * time.Millisecond)
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Welcome to our API!"})
}

func userHandler(w http.ResponseWriter, r *http.Request) {
	time.Sleep(time.Duration(rand.Intn(2000)) * time.Millisecond)
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"user": "John Doe", "email": "john@example.com"})
}

func simulateActiveUsers() {
	for {
		activeUsers.Set(float64(rand.Intn(100)))
		time.Sleep(5 * time.Second)
	}
}

func main() {
	go simulateActiveUsers()

	http.HandleFunc("/v1", metricsMiddleware(homeHandler))
	http.HandleFunc("/v1/user", metricsMiddleware(userHandler))
	http.Handle("/metrics", promhttp.Handler())

	log.Println("Server starting on port 8085...")
	log.Fatal(http.ListenAndServe(":8085", nil))
}
