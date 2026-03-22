package repository

import (
	"database/sql"
	"errors"

	"github.com/g123udini/gophkeeper/internal/server/model"
)

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) CreateUser(login, passwordHash, masterPasswordHash string) error {
	_, err := r.db.Exec(
		"INSERT INTO user (login, password_hash, master_password_hash) VALUES (?, ?, ?)",
		login, passwordHash, masterPasswordHash,
	)
	return err
}

func (r *UserRepository) GetUserByLogin(login string) (*model.User, error) {
	u := &model.User{}
	err := r.db.QueryRow(
		"SELECT id, login, password_hash, master_password_hash FROM user WHERE login = ?",
		login,
	).Scan(&u.ID, &u.Login, &u.PasswordHash, &u.MasterPasswordHash)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	return u, err
}
