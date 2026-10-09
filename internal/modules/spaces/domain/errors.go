package domain

import "errors"

var (
	ErrSpaceNotFound     = errors.New("space not found")
	ErrSlugAlreadyExists = errors.New("space slug already exists")
	ErrForbidden         = errors.New("space access forbidden")
	ErrInvalidInput      = errors.New("invalid space input")
	ErrSpaceArchived     = errors.New("space archived")
)
