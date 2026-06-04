package token

import (
	"context"
	"crypto/rand"
	"encoding/base64"
)

type RandomGenerator struct {
	bytes int
}

func NewRandomGenerator(bytes int) *RandomGenerator {
	if bytes <= 0 {
		bytes = 32
	}
	return &RandomGenerator{bytes: bytes}
}

func (g *RandomGenerator) Generate(ctx context.Context) (string, error) {
	_ = ctx
	buffer := make([]byte, g.bytes)
	if _, err := rand.Read(buffer); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(buffer), nil
}
