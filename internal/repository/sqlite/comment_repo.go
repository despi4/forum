package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"01.tomorrow-school.ai/git/amadiuly/forum/internal/db"
	domain "01.tomorrow-school.ai/git/amadiuly/forum/internal/domain/comment"
	"01.tomorrow-school.ai/git/amadiuly/forum/internal/repository"
	"github.com/google/uuid"
)

type CommentRepo struct {
	db *db.ConnDB
}

func NewCommentRepo(db *db.ConnDB) *CommentRepo {
	return &CommentRepo{db: db}
}

var _ domain.CommentRepository = (*CommentRepo)(nil)

func (r *CommentRepo) Create(ctx context.Context, comment *domain.Comment) error {
	conn := r.db.GetDB()

	query := `
		insert into comments (id, user_id, post_id, content)
		values (?, ?, ?, ?)
		returning id, user_id, post_id, content, created_at, updated_at;
	`

	if comment.ID == uuid.Nil {
		comment.ID = uuid.New()
	}

	if comment.UserID == uuid.Nil {
		return fmt.Errorf("create comment failed: user_id is required")
	}

	if comment.PostID == uuid.Nil {
		return fmt.Errorf("create comment failed: post_id is required")
	}

	content := strings.TrimSpace(comment.Content)
	if content == "" {
		return fmt.Errorf("create comment failed: content is required")
	}

	err := conn.QueryRowContext(ctx, query, comment.ID, comment.UserID, comment.PostID, content).Scan(
		&comment.ID,
		&comment.UserID,
		&comment.PostID,
		&comment.Content,
		&comment.CreatedAt,
		&comment.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("create comment failed comment_id=(%s): %w", comment.ID, err)
	}

	return nil
}

func (r *CommentRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Comment, error) {
	var comment domain.Comment
	conn := r.db.GetDB()

	query := `
		select id, user_id, post_id, content, created_at, updated_at
		from comments
		where id = ?;
	`

	err := conn.QueryRowContext(ctx, query, id).Scan(
		&comment.ID,
		&comment.UserID,
		&comment.PostID,
		&comment.Content,
		&comment.CreatedAt,
		&comment.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrCommentNotFound
		}

		return nil, fmt.Errorf("get comment by id failed comment_id=(%s): %w", id, err)
	}

	return &comment, nil
}

func (r *CommentRepo) Update(ctx context.Context, commentID uuid.UUID, content string) error {
	conn := r.db.GetDB()

	if commentID == uuid.Nil {
		return fmt.Errorf("update comment failed: comment_id is required")
	}

	content = strings.TrimSpace(content)
	if content == "" {
		return fmt.Errorf("update comment failed by comment_id=(%s): content is required", commentID)
	}

	query := `
		update comments
		set content = ?, updated_at = CURRENT_TIMESTAMP
		where id = ?;
	`

	cmd, err := conn.ExecContext(ctx, query, content, commentID)
	if err != nil {
		return fmt.Errorf("update comment failed by comment_id=(%s): %w", commentID, err)
	}

	rows, err := cmd.RowsAffected()
	if err != nil {
		return fmt.Errorf("update comment failed by comment_id=(%s): %w", commentID, err)
	}

	if rows == 0 {
		return domain.ErrCommentNotFound
	}

	return nil
}

func (r *CommentRepo) Delete(ctx context.Context, commentID uuid.UUID) error {
	conn := r.db.GetDB()

	if commentID == uuid.Nil {
		return fmt.Errorf("delete comment failed: comment_id is required")
	}

	query := `
		delete from comments
		where id = ?;
	`

	res, err := conn.ExecContext(ctx, query, commentID)
	if err != nil {
		return fmt.Errorf("delete comment failed by comment_id=(%s): %w", commentID, err)
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("get comment affected rows failed: %w", err)
	}

	if rows == 0 {
		return domain.ErrCommentNotFound
	}

	return nil
}

func (r *CommentRepo) ListByPostID(ctx context.Context, postID uuid.UUID, filter domain.CommentFilter) ([]domain.Comment, error) {
	conn := r.db.GetDB()

	if postID == uuid.Nil {
		return nil, fmt.Errorf("list comments failed: post_id is required")
	}

	query, args := buildListCommentsByPostIDQuery(postID, filter)

	rows, err := conn.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("get comments by post failed post_id=(%s): %w", postID, err)
	}
	defer rows.Close()

	capacity := filter.Limit
	if capacity <= 0 {
		capacity = 50
	}

	comments := make([]domain.Comment, 0, capacity)

	for rows.Next() {
		var comment domain.Comment

		err := rows.Scan(
			&comment.ID,
			&comment.UserID,
			&comment.PostID,
			&comment.Content,
			&comment.CreatedAt,
			&comment.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scan comment rows failed: %w", err)
		}

		comments = append(comments, comment)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate comment rows failed: %w", err)
	}

	return comments, nil
}

func buildListCommentsByPostIDQuery(postID uuid.UUID, filter domain.CommentFilter) (string, []any) {
	var sb strings.Builder
	q := repository.NewQueryParts()

	sb.WriteString(`
		select
			c.id,
			c.user_id,
			c.post_id,
			c.content,
			c.created_at,
			c.updated_at
		from comments c
		left join comment_reactions cr on cr.comment_id = c.id
	`)

	q.AddWhere("c.post_id = ?", postID)

	if filter.Limit <= 0 {
		filter.Limit = 50
	}

	if filter.Offset < 0 {
		filter.Offset = 0
	}

	sb.WriteString(q.BuildWhere())
	sb.WriteString(`
		group by c.id, c.user_id, c.post_id, c.content, c.created_at, c.updated_at
	`)
	sb.WriteString(buildCommentOrderBy(filter.CommentSort))
	sb.WriteString(` limit ? offset ?;`)

	args := append(q.WhereArgs, filter.Limit, filter.Offset)

	return sb.String(), args
}

func buildCommentOrderBy(sort domain.CommentSort) string {
	switch sort {
	case domain.CommentSortCreatedAsc:
		return ` order by c.created_at asc, c.id asc`
	case domain.CommentSortReactionsCount, domain.CommentSort("likes_count"):
		return `
			order by
				coalesce(sum(case when cr.reaction_type = 1 then 1 else 0 end), 0) desc,
				c.created_at desc,
				c.id desc
		`
	default:
		return ` order by c.created_at desc, c.id desc`
	}
}
