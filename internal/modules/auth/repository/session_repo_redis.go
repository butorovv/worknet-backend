package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/butorovv/worknet-backend/internal/modules/auth/domain"
	"github.com/redis/go-redis/v9"
)

type SessionRepoRedis struct {
	client *redis.Client
	prefix string
}

type redisSessionPayload struct {
	UserID    int64     `json:"user_id"`
	CreatedAt time.Time `json:"created_at"`
	ExpiresAt time.Time `json:"expires_at"`
}

func NewSessionRepoRedis(client *redis.Client, prefix string) *SessionRepoRedis {
	if prefix == "" {
		prefix = "auth:session:"
	}
	return &SessionRepoRedis{
		client: client,
		prefix: prefix,
	}
}

func (s *SessionRepoRedis) Create(ctx context.Context, session domain.Session) error {
	ttl := time.Until(session.ExpiresAt)
	if ttl <= 0 {
		return domain.ErrInvalidInput
	}

	payload, err := json.Marshal(redisSessionPayload{
		UserID:    int64(session.UserID),
		CreatedAt: session.CreatedAt,
		ExpiresAt: session.ExpiresAt,
	})
	if err != nil {
		return fmt.Errorf("marshal redis session: %w", err)
	}

	if err := s.client.Set(ctx, s.key(session.ID), payload, ttl).Err(); err != nil {
		return fmt.Errorf("store redis session: %w", err)
	}
	return nil
}

func (s *SessionRepoRedis) Get(ctx context.Context, sessionID string) (domain.Session, error) {
	value, err := s.client.Get(ctx, s.key(sessionID)).Bytes()
	if err != nil {
		if err == redis.Nil {
			return domain.Session{}, domain.ErrSessionNotFound
		}
		return domain.Session{}, fmt.Errorf("get redis session: %w", err)
	}

	var payload redisSessionPayload
	if err := json.Unmarshal(value, &payload); err != nil {
		return domain.Session{}, fmt.Errorf("unmarshal redis session: %w", err)
	}

	return domain.Session{
		ID:        sessionID,
		UserID:    domain.UserID(payload.UserID),
		CreatedAt: payload.CreatedAt,
		ExpiresAt: payload.ExpiresAt,
	}, nil
}

func (s *SessionRepoRedis) Delete(ctx context.Context, sessionID string) error {
	deleted, err := s.client.Del(ctx, s.key(sessionID)).Result()
	if err != nil {
		return fmt.Errorf("delete redis session: %w", err)
	}
	if deleted == 0 {
		return domain.ErrSessionNotFound
	}
	return nil
}

func (s *SessionRepoRedis) key(sessionID string) string {
	return s.prefix + sessionID
}
