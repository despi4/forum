package utils

import (
	"database/sql"
	"errors"
	"log"
	"os"
	"sort"
	"strings"
)

// transaction errors
var (
	ErrCreateTx error = errors.New("transaction creating is failed")
	ErrRollback error = errors.New("transaction rollback is failed")
)

func createIfNotExist(db *sql.DB) error {
	query := `CREATE TABLE IF NOT EXISTS schema_migrations (
    			id INTEGER PRIMARY KEY AUTOINCREMENT,
   				migration_version TEXT NOT NULL UNIQUE,
    			applied_at DATETIME DEFAULT (datetime('now', 'localtime'))
	);`

	_, err := db.Exec(query)
	if err != nil {
		return err
	}

	return nil
}

func addNewMigration(tx *sql.Tx, migrationName string) error {
	query := `
		insert into schema_migrations (migration_version)
		values (?);
	`

	_, err := tx.Exec(query, migrationName)
	if err != nil {
		return err
	}

	return nil
}

func readDir(directoryPath string) ([]string, error) {
	files, err := os.ReadDir(directoryPath)
	if err != nil {
		return []string{}, err
	}

	var migrationFile []string

	for _, file := range files {
		if strings.Contains(file.Name(), ".up.sql") {
			migrationFile = append(migrationFile, file.Name())
		}
	}

	sort.Slice(migrationFile, func(i, j int) bool {
		return migrationFile[i] < migrationFile[j]
	})

	return migrationFile, nil
}

func checkMigration(db *sql.DB, fileName string) (bool, error) {
	var exists bool
	query := `
	select exists(
		select 1 
		from schema_migrations 
		where migration_version = ?
	)`

	err := db.QueryRow(query, fileName).Scan(&exists)
	if err != nil {
		return false, err
	}

	return exists, nil
}

func readFile(fileName string) (string, error) {
	file, err := os.ReadFile(fileName)
	if err != nil {
		return "", err
	}

	return string(file), nil
}

func RunMigrator(db *sql.DB, migrationDir string) error {
	if err := createIfNotExist(db); err != nil {
		return err
	}

	migrationFiles, err := readDir(migrationDir)
	if err != nil {
		return err
	}

	for _, migrationFile := range migrationFiles {
		exist, err := checkMigration(db, migrationFile)
		if err != nil {
			return err
		}

		if !exist {
			tx, err := db.Begin()
			if err != nil {
				return ErrCreateTx
			}

			query, err := readFile(migrationDir + "/" + migrationFile)
			if err != nil {
				txErr := tx.Rollback()
				if txErr != nil {
					return ErrRollback
				}

				return err
			}

			_, err = tx.Exec(query)
			if err != nil {
				txErr := tx.Rollback()
				if txErr != nil {
					return ErrRollback
				}

				return err
			}

			if err := addNewMigration(tx, migrationFile); err != nil {
				txErr := tx.Rollback()
				if txErr != nil {
					return ErrRollback
				}

				return err
			}

			err = tx.Commit()
			if err != nil {
				return err
			}
		}
	}

	log.Println("Migration executed successfully")

	return nil
}
