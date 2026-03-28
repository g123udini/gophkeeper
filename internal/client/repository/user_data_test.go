package repository

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/g123udini/gophkeeper/internal/client/model"
	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupUserDataTestDB(t *testing.T) *sql.DB {
	db, err := sql.Open("sqlite3", ":memory:")
	require.NoError(t, err)
	require.NotNil(t, db)
	return db
}

func TestNewUserDataRepository(t *testing.T) {
	db := setupUserDataTestDB(t)
	defer db.Close()

	repo, err := NewUserDataRepository(db)
	require.NoError(t, err)
	require.NotNil(t, repo)

	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='user_data'").Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 1, count)
}

func TestUserDataRepository_Upsert(t *testing.T) {
	db := setupUserDataTestDB(t)
	defer db.Close()

	repo, err := NewUserDataRepository(db)
	require.NoError(t, err)

	ctx := context.Background()

	now := time.Now().UTC()
	data := &model.UserData{
		DataKey:   "test-key",
		DataValue: []byte("test-value"),
		UpdatedAt: now,
		DeletedAt: time.Unix(0, 0),
	}

	err = repo.Upsert(ctx, data)
	require.NoError(t, err)

	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM user_data WHERE data_key = ?", data.DataKey).Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 1, count)

	data.DataValue = []byte("updated-value")
	data.UpdatedAt = now.Add(time.Hour)

	err = repo.Upsert(ctx, data)
	require.NoError(t, err)

	var value []byte
	var updatedAt int64
	err = db.QueryRow("SELECT data_value, updated_at FROM user_data WHERE data_key = ?", data.DataKey).Scan(&value, &updatedAt)
	require.NoError(t, err)
	assert.Equal(t, []byte("updated-value"), value)
	assert.Equal(t, data.UpdatedAt.Unix(), updatedAt)
}

func TestUserDataRepository_Get(t *testing.T) {
	db := setupUserDataTestDB(t)
	defer db.Close()

	repo, err := NewUserDataRepository(db)
	require.NoError(t, err)

	ctx := context.Background()

	now := time.Now().UTC()
	data := &model.UserData{
		DataKey:   "test-key",
		DataValue: []byte("test-value"),
		UpdatedAt: now,
		DeletedAt: time.Unix(0, 0),
	}

	err = repo.Upsert(ctx, data)
	require.NoError(t, err)

	result, err := repo.Get(ctx, data.DataKey)
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, data.DataKey, result.DataKey)
	assert.Equal(t, data.DataValue, result.DataValue)
	assert.Equal(t, data.UpdatedAt.Unix(), result.UpdatedAt.Unix())
	assert.Equal(t, data.DeletedAt.Unix(), result.DeletedAt.Unix())

	nonExisting, err := repo.Get(ctx, "non-existing-key")
	require.NoError(t, err)
	assert.Nil(t, nonExisting)
}

func TestUserDataRepository_GetUpdates(t *testing.T) {
	db := setupUserDataTestDB(t)
	defer db.Close()

	repo, err := NewUserDataRepository(db)
	require.NoError(t, err)

	ctx := context.Background()

	baseTime := time.Now().UTC()
	testData := []*model.UserData{
		{
			DataKey:   "key1",
			DataValue: []byte("value1"),
			UpdatedAt: baseTime,
			DeletedAt: time.Unix(0, 0),
		},
		{
			DataKey:   "key2",
			DataValue: []byte("value2"),
			UpdatedAt: baseTime.Add(time.Hour),
			DeletedAt: time.Unix(0, 0),
		},
		{
			DataKey:   "key3",
			DataValue: []byte("value3"),
			UpdatedAt: baseTime.Add(2 * time.Hour),
			DeletedAt: time.Unix(0, 0),
		},
	}

	for _, data := range testData {
		err = repo.Upsert(ctx, data)
		require.NoError(t, err)
	}

	updates, err := repo.GetUpdates(ctx, baseTime.Add(30*time.Minute))
	require.NoError(t, err)
	assert.Len(t, updates, 2)

	keys := make(map[string]bool)
	for _, update := range updates {
		keys[update.DataKey] = true
	}
	assert.True(t, keys["key2"])
	assert.True(t, keys["key3"])
	assert.False(t, keys["key1"])

	allUpdates, err := repo.GetUpdates(ctx, baseTime.Add(-time.Hour))
	require.NoError(t, err)
	assert.Len(t, allUpdates, 3)
}

func TestUserDataRepository_Init(t *testing.T) {
	db := setupUserDataTestDB(t)
	defer db.Close()

	repo := &UserDataRepository{db: db}
	err := repo.init()
	require.NoError(t, err)

	var tableExists int
	err = db.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='user_data'").Scan(&tableExists)
	require.NoError(t, err)
	assert.Equal(t, 1, tableExists)

	var indexCount int
	err = db.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='index' AND name IN ('idx_updated_at', 'idx_deleted_at')").Scan(&indexCount)
	require.NoError(t, err)
	assert.Equal(t, 2, indexCount)

	err = repo.init()
	require.NoError(t, err)
}
