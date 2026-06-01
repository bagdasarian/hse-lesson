package main

import (
	"context"
	_ "embed"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	appkafka "labago/internal/kafka"
	"labago/internal/metrics"
	"labago/internal/models"

	"labago/internal/handler"
	"labago/internal/repository"
	"labago/internal/service"
)

//go:embed static/index.html
var indexHTML []byte

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	dsn := os.Getenv("DB_DSN")
	if dsn == "" {
		dsn = "postgres://postgres:postgres@localhost:5433/labago?sslmode=disable"
	}
	port := os.Getenv("APP_PORT")
	if port == "" {
		port = "8080"
	}
	kafkaBrokers := strings.Split(envOr("KAFKA_BROKERS", "localhost:9092"), ",")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// --- БД ---
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		log.Fatalf("Failed to create pool: %v", err)
	}
	defer pool.Close()
	if err := pool.Ping(ctx); err != nil {
		log.Fatalf("Failed to connect to DB: %v", err)
	}
	log.Println("Connected to DB")

	// --- Kafka producer (bookings.new) ---
	producer := appkafka.NewProducer(kafkaBrokers, "bookings.new")
	defer producer.Close()

	// --- Слои приложения ---
	repo := repository.New(pool)
	svc := service.New(repo, producer)
	h := handler.New(svc)

	// --- Kafka consumer (bookings.status) — меняет статус бронирований в БД ---
	go consumeBookingStatuses(ctx, kafkaBrokers, svc)

	// --- HTTP маршруты ---
	mux := http.NewServeMux()

	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write(indexHTML)
	})
	mux.Handle("GET /metrics", promhttp.Handler())    // Prometheus
	mux.HandleFunc("GET /test", h.Test)               // ЛР1
	mux.HandleFunc("POST /dbtest", h.DBTest)          // ЛР2
	mux.HandleFunc("POST /auth/register", h.RegisterUser) // ЛР3
	mux.HandleFunc("POST /auth/login", h.LoginUser)       // ЛР3
	mux.HandleFunc("POST /bookings/create", handler.AuthMiddleware(h.CreateBooking)) // ЛР4+5
	mux.HandleFunc("GET /bookings/list", handler.AuthMiddleware(h.ListBookings))     // ЛР4+5

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: metrics.Middleware(mux),
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		log.Printf("Server started on :%s", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	<-quit
	log.Println("Shutting down...")
	cancel() // останавливаем kafka consumer

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("Forced shutdown: %v", err)
	}
	log.Println("Server stopped")
}

// consumeBookingStatuses слушает топик bookings.status и обновляет статусы в БД.
func consumeBookingStatuses(ctx context.Context, brokers []string, svc service.Service) {
	consumer := appkafka.NewConsumer(brokers, "bookings.status")
	defer consumer.Close()
	log.Println("Kafka consumer started: bookings.status")

	for {
		var event models.BookingEvent
		if err := consumer.Read(ctx, &event); err != nil {
			if ctx.Err() != nil {
				return // graceful shutdown
			}
			log.Printf("kafka consumer error: %v", err)
			time.Sleep(time.Second)
			continue
		}
		metrics.KafkaMessagesConsumed.WithLabelValues("bookings.status").Inc()
		if err := svc.UpdateBookingStatus(ctx, event.BookingID, event.Status); err != nil {
			log.Printf("UpdateBookingStatus(%s, %s): %v", event.BookingID, event.Status, err)
		} else {
			log.Printf("Booking %s → %s", event.BookingID, event.Status)
		}
	}
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
