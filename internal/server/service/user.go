package service

import (
	"errors"

	"github.com/g123udini/gophkeeper/internal/server/jwt"
	"github.com/g123udini/gophkeeper/internal/server/model"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrUserExists              = errors.New("user already exists")
	ErrInvalidCredentials      = errors.New("invalid credentials")
	ErrUserNotFoundAfterCreate = errors.New("user was created but not found")
)

type UserRepository interface {
	GetUserByLogin(login string) (*model.User, error)
	CreateUser(login, passwordHash, masterPasswordHash string) error
}

type TokenService interface {
	Encode(userID uint32, login string) (string, error)
	Decode(token string) (*jwt.Claims, error)
}

type UserService struct {
	userRepo UserRepository
	jwt      TokenService
}

func NewUserManager(
	userRepo UserRepository,
	tokenMgr TokenService,
) *UserService {
	return &UserService{
		userRepo: userRepo,
		jwt:      tokenMgr,
	}
}

func (s *UserService) Register(login, password, masterPassword string) (string, error) {
	existing, err := s.userRepo.GetUserByLogin(login)
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

	err = s.userRepo.CreateUser(login, string(passwordHash), string(masterPasswordHash))
	if err != nil {
		return "", err
	}

	user, err := s.userRepo.GetUserByLogin(login)
	if err != nil {
		return "", err
	}
	if user == nil {
		return "", ErrUserNotFoundAfterCreate
	}

	return s.jwt.Encode(user.ID, login)
}

func (s *UserService) Login(login, password, masterPassword string) (string, error) {
	user, err := s.userRepo.GetUserByLogin(login)
	if err != nil {
		return "", err
	}
	if user == nil {
		return "", ErrInvalidCredentials
	}

	if err = bcrypt.CompareHashAndPassword(
		[]byte(user.PasswordHash),
		[]byte(password),
	); err != nil {
		return "", ErrInvalidCredentials
	}

	if err = bcrypt.CompareHashAndPassword(
		[]byte(user.MasterPasswordHash),
		[]byte(masterPassword),
	); err != nil {
		return "", ErrInvalidCredentials
	}

	return s.jwt.Encode(user.ID, login)
}

func (s *UserService) DecodeToken(token string) (*jwt.Claims, error) {
	return s.jwt.Decode(token)
}
