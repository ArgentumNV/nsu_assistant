package user

import "time"

type User struct {
	ID         int64     `json:"id"`
	TelegramID int64     `json:"telegram_id"`
	GroupID    *string   `json:"group_id"`
	CreatedAt  time.Time `json:"created_at"`
}
