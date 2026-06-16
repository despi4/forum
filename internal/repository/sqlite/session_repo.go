package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"01.tomorrow-school.ai/git/amadiuly/forum/internal/db"
	domain "01.tomorrow-school.ai/git/amadiuly/forum/internal/domain/auth"
	user "01.tomorrow-school.ai/git/amadiuly/forum/internal/domain/user"
	"github.com/google/uuid"
	gosqlite3 "github.com/mattn/go-sqlite3"
)

type SessionRepo struct {
	db *db.ConnDB
}

var _ domain.SessionRepository = (*SessionRepo)(nil)

func NewSessionRepo(db *db.ConnDB) *SessionRepo {
	return &SessionRepo{db: db}
}

func (r *SessionRepo) Create(ctx context.Context, session *domain.Session, tx *sql.Tx) error {
	query := `
		insert into sessions (id, user_id, lastseen_at, expires_at)
		values (?, ?, ?, ?)
		returning id, user_id, created_at, lastseen_at, expires_at;
	`

	if session.ID == uuid.Nil {
		session.ID = uuid.New()
	}

	if session.LastSeenAt.IsZero() {
		session.LastSeenAt = time.Now().UTC()
	}

	if session.ExpiresAt.IsZero() {
		return fmt.Errorf("expires_at is required")
	}

	if tx == nil {
		db := r.db.GetDB()

		err := db.QueryRowContext(ctx, query, session.ID, session.UserID, session.LastSeenAt, session.ExpiresAt).Scan(
			&session.ID,
			&session.UserID,
			&session.CreatedAt,
			&session.LastSeenAt,
			&session.ExpiresAt,
		)
		if err != nil {
			var sqlErr gosqlite3.Error
			if errors.As(err, &sqlErr) {
				if sqlErr.ExtendedCode == gosqlite3.ErrConstraintUnique {
					return domain.ErrSessionAlreadyExists
				}
			}

			return fmt.Errorf("create session failed session_id=(%s): %w", session.ID, err)
		}
	} else {
		err := tx.QueryRowContext(ctx, query, session.ID, session.UserID, session.LastSeenAt, session.ExpiresAt).Scan(
			&session.ID,
			&session.UserID,
			&session.CreatedAt,
			&session.LastSeenAt,
			&session.ExpiresAt,
		)
		if err != nil {
			var sqlErr gosqlite3.Error
			if errors.As(err, &sqlErr) {
				if sqlErr.ExtendedCode == gosqlite3.ErrConstraintUnique {
					return domain.ErrSessionAlreadyExists
				}
			}

			return fmt.Errorf("create session failed session_id=(%s): %w", session.ID, err)
		}
	}

	return nil
}

func (r *SessionRepo) GetByID(ctx context.Context, sessionID uuid.UUID) (*domain.Session, error) {
	db := r.db.GetDB()

	query := `
		select id, user_id, created_at, lastseen_at, expires_at
		from sessions
		where id = ?;
	`

	var session domain.Session
	err := db.QueryRowContext(ctx, query, sessionID).Scan(
		&session.ID,
		&session.UserID,
		&session.CreatedAt,
		&session.LastSeenAt,
		&session.ExpiresAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrSessionNotFound
		}

		return nil, fmt.Errorf("get session by id failed session_id=(%s): %w", sessionID, err)
	}

	return &session, nil
}

func (r *SessionRepo) UpdateLastSeen(ctx context.Context, sessionID uuid.UUID, lastSeenAt time.Time) error {
	db := r.db.GetDB()

	query := `
		update sessions
		set lastseen_at = ?
		where id = ?;
	`

	res, err := db.ExecContext(ctx, query, lastSeenAt, sessionID)
	if err != nil {
		return fmt.Errorf("update last seen failed session_id=(%s): %w", sessionID, err)
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("get last seen affected rows failed: %w", err)
	}
	if rows == 0 {
		return domain.ErrSessionNotFound
	}

	return nil
}

func (r *SessionRepo) DeleteByID(ctx context.Context, sessionID uuid.UUID) error {
	db := r.db.GetDB()

	query := `
		delete from sessions
		where id = ?;
	`

	res, err := db.ExecContext(ctx, query, sessionID)
	if err != nil {
		return fmt.Errorf("delete session failed by session_id=(%s): %w", sessionID, err)
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("get session affected rows failed: %w", err)
	} else {
		if rows == 0 {
			return domain.ErrSessionNotFound
		}
	}

	return nil
}

func (r *SessionRepo) DeleteAllUserSessions(ctx context.Context, userID uuid.UUID) error {
	db := r.db.GetDB()

	query := `
		delete from sessions
		where user_id = ?;
	`

	_, err := db.ExecContext(ctx, query, userID)
	if err != nil {
		return fmt.Errorf("delete sessions failed by user_id(%s): %w", userID, err)
	}

	return nil
}

func (r *SessionRepo) DeleteExpiredSessions(ctx context.Context, expiresAt, lastSeenAt time.Time) error {
	db := r.db.GetDB()

	query := `
		delete from sessions
		where expires_at <= ?
			or lastseen_at <= ?;
	`

	_, err := db.ExecContext(ctx, query, expiresAt, lastSeenAt)
	if err != nil {
		return fmt.Errorf("delete expired session failed: %w", err)
	}

	return nil
}

func (r *SessionRepo) CreateUserWithSession(
	ctx context.Context,
	sessionDuration time.Duration,
	newUser *user.User,
	userRepo user.UserRepository,
) (domain.Session, error) {
	db := r.db.GetDB()

	// создаем транзакцию
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return domain.Session{}, err
	}

	// до return, tx сделает RollBack
	// мы не расматриваем ее ошибку потому что если commit будет успешен он вернет ошибку то что транзакция закрыта
	// он нужен для того чтобы при ошибке сделать откат
	// так писать удобнее, не надо перед каждым err делать откат
	defer func() {
		_ = tx.Rollback()
	}()

	createdUser, err := userRepo.Create(ctx, newUser, tx)
	if err != nil {
		return domain.Session{}, err
	}

	now := time.Now().UTC()

	session := domain.Session{
		UserID:     createdUser.ID,
		LastSeenAt: now,
		ExpiresAt:  now.Add(sessionDuration),
	}

	err = r.Create(ctx, &session, tx)
	if err != nil {
		return domain.Session{}, err
	}

	if err := tx.Commit(); err != nil {
		return domain.Session{}, nil
	}

	return session, nil
}
