package category

import (
	"context"

	"github.com/google/uuid"
)

type CategoryRepository interface {
	List(ctx context.Context, search *string) ([]Category, error)

	// admin only
	Create(ctx context.Context, category *Category) error
	Update(ctx context.Context, name string, categoryID uuid.UUID) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type CategoryService interface {
	ListCategories(ctx context.Context) ([]Category, error)

	// admin only
	CreateCategory(ctx context.Context, categoryName string) error
	UpdateCategory(ctx context.Context, categoryName string) error
	DeleteCategory(ctx context.Context, id uuid.UUID) error
}
