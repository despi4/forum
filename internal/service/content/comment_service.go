package contentsvc

import (
	"context"
	"fmt"
	"strings"

	domain "01.tomorrow-school.ai/git/amadiuly/forum/internal/domain/comment"
	"01.tomorrow-school.ai/git/amadiuly/forum/internal/domain/user"
	"01.tomorrow-school.ai/git/amadiuly/forum/internal/service"
	"github.com/google/uuid"
)

type CommentService struct {
	commentRepo domain.CommentRepository
}

func NewCommentService(repo domain.CommentRepository) *CommentService {
	return &CommentService{commentRepo: repo}
}

var _ domain.CommentService = (*CommentService)(nil)

func (s *CommentService) CreateComment(ctx context.Context, comment *domain.CreateComment) error {
	if comment == nil {
		return domain.ErrInvalidArgument
	}

	err := validBody(&comment.Content)
	if err != nil {
		return err
	}

	if isNil := service.CheckUUID(comment.PostID, comment.UserID); !isNil {
		return domain.ErrInvalidArgument
	}

	newComment := domain.Comment{
		UserID:  comment.UserID,
		PostID:  comment.PostID,
		Content: comment.Content,
	}

	err = s.commentRepo.Create(ctx, &newComment)
	if err != nil {
		return err
	}

	return nil
}

func (s *CommentService) GetById(ctx context.Context, id uuid.UUID) (*domain.Comment, error) {
	if isNil := service.CheckUUID(id); !isNil {
		return nil, domain.ErrInvalidArgument
	}

	comment, err := s.commentRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if comment == nil {
		return nil, domain.ErrCommentNotFound
	}

	return comment, nil
}

func (s *CommentService) UpdateMyComment(ctx context.Context, userID uuid.UUID, commentId uuid.UUID, content string) error {
	if err := validBody(&content); err != nil {
		return err
	}

	if isNil := service.CheckUUID(userID, commentId); !isNil {
		return domain.ErrInvalidArgument
	}

	comment, err := s.commentRepo.GetByID(ctx, commentId)
	if err != nil {
		return err
	}

	if comment == nil {
		return domain.ErrCommentNotFound
	}

	if comment.UserID != userID {
		return domain.ErrForbidden
	}

	if err := s.commentRepo.Update(ctx, commentId, content); err != nil {
		return err
	}

	return nil
}

func (s *CommentService) DeleteComment(ctx context.Context, actorID uuid.UUID, actorRole user.Role, commentID uuid.UUID) error {
	if isNil := service.CheckUUID(actorID, commentID); !isNil {
		return domain.ErrInvalidArgument
	}

	comment, err := s.commentRepo.GetByID(ctx, commentID)
	if err != nil {
		return err
	}

	if comment == nil {
		return domain.ErrCommentNotFound
	}

	if comment.UserID != actorID && actorRole != user.RoleAdmin {
		return domain.ErrForbidden
	}

	return s.commentRepo.Delete(ctx, commentID)
}

func (s *CommentService) ListComments(ctx context.Context, postID uuid.UUID, filter domain.CommentFilter) ([]domain.Comment, error) {
	if isNil := service.CheckUUID(postID); isNil {
		return nil, domain.ErrInvalidArgument
	}

	if filter.Limit <= 0 {
		filter.Limit = 10
	}

	if filter.Limit > 100 {
		filter.Limit = 100
	}

	if filter.Offset < 0 {
		filter.Offset = 0
	}

	if filter.CommentSort == nil {
		defaultSort := domain.CommentSortReactionsCount

		filter.CommentSort = &defaultSort
	}

	list, err := s.commentRepo.ListByPostID(ctx, postID, filter)
	if err != nil {
		return nil, err
	}

	return list, nil
}

func validBody(content *string) error {
	if content == nil {
		return domain.ErrInvalidArgument
	}

	*content = strings.TrimSpace(*content)

	if *content == "" {
		return domain.ErrInvalidArgument
	}

	if len(*content) < 10 || len(*content) > 300 {
		return fmt.Errorf("comment must contain at least 10 characters and not more than 300 characters: %w", domain.ErrInvalidArgument)
	}

	return nil
}
