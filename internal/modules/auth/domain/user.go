package domain

import "time"

type UserID int64

type UserStatus string

const (
	UserStatusActive  UserStatus = "active"
	UserStatusBlocked UserStatus = "blocked"
	UserStatusDeleted UserStatus = "deleted"
)

type User struct {
	ID        UserID
	Username  string
	Email     string
	Status    UserStatus
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (u User) IsActive() bool {
	return u.Status == UserStatusActive
}
