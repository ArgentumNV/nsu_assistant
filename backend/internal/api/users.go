package api

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"

	"github.com/ArgentumNV/nsu_assistant/backend/internal/user"
	"github.com/jackc/pgx/v5"
)

type UserHandler struct {
	repo *user.Repository
}

// UserHandler construstor
func NewUserHandler(repo *user.Repository) *UserHandler {
	return &UserHandler{
		repo: repo,
	}
}

type CreateUserRequest struct {
	TelegramID int64   `json:"telegram_id"`
	GroupID    *string `json:"group_id"`
}

func (u *UserHandler) CreateUser(
	w http.ResponseWriter,
	r *http.Request,
) {
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", "POST")
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var request CreateUserRequest

	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		log.Printf("CreateUser failed: %v", err)
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if request.TelegramID <= 0 {
		http.Error(w, "Invalid telegram_id", http.StatusBadRequest)
		return
	}

	createdUser, err := u.repo.CreateUser(
		r.Context(),
		request.TelegramID,
		request.GroupID,
	)
	if err != nil {
		log.Printf("CreateUser failed: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	if err := json.NewEncoder(w).Encode(createdUser); err != nil {
		log.Printf("JSON encoding error: %v", err)
	}
}

func (u *UserHandler) GetUserByTelegramID(
	w http.ResponseWriter,
	r *http.Request,
) {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", "GET")
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	telegramIDStr := r.PathValue("telegram_id")
	telegramID, err := strconv.ParseInt(telegramIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid telegram_id", http.StatusBadRequest)
		return
	}
	if telegramID <= 0 {
		http.Error(w, "telegram_id must be > 0", http.StatusBadRequest)
		return
	}

	user, err := u.repo.GetUserByTelegramID(r.Context(), telegramID)

	// User not found
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			http.Error(w, "User not found", http.StatusNotFound)
			return
		}
		log.Printf("GetUserByTelegramID failed: %v", err)
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	if err = json.NewEncoder(w).Encode(user); err != nil {
		log.Println("JSON encoding error:", err)
	}

}

func (u *UserHandler) DeleteUser(
	w http.ResponseWriter,
	r *http.Request,
) {
	if r.Method != http.MethodDelete {
		w.Header().Set("Allow", "DELETE")
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	telegramIDStr := r.PathValue("telegram_id")
	telegramID, err := strconv.ParseInt(telegramIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid telegram_id", http.StatusBadRequest)
		return
	}
	if telegramID <= 0 {
		http.Error(w, "telegram_id must be > 0", http.StatusBadRequest)
		return
	}

	err = u.repo.DeleteUser(r.Context(), telegramID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			http.Error(w, "User not found", http.StatusNotFound)
			return
		}
		log.Printf("DeleteUser failed: %v", err)
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
