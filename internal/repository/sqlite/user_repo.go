package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	domain "01.tomorrow-school.ai/git/amadiuly/forum/internal/domain/user"
	"01.tomorrow-school.ai/git/amadiuly/forum/internal/repository"

	"01.tomorrow-school.ai/git/amadiuly/forum/internal/db"
	"github.com/google/uuid"

	gosqlite3 "github.com/mattn/go-sqlite3"
)

type UserRepo struct {
	db *db.ConnDB
}

func NewUserRepo(db *db.ConnDB) *UserRepo {
	return &UserRepo{db: db}
}

func (r *UserRepo) Create(ctx context.Context, user *domain.User) (domain.User, error) {
	db := r.db.GetDB()

	query := `
		insert into users (id, username, email, role, password_hash, visibility)
		values (?, ?, ?, ?, ?, ?)
		returning id, username, email, created_at, updated_at, role, password_hash, visibility;
	`

	if user.ID == uuid.Nil {
		user.ID = uuid.New()
	}

	if user.Role == "" {
		user.Role = domain.RoleUser
	}

	if user.Visibility == "" {
		user.Visibility = domain.VisibilityPublic
	}

	user.Email = strings.ToLower(strings.TrimSpace(user.Email))

	user.Username = strings.ToLower(strings.TrimSpace(user.Username))

	err := db.QueryRowContext(ctx, query, user.ID, user.Username, user.Email, user.Role, user.PasswordHash, user.Visibility).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.Role,
		&user.PasswordHash,
		&user.Visibility,
	)
	if err != nil {
		var sqlErr gosqlite3.Error
		if errors.As(err, &sqlErr) {
			if sqlErr.ExtendedCode == gosqlite3.ErrConstraintUnique {
				return domain.User{}, domain.ErrUserAlreadyExists
			}
		}

		return domain.User{}, fmt.Errorf("create user failed user_id=(%s): %w", user.ID, err)
	}

	return *user, nil
}

func (r *UserRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	db := r.db.GetDB()

	query := `
		select id, username, email, role, password_hash, created_at, updated_at, visibility
		from users
		where id = ?;
	`

	var user domain.User
	err := db.QueryRowContext(ctx, query, id).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.Role,
		&user.PasswordHash,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.Visibility,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrUserNotFound
		}

		return nil, fmt.Errorf("get user by id failed user_id=(%s): %w", id, err)
	}

	return &user, nil
}

func (r *UserRepo) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	db := r.db.GetDB()

	query := `
		select id, username, email, role, password_hash, created_at, updated_at, visibility
		from users
		where email = ?;
	`

	email = strings.ToLower(strings.TrimSpace(email))

	var user domain.User
	err := db.QueryRowContext(ctx, query, email).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.Role,
		&user.PasswordHash,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.Visibility,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrUserNotFound
		}

		return nil, fmt.Errorf("get user by email failed user_email=(%s): %w", email, err)
	}

	return &user, nil
}

func (r *UserRepo) GetByUsername(ctx context.Context, username string) (*domain.User, error) {
	var (
		user domain.User
		err  error
	)

	username = strings.TrimSpace(username)

	db := r.db.GetDB()

	query := `
		select id, username, email, role, password_hash, created_at, updated_at, visibility
		from users
		where username = ?;
	`

	err = db.QueryRowContext(ctx, query, username).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.Role,
		&user.PasswordHash,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.Visibility,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrUserNotFound
		}

		return nil, fmt.Errorf("get user by username failed username=(%s): %w", username, err)
	}

	return &user, err
}

func (r *UserRepo) Update(ctx context.Context, updateRows domain.UserUpdate, userID uuid.UUID) error {
	db := r.db.GetDB()

	query, args, err := buildUpdateUserQuery(updateRows, userID)
	if err != nil {
		return fmt.Errorf("updated user failed user_id=(%s): %w", userID, err)
	}

	cmd, err := db.ExecContext(ctx, query, args...)
	if err != nil {
		var sqlErr gosqlite3.Error
		if errors.As(err, &sqlErr) {
			if sqlErr.ExtendedCode == gosqlite3.ErrNoExtended(gosqlite3.ErrNotFound) {
				return domain.ErrUserNotFound
			}

			if sqlErr.ExtendedCode == gosqlite3.ErrConstraintUnique {
				return domain.ErrUserAlreadyExists
			}
		}

		return fmt.Errorf("update user failed by user_id=(%s), %w", userID, err)
	}

	rows, err := cmd.RowsAffected()
	if err != nil {
		return fmt.Errorf("update user failed by user_id=(%s), %w", userID, err)
	}

	if rows == 0 {
		return domain.ErrUserNotFound
	}

	return nil
}

func (r *UserRepo) Delete(ctx context.Context, id uuid.UUID) error {
	db := r.db.GetDB()

	query := `
		delete from users
		where id = ?;
	`

	res, err := db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("delete user failed by user_id=(%s): %w", id, err)
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("get user affected rows failed: %w", err)
	} else {
		if rows == 0 {
			return domain.ErrUserNotFound
		}
	}

	return nil
}

func (r *UserRepo) List(ctx context.Context, filter domain.UserFilter) ([]domain.User, error) {
	db := r.db.GetDB()

	query, args := buildListUsersQuery(filter)

	rows, err := db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("get users failed: %w", err)
	}
	defer rows.Close()

	capacity := 0
	if filter.Limit > 0 {
		capacity = filter.Limit
	}

	users := make([]domain.User, 0, capacity)

	for rows.Next() {
		var user domain.User

		err := rows.Scan(
			&user.ID,
			&user.Username,
			&user.Email,
			&user.Role,
			&user.Visibility,
			&user.CreatedAt,
			&user.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scan user rows failed %w", err)
		}

		users = append(users, user)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("increate rows failed %w", err)
	}

	return users, nil
}

func buildListUsersQuery(filter domain.UserFilter) (string, []any) {
	var sb strings.Builder
	q := repository.NewQueryParts()

	sb.WriteString(`
		select id, username, email, role, visibility, created_at, updated_at
		from users
	`)

	if filter.Role != nil {
		q.AddWhere("role = ?", *filter.Role)
	}

	if filter.Search != nil && strings.TrimSpace(*filter.Search) != "" {
		search := "%" + strings.TrimSpace(*filter.Search) + "%"
		q.AddWhere("(username like ? or email like ?)", search, search)
	}

	if filter.Limit <= 0 {
		filter.Limit = 50
	}

	if filter.Offset < 0 {
		filter.Offset = 0
	}

	sb.WriteString(q.BuildWhere())
	sb.WriteString(` order by created_at desc, id desc`)
	sb.WriteString(` limit ? offset ?;`)

	args := append(q.WhereArgs, filter.Limit, filter.Offset)

	return sb.String(), args
}

func buildUpdateUserQuery(params domain.UserUpdate, id uuid.UUID) (string, []any, error) {
	var sb strings.Builder // strings.Builder more effective and cheeper than += or connect strings with loop
	q := repository.NewQueryParts()

	sb.WriteString(`update users`)

	if params.Username != nil {
		q.AddSet("username = ?", *params.Username)
	}

	if params.Email != nil {
		q.AddSet("email = ?", *params.Email)
	}

	if params.Role != nil {
		q.AddSet("role = ?", *params.Role)
	}

	if params.Visibility != nil {
		q.AddSet("visibility = ?", *params.Visibility)
	}

	if params.Visibility != nil {
		q.AddSet("password_hash = ?", *params.PasswordHash)
	}

	if len(q.Set) == 0 {
		return "", nil, fmt.Errorf("no fields to update user")
	}

	q.AddWhere("id = ?", id)

	sb.WriteString(q.BuildSet())
	sb.WriteString(q.BuildWhere())

	args := append(q.SetArgs, q.WhereArgs...)

	return sb.String(), args, nil
}
