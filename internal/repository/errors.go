package repository

import "errors"

var (
	ErrNotFoundShort   = errors.New("short url not found")
	ErrNotFoundOrigin  = errors.New("origin url not found")
	ErrDuplicateShort  = errors.New("short url already exists")
	ErrDuplicateOrigin = errors.New("origin url already exists")
)
