package repository

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"

	"labago/internal/models"
)

//go:generate mockgen -source=repository.go -destination=mock_repository.go -package=repository

// Repository — интерфейс слоя работы с БД для пользователей и сессий.
type Repository interface {
	RegisterUser(ctx context.Context, login, password string) (*models.UserData, error)
	GetUserByLogin(ctx context.Context, login string) (*models.UserData, error)
	CreateSession(ctx context.Context, sessionID, userID string) error
	GetSessionByUserID(ctx context.Context, userID string) (*models.Session, error)
	UpdateSessionExpiry(ctx context.Context, sessionID string) error
}

type repo struct {
	db *pgxpool.Pool
}

func New(db *pgxpool.Pool) Repository {
	return &repo{db: db}
}

// RegisterUser регистрирует пользователя: хэширует пароль, вставляет в БД.
func (r *repo) RegisterUser(ctx context.Context, login, password string) (*models.UserData, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	var user models.UserData
	err = r.db.QueryRow(ctx,
		`INSERT INTO users (login, password, created_at, updated_at) VALUES ($1, $2, NOW(), NOW())
		 RETURNING id, login, password, created_at`,
		login, string(hashedPassword),
	).Scan(&user.ID, &user.Login, &user.Password, &user.CreatedAt)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

// GetUserByLogin находит пользователя по логину.
func (r *repo) GetUserByLogin(ctx context.Context, login string) (*models.UserData, error) {
	user := &models.UserData{}
	err := r.db.QueryRow(ctx,
		`SELECT id, login, password, created_at FROM users WHERE login = $1`,
		login,
	).Scan(&user.ID, &user.Login, &user.Password, &user.CreatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil // пользователь не найден
		}
		return nil, err
	}
	return user, nil
}

// CreateSession создаёт сессию с TTL 24 часа.
func (r *repo) CreateSession(ctx context.Context, sessionID, userID string) error {
	_, err := r.db.Exec(ctx,
		`INSERT INTO sessions (session_id, user_id, created_at, expires_at) VALUES ($1, $2, NOW(), NOW() + INTERVAL '24 hours')`,
		sessionID, userID,
	)
	return err
}

// GetSessionByUserID возвращает сессию по ID пользователя.
func (r *repo) GetSessionByUserID(ctx context.Context, userID string) (*models.Session, error) {
	session := &models.Session{}
	err := r.db.QueryRow(ctx,
		`SELECT session_id, user_id, created_at, expires_at FROM sessions WHERE user_id = $1`,
		userID,
	).Scan(&session.SessionID, &session.UserID, &session.CreatedAt, &session.ExpiresAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return session, nil
}

// UpdateSessionExpiry продлевает время жизни сессии.
func (r *repo) UpdateSessionExpiry(ctx context.Context, sessionID string) error {
	_, err := r.db.Exec(ctx,
		`UPDATE sessions SET expires_at = NOW() + INTERVAL '24 hours' WHERE session_id = $1`,
		sessionID,
	)
	return err
}

// Message — вспомогательная структура для обратной совместимости (оставлена для примера).
type Message struct {
	ID        int       `json:"id"`
	Text      string    `json:"text"`
	CreatedAt time.Time `json:"created_at"`
}