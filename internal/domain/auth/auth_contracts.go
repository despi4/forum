package auth

import (
	"context"
	"database/sql"
	"time"

	user "01.tomorrow-school.ai/git/amadiuly/forum/internal/domain/user"

	"github.com/google/uuid"
)

type SessionRepository interface {
	Create(ctx context.Context, session *Session, tx *sql.Tx) error
	CreateUserWithSession(ctx context.Context, sessionDuration time.Duration, newUser *user.User, userRepo user.UserRepository) (Session, error)
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
	ChangePassword(ctx context.Context, userID uuid.UUID, oldPassword, newPassword string) (Session, error)
}
