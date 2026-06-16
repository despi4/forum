package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"01.tomorrow-school.ai/git/amadiuly/forum/internal/db"
	domain "01.tomorrow-school.ai/git/amadiuly/forum/internal/domain/reaction"
	"github.com/google/uuid"
	"github.com/mattn/go-sqlite3"
	gosqlite3 "github.com/mattn/go-sqlite3"
)

type PostReactionRepo struct {
	db *db.ConnDB
}

func NewPostReactionRepo(db *db.ConnDB) *PostReactionRepo {
	return &PostReactionRepo{db: db}
}

var _ domain.PostReactionRepository = (*PostReactionRepo)(nil)

func (r *PostReactionRepo) Set(ctx context.Context, reaction *domain.PostReaction) error {
	db := r.db.GetDB()

	query := `
		insert into post_reactions (id, user_id, comment_id, reaction_type)
		values (?, ?, ?, ?)
		returning id, user_id, comment_id, reaction_type, created_at;
	`

	if reaction.ID == uuid.Nil {
		reaction.ID = uuid.New()
	}

	if reaction.UserID == uuid.Nil {
		return fmt.Errorf("user_id is reqruied")
	}

	if reaction.PostID == uuid.Nil {
		return fmt.Errorf("comment_id is required")
	}

	if reaction.Type == 0 {
		reaction.Type = 1
	}

	err := db.QueryRowContext(ctx, query, reaction.ID, reaction.UserID, reaction.PostID, reaction.Type).Scan(
		&reaction.ID,
		&reaction.UserID,
		&reaction.PostID,
		&reaction.Type,
		&reaction.CreatedAt,
	)
	if err != nil {
		var sqlErr gosqlite3.Error
		if errors.As(err, &sqlErr) {
			if sqlErr.ExtendedCode == gosqlite3.ErrConstraintUnique {
				return domain.ErrReactionAlreadyExists
			}
		}

		return fmt.Errorf("create post reaction failed reaction_id=(%s): %w", reaction.ID, err)
	}

	return nil
}

func (r *PostReactionRepo) GetByUserAndPost(ctx context.Context, userID, postID uuid.UUID) (*domain.PostReaction, error) {
	var reaction domain.PostReaction
	conn := r.db.GetDB()

	query := `
		select id, user_id, comment_id, reaction_type, created_at
		from post_reactions
		where user_id = ? and comment_id = ?;
	`

	err := conn.QueryRowContext(ctx, query, userID, postID).Scan(
		&reaction.ID,
		&reaction.UserID,
		&reaction.PostID,
		&reaction.Type,
		&reaction.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrReactionNotFound
		}

		return nil, fmt.Errorf("get post reaction by IDs failed user_id=(%s), post_id=(%s): %w", userID, postID, err)
	}

	return &reaction, nil
}

func (r *PostReactionRepo) Update(ctx context.Context, reactionID uuid.UUID, reactionType domain.ReactionType) error {
	conn := r.db.GetDB()

	query := `
		update post_reactions
		set reaction_type = ?
		where id = ?;
	`

	req, err := conn.ExecContext(ctx, query, reactionType, reactionID)
	if err != nil {
		var sqlErr sqlite3.Error
		if errors.As(err, &sqlErr) {
			if sqlErr.ExtendedCode == sqlite3.ErrNoExtended(sqlite3.ErrNotFound) {
				return domain.ErrReactionNotFound
			}
		}

		return fmt.Errorf("update post reaction failed by reaction_id=(%s): %w", reactionID, err)
	}

	rows, err := req.RowsAffected()
	if err != nil {
		return fmt.Errorf("update post reaction failed by reaction_id=(%s): %w", reactionID, err)
	}

	if rows == 0 {
		return domain.ErrReactionNotFound
	}

	return nil
}

func (r *PostReactionRepo) Delete(ctx context.Context, reactionID uuid.UUID) error {
	conn := r.db.GetDB()

	query := `
		delete from post_reactions
		where id = ?;
	`

	res, err := conn.ExecContext(ctx, query, reactionID)
	if err != nil {
		return fmt.Errorf("delete post reaction failed by reaction_id=(%s): %w", reactionID, err)
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("delete post reaction affected rows failed: %w", err)
	} else {
		if rows == 0 {
			return domain.ErrReactionNotFound
		}
	}

	return nil
}

func (r *PostReactionRepo) Count(ctx context.Context, postID uuid.UUID) (domain.ReactionsCount, error) {
	var reactions domain.ReactionsCount
	conn := r.db.GetDB()

	query := `
        SELECT 
            COUNT(*) FILTER (WHERE reaction_type = 1),
            COUNT(*) FILTER (WHERE reaction_type = -1)
        FROM post_reactions
        WHERE comment_id = ?
    `

	err := conn.QueryRowContext(ctx, query, postID).Scan(
		&reactions.Likes,
		&reactions.Dislikes,
	)
	if err != nil {
		return domain.ReactionsCount{}, fmt.Errorf("get count of post reactions failed comment_id=(%s): %w", postID, err)
	}

	return reactions, nil
}
