package repository

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/g123udini/gophkeeper/internal/server/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUserDataRepository_Upsert_Insert(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewUserDataRepository(db)
	ctx := context.Background()

	now := time.Now()
	deletedAt := time.Now().Add(time.Hour)
	data := &model.UserData{
		UserID:    1,
		DataKey:   "test-key",
		DataValue: []byte("test-value"),
		UpdatedAt: now,
		DeletedAt: deletedAt,
	}

	mock.ExpectBegin()
	mock.ExpectQuery("SELECT updated_at FROM user_data WHERE user_id").
		WithArgs(data.UserID, data.DataKey).
		WillReturnError(sql.ErrNoRows)

	mock.ExpectExec("INSERT INTO user_data").
		WithArgs(data.UserID, data.DataKey, data.DataValue, data.UpdatedAt.Format(time.DateTime), data.DeletedAt.Format(time.DateTime)).
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectCommit()

	err = repo.Upsert(ctx, data)
	assert.NoError(t, err)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %s", err)
	}
}

func TestUserDataRepository_Upsert_Update(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewUserDataRepository(db)
	ctx := context.Background()

	oldTime := time.Now().Add(-time.Hour)
	now := time.Now()
	deletedAt := time.Now().Add(time.Hour)
	data := &model.UserData{
		UserID:    1,
		DataKey:   "test-key",
		DataValue: []byte("new-value"),
		UpdatedAt: now,
		DeletedAt: deletedAt,
	}

	rows := sqlmock.NewRows([]string{"updated_at"}).
		AddRow(oldTime)

	mock.ExpectBegin()
	mock.ExpectQuery("SELECT updated_at FROM user_data WHERE user_id").
		WithArgs(data.UserID, data.DataKey).
		WillReturnRows(rows)

	mock.ExpectExec("UPDATE user_data").
		WithArgs(data.DataValue, data.UpdatedAt.Format(time.DateTime), data.DeletedAt.Format(time.DateTime), data.UserID, data.DataKey).
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectCommit()

	err = repo.Upsert(ctx, data)
	assert.NoError(t, err)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %s", err)
	}
}

func TestUserDataRepository_Upsert_OlderUpdate(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewUserDataRepository(db)
	ctx := context.Background()

	newTime := time.Now()
	oldTime := time.Now().Add(-time.Hour)
	data := &model.UserData{
		UserID:    1,
		DataKey:   "test-key",
		DataValue: []byte("old-value"),
		UpdatedAt: oldTime,
		DeletedAt: oldTime,
	}

	rows := sqlmock.NewRows([]string{"updated_at"}).
		AddRow(newTime)

	mock.ExpectBegin()
	mock.ExpectQuery("SELECT updated_at FROM user_data WHERE user_id").
		WithArgs(data.UserID, data.DataKey).
		WillReturnRows(rows)

	mock.ExpectCommit()

	err = repo.Upsert(ctx, data)
	assert.NoError(t, err)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %s", err)
	}
}

func TestUserDataRepository_Upsert_QueryError(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewUserDataRepository(db)
	ctx := context.Background()

	now := time.Now()
	data := &model.UserData{
		UserID:    1,
		DataKey:   "test-key",
		DataValue: []byte("test-value"),
		UpdatedAt: now,
	}

	expectedError := errors.New("db error")
	mock.ExpectBegin()
	mock.ExpectQuery("SELECT updated_at FROM user_data WHERE user_id").
		WithArgs(data.UserID, data.DataKey).
		WillReturnError(expectedError)

	mock.ExpectRollback()

	err = repo.Upsert(ctx, data)
	assert.Error(t, err)
	assert.Equal(t, expectedError, err)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %s", err)
	}
}

func TestUserDataRepository_GetUpdates(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewUserDataRepository(db)
	ctx := context.Background()
	userID := uint32(1)
	since := time.Now().Add(-24 * time.Hour)

	now := time.Now()
	deletedAt := time.Now().Add(time.Hour)

	rows := sqlmock.NewRows([]string{"id", "user_id", "data_key", "data_value", "updated_at", "deleted_at"}).
		AddRow(1, userID, "key1", []byte("value1"), now, deletedAt).
		AddRow(2, userID, "key2", []byte("value2"), now, deletedAt)

	mock.ExpectQuery("SELECT id, user_id, data_key, data_value, updated_at, deleted_at FROM user_data WHERE").
		WithArgs(userID, since.Format(time.DateTime)).
		WillReturnRows(rows)

	results, err := repo.GetUpdates(ctx, userID, since)
	assert.NoError(t, err)
	assert.Len(t, results, 2)
	assert.Equal(t, uint32(1), results[0].ID)
	assert.Equal(t, "key1", results[0].DataKey)
	assert.Equal(t, []byte("value1"), results[0].DataValue)
	assert.Equal(t, now, results[0].UpdatedAt)
	assert.Equal(t, deletedAt, results[0].DeletedAt)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %s", err)
	}
}

func TestUserDataRepository_GetUpdates_Error(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewUserDataRepository(db)
	ctx := context.Background()
	userID := uint32(1)
	since := time.Now().Add(-24 * time.Hour)

	expectedError := errors.New("db error")
	mock.ExpectQuery("SELECT id, user_id, data_key, data_value, updated_at, deleted_at FROM user_data WHERE").
		WithArgs(userID, since.Format(time.DateTime)).
		WillReturnError(expectedError)

	results, err := repo.GetUpdates(ctx, userID, since)
	assert.Error(t, err)
	assert.Equal(t, expectedError, err)
	assert.Nil(t, results)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %s", err)
	}
}

func TestUserDataRepository_GetUpdates_ScanError(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewUserDataRepository(db)
	ctx := context.Background()
	userID := uint32(1)
	since := time.Now().Add(-24 * time.Hour)

	// Ошибка при сканировании из-за несоответствия типов
	rows := sqlmock.NewRows([]string{"id", "user_id", "data_key", "data_value", "updated_at", "deleted_at"}).
		AddRow("not-a-number", userID, "key1", []byte("value1"), time.Now(), time.Now())

	mock.ExpectQuery("SELECT id, user_id, data_key, data_value, updated_at, deleted_at FROM user_data WHERE").
		WithArgs(userID, since.Format(time.DateTime)).
		WillReturnRows(rows)

	results, err := repo.GetUpdates(ctx, userID, since)
	assert.Error(t, err)
	assert.Nil(t, results)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %s", err)
	}
}
