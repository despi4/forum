package reactionsvc

import (
	"context"

	domain "01.tomorrow-school.ai/git/amadiuly/forum/internal/domain/reaction"
	"01.tomorrow-school.ai/git/amadiuly/forum/internal/service"
	"github.com/google/uuid"
)

type CommentReactionService struct {
	commentReactionRepo domain.CommentReactionRepository
}

func NewCommentReactionService(repo domain.CommentReactionRepository) *CommentReactionService {
	return &CommentReactionService{commentReactionRepo: repo}
}

var _ domain.CommentReactionService = (*CommentReactionService)(nil)

func (s *CommentReactionService) SetCommentReaction(ctx context.Context, reactionType domain.ReactionType, userID, commentID uuid.UUID) error {
	if isNil := service.CheckUUID(userID, commentID); !isNil {
		return domain.ErrInvalidArgument
	}

	reaction := domain.CommentReaction{
		UserID:    userID,
		CommentID: commentID,
		Type:      reactionType,
	}

	return s.commentReactionRepo.Set(ctx, &reaction)
}

func (s *CommentReactionService) GetCommentReaction(ctx context.Context, userID, commentID uuid.UUID) (*domain.CommentReaction, error) {
	if isNil := service.CheckUUID(userID, commentID); !isNil {
		return nil, domain.ErrInvalidArgument
	}

	commentReaction, err := s.commentReactionRepo.GetByUserAndComment(ctx, userID, commentID)
	if err != nil {
		return nil, err
	}

	if commentReaction == nil {
		return nil, domain.ErrReactionNotFound
	}

	return commentReaction, nil
}

func (s *CommentReactionService) DeleteCommentReaction(ctx context.Context, userID, commentID uuid.UUID) error {
	if isNil := service.CheckUUID(userID, commentID); !isNil {
		return domain.ErrInvalidArgument
	}

	reaction, err := s.GetCommentReaction(ctx, userID, commentID)
	if err != nil {
		return err
	}

	if reaction == nil {
		return domain.ErrInvalidArgument
	}

	if userID != reaction.UserID {
		return domain.ErrForbidden
	}

	return s.commentReactionRepo.Delete(ctx, reaction.ID)
}

func (s *CommentReactionService) CountCommentReactions(ctx context.Context, commentID uuid.UUID) (*domain.ReactionsCount, error) {
	if isNil := service.CheckUUID(commentID); !isNil {
		return nil, domain.ErrInvalidArgument
	}

	count, err := s.commentReactionRepo.Count(ctx, commentID)
	if err != nil {
		return nil, err
	}

	return &count, nil
}
