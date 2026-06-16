package post

import (
	"context"

	"01.tomorrow-school.ai/git/amadiuly/forum/internal/domain/user"
	"github.com/google/uuid"
)

type PostRepository interface {
	Create(ctx context.Context, post *Post) error
	GetByID(ctx context.Context, id uuid.UUID) (*Post, error)
	Update(ctx context.Context, updatePost UpdatePost, postID uuid.UUID) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, filter PostFilter) ([]Post, error)
}

type PostService interface {
	CreatePost(ctx context.Context, post *CreatePost) error
	GetByID(ctx context.Context, id uuid.UUID) (*Post, error)
	UpdateMyPost(ctx context.Context, actorID uuid.UUID, postID uuid.UUID, updatedPost UpdatePost) error
	DeletePost(ctx context.Context, actorID uuid.UUID, actorRole user.Role, postID uuid.UUID) error
	ListPosts(ctx context.Context, filter PostFilter) ([]Post, error)
}
