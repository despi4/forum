package reactionsvc

import (
	"context"

	domain "01.tomorrow-school.ai/git/amadiuly/forum/internal/domain/reaction"
	"01.tomorrow-school.ai/git/amadiuly/forum/internal/service"
	"github.com/google/uuid"
)

type PostReactionService struct {
	postReactionRepo domain.PostReactionRepository
}

func NewPostReactionService(repo domain.PostReactionRepository) *PostReactionService {
	return &PostReactionService{postReactionRepo: repo}
}

var _ domain.PostReactionService = (*PostReactionService)(nil)

func (s *PostReactionService) SetPostReaction(ctx context.Context, reactionType domain.ReactionType, userID, postID uuid.UUID) error {
	if isNil := service.CheckUUID(userID, postID); !isNil {
		return domain.ErrInvalidArgument
	}

	reaction := domain.PostReaction{
		UserID: userID,
		PostID: postID,
		Type:   reactionType,
	}

	return s.postReactionRepo.Set(ctx, &reaction)
}

func (s *PostReactionService) GetPostReaction(ctx context.Context, userID, postID uuid.UUID) (*domain.PostReaction, error) {
	if isNil := service.CheckUUID(userID, postID); !isNil {
		return nil, domain.ErrInvalidArgument
	}

	postReaction, err := s.postReactionRepo.GetByUserAndPost(ctx, userID, postID)
	if err != nil {
		return nil, err
	}

	if postReaction == nil {
		return nil, domain.ErrReactionNotFound
	}

	return postReaction, nil
}

func (s *PostReactionService) DeletePostReaction(ctx context.Context, userID, commentID uuid.UUID) error {
	if isNil := service.CheckUUID(userID, commentID); !isNil {
		return domain.ErrInvalidArgument
	}

	reaction, err := s.GetPostReaction(ctx, userID, commentID)
	if err != nil {
		return err
	}

	if reaction == nil {
		return domain.ErrInvalidArgument
	}

	if userID != reaction.UserID {
		return domain.ErrForbidden
	}

	return s.postReactionRepo.Delete(ctx, reaction.ID)
}

func (s *PostReactionService) CountPostReactions(ctx context.Context, commentID uuid.UUID) (*domain.ReactionsCount, error) {
	if isNil := service.CheckUUID(commentID); !isNil {
		return nil, domain.ErrInvalidArgument
	}

	count, err := s.postReactionRepo.Count(ctx, commentID)
	if err != nil {
		return nil, err
	}

	return &count, nil
}
