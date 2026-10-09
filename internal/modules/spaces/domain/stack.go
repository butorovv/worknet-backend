package domain

import (
	"strings"
	"time"
)

type SpaceStackItem struct {
	ID        int64
	SpaceID   SpaceID
	Name      string
	CreatedAt time.Time
}

func NormalizeStack(items []string) ([]SpaceStackItem, error) {
	if len(items) > 12 {
		return nil, ErrInvalidInput
	}

	result := make([]SpaceStackItem, 0, len(items))
	seen := make(map[string]struct{}, len(items))
	for _, item := range items {
		name := strings.TrimSpace(item)
		if !validLength(name, 1, 32) {
			return nil, ErrInvalidInput
		}
		key := strings.ToLower(name)
		if _, exists := seen[key]; exists {
			return nil, ErrInvalidInput
		}
		seen[key] = struct{}{}
		result = append(result, SpaceStackItem{Name: name})
	}
	return result, nil
}
