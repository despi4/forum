package comment

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

type Comment struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	PostID    uuid.UUID
	Content   string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type CreateComment struct {
	UserID  uuid.UUID
	PostID  uuid.UUID
	Content string
}

type CommentSort string

const (
	CommentSortCreatedDesc    CommentSort = "created_desc"
	CommentSortCreatedAsc     CommentSort = "created_asc"
	CommentSortReactionsCount CommentSort = "reactions_count"
)

type CommentFilter struct {
	Offset      int
	Limit       int
	CommentSort *CommentSort
}

var ErrCommentNotFound error = errors.New("comment not found")
var ErrInvalidArgument error = errors.New("invalid argument")
var ErrForbidden error = errors.New("forbidden")
