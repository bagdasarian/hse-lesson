package repository

import (
	"context"

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
	CreateBooking(ctx context.Context, userID string) (*models.Booking, error)
	ListBookingsByUserID(ctx context.Context, userID string) ([]models.Booking, error)
	UpdateBookingStatus(ctx context.Context, bookingID, status string) error
	SaveMessage(ctx context.Context, text string) (*models.Message, error)
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

// CreateBooking создаёт новое бронирование для пользователя.
func (r *repo) CreateBooking(ctx context.Context, userID string) (*models.Booking, error) {
	var b models.Booking
	err := r.db.QueryRow(ctx,
		`INSERT INTO bookings (user_id, status, created_at, updated_at)
		 VALUES ($1, 'new', NOW(), NOW())
		 RETURNING id, user_id, status, created_at, updated_at`,
		userID,
	).Scan(&b.ID, &b.UserID, &b.Status, &b.CreatedAt, &b.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &b, nil
}

// ListBookingsByUserID возвращает все бронирования пользователя.
func (r *repo) ListBookingsByUserID(ctx context.Context, userID string) ([]models.Booking, error) {
	rows, err := r.db.Query(ctx,
		`SELECT id, user_id, status, created_at, updated_at FROM bookings WHERE user_id = $1 ORDER BY created_at DESC`,
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var bookings []models.Booking
	for rows.Next() {
		var b models.Booking
		if err := rows.Scan(&b.ID, &b.UserID, &b.Status, &b.CreatedAt, &b.UpdatedAt); err != nil {
			return nil, err
		}
		bookings = append(bookings, b)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return bookings, nil
}

// UpdateBookingStatus меняет статус бронирования и обновляет updated_at.
func (r *repo) UpdateBookingStatus(ctx context.Context, bookingID, status string) error {
	_, err := r.db.Exec(ctx,
		`UPDATE bookings SET status = $1, updated_at = NOW() WHERE id = $2`,
		status, bookingID,
	)
	return err
}

// SaveMessage сохраняет произвольный текст в таблицу messages.
func (r *repo) SaveMessage(ctx context.Context, text string) (*models.Message, error) {
	var m models.Message
	err := r.db.QueryRow(ctx,
		`INSERT INTO messages (text) VALUES ($1) RETURNING id, text, created_at`,
		text,
	).Scan(&m.ID, &m.Text, &m.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &m, nil
}
