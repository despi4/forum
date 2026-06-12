package auth

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type SessionRepository interface {
	Create(ctx context.Context, session *Session) error
	GetByID(ctx context.Context, sessionID uuid.UUID) (*Session, error)
	UpdateLastSeen(ctx context.Context, sessionID uuid.UUID, lastSeenAt time.Time) error
	DeleteByID(ctx context.Context, sessionID uuid.UUID) error
	DeleteAllUserSessions(ctx context.Context, userID uuid.UUID) error
	DeleteExpiredSessions(ctx context.Context, expiresAt, lastSeenAt time.Time) error
}

type AuthService interface {
	Register(ctx context.Context, userInput *UserInput) (Session, error)
	Login(ctx context.Context, userInput *UserInput) (Session, error)
	Logout(ctx context.Context, sessionID uuid.UUID) error
	ValidateSession(ctx context.Context, sessionID uuid.UUID) (Session, error)
	ChangePassword(ctx context.Context, userID uuid.UUID, oldPassword, newPassword string) error
}
