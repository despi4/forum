package contentsvc

import (
	"context"
	"strings"

	domain "01.tomorrow-school.ai/git/amadiuly/forum/internal/domain/category"
	"01.tomorrow-school.ai/git/amadiuly/forum/internal/service"
	"github.com/google/uuid"
)

type CategoryService struct {
	categoryRepo domain.CategoryRepository
}

func NewAuthService(categoryRepo domain.CategoryRepository) *CategoryService {
	return &CategoryService{categoryRepo: categoryRepo}
}

var _ domain.CategoryService = (*CategoryService)(nil)

func (s *CategoryService) ListCategories(ctx context.Context, search string) ([]domain.Category, error) {
	search = strings.TrimSpace(search)

	var searchPtr *string
	if search != "" {
		searchPtr = &search
	}

	listOfCategories, err := s.categoryRepo.List(ctx, searchPtr)
	if err != nil {
		return nil, err
	}

	return listOfCategories, nil
}

// admin only
func (s *CategoryService) CreateCategory(ctx context.Context, categoryName string) error {
	newCategory := domain.Category{
		Name: service.Capitalize(strings.TrimSpace(categoryName)),
	}

	err := s.categoryRepo.Create(ctx, &newCategory)
	if err != nil {
		return err
	}

	return nil
}

func (s *CategoryService) UpdateCategory(ctx context.Context, categoryName string, categoryID uuid.UUID) error {
	categoryName = service.Capitalize(strings.TrimSpace(categoryName))

	if isNil := service.CheckUUID(categoryID); !isNil {
		return domain.ErrInvalidArgument
	}

	err := s.categoryRepo.Update(ctx, categoryName, categoryID)
	if err != nil {
		return err
	}

	return nil
}

func (s *CategoryService) DeleteCategory(ctx context.Context, id uuid.UUID) error {
	if isNil := service.CheckUUID(id); !isNil {
		return domain.ErrInvalidArgument
	}

	err := s.categoryRepo.Delete(ctx, id)
	if err != nil {
		return err
	}

	return nil
}
