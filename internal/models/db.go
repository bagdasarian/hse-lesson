package models

import "time"

// UserData — модель пользователя в БД. Отображает запись в таблице users.
// Password хранится уже хэшированным (bcrypt), никогда не в открытом виде.
type UserData struct {
	ID        string    `json:"id,omitempty" db:"id"`
	Login     string    `json:"login" db:"login" validate:"required"`
	Password  string    `json:"-" db:"password" validate:"required"`
	Email     string    `json:"email,omitempty" db:"email"`
	Phone     string    `json:"phone,omitempty" db:"phone"`
	IsActive  bool      `json:"is_active,omitempty" db:"is_active"`
	CreatedAt time.Time `json:"created_at,omitempty" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at,omitempty" db:"updated_at"`
}

// Session — модель сессии пользователя. Связывает JWT-токен с пользователем в БД.
type Session struct {
	SessionID string    `db:"session_id" json:"sessionID"`
	UserID    string    `db:"user_id" json:"userID"`
	ExpiresAt time.Time `db:"expires_at" json:"expiresAt"`
	CreatedAt time.Time `db:"created_at" json:"createdAt"`
}

// Booking — модель бронирования.
type Booking struct {
	ID        string    `json:"id" db:"id"`
	UserID    string    `json:"user_id" db:"user_id"`
	Status    string    `json:"status" db:"status"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// Message — модель записи из таблицы messages (используется для /dbtest).
type Message struct {
	ID        int       `json:"id" db:"id"`
	Text      string    `json:"text" db:"text"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// BookingEvent — событие, передаваемое через Kafka между сервисами.
type BookingEvent struct {
	BookingID string `json:"booking_id"`
	UserID    string `json:"user_id,omitempty"`
	Status    string `json:"status,omitempty"`
}