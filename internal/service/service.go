package service

import (
	"context"
	"fmt"
	"log"

	"labago/internal/metrics"
	"labago/internal/models"
	"labago/internal/repository"
)

// Publisher — интерфейс для публикации событий (реализуется kafka.Producer).
type Publisher interface {
	Publish(ctx context.Context, key string, value any) error
	Topic() string
}

// Service — интерфейс бизнес-логики.
type Service interface {
	RegisterUser(ctx context.Context, login, password string) (*models.UserData, error)
	GetUserByLogin(ctx context.Context, login string) (*models.UserData, error)
	CreateSession(ctx context.Context, sessionID, userID string) error
	CreateBooking(ctx context.Context, userID string) (*models.Booking, error)
	ListBookingsByUserID(ctx context.Context, userID string) ([]models.Booking, error)
	UpdateBookingStatus(ctx context.Context, bookingID, status string) error
	SaveMessage(ctx context.Context, text string) (*models.Message, error)
}

type svc struct {
	repo      repository.Repository
	publisher Publisher // может быть nil (тесты без Kafka)
}

func New(repo repository.Repository, publisher Publisher) Service {
	return &svc{repo: repo, publisher: publisher}
}

func (s *svc) RegisterUser(ctx context.Context, login, password string) (*models.UserData, error) {
	existing, err := s.repo.GetUserByLogin(ctx, login)
	if err != nil {
		return nil, fmt.Errorf("failed to check existing user: %w", err)
	}
	if existing != nil {
		return nil, fmt.Errorf("user with login %s already exists", login)
	}
	return s.repo.RegisterUser(ctx, login, password)
}

func (s *svc) GetUserByLogin(ctx context.Context, login string) (*models.UserData, error) {
	return s.repo.GetUserByLogin(ctx, login)
}

func (s *svc) CreateSession(ctx context.Context, sessionID, userID string) error {
	return s.repo.CreateSession(ctx, sessionID, userID)
}

// CreateBooking создаёт бронирование и публикует событие в Kafka.
func (s *svc) CreateBooking(ctx context.Context, userID string) (*models.Booking, error) {
	booking, err := s.repo.CreateBooking(ctx, userID)
	if err != nil {
		return nil, err
	}
	if s.publisher != nil {
		event := models.BookingEvent{BookingID: booking.ID, UserID: userID}
		if err := s.publisher.Publish(ctx, booking.ID, event); err != nil {
			log.Printf("kafka publish booking.created failed: %v", err)
		} else {
			metrics.KafkaMessagesProduced.WithLabelValues(s.publisher.Topic()).Inc()
		}
	}
	return booking, nil
}

func (s *svc) ListBookingsByUserID(ctx context.Context, userID string) ([]models.Booking, error) {
	return s.repo.ListBookingsByUserID(ctx, userID)
}

func (s *svc) UpdateBookingStatus(ctx context.Context, bookingID, status string) error {
	return s.repo.UpdateBookingStatus(ctx, bookingID, status)
}

func (s *svc) SaveMessage(ctx context.Context, text string) (*models.Message, error) {
	return s.repo.SaveMessage(ctx, text)
}
