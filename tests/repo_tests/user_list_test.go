package tests

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"math/rand/v2"
	"testing"
	"time"

	dsqlite "01.tomorrow-school.ai/git/amadiuly/forum/internal/db"

	"01.tomorrow-school.ai/git/amadiuly/forum/internal/domain/user"
	"01.tomorrow-school.ai/git/amadiuly/forum/internal/repository/sqlite"
	"01.tomorrow-school.ai/git/amadiuly/forum/utils"
)

const (
	dsn           string = ":memory:"
	migrationsDir string = "../../migrations"
)

var (
	ErrFailedToConnect  = errors.New("failed to open sqlite connection: %v")
	ErrFailedMigrate    = errors.New("failed to run migrations: %v")
	ErrFailedInsertData = errors.New("failed to run migrations: %v")
	ErrRepoList         = errors.New("repo list error: %v")
)

func TestUserRepo_List_ByRole(t *testing.T) {
	conn, err := dsqlite.NewConnDB(dsn)
	if err != nil {
		t.Fatalf(ErrFailedToConnect.Error(), err)
	}
	defer conn.Close()

	db := conn.GetDB()

	err = utils.RunMigrator(db, migrationsDir)
	if err != nil {
		t.Fatalf(ErrFailedMigrate.Error(), err)
	}

	err = seedManyUsers(db, nil)
	if err != nil {
		t.Fatalf(ErrFailedInsertData.Error(), err)
	}

	role := user.RoleAdmin
	filter := user.UserFilter{
		Role:   &role,
		Limit:  100,
		Offset: 0,
	}

	repo := sqlite.NewUserRepo(conn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	users, err := repo.List(ctx, filter)
	if err != nil {
		t.Fatalf(ErrRepoList.Error(), err)
	}

	if len(users) == 0 {
		t.Fatal("expected at least one admin user, got zero")
	}

	for _, u := range users {
		if u.Role != user.RoleAdmin {
			t.Fatalf("expected role %q, got %q for user %s", user.RoleAdmin, u.Role, u.Username)
		}
	}
}

func TestUserRepo_List_BySearch(t *testing.T) {
	conn, err := dsqlite.NewConnDB(dsn)
	if err != nil {
		t.Fatalf(ErrFailedToConnect.Error(), err)
	}

	db := conn.GetDB()
	defer conn.Close()

	if err = utils.RunMigrator(db, migrationsDir); err != nil {
		t.Fatalf(ErrFailedMigrate.Error(), err)
	}

	search := "almadi"

	if err = seedManyUsers(db, &search); err != nil {
		t.Fatalf(ErrFailedInsertData.Error(), err)
	}

	filter := user.UserFilter{
		Search: &search,
		Limit:  100,
		Offset: 0,
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	repo := sqlite.NewUserRepo(conn)

	users, err := repo.List(ctx, filter)
	if err != nil {
		t.Fatalf(ErrRepoList.Error(), err)
	}

	if users == nil {
		t.Fatalf("expected at least one %s user, got zero", search)
	}

	for i, u := range users {
		contains := u.Username == fmt.Sprint(search, i) || u.Email == fmt.Sprint(search, i, "@gmail.com") || u.Email == fmt.Sprint(search, i, "@mail.com") || u.Email == fmt.Sprint(search, i, "@example.com")
		fmt.Printf("%d. Username: %s | Email: %s | Role: %s | PasswordHash: %s\n", i, u.Username, u.Email, u.Role, u.PasswordHash)

		if !contains {
			t.Fatalf("unexpected user returned for search %+v", u)
		}
	}

	fmt.Println(len(users))
}

func seedManyUsers(db *sql.DB, name *string) error {
	randNumber := rand.IntN(10)

	const (
		gmail   = "@gmail.com"
		example = "@example.com"
		mail    = "@mail.com"
	)

	switch randNumber {
	case 0:
		randNumber = 10
	default:
		randNumber *= 10
	}

	for i := 0; i < randNumber; i++ {
		var (
			username string
			email    string
		)

		if name == nil {
			username = fmt.Sprintf("user_%d", i)

			if i%3 == 0 {
				email = fmt.Sprintf("test_%d%s", i, gmail)
			} else if i%2 == 0 {
				email = fmt.Sprintf("test_%d%s", i, example)
			} else {
				email = fmt.Sprintf("test_%d%s", i, mail)
			}
		} else {
			username = fmt.Sprint(*name, i)

			if i%3 == 0 {
				email = fmt.Sprintf(username, i, gmail)
			} else if i%2 == 0 {
				email = fmt.Sprintf(username, i, example)
			} else {
				email = fmt.Sprintf(username, i, mail)
			}
		}

		password_hash := fmt.Sprintf(username+"%d", i)

		role := "user"
		if i%5 == 0 {
			role = "admin"
		}

		_, err := db.Exec(`INSERT INTO users (username, email, role, password_hash) VALUES (?, ?, ?, ?)`, username, email, role, password_hash)
		if err != nil {
			return fmt.Errorf("insert user failed: %w", err)
		}
	}

	return nil
}
