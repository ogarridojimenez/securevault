package model

import "errors"

var (
	ErrNotFound       = errors.New("resource not found")
	ErrAlreadyExists  = errors.New("resource already exists")
	ErrUnauthorized   = errors.New("unauthorized")
	ErrForbidden      = errors.New("forbidden")
	ErrInvalidInput   = errors.New("invalid input")
	ErrTokenExpired   = errors.New("token expired")
	ErrRateLimited    = errors.New("rate limit exceeded")
	ErrVaultNotEmpty  = errors.New("vault is not empty")
)
