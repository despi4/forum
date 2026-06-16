package reaction

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

type ReactionType int

const (
	Like    ReactionType = 1
	Dislike ReactionType = -1
)

type PostReaction struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	PostID    uuid.UUID
	Type      ReactionType
	CreatedAt time.Time
}

type CommentReaction struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	CommentID uuid.UUID
	Type      ReactionType
	CreatedAt time.Time
}

type ReactionsCount struct {
	Likes    int
	Dislikes int
}

var (
	ErrReactionAlreadyExists error = errors.New("reaction already exists")
	ErrReactionNotFound      error = errors.New("reaction not found")
	ErrInvalidArgument       error = errors.New("invalid input")
	ErrForbidden error = errors.New("forbidden")
)
