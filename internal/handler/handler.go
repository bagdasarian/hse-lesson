package handler

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"labago/internal/jwt"
	"labago/internal/metrics"
	"labago/internal/service"
)

type Handler struct {
	svc service.Service
}

func New(svc service.Service) *Handler {
	return &Handler{svc: svc}
}

// Test — тестовый эндпоинт ЛР1. GET /test
func (h *Handler) Test(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Hello!"))
}

// DBTest — тестовый эндпоинт ЛР2. POST /dbtest
// Принимает {"text":"..."} и сохраняет строку в БД.
func (h *Handler) DBTest(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Text string `json:"text"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Text == "" {
		http.Error(w, "field 'text' is required", http.StatusBadRequest)
		return
	}

	msg, err := h.svc.SaveMessage(r.Context(), req.Text)
	if err != nil {
		log.Printf("DBTest SaveMessage failed: %v", err)
		http.Error(w, "failed to save message", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(msg)
}

// RegisterUser — хэндлер регистрации. POST /auth/register
func (h *Handler) RegisterUser(w http.ResponseWriter, r *http.Request) {
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

// LoginUser — хэндлер авторизации. POST /auth/login, возвращает JWT-токен.
func (h *Handler) LoginUser(w http.ResponseWriter, r *http.Request) {
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

	user, err := h.svc.GetUserByLogin(r.Context(), req.Login)
	if err != nil || user == nil {
		http.Error(w, "invalid login or password", http.StatusUnauthorized)
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		http.Error(w, "invalid login or password", http.StatusUnauthorized)
		return
	}

	token, err := jwt.GenerateToken(user.ID, user.Login)
	if err != nil {
		log.Printf("GenerateToken failed: %v", err)
		http.Error(w, "failed to generate token", http.StatusInternalServerError)
		return
	}

	sessionID := uuid.New().String()
	if err := h.svc.CreateSession(r.Context(), sessionID, user.ID); err != nil {
		log.Printf("CreateSession failed: %v", err)
		http.Error(w, "failed to create session", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"token": token})
}

// CreateBooking — хендлер создания бронирования. POST /bookings/create
// Требует AuthMiddleware: берёт userID из контекста.
func (h *Handler) CreateBooking(w http.ResponseWriter, r *http.Request) {
	userID, _ := r.Context().Value(UserIDKey).(string)

	booking, err := h.svc.CreateBooking(r.Context(), userID)
	if err != nil {
		log.Printf("CreateBooking failed: %v", err)
		http.Error(w, "failed to create booking", http.StatusInternalServerError)
		return
	}

	metrics.BookingsCreatedTotal.Inc()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(booking)
}

// ListBookings — хендлер получения списка бронирований. GET /bookings/list
// Требует AuthMiddleware: берёт userID из контекста.
func (h *Handler) ListBookings(w http.ResponseWriter, r *http.Request) {
	userID, _ := r.Context().Value(UserIDKey).(string)

	bookings, err := h.svc.ListBookingsByUserID(r.Context(), userID)
	if err != nil {
		log.Printf("ListBookings failed: %v", err)
		http.Error(w, "failed to list bookings", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(bookings)
}
