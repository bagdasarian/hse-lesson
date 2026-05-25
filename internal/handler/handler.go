package handler

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"labago/internal/jwt"
	"labago/internal/service"
)

type Handler struct {
	svc service.Service
}

func New(svc service.Service) *Handler {
	return &Handler{svc: svc}
}

// RegisterUser — хэндлер регистрации. POST /auth/register
func (h *Handler) RegisterUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Login    string `json:"login"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.Login == "" || req.Password == "" {
		http.Error(w, "login and password are required", http.StatusBadRequest)
		return
	}

	user, err := h.svc.RegisterUser(r.Context(), req.Login, req.Password)
	if err != nil {
		log.Printf("RegisterUser failed: %v", err)
		http.Error(w, err.Error(), http.StatusConflict)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(user)
}

// LoginUser — хэндлер авторизации. POST /auth/login, возвращает JWT-токен
func (h *Handler) LoginUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Login    string `json:"login"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.Login == "" || req.Password == "" {
		http.Error(w, "login and password are required", http.StatusBadRequest)
		return
	}

	// 1. Находим пользователя по логину
	user, err := h.svc.GetUserByLogin(r.Context(), req.Login)
	if err != nil || user == nil {
		// Безопасность: не различаем "пользователь не найден" и "неверный пароль"
		http.Error(w, "invalid login or password", http.StatusUnauthorized)
		return
	}

	// 2. Проверяем пароль через bcrypt
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		http.Error(w, "invalid login or password", http.StatusUnauthorized)
		return
	}

	// 3. Генерируем JWT-токен
	token, err := jwt.GenerateToken(user.ID, user.Login)
	if err != nil {
		log.Printf("GenerateToken failed: %v", err)
		http.Error(w, "failed to generate token", http.StatusInternalServerError)
		return
	}

	// 4. Создаём сессию в БД
	sessionID := uuid.New().String()
	if err := h.svc.CreateSession(r.Context(), sessionID, user.ID); err != nil {
		log.Printf("CreateSession failed: %v", err)
		http.Error(w, "failed to create session", http.StatusInternalServerError)
		return
	}

	// 5. Возвращаем токен клиенту
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"token": token})
}