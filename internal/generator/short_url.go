package generator

import (
	"crypto/rand"
	"fmt"
	"strings"
)

func GenerateShortUrl() (string, error) {
	const (
		length  = 10
		charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789_"
	)

	b := make([]byte, length)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("error generate bytes: %w", err)
	}

	var sb strings.Builder
	sb.Grow(length)

	for i := 0; i < length; i++ {
		idx := int(b[i]) % len(charset)
		sb.WriteByte(charset[idx])
	}

	return sb.String(), nil
}
