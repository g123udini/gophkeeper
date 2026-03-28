package repository

import (
	"context"
	"database/sql"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestDB(t *testing.T) *sql.DB {
	db, err := sql.Open("sqlite3", ":memory:")
	require.NoError(t, err)
	require.NotNil(t, db)
	return db
}

func TestNewMetaRepository(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo, err := NewMetaRepository(db)
	require.NoError(t, err)
	require.NotNil(t, repo)

	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM meta WHERE id = 0").Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 1, count)
}

func TestMetaRepository_GetSetLastSync(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo, err := NewMetaRepository(db)
	require.NoError(t, err)

	ctx := context.Background()

	initialTime, err := repo.GetLastSync(ctx)
	require.NoError(t, err)
	assert.Equal(t, time.Unix(0, 0).UTC(), initialTime)

	testTime := time.Date(2023, 5, 15, 10, 30, 0, 0, time.UTC)
	err = repo.SetLastSync(ctx, testTime)
	require.NoError(t, err)

	updatedTime, err := repo.GetLastSync(ctx)
	require.NoError(t, err)
	assert.Equal(t, testTime.Unix(), updatedTime.Unix())
}

func TestMetaRepository_GetSetMasterPasswordHash(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo, err := NewMetaRepository(db)
	require.NoError(t, err)

	ctx := context.Background()

	initialHash, err := repo.GetMasterPasswordHash(ctx)
	require.NoError(t, err)
	assert.Equal(t, "", initialHash)

	testHash := "hash123456"
	err = repo.SetMasterPasswordHash(ctx, testHash)
	require.NoError(t, err)

	updatedHash, err := repo.GetMasterPasswordHash(ctx)
	require.NoError(t, err)
	assert.Equal(t, testHash, updatedHash)
}

func TestMetaRepository_Init(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := &MetaRepository{db: db}
	err := repo.init()
	require.NoError(t, err)

	var tableExists bool
	err = db.QueryRow("SELECT 1 FROM sqlite_master WHERE type='table' AND name='meta'").Scan(&tableExists)
	require.NoError(t, err)
	assert.True(t, tableExists)

	var lastSync int64
	var masterPasswordHash string
	err = db.QueryRow("SELECT last_sync, master_password_hash FROM meta WHERE id = 0").Scan(&lastSync, &masterPasswordHash)
	require.NoError(t, err)
	assert.Equal(t, int64(0), lastSync)
	assert.Equal(t, "", masterPasswordHash)

	err = repo.init()
	require.NoError(t, err)
}
