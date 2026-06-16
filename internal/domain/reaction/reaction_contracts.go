package reaction

import (
	"context"

	"github.com/google/uuid"
)

type PostReactionRepository interface {
	Create(ctx context.Context, reaction *PostReaction) error
	GetByUserAndPost(ctx context.Context, userID, postID uuid.UUID) (*PostReaction, error)
	Update(ctx context.Context, reactionID uuid.UUID, reactionType ReactionType) error
	Delete(ctx context.Context, reactionID uuid.UUID) error
	Count(ctx context.Context, postID uuid.UUID) (ReactionsCount, error)
}

type CommentReactionRepository interface {
	Create(ctx context.Context, reaction *CommentReaction) error
	GetByUserAndComment(ctx context.Context, userID, commentID uuid.UUID) (*CommentReaction, error)
	Update(ctx context.Context, reactionID uuid.UUID, reactionType ReactionType) error
	Delete(ctx context.Context, reactionID uuid.UUID) error
	Count(ctx context.Context, commentID uuid.UUID) (ReactionsCount, error)
}

type PostReactionService interface {
	SetPostReaction(ctx context.Context, reactionType ReactionType, userID, postID uuid.UUID) error
	GetPostReactionByID(ctx context.Context, userID, postID uuid.UUID) (*PostReaction, error)
	DeletePostReaction(ctx context.Context, userID, postID uuid.UUID) error
	ListPostReactions(ctx context.Context, postID uuid.UUID) error
}

type CommentReactionService interface {
	SetCommentReaction(ctx context.Context, reactionType ReactionType, userID, commentID uuid.UUID) error
	GetCommentReactionByID(ctx context.Context, userID, commentID uuid.UUID) (*CommentReaction, error)
	DeleteCommentReaction(ctx context.Context, userID, commentID uuid.UUID) error
	ListCommentReactions(ctx context.Context, commentID uuid.UUID) error
}
