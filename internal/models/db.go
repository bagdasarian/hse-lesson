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