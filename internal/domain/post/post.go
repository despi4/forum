package post

import (
	"errors"
	"time"

	"01.tomorrow-school.ai/git/amadiuly/forum/internal/domain/category"
	"github.com/google/uuid"
)

type Post struct {
	ID         uuid.UUID
	AuthorID   uuid.UUID
	CategoryID uuid.UUID
	Title      string
	Content    string
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

type PostSort string

const (
	PostSortCreatedDesc PostSort = "created_desc"
	PostSortCreatedAsc  PostSort = "created_asc"
)

type PostFilter struct {
	Search     *string
	CategoryID *uuid.UUID
	AuthorID   *uuid.UUID
	Offset     int
	Limit      int
	Sort       PostSort
}

type UpdatePost struct {
	Title      *string
	Content    *string
	CategoryID *uuid.UUID
}

type CreatePost struct {
	UserID   uuid.UUID
	Title    string
	Content  string
	Category category.Category
}

var ErrPostNotFound = errors.New("post not found")
