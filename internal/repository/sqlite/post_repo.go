package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"01.tomorrow-school.ai/git/amadiuly/forum/internal/db"
	domain "01.tomorrow-school.ai/git/amadiuly/forum/internal/domain/post"
	"01.tomorrow-school.ai/git/amadiuly/forum/internal/repository"
	"github.com/google/uuid"
	"github.com/mattn/go-sqlite3"
)

type PostRepo struct {
	db *db.ConnDB
}

func NewPostRepo(db *db.ConnDB) *PostRepo {
	return &PostRepo{db: db}
}

func (r *PostRepo) Create(ctx context.Context, post *domain.Post) error {
	db := r.db.GetDB()

	if post.ID == uuid.Nil {
		post.ID = uuid.New()
	}

	title := strings.ToLower(post.Title)
	content := strings.ToLower(post.Content)

	query := `
		insert into posts (id, author_id, category_id, title, content)
		values (?, ?, ?, ?)
		returning id, author_id, category_id, title, content, created_at, updated_at;
	`

	err := db.QueryRowContext(ctx, query, post.ID, post.AuthorID, title, content).Scan(
		&post.ID,
		&post.AuthorID,
		&post.CategoryID,
		&post.Title,
		&post.Content,
		&post.CreatedAt,
		&post.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("create post failed by post_id=(%s): %w", post.ID, err)
	}

	return nil
}

func (r *PostRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Post, error) {
	db := r.db.GetDB()

	query := `
		select id, author_id, category_id, title, content, created_at, updated_at
		from posts
		where id = ?;
	`

	var post domain.Post
	err := db.QueryRowContext(ctx, query, id).Scan(
		&post.ID,
		&post.AuthorID,
		&post.Title,
		&post.Content,
		&post.CreatedAt,
		&post.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrPostNotFound
		}

		return nil, fmt.Errorf("get post by id failed post_id=(%s): %w", id, err)
	}

	return &post, nil
}

func (r *PostRepo) Update(ctx context.Context, updatePost domain.UpdatePost, postID uuid.UUID) error {
	db := r.db.GetDB()

	query, args, err := buildUpdatePostQuery(updatePost, postID)
	if err != nil {
		return fmt.Errorf("updated user failed user_id=(%s): %w", postID, err)
	}

	cmd, err := db.ExecContext(ctx, query, args...)
	if err != nil {
		var sqlErr sqlite3.Error
		if errors.As(err, &sqlErr) {
			if sqlErr.ExtendedCode == sqlite3.ErrNoExtended(sqlite3.ErrNotFound) {
				return domain.ErrPostNotFound
			}
		}

		return fmt.Errorf("update post failed by post_id=(%s): %w", postID, err)
	}

	rows, err := cmd.RowsAffected()
	if err != nil {
		return fmt.Errorf("update post failed by post_id(=%s): %w", postID, err)
	}

	if rows == 0 {
		return domain.ErrPostNotFound
	}

	return nil
}

func (r *PostRepo) Delete(ctx context.Context, postID uuid.UUID) error {
	db := r.db.GetDB()

	query := `
		delete from posts
		where id = ?;
	`

	res, err := db.ExecContext(ctx, query, postID)
	if err != nil {
		return fmt.Errorf("delete post failed by post_id=(%s): %w", postID, err)
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("get post affected rows failed: %w", err)
	} else {
		if rows == 0 {
			return domain.ErrPostNotFound
		}
	}

	return nil
}

func (r *PostRepo) List(ctx context.Context, filter domain.PostFilter) ([]domain.Post, error) {
	db := r.db.GetDB()

	query, args := buildListPostsQuery(filter)

	rows, err := db.QueryContext(ctx, query, args)
	if err != nil {
		return nil, fmt.Errorf("get posts failed: %w", err)
	}
	defer rows.Close()

	return nil, nil
}

func buildUpdatePostQuery(params domain.UpdatePost, id uuid.UUID) (string, []any, error) {
	var sb strings.Builder
	q := repository.NewQueryParts()

	sb.WriteString(`update posts`)

	if params.CategoryID != nil {
		q.AddSet("category_id = ?", *params.CategoryID)
	}

	if params.Content != nil {
		q.AddSet("content = ?", *params.Content)
	}

	if params.Title != nil {
		q.AddSet("title = ?", *params.Title)
	}

	if len(q.Set) == 0 {
		return "", nil, fmt.Errorf("no fields to update post")
	}

	q.AddWhere("id = ?", id)

	sb.WriteString(q.BuildSet())
	sb.WriteString(q.BuildWhere())

	args := append(q.SetArgs, q.WhereArgs...)

	return sb.String(), args, nil
}

func buildListPostsQuery(filter domain.PostFilter) (string, []any) {
	var sb strings.Builder
	q := repository.NewQueryParts()

	sb.WriteString(`
		select id, author_id, category_id, title, content, created_at, updated_at
		from posts
	`)

	if filter.Search != nil && strings.TrimSpace(*filter.Search) != "" {
		search := "%" + strings.TrimSpace(*filter.Search) + "%"
		q.AddWhere("(username like ?)", search)
	}

	if filter.Limit <= 0 {
		filter.Limit = 50
	}

	if filter.Offset < 0 {
		filter.Offset = 0
	}

	sb.WriteString(q.BuildWhere())
	if filter.Sort == domain.PostSortCreatedAsc {
		sb.WriteString("order by created_at asc, id desc")
	} else {
		sb.WriteString("order by created_at desc, id desc")
	}
	sb.WriteString(` limit ? offset ?;`)

	args := append(q.WhereArgs, filter.Limit, filter.Offset)

	return sb.String(), args
}
