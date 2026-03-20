package repository

import "errors"

var ErrNotFoundShort = errors.New("short url not found")
var ErrNotFoundOrigin = errors.New("origin url not found")
var ErrDuplicateShort = errors.New("short url already exists")
var ErrDuplicateOrigin = errors.New("origin url already exists")
