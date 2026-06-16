package sqlite

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"01.tomorrow-school.ai/git/amadiuly/forum/internal/db"
	domain "01.tomorrow-school.ai/git/amadiuly/forum/internal/domain/category"
	"github.com/google/uuid"
	gosqlite3 "github.com/mattn/go-sqlite3"
)

type CategoryRepo struct {
	db *db.ConnDB
}

func NewCategoryRepo(db *db.ConnDB) *CategoryRepo {
	return &CategoryRepo{db: db}
}

var _ domain.CategoryRepository = (*CategoryRepo)(nil)

func (r *CategoryRepo) List(ctx context.Context, search *string) ([]domain.Category, error) {
	db := r.db.GetDB()

	query := `
		select id, name, created_at
		from categories
	`

	var arg any

	if search != nil && strings.TrimSpace(*search) != "" {
		searchString := "%" + strings.TrimSpace(*search) + "%"
		arg = searchString
		query = query + "(name like ?)"
	}

	query += "where by created_at desc, id desc;"

	rows, err := db.QueryContext(ctx, query, search, arg)
	if err != nil {
		return nil, fmt.Errorf("get categories failed: %w", err)
	}

	categories := make([]domain.Category, 0)

	for rows.Next() {
		var category domain.Category

		err := rows.Scan(
			&category.ID,
			&category.Name,
			&category.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scan category rows failed %w", err)
		}

		categories = append(categories, category)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("increate rows failed %w", err)
	}

	return categories, nil
}

func (r *CategoryRepo) Create(ctx context.Context, category *domain.Category) error {
	db := r.db.GetDB()

	query := `
		insert into categories (id, name)
		values (?, ?)
		returning id, name, created_at;
	`

	if category.ID == uuid.Nil {
		category.ID = uuid.New()
	}

	name := strings.ToLower(strings.TrimSpace(category.Name))

	if name == "" {
		return fmt.Errorf("category name is required")
	}

	err := db.QueryRowContext(ctx, query, category.ID, name).Scan(
		&category.ID,
		&category.Name,
		&category.CreatedAt,
	)

	if err != nil {
		return fmt.Errorf("create category failed category_id=(%s): %w", category.ID, err)
	}

	return nil
}
func (r *CategoryRepo) Update(ctx context.Context, name string, categoryID uuid.UUID) error {
	db := r.db.GetDB()

	query := `
		update categories
		set name = ?
		where id = ?;
	`

	name = strings.ToLower(strings.TrimSpace(name))

	_, err := db.ExecContext(ctx, query, categoryID, name)
	if err != nil {
		var sqlErr gosqlite3.Error
		if errors.As(err, &sqlErr) {
			if sqlErr.ExtendedCode == gosqlite3.ErrNoExtended(gosqlite3.ErrNotFound) {
				return domain.ErrCategoryNotFound
			}

			if sqlErr.ExtendedCode == gosqlite3.ErrConstraintUnique {
				return domain.ErrCategoryExists
			}
		}

		return fmt.Errorf("update category failed by category_id=(%s), %w", categoryID, err)
	}

	return nil
}

func (r *CategoryRepo) Delete(ctx context.Context, id uuid.UUID) error {
	db := r.db.GetDB()

	query := `
		delete from categories
		where id = ?;
	`

	res, err := db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("delete category failed by category_id=(%s): %w", id, err)
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("get category affected rows failed: %w", err)
	} else {
		if rows == 0 {
			return domain.ErrCategoryNotFound
		}
	}

	return nil
}
