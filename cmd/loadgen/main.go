package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"
)

var (
	baseURL    string
	httpClient = &http.Client{Timeout: 5 * time.Second}
)

func main() {
	_ = godotenv.Load()

	baseURL = envOr("TARGET_URL", "http://app:8080")

	// Ждём запуска основного сервиса
	log.Printf("Load generator: target=%s, waiting 10s for server...", baseURL)
	time.Sleep(10 * time.Second)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	ticker := time.NewTicker(200 * time.Millisecond)
	defer ticker.Stop()

	log.Println("Load generator started")

	for {
		select {
		case <-quit:
			log.Println("Load generator stopped")
			return
		case <-ticker.C:
			go runScenario()
		}
	}
}

// runScenario выполняет один полный сценарий: регистрация → логин → бронирование(я) → список.
func runScenario() {
	login := fmt.Sprintf("load_%d_%04d", time.Now().UnixNano()/1e6, rand.Intn(9999))
	password := "loadtest123"

	if err := register(login, password); err != nil {
		log.Printf("register failed: %v", err)
		return
	}

	token, err := loginUser(login, password)
	if err != nil {
		log.Printf("login failed: %v", err)
		return
	}

	count := rand.Intn(3) + 1
	for i := 0; i < count; i++ {
		if err := createBooking(token); err != nil {
			log.Printf("createBooking failed: %v", err)
		}
	}

	if err := listBookings(token); err != nil {
		log.Printf("listBookings failed: %v", err)
	}
}

func register(login, password string) error {
	body, _ := json.Marshal(map[string]string{"login": login, "password": password})
	resp, err := httpClient.Post(baseURL+"/auth/register", "application/json", bytes.NewReader(body))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("register: status %d", resp.StatusCode)
	}
	return nil
}

func loginUser(login, password string) (string, error) {
	body, _ := json.Marshal(map[string]string{"login": login, "password": password})
	resp, err := httpClient.Post(baseURL+"/auth/login", "application/json", bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	var result map[string]string
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}
	token, ok := result["token"]
	if !ok {
		return "", fmt.Errorf("no token in response")
	}
	return token, nil
}

func createBooking(token string) error {
	req, _ := http.NewRequest(http.MethodPost, baseURL+"/bookings/create", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("createBooking: status %d", resp.StatusCode)
	}
	return nil
}

func listBookings(token string) error {
	req, _ := http.NewRequest(http.MethodGet, baseURL+"/bookings/list", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("listBookings: status %d", resp.StatusCode)
	}
	return nil
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
