package domain

import "errors"

var (
	ErrInvalidInput       = errors.New("invalid input")
	ErrUserAlreadyExists  = errors.New("user already exists")
	ErrUserNotFound       = errors.New("user not found")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrSessionNotFound    = errors.New("session not found")
	ErrUserInactive       = errors.New("user inactive")
)
