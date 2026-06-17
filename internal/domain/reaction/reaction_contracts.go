package reaction

import (
	"context"

	"github.com/google/uuid"
)

type PostReactionRepository interface {
	Set(ctx context.Context, reaction *PostReaction) error
	GetByUserAndPost(ctx context.Context, userID, postID uuid.UUID) (*PostReaction, error)
	Update(ctx context.Context, reactionID uuid.UUID, reactionType ReactionType) error
	Delete(ctx context.Context, reactionID uuid.UUID) error
	Count(ctx context.Context, postID uuid.UUID) (ReactionsCount, error)
}

type CommentReactionRepository interface {
	Set(ctx context.Context, reaction *CommentReaction) error
	GetByUserAndComment(ctx context.Context, userID, commentID uuid.UUID) (*CommentReaction, error)
	Update(ctx context.Context, reactionID uuid.UUID, reactionType ReactionType) error
	Delete(ctx context.Context, reactionID uuid.UUID) error
	Count(ctx context.Context, commentID uuid.UUID) (ReactionsCount, error)
}

type PostReactionService interface {
	SetPostReaction(ctx context.Context, reactionType ReactionType, userID, postID uuid.UUID) error
	GetPostReaction(ctx context.Context, userID, postID uuid.UUID) (*PostReaction, error)
	DeletePostReaction(ctx context.Context, userID, postID uuid.UUID) error
	CountPostReactions(ctx context.Context, postID uuid.UUID) (*ReactionsCount, error)
}

type CommentReactionService interface {
	SetCommentReaction(ctx context.Context, reactionType ReactionType, userID, commentID uuid.UUID) error
	GetCommentReaction(ctx context.Context, userID, commentID uuid.UUID) (*CommentReaction, error)
	DeleteCommentReaction(ctx context.Context, userID, commentID uuid.UUID) error
	CountCommentReactions(ctx context.Context, commentID uuid.UUID) (*ReactionsCount, error)
}
