package service

import (
	"errors"
	"testing"

	"github.com/g123udini/gophkeeper/internal/server/jwt"
	"github.com/g123udini/gophkeeper/internal/server/model"
	"github.com/g123udini/gophkeeper/internal/server/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) GetUserByLogin(login string) (*model.User, error) {
	args := m.Called(login)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}

func (m *MockUserRepository) CreateUser(login, passwordHash, masterPasswordHash string) error {
	args := m.Called(login, passwordHash, masterPasswordHash)
	return args.Error(0)
}

func verifyPasswordHash(t *testing.T, password, hash string) {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	assert.NoError(t, err, "Password hash verification failed")
}

func TestUserManager_Register_Success(t *testing.T) {
	mockRepo := new(MockUserRepository)
	jwtContainer := jwt.New("secret")

	manager := NewUserManager((*repository.UserRepository)(nil), nil)
	manager.userRepo = mockRepo
	manager.jwt = jwtContainer

	login := "testuser"
	password := "password123"
	masterPassword := "master123"
	userID := uint32(1)

	mockRepo.On("GetUserByLogin", login).Return(nil, nil).Once()

	mockRepo.On("CreateUser", login, mock.MatchedBy(func(hash string) bool {
		verifyPasswordHash(t, password, hash)
		return true
	}), mock.MatchedBy(func(hash string) bool {
		verifyPasswordHash(t, masterPassword, hash)
		return true
	})).Return(nil).Once()

	mockRepo.On("GetUserByLogin", login).Return(&model.User{ID: userID, Login: login}, nil).Once()

	token, err := manager.Register(login, password, masterPassword)

	assert.NoError(t, err)
	require.NotEmpty(t, token)

	claims, err := jwtContainer.Decode(token)
	assert.NoError(t, err)
	assert.Equal(t, userID, claims.SubjectID)
	assert.Equal(t, login, claims.Subject)

	mockRepo.AssertExpectations(t)
}

func TestUserManager_Register_UserExists(t *testing.T) {
	mockRepo := new(MockUserRepository)
	jwtContainer := jwt.New("secret")

	manager := NewUserManager((*repository.UserRepository)(nil), nil)
	manager.userRepo = mockRepo
	manager.jwt = jwtContainer

	login := "testuser"
	password := "password123"
	masterPassword := "master123"

	existingUser := &model.User{ID: 1, Login: login}
	mockRepo.On("GetUserByLogin", login).Return(existingUser, nil).Once()

	token, err := manager.Register(login, password, masterPassword)

	assert.Equal(t, "", token)
	assert.Equal(t, ErrUserExists, err)
	mockRepo.AssertExpectations(t)
}

func TestUserManager_Register_RepoError(t *testing.T) {
	mockRepo := new(MockUserRepository)
	jwtContainer := jwt.New("secret")

	manager := NewUserManager((*repository.UserRepository)(nil), nil)
	manager.userRepo = mockRepo
	manager.jwt = jwtContainer

	login := "testuser"
	password := "password123"
	masterPassword := "master123"

	dbError := errors.New("db error")
	mockRepo.On("GetUserByLogin", login).Return(nil, dbError).Once()

	token, err := manager.Register(login, password, masterPassword)

	assert.Equal(t, "", token)
	assert.Error(t, err)
	assert.Equal(t, dbError, err)
	mockRepo.AssertExpectations(t)
}

func TestUserManager_Login_Success(t *testing.T) {
	mockRepo := new(MockUserRepository)
	jwtContainer := jwt.New("secret")

	manager := NewUserManager((*repository.UserRepository)(nil), nil)
	manager.userRepo = mockRepo
	manager.jwt = jwtContainer

	login := "testuser"
	password := "password123"
	masterPassword := "master123"
	userID := uint32(1)

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	require.NoError(t, err)
	masterPasswordHash, err := bcrypt.GenerateFromPassword([]byte(masterPassword), bcrypt.DefaultCost)
	require.NoError(t, err)

	user := &model.User{
		ID:                 userID,
		Login:              login,
		PasswordHash:       string(passwordHash),
		MasterPasswordHash: string(masterPasswordHash),
	}

	mockRepo.On("GetUserByLogin", login).Return(user, nil).Once()

	token, err := manager.Login(login, password, masterPassword)

	assert.NoError(t, err)
	require.NotEmpty(t, token)

	claims, err := jwtContainer.Decode(token)
	assert.NoError(t, err)
	assert.Equal(t, userID, claims.SubjectID)
	assert.Equal(t, login, claims.Subject)

	mockRepo.AssertExpectations(t)
}

func TestUserManager_Login_InvalidPassword(t *testing.T) {
	mockRepo := new(MockUserRepository)
	jwtContainer := jwt.New("secret")

	manager := NewUserManager((*repository.UserRepository)(nil), nil)
	manager.userRepo = mockRepo
	manager.jwt = jwtContainer

	login := "testuser"
	password := "password123"
	wrongPassword := "wrongpassword"
	masterPassword := "master123"
	userID := uint32(1)

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	require.NoError(t, err)
	masterPasswordHash, err := bcrypt.GenerateFromPassword([]byte(masterPassword), bcrypt.DefaultCost)
	require.NoError(t, err)

	user := &model.User{
		ID:                 userID,
		Login:              login,
		PasswordHash:       string(passwordHash),
		MasterPasswordHash: string(masterPasswordHash),
	}

	mockRepo.On("GetUserByLogin", login).Return(user, nil).Once()

	token, err := manager.Login(login, wrongPassword, masterPassword)

	assert.Equal(t, "", token)
	assert.Equal(t, ErrInvalidCredentials, err)
	mockRepo.AssertExpectations(t)
}

func TestUserManager_Login_InvalidMasterPassword(t *testing.T) {
	mockRepo := new(MockUserRepository)
	jwtContainer := jwt.New("secret")

	manager := NewUserManager((*repository.UserRepository)(nil), nil)
	manager.userRepo = mockRepo
	manager.jwt = jwtContainer

	login := "testuser"
	password := "password123"
	masterPassword := "master123"
	wrongMasterPassword := "wrongmaster"
	userID := uint32(1)

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	require.NoError(t, err)
	masterPasswordHash, err := bcrypt.GenerateFromPassword([]byte(masterPassword), bcrypt.DefaultCost)
	require.NoError(t, err)

	user := &model.User{
		ID:                 userID,
		Login:              login,
		PasswordHash:       string(passwordHash),
		MasterPasswordHash: string(masterPasswordHash),
	}

	mockRepo.On("GetUserByLogin", login).Return(user, nil).Once()

	token, err := manager.Login(login, password, wrongMasterPassword)

	assert.Equal(t, "", token)
	assert.Equal(t, ErrInvalidCredentials, err)
	mockRepo.AssertExpectations(t)
}

func TestUserManager_Login_UserNotFound(t *testing.T) {
	mockRepo := new(MockUserRepository)
	jwtContainer := jwt.New("secret")

	manager := NewUserManager((*repository.UserRepository)(nil), nil)
	manager.userRepo = mockRepo
	manager.jwt = jwtContainer

	login := "nonexistentuser"
	password := "password123"
	masterPassword := "master123"

	mockRepo.On("GetUserByLogin", login).Return(nil, nil).Once()

	token, err := manager.Login(login, password, masterPassword)

	assert.Equal(t, "", token)
	assert.Equal(t, ErrInvalidCredentials, err)
	mockRepo.AssertExpectations(t)
}

func TestUserManager_DecodeToken(t *testing.T) {
	mockRepo := new(MockUserRepository)
	jwtContainer := jwt.New("secret")

	manager := NewUserManager((*repository.UserRepository)(nil), nil)
	manager.userRepo = mockRepo
	manager.jwt = jwtContainer

	userID := uint32(1)
	login := "testuser"
	token, err := jwtContainer.Encode(userID, login)
	require.NoError(t, err)

	claims, err := manager.DecodeToken(token)

	assert.NoError(t, err)
	assert.NotNil(t, claims)
	assert.Equal(t, userID, claims.SubjectID)
	assert.Equal(t, login, claims.Subject)
}

func TestUserManager_DecodeToken_Invalid(t *testing.T) {
	mockRepo := new(MockUserRepository)
	jwtContainer := jwt.New("secret")

	manager := NewUserManager((*repository.UserRepository)(nil), nil)
	manager.userRepo = mockRepo
	manager.jwt = jwtContainer

	token := "invalid.token.format"

	claims, err := manager.DecodeToken(token)

	assert.Nil(t, claims)
	assert.Error(t, err)
}
