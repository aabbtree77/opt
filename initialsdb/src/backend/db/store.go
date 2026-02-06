package db

import (
	"database/sql"
)

// Store is a thin wrapper around all sqlc-generated query sets.
// You can add more interfaces here in future (sessions, logs, metadata, etc.)
type Store struct {
	*Queries
	db *sql.DB
}

// NewStore returns a new Store object with a live DB handle.
func NewStore(db *sql.DB) *Store {
	return &Store{
		Queries: New(db),
		db:      db,
	}
}

// DB returns the underlying *sql.DB (useful for migrations)
func (s *Store) DB() *sql.DB {
	return s.db
}

func (s *Store) WithTx(tx *sql.Tx) *Store {
	return &Store{
		Queries: New(tx),
		db:      s.db,
	}
}
