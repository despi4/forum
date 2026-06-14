package user

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

type PasswordHash string

type Role string

type Visibility string

const (
	RoleAdmin Role = "admin"
	RoleUser  Role = "user"

	VisibilityPrivate Visibility = "private"
	VisibilityPublic  Visibility = "public"
)

type User struct {
	ID           uuid.UUID
	Username     string
	Role         Role
	Visibility   Visibility
	Email        string
	PasswordHash PasswordHash
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type UserUpdate struct {
	Username     *string
	Email        *string
	Role         *Role
	Visibility   *Visibility
}

type UserFilter struct {
	Role   *Role
	Search *string
	Limit  int // кол-во записей на странице
	Offset int // сколько записей пропустить
}

var (
	ErrUserNotFound       = errors.New("user not found")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUserAlreadyExists  = errors.New("user already exists")
	ErrUnauthorized       = errors.New("unauthorized")
	ErrForbidden          = errors.New("forbidden")
	ErrInvalidArgument = errors.New("invalid argument")
)
