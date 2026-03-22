package service

import (
	"errors"

	"github.com/g123udini/gophkeeper/internal/server/jwt"
	"github.com/g123udini/gophkeeper/internal/server/model"
	"github.com/g123udini/gophkeeper/internal/server/repository"

	"golang.org/x/crypto/bcrypt"
)

var (
	ErrUserExists         = errors.New("user already exists")
	ErrInvalidCredentials = errors.New("invalid credentials")
)

type UserRepository interface {
	GetUserByLogin(login string) (*model.User, error)
	CreateUser(login, passwordHash, masterPasswordHash string) error
}

type UserService struct {
	userRepo UserRepository
	jwt      *jwt.Container
}

func NewUserManager(
	userRepo *repository.UserRepository,
	jwt *jwt.Container,
) *UserService {
	return &UserService{
		userRepo: userRepo,
		jwt:      jwt,
	}
}

func (m *UserService) Register(login, password, masterPassword string) (string, error) {
	existing, err := m.userRepo.GetUserByLogin(login)
	if err != nil {
		return "", err
	}
	if existing != nil {
		return "", ErrUserExists
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}

	masterPasswordHash, err := bcrypt.GenerateFromPassword([]byte(masterPassword), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}

	err = m.userRepo.CreateUser(login, string(passwordHash), string(masterPasswordHash))
	if err != nil {
		return "", err
	}

	user, err := m.userRepo.GetUserByLogin(login)
	if err != nil || user == nil {
		return "", err
	}

	return m.jwt.Encode(user.ID, login)
}

func (m *UserService) Login(login, password, masterPassword string) (string, error) {
	user, err := m.userRepo.GetUserByLogin(login)
	if err != nil {
		return "", err
	}
	if user == nil {
		return "", ErrInvalidCredentials
	}

	if err := bcrypt.CompareHashAndPassword(
		[]byte(user.PasswordHash),
		[]byte(password),
	); err != nil {
		return "", ErrInvalidCredentials
	}

	if err := bcrypt.CompareHashAndPassword(
		[]byte(user.MasterPasswordHash),
		[]byte(masterPassword),
	); err != nil {
		return "", ErrInvalidCredentials
	}

	return m.jwt.Encode(user.ID, login)
}

func (m *UserService) DecodeToken(token string) (*jwt.Claims, error) {
	return m.jwt.Decode(token)
}
