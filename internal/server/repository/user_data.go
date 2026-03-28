package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/g123udini/gophkeeper/internal/server/model"
)

type UserDataRepository struct {
	db *sql.DB
}

func NewUserDataRepository(db *sql.DB) *UserDataRepository {
	return &UserDataRepository{db: db}
}

func (r *UserDataRepository) Upsert(ctx context.Context, data *model.UserData) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		} else {
			tx.Commit()
		}
	}()

	var currentVersion time.Time
	err = tx.QueryRowContext(ctx,
		"SELECT updated_at FROM user_data WHERE user_id = ? AND data_key = ?",
		data.UserID, data.DataKey,
	).Scan(&currentVersion)

	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return err
	}

	if errors.Is(err, sql.ErrNoRows) {
		_, err = tx.ExecContext(ctx, `
			INSERT INTO user_data 
				(user_id, data_key, data_value, updated_at, deleted_at) 
			VALUES 
				(?, ?, ?, ?, ?)
		`, data.UserID, data.DataKey, data.DataValue, data.UpdatedAt.Format(time.DateTime), data.DeletedAt.Format(time.DateTime))

		if err != nil {
			return err
		}
	} else {
		if data.UpdatedAt.Before(currentVersion) {
			return nil
		}

		_, err = tx.ExecContext(ctx, `
		UPDATE user_data 
		SET 
			data_value = ?,
			updated_at = ?,
			deleted_at = ?
		WHERE user_id = ? AND data_key = ?
	`, data.DataValue, data.UpdatedAt.Format(time.DateTime), data.DeletedAt.Format(time.DateTime), data.UserID, data.DataKey)
		if err != nil {
			return err
		}
	}

	return nil
}

func (r *UserDataRepository) GetUpdates(ctx context.Context, userID uint32, since time.Time) ([]*model.UserData, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, user_id, data_key, data_value, updated_at, deleted_at
		 FROM user_data
		 WHERE user_id = ? AND srv_updated_at > ?`,
		userID, since.Format(time.DateTime),
	)
	if err != nil {
		return nil, err
	}
	if rows.Err() != nil {
		return nil, rows.Err()
	}

	defer rows.Close()

	var result []*model.UserData
	for rows.Next() {
		d := &model.UserData{}
		err = rows.Scan(
			&d.ID,
			&d.UserID,
			&d.DataKey,
			&d.DataValue,
			&d.UpdatedAt,
			&d.DeletedAt,
		)
		if err != nil {
			return nil, err
		}
		result = append(result, d)
	}
	return result, nil
}
