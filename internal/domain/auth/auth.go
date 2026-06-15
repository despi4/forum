package auth

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

type Session struct {
	ID         uuid.UUID
	UserID     uuid.UUID
	CreatedAt  time.Time
	LastSeenAt time.Time
	ExpiresAt  time.Time
}

type UserInput struct {
	Username *string
	Email    *string
	Password string
} 

var (
	ErrSessionNotFound      = errors.New("session not found")
	ErrSessionAlreadyExists = errors.New("user already exists")
	ErrSessionExpired = errors.New("session expired")
)
