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

type CommentReactionRepo struct {
	db *db.ConnDB
}

func NewCommentReactionRepo(db *db.ConnDB) *CommentReactionRepo {
	return &CommentReactionRepo{db: db}
}

func (r *CommentReactionRepo) Create(ctx context.Context, reaction *domain.CommentReaction) error {
	db := r.db.GetDB()

	query := `
		insert into comment_reactions (id, user_id, comment_id, reaction_type)
		values (?, ?, ?, ?)
		returning id, user_id, comment_id, reaction_type, created_at;
	`

	if reaction.ID == uuid.Nil {
		reaction.ID = uuid.New()
	}

	if reaction.UserID == uuid.Nil {
		return fmt.Errorf("user_id is reqruied")
	}

	if reaction.CommentID == uuid.Nil {
		return fmt.Errorf("comment_id is required")
	}

	if reaction.Type == 0 {
		reaction.Type = 1
	}

	err := db.QueryRowContext(ctx, query, reaction.ID, reaction.UserID, reaction.CommentID, reaction.Type).Scan(
		&reaction.ID,
		&reaction.UserID,
		&reaction.CommentID,
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

		return fmt.Errorf("create comment reaction failed reaction_id=(%s): %w", reaction.ID, err)
	}

	return nil
}

func (r *CommentReactionRepo) GetByUserAndComment(ctx context.Context, userID, commentID uuid.UUID) (*domain.CommentReaction, error) {
	var reaction domain.CommentReaction
	conn := r.db.GetDB()

	query := `
		select id, user_id, comment_id, reaction_type, created_at
		from comment_reactions
		where user_id = ? and comment_id = ?;
	`

	err := conn.QueryRowContext(ctx, query, userID, commentID).Scan(
		&reaction.ID,
		&reaction.UserID,
		&reaction.CommentID,
		&reaction.Type,
		&reaction.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrReactionNotFound
		}

		return nil, fmt.Errorf("get comment reaction by IDs failed user_id=(%s), comment_id=(%s): %w", userID, commentID, err)
	}

	return &reaction, nil
}

func (r *CommentReactionRepo) Update(ctx context.Context, reactionID uuid.UUID, reactionType domain.ReactionType) error {
	conn := r.db.GetDB()

	query := `
		update comment_reactions
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

		return fmt.Errorf("update comment reaction failed by reaction_id=(%s): %w", reactionID, err)
	}

	rows, err := req.RowsAffected()
	if err != nil {
		return fmt.Errorf("update comment reaction failed by reaction_id=(%s): %w", reactionID, err)
	}

	if rows == 0 {
		return domain.ErrReactionNotFound
	}

	return nil
}

func (r *CommentReactionRepo) Delete(ctx context.Context, reactionID uuid.UUID) error {
	conn := r.db.GetDB()

	query := `
		delete from comment_reactions
		where id = ?;
	`

	res, err := conn.ExecContext(ctx, query, reactionID)
	if err != nil {
		return fmt.Errorf("delete comment reaction failed by reaction_id=(%s): %w", reactionID, err)
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("delete comment reaction affected rows failed: %w", err)
	} else {
		if rows == 0 {
			return domain.ErrReactionNotFound
		}
	}

	return nil
}

func (r *CommentReactionRepo) Count(ctx context.Context, commentID uuid.UUID) (*domain.ReactionsCount, error) {
	var reactions domain.ReactionsCount
	conn := r.db.GetDB()

	query := `
        SELECT 
            COUNT(*) FILTER (WHERE reaction_type = 1),
            COUNT(*) FILTER (WHERE reaction_type = -1)
        FROM comment_reactions
        WHERE comment_id = ?
    `

	err := conn.QueryRowContext(ctx, query, commentID).Scan(
		&reactions.Likes,
		&reactions.Dislikes,
	)
	if err != nil {
		return nil, fmt.Errorf("get count of comment reactions failed comment_id=(%s): %w", commentID, err)
	}

	return &reactions, nil
}
