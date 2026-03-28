package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/g123udini/gophkeeper/internal/client/model"

	_ "github.com/mattn/go-sqlite3"
)

type UserDataRepository struct {
	db *sql.DB
}

func NewUserDataRepository(db *sql.DB) (*UserDataRepository, error) {
	repo := &UserDataRepository{db: db}
	if err := repo.init(); err != nil {
		return nil, err
	}

	return repo, nil
}

func (r *UserDataRepository) Upsert(ctx context.Context, data *model.UserData) error {
	query := `
		INSERT INTO user_data (data_key, data_value, updated_at, deleted_at)
		VALUES (?, ?, ?, ?)
		ON CONFLICT(data_key) DO UPDATE SET
			data_value=excluded.data_value,
			updated_at=excluded.updated_at,
			deleted_at=excluded.deleted_at
	`

	_, err := r.db.ExecContext(
		ctx, query,
		data.DataKey,
		data.DataValue,
		data.UpdatedAt.Unix(),
		data.DeletedAt.Unix(),
	)

	return err
}

func (r *UserDataRepository) Get(ctx context.Context, key string) (*model.UserData, error) {
	query := `SELECT id, data_key, data_value, updated_at, deleted_at FROM user_data WHERE data_key=?`
	row := r.db.QueryRowContext(ctx, query, key)
	d := &model.UserData{}

	var updatedAt, deletedAt int64
	err := row.Scan(&d.ID, &d.DataKey, &d.DataValue, &updatedAt, &deletedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	d.UpdatedAt = time.Unix(updatedAt, 0)
	d.DeletedAt = time.Unix(deletedAt, 0)

	return d, nil
}

func (r *UserDataRepository) GetUpdates(ctx context.Context, after time.Time) ([]*model.UserData, error) {
	rows, err := r.db.QueryContext(ctx, `
	SELECT id, data_key, data_value, updated_at, deleted_at 
	FROM user_data 
	WHERE updated_at > ?
`, after.Unix())
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var result []*model.UserData
	for rows.Next() {
		var ud model.UserData
		var updatedAt, deletedAt int64
		if err := rows.Scan(&ud.ID, &ud.DataKey, &ud.DataValue, &updatedAt, &deletedAt); err != nil {
			return nil, err
		}
		ud.UpdatedAt = time.Unix(updatedAt, 0)
		ud.DeletedAt = time.Unix(deletedAt, 0)
		result = append(result, &ud)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return result, nil
}

func (r *UserDataRepository) init() error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS user_data (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			data_key TEXT NOT NULL UNIQUE,
			data_value BLOB NOT NULL,
			updated_at INTEGER NOT NULL,
			deleted_at INTEGER NOT NULL
		)`,
		`CREATE INDEX IF NOT EXISTS idx_updated_at ON user_data(updated_at)`,
		`CREATE INDEX IF NOT EXISTS idx_deleted_at ON user_data(deleted_at)`,
	}

	for _, query := range queries {
		_, err := r.db.Exec(query)

		if err != nil {
			return err
		}
	}

	return nil
}
