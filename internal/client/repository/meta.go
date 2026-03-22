package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type MetaRepository struct {
	db *sql.DB
}

func NewMetaRepository(db *sql.DB) (*MetaRepository, error) {
	repo := &MetaRepository{db: db}
	if err := repo.init(); err != nil {
		return nil, err
	}

	return repo, nil
}

func (r *MetaRepository) GetToken(ctx context.Context) (string, error) {
	var token string
	err := r.db.QueryRowContext(ctx, "SELECT token FROM meta WHERE id = 0").Scan(&token)
	if err != nil {
		return "", err
	}

	return token, nil
}

func (r *MetaRepository) SetToken(ctx context.Context, token string) error {
	_, err := r.db.ExecContext(ctx, "UPDATE meta SET token = ? WHERE id = 0", token)

	return err
}

func (r *MetaRepository) GetLastSync(ctx context.Context) (time.Time, error) {
	var tsInt int64
	err := r.db.QueryRowContext(ctx, "SELECT last_sync FROM meta WHERE id = 0").Scan(&tsInt)
	if err != nil {
		return time.Time{}, err
	}

	return time.Unix(tsInt, 0).UTC(), nil
}

func (r *MetaRepository) SetLastSync(ctx context.Context, t time.Time) error {
	_, err := r.db.ExecContext(ctx, "UPDATE meta SET last_sync = ? WHERE id = 0", t.Unix())

	return err
}

func (r *MetaRepository) GetMasterPasswordHash(ctx context.Context) (string, error) {
	var h string
	err := r.db.QueryRowContext(ctx, "SELECT master_password_hash FROM meta WHERE id = 0").Scan(&h)
	if err != nil {
		return "", err
	}

	return h, nil
}

func (r *MetaRepository) SetMasterPasswordHash(ctx context.Context, h string) error {
	_, err := r.db.ExecContext(ctx, "UPDATE meta SET master_password_hash = ? WHERE id = 0", h)

	return err
}

func (r *MetaRepository) init() error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS meta (
			id INTEGER PRIMARY KEY CHECK (id = 0),
			last_sync INTEGER NOT NULL,
			master_password_hash TEXT NOT NULL,
			token TEXT NOT NULL DEFAULT ""
		)`,
		`INSERT OR IGNORE INTO meta (id, last_sync, master_password_hash, token) VALUES (0, 0, "", "")`,
	}

	for _, query := range queries {
		_, err := r.db.Exec(query)
		if err != nil {
			return err
		}
	}

	if err := r.ensureTokenColumn(); err != nil {
		return err
	}

	return nil
}

func (r *MetaRepository) ensureTokenColumn() error {
	_, err := r.db.Exec(`ALTER TABLE meta ADD COLUMN token TEXT NOT NULL DEFAULT ""`)
	if err != nil {
		// SQLite вернёт ошибку, если колонка уже есть — это нормальный случай
		if errors.Is(err, sql.ErrNoRows) {
			return nil
		}

		// SQLite не даёт typed error для duplicate column, поэтому гасим по тексту не будем.
		// Раз ALTER TABLE не обязателен для новой схемы, просто проверим, что колонка уже читается.
		var token string
		checkErr := r.db.QueryRow(`SELECT token FROM meta WHERE id = 0`).Scan(&token)
		if checkErr == nil {
			return nil
		}

		return err
	}

	return nil
}
