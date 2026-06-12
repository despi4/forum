package comment

import (
	"context"

	"github.com/google/uuid"
)

type CommentRepository interface {
	Create(ctx context.Context, comment *Comment) error
	GetByID(ctx context.Context, id uuid.UUID) (*Comment, error)
	Update(ctx context.Context, commentID uuid.UUID, content string) error
	Delete(ctx context.Context, id uuid.UUID) error
	ListByPostID(ctx context.Context, postID uuid.UUID, filter CommentFilter) ([]Comment, error)
}

type CommentService interface {
	CreateComment(ctx context.Context, comment *CreateComment) error
	GetById(ctx context.Context, id uuid.UUID) (*Comment, error)
	UpdateMyComment(ctx context.Context, id uuid.UUID, userID uuid.UUID, content string) error
	DeleteComment(ctx context.Context, actor uuid.UUID, target uuid.UUID) error
	ListComments(ctx context.Context, postID uuid.UUID) ([]Comment, error)
}
