package contentsvc

import (
	"context"
	"fmt"
	"strings"

	domain "01.tomorrow-school.ai/git/amadiuly/forum/internal/domain/post"
	"01.tomorrow-school.ai/git/amadiuly/forum/internal/domain/user"
	"01.tomorrow-school.ai/git/amadiuly/forum/internal/service"
	"github.com/google/uuid"
)

type PostService struct {
	postRepo domain.PostRepository
}

func NewPostService(repo domain.PostRepository) *PostService {
	return &PostService{postRepo: repo}
}

var _ domain.PostService = (*PostService)(nil)

func (s *PostService) CreatePost(ctx context.Context, post *domain.CreatePost) error {
	if post == nil {
		return domain.ErrInvalidArgument
	}

	newPost := domain.Post{
		AuthorID:   post.UserID,
		CategoryID: post.Category.ID,
		Content:    post.Content,
		Title:      post.Title,
	}

	if err := validPost(&newPost); err != nil {
		return err
	}

	err := s.postRepo.Create(ctx, &newPost)

	return err
}

func (s *PostService) GetByID(ctx context.Context, id uuid.UUID) (*domain.Post, error) {
	if isNil := service.CheckUUID(id); !isNil {
		return nil, domain.ErrInvalidArgument
	}

	post, err := s.postRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if post == nil {
		return nil, domain.ErrPostNotFound
	}

	return post, nil
}

func (s *PostService) UpdateMyPost(ctx context.Context, actorID uuid.UUID, postID uuid.UUID, updatedPost domain.UpdatePost) error {
	if isValid := service.CheckUUID(actorID, postID); !isValid {
		return domain.ErrInvalidArgument
	}

	post, err := s.postRepo.GetByID(ctx, postID)
	if err != nil {
		return err
	}

	if post == nil {
		return domain.ErrPostNotFound
	}

	if post.AuthorID != actorID {
		return domain.ErrForbidden
	}

	mergedPost := *post

	if updatedPost.CategoryID != nil {
		if isValid := service.CheckUUID(*updatedPost.CategoryID); !isValid {
			return domain.ErrInvalidArgument
		}

		mergedPost.CategoryID = *updatedPost.CategoryID
	}

	if updatedPost.Title != nil {
		title := service.Capitalize(strings.TrimSpace(*updatedPost.Title))
		updatedPost.Title = &title
		mergedPost.Title = title
	}

	if updatedPost.Content != nil {
		content := strings.TrimSpace(*updatedPost.Content)
		updatedPost.Content = &content
		mergedPost.Content = content
	}

	if err := validPost(&mergedPost); err != nil {
		return err
	}

	return s.postRepo.Update(ctx, updatedPost, postID)
}

func (s *PostService) DeletePost(ctx context.Context, actorID uuid.UUID, actorRole user.Role, postID uuid.UUID) error {
	if isNil := service.CheckUUID(actorID, postID); !isNil {
		return domain.ErrInvalidArgument
	}

	post, err := s.postRepo.GetByID(ctx, postID)
	if err != nil {
		return err
	}

	if post == nil {
		return domain.ErrPostNotFound
	}

	if post.AuthorID != actorID && actorRole != user.RoleAdmin {
		return domain.ErrForbidden
	}

	return s.postRepo.Delete(ctx, postID)
}

func (s *PostService) ListPosts(ctx context.Context, filter domain.PostFilter) ([]domain.Post, error) {
	if filter.Search != nil {
		filter.Search = service.Ptr(strings.TrimSpace(*filter.Search))
	}

	if filter.CategoryID != nil {
		if isNil := service.CheckUUID(*filter.CategoryID); !isNil {
			return nil, domain.ErrInvalidArgument
		}
	}

	if filter.AuthorID != nil {
		if isNil := service.CheckUUID(*filter.AuthorID); !isNil {
			return nil, domain.ErrInvalidArgument
		}
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

	if filter.Sort == "" {
		filter.Sort = domain.PostSortCreatedDesc
	}

	list, err := s.postRepo.List(ctx, filter)
	if err != nil {
		return nil, err
	}

	return list, nil
}

func validPost(post *domain.Post) error {
	if post == nil {
		return domain.ErrInvalidArgument
	}

	if isNil := service.CheckUUID(post.AuthorID, post.CategoryID); !isNil {
		return domain.ErrInvalidArgument
	}

	post.Title = service.Capitalize(strings.TrimSpace(post.Title))

	if len(post.Title) < 5 || len(post.Title) > 50 {
		return fmt.Errorf("post title contain at least 5 and not more than 50 characters: %w", domain.ErrInvalidArgument)
	}

	post.Content = strings.TrimSpace(post.Content)

	if len(post.Content) < 10 || len(post.Content) > 400 {
		return fmt.Errorf("post content must contain at least 10 and not more than 400 characters: %w", domain.ErrInvalidArgument)
	}

	return nil
}
