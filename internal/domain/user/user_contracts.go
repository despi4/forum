package user

import (
	"context"

	"github.com/google/uuid"
)

type UserRepository interface {
	Create(ctx context.Context, user *User) (User, error)
	GetByID(ctx context.Context, id uuid.UUID) (*User, error)
	GetByEmail(ctx context.Context, email string) (*User, error)
	GetByUsername(ctx context.Context, username string) (*User, error)
	Update(ctx context.Context, user UserUpdate, userID uuid.UUID) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, params UserFilter) ([]User, error)
}

type UserService interface {
	GetMe(ctx context.Context, id uuid.UUID) (*User, error)
	UpdateMe(ctx context.Context, id uuid.UUID, updatedUser UserUpdate) error
	DeleteUser(ctx context.Context, actor, target uuid.UUID) error
	ListUsers(ctx context.Context, params UserFilter) ([]User, error)

	// admin method
	GetUserByID(ctx context.Context, id uuid.UUID) (*User, error)
	ChangeRole(ctx context.Context, actor, target uuid.UUID, role Role) error
}
