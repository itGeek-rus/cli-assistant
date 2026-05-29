package domain

import "errors"

var (
	ErrNotFound     = errors.New("not found")
	ErrUnavailable  = errors.New("unavailable")
	ErrUnauthorized = errors.New("unauthorized")
	ErrInvalidInput = errors.New("invalid input")
)
