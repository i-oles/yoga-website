package token

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
)

type Generator struct{}

func NewGenerator() Generator {
	return Generator{}
}

func (g Generator) Generate(length int) (string, error) {
	b := make([]byte, length)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("error generating token: %w", err)
	}

	return base64.URLEncoding.EncodeToString(b), nil
}
