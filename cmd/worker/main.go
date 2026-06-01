package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/joho/godotenv"

	appkafka "labago/internal/kafka"
	"labago/internal/models"
)

// Worker имитирует обработку бронирований:
//  1. Получает событие из топика bookings.new
//  2. Через 3 сек публикует статус "processing"
//  3. Через ещё 5 сек публикует статус "done"
func main() {
	_ = godotenv.Load()

	brokers := strings.Split(envOr("KAFKA_BROKERS", "localhost:9092"), ",")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	consumer := appkafka.NewConsumer(brokers, "bookings.new")
	defer consumer.Close()

	producer := appkafka.NewProducer(brokers, "bookings.status")
	defer producer.Close()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-quit
		log.Println("Worker shutting down...")
		cancel()
	}()

	log.Println("Worker started, waiting for bookings...")

	for {
		var event models.BookingEvent
		if err := consumer.Read(ctx, &event); err != nil {
			if ctx.Err() != nil {
				log.Println("Worker stopped")
				return
			}
			log.Printf("consumer read error: %v", err)
			time.Sleep(time.Second)
			continue
		}

		log.Printf("Processing booking %s (user %s)", event.BookingID, event.UserID)

		// Горутина на каждый заказ — не блокируем основной цикл
		go processBooking(ctx, producer, event.BookingID)
	}
}

// processBooking имитирует конвейер обработки заказа.
func processBooking(ctx context.Context, producer *appkafka.Producer, bookingID string) {
	publish := func(status string) {
		event := models.BookingEvent{BookingID: bookingID, Status: status}
		if err := producer.Publish(ctx, bookingID, event); err != nil {
			log.Printf("publish %s for %s: %v", status, bookingID, err)
		} else {
			log.Printf("Booking %s → %s", bookingID, status)
		}
	}

	select {
	case <-time.After(3 * time.Second):
		publish("processing")
	case <-ctx.Done():
		return
	}

	select {
	case <-time.After(5 * time.Second):
		publish("done")
	case <-ctx.Done():
		return
	}
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
