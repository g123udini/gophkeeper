package repository

import (
	"database/sql"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/g123udini/gophkeeper/internal/server/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUserRepository_CreateUser(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewUserRepository(db)

	login := "testuser"
	passwordHash := "hashed_password"
	masterPasswordHash := "hashed_master_password"

	mock.ExpectExec("INSERT INTO user").
		WithArgs(login, passwordHash, masterPasswordHash).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = repo.CreateUser(login, passwordHash, masterPasswordHash)
	assert.NoError(t, err)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %s", err)
	}
}

func TestUserRepository_CreateUser_Error(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewUserRepository(db)

	login := "testuser"
	passwordHash := "hashed_password"
	masterPasswordHash := "hashed_master_password"

	expectedError := errors.New("db error")
	mock.ExpectExec("INSERT INTO user").
		WithArgs(login, passwordHash, masterPasswordHash).
		WillReturnError(expectedError)

	err = repo.CreateUser(login, passwordHash, masterPasswordHash)
	assert.Error(t, err)
	assert.Equal(t, expectedError, err)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %s", err)
	}
}

func TestUserRepository_GetUserByLogin_Found(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewUserRepository(db)
	login := "testuser"

	rows := sqlmock.NewRows([]string{"id", "login", "password_hash", "master_password_hash"}).
		AddRow(uint32(1), login, "hashed_password", "hashed_master_password")

	mock.ExpectQuery("SELECT id, login, password_hash, master_password_hash FROM user WHERE login").
		WithArgs(login).
		WillReturnRows(rows)

	user, err := repo.GetUserByLogin(login)
	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, uint32(1), user.ID)
	assert.Equal(t, login, user.Login)
	assert.Equal(t, "hashed_password", user.PasswordHash)
	assert.Equal(t, "hashed_master_password", user.MasterPasswordHash)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %s", err)
	}
}

func TestUserRepository_GetUserByLogin_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewUserRepository(db)
	login := "nonexistentuser"

	mock.ExpectQuery("SELECT id, login, password_hash, master_password_hash FROM user WHERE login").
		WithArgs(login).
		WillReturnError(sql.ErrNoRows)

	user, err := repo.GetUserByLogin(login)
	assert.NoError(t, err)
	assert.Nil(t, user)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %s", err)
	}
}

func TestUserRepository_GetUserByLogin_Error(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewUserRepository(db)
	login := "testuser"
	expectedError := errors.New("db error")

	mock.ExpectQuery("SELECT id, login, password_hash, master_password_hash FROM user WHERE login").
		WithArgs(login).
		WillReturnError(expectedError)

	user, err := repo.GetUserByLogin(login)
	assert.Error(t, err)
	assert.Equal(t, user, &model.User{})
	assert.Equal(t, expectedError, err)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %s", err)
	}
}
