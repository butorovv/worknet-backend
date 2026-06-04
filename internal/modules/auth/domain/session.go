package domain

import "time"

type Session struct {
	ID        string
	UserID    UserID
	CreatedAt time.Time
	ExpiresAt time.Time
}
