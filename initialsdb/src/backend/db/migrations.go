package db

import (
	"context"
	"database/sql"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

func RunMigrations(db *sql.DB) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Ensure migrations table exists
	_, err := db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS migrations (
			id TEXT PRIMARY KEY,
			applied_at TIMESTAMPTZ NOT NULL DEFAULT now()
		);
	`)
	if err != nil {
		return fmt.Errorf("create migrations table: %w", err)
	}

	files, err := readMigrationFiles(migrationsDir())
	if err != nil {
		return err
	}

	for _, path := range files {
		if err := applyOne(ctx, db, path); err != nil {
			return err
		}
	}

	return nil
}

func migrationsDir() string {
	// Container runtime (dev & prod)
	if _, err := os.Stat("/app/migrations"); err == nil {
		return "/app/migrations"
	}

	// Local development (run from src/)
	if _, err := os.Stat("migrations"); err == nil {
		return "migrations"
	}

	panic("migrations directory not found (expected /app/migrations or ./migrations)")
}

func readMigrationFiles(dir string) ([]string, error) {
	var out []string

	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if strings.HasSuffix(d.Name(), ".sql") {
			out = append(out, path)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	sort.Strings(out)
	return out, nil
}

func applyOne(ctx context.Context, db *sql.DB, path string) error {
	id := filepath.Base(path)

	var exists bool
	err := db.QueryRowContext(
		ctx,
		`SELECT EXISTS (SELECT 1 FROM migrations WHERE id = $1)`,
		id,
	).Scan(&exists)
	if err != nil {
		return fmt.Errorf("checking migration %s: %w", id, err)
	}
	if exists {
		return nil
	}

	sqlBytes, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read migration %s: %w", path, err)
	}

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx for migration %s: %w", id, err)
	}

	if _, err := tx.ExecContext(ctx, string(sqlBytes)); err != nil {
		tx.Rollback()
		return fmt.Errorf("apply migration %s: %w", id, err)
	}

	if _, err := tx.ExecContext(
		ctx,
		`INSERT INTO migrations (id) VALUES ($1)`,
		id,
	); err != nil {
		tx.Rollback()
		return fmt.Errorf("record migration %s: %w", id, err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit migration %s: %w", id, err)
	}

	return nil
}
