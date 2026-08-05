package domain

import "errors"

var (
	ErrInvalidInput   = errors.New("invalid requests input")
	ErrSpaceNotFound  = errors.New("space not found")
	ErrSpaceForbidden = errors.New("space access forbidden")
	ErrSpaceArchived  = errors.New("space archived")
)
