package service

import (
	"context"
	"fmt"

	"labago/internal/models"
	"labago/internal/repository"
)

// Service — интерфейс бизнес-логики для работы с пользователями.
type Service interface {
	RegisterUser(ctx context.Context, login, password string) (*models.UserData, error)
	GetUserByLogin(ctx context.Context, login string) (*models.UserData, error)
	CreateSession(ctx context.Context, sessionID, userID string) error
}

type svc struct {
	repo repository.Repository
}

func New(repo repository.Repository) Service {
	return &svc{repo: repo}
}

// RegisterUser проверяет дубликат и регистрирует пользователя.
func (s *svc) RegisterUser(ctx context.Context, login, password string) (*models.UserData, error) {
	// Проверка дубликата логина
	existing, err := s.repo.GetUserByLogin(ctx, login)
	if err != nil {
		return nil, fmt.Errorf("failed to check existing user: %w", err)
	}
	if existing != nil {
		return nil, fmt.Errorf("user with login %s already exists", login)
	}

	return s.repo.RegisterUser(ctx, login, password)
}

// GetUserByLogin передаёт вызов напрямую в Repository.
func (s *svc) GetUserByLogin(ctx context.Context, login string) (*models.UserData, error) {
	return s.repo.GetUserByLogin(ctx, login)
}

// CreateSession передаёт вызов напрямую в Repository.
func (s *svc) CreateSession(ctx context.Context, sessionID, userID string) error {
	return s.repo.CreateSession(ctx, sessionID, userID)
}