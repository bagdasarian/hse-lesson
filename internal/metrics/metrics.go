package metrics

import (
	"net/http"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	HttpRequestsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "http_requests_total",
		Help: "Общее количество HTTP-запросов",
	}, []string{"method", "path", "status"})

	HttpRequestDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "http_request_duration_seconds",
		Help:    "Время обработки HTTP-запроса",
		Buckets: prometheus.DefBuckets,
	}, []string{"method", "path"})

	BookingsCreatedTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "bookings_created_total",
		Help: "Общее количество созданных бронирований",
	})

	KafkaMessagesProduced = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "kafka_messages_produced_total",
		Help: "Количество опубликованных сообщений Kafka",
	}, []string{"topic"})

	KafkaMessagesConsumed = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "kafka_messages_consumed_total",
		Help: "Количество потреблённых сообщений Kafka",
	}, []string{"topic"})
)

// statusRecorder перехватывает HTTP-статус для последующей записи в метрику.
type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(code int) {
	r.status = code
	r.ResponseWriter.WriteHeader(code)
}

// Middleware оборачивает mux и записывает HTTP-метрики для каждого запроса.
func Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(rec, r)
		duration := time.Since(start).Seconds()
		status := strconv.Itoa(rec.status)
		HttpRequestsTotal.WithLabelValues(r.Method, r.URL.Path, status).Inc()
		HttpRequestDuration.WithLabelValues(r.Method, r.URL.Path).Observe(duration)
	})
}
