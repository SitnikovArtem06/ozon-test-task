package service

import "errors"

var (
	ErrNotFound         error = errors.New("not found any origin url")
	ErrGenerateShortUrl       = errors.New("failed generated short url")
	ErrGenerateTimeout        = errors.New("short url generation timeout")
	ErrPersistLink            = errors.New("failed to persist link")
	ErrInvalidShortURL        = errors.New("short url must be 10 chars")
)
