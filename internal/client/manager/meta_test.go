package manager

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

type MockMetaRepository struct {
	mock.Mock
}

func (m *MockMetaRepository) GetLastSync(ctx context.Context) (time.Time, error) {
	args := m.Called(ctx)
	return args.Get(0).(time.Time), args.Error(1)
}

func (m *MockMetaRepository) SetLastSync(ctx context.Context, t time.Time) error {
	args := m.Called(ctx, t)
	return args.Error(0)
}

func (m *MockMetaRepository) GetMasterPasswordHash(ctx context.Context) (string, error) {
	args := m.Called(ctx)
	return args.String(0), args.Error(1)
}

func (m *MockMetaRepository) SetMasterPasswordHash(ctx context.Context, h string) error {
	args := m.Called(ctx, h)
	return args.Error(0)
}

func (m *MockMetaRepository) GetToken(ctx context.Context) (string, error) {
	args := m.Called(ctx)
	return args.String(0), args.Error(1)
}

func (m *MockMetaRepository) SetToken(ctx context.Context, token string) error {
	args := m.Called(ctx, token)
	return args.Error(0)
}

func TestNewMetaManager(t *testing.T) {
	mockRepo := new(MockMetaRepository)
	manager := NewMetaManager(mockRepo)

	assert.NotNil(t, manager)
}

func TestGetLastSync(t *testing.T) {
	mockRepo := new(MockMetaRepository)
	manager := &MetaManager{repo: mockRepo}

	ctx := context.Background()
	expectedTime := time.Now().UTC()
	testError := errors.New("test error")

	t.Run("Success", func(t *testing.T) {
		mockRepo.On("GetLastSync", ctx).Return(expectedTime, nil).Once()

		actualTime, err := manager.GetLastSync(ctx)

		require.NoError(t, err)
		assert.Equal(t, expectedTime, actualTime)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Error", func(t *testing.T) {
		mockRepo.On("GetLastSync", ctx).Return(time.Time{}, testError).Once()

		_, err := manager.GetLastSync(ctx)

		require.Error(t, err)
		assert.Equal(t, testError, err)
		mockRepo.AssertExpectations(t)
	})
}

func TestSetLastSync(t *testing.T) {
	mockRepo := new(MockMetaRepository)
	manager := &MetaManager{repo: mockRepo}

	ctx := context.Background()
	syncTime := time.Now().UTC()
	testError := errors.New("test error")

	t.Run("Success", func(t *testing.T) {
		mockRepo.On("SetLastSync", ctx, syncTime).Return(nil).Once()

		err := manager.SetLastSync(ctx, syncTime)

		require.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Error", func(t *testing.T) {
		mockRepo.On("SetLastSync", ctx, syncTime).Return(testError).Once()

		err := manager.SetLastSync(ctx, syncTime)

		require.Error(t, err)
		assert.Equal(t, testError, err)
		mockRepo.AssertExpectations(t)
	})
}

func TestGetToken(t *testing.T) {
	mockRepo := new(MockMetaRepository)
	manager := &MetaManager{repo: mockRepo}

	ctx := context.Background()
	testError := errors.New("test error")

	t.Run("Success", func(t *testing.T) {
		mockRepo.On("GetToken", ctx).Return("jwt-token", nil).Once()

		token, err := manager.GetToken(ctx)

		require.NoError(t, err)
		assert.Equal(t, "jwt-token", token)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Error", func(t *testing.T) {
		mockRepo.On("GetToken", ctx).Return("", testError).Once()

		_, err := manager.GetToken(ctx)

		require.Error(t, err)
		assert.Equal(t, testError, err)
		mockRepo.AssertExpectations(t)
	})
}

func TestSetToken(t *testing.T) {
	mockRepo := new(MockMetaRepository)
	manager := &MetaManager{repo: mockRepo}

	ctx := context.Background()
	testError := errors.New("test error")

	t.Run("Success", func(t *testing.T) {
		mockRepo.On("SetToken", ctx, "jwt-token").Return(nil).Once()

		err := manager.SetToken(ctx, "jwt-token")

		require.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Error", func(t *testing.T) {
		mockRepo.On("SetToken", ctx, "jwt-token").Return(testError).Once()

		err := manager.SetToken(ctx, "jwt-token")

		require.Error(t, err)
		assert.Equal(t, testError, err)
		mockRepo.AssertExpectations(t)
	})
}

func TestHasToken(t *testing.T) {
	mockRepo := new(MockMetaRepository)
	manager := &MetaManager{repo: mockRepo}

	ctx := context.Background()
	testError := errors.New("test error")

	t.Run("Has token", func(t *testing.T) {
		mockRepo.On("GetToken", ctx).Return("jwt-token", nil).Once()

		hasToken, err := manager.HasToken(ctx)

		require.NoError(t, err)
		assert.True(t, hasToken)
		mockRepo.AssertExpectations(t)
	})

	t.Run("No token", func(t *testing.T) {
		mockRepo.On("GetToken", ctx).Return("", nil).Once()

		hasToken, err := manager.HasToken(ctx)

		require.NoError(t, err)
		assert.False(t, hasToken)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Error", func(t *testing.T) {
		mockRepo.On("GetToken", ctx).Return("", testError).Once()

		hasToken, err := manager.HasToken(ctx)

		require.Error(t, err)
		assert.False(t, hasToken)
		assert.Equal(t, testError, err)
		mockRepo.AssertExpectations(t)
	})
}

func TestMasterPasswordHashDefined(t *testing.T) {
	mockRepo := new(MockMetaRepository)
	manager := &MetaManager{repo: mockRepo}

	ctx := context.Background()
	testError := errors.New("test error")

	t.Run("Defined", func(t *testing.T) {
		mockRepo.On("GetMasterPasswordHash", ctx).Return("hashedpassword", nil).Once()

		defined, err := manager.MasterPasswordHashDefined(ctx)

		require.NoError(t, err)
		assert.True(t, defined)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Not Defined", func(t *testing.T) {
		mockRepo.On("GetMasterPasswordHash", ctx).Return("", nil).Once()

		defined, err := manager.MasterPasswordHashDefined(ctx)

		require.NoError(t, err)
		assert.False(t, defined)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Error", func(t *testing.T) {
		mockRepo.On("GetMasterPasswordHash", ctx).Return("", testError).Once()

		_, err := manager.MasterPasswordHashDefined(ctx)

		require.Error(t, err)
		assert.Equal(t, testError, err)
		mockRepo.AssertExpectations(t)
	})
}

func TestValidateMasterPassword(t *testing.T) {
	mockRepo := new(MockMetaRepository)
	manager := &MetaManager{repo: mockRepo}

	ctx := context.Background()
	password := "secret"
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	wrongPassword := "wrong"
	testError := errors.New("test error")

	t.Run("Valid Password", func(t *testing.T) {
		mockRepo.On("GetMasterPasswordHash", ctx).Return(string(hashedPassword), nil).Once()

		err := manager.ValidateMasterPassword(ctx, password)

		require.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Invalid Password", func(t *testing.T) {
		mockRepo.On("GetMasterPasswordHash", ctx).Return(string(hashedPassword), nil).Once()

		err := manager.ValidateMasterPassword(ctx, wrongPassword)

		require.Error(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Repository Error", func(t *testing.T) {
		mockRepo.On("GetMasterPasswordHash", ctx).Return("", testError).Once()

		err := manager.ValidateMasterPassword(ctx, password)

		require.Error(t, err)
		assert.Equal(t, testError, err)
		mockRepo.AssertExpectations(t)
	})
}

func TestSetMasterPassword(t *testing.T) {
	mockRepo := new(MockMetaRepository)
	manager := &MetaManager{repo: mockRepo}

	ctx := context.Background()
	password := "secret"
	testError := errors.New("test error")

	t.Run("Success", func(t *testing.T) {
		mockRepo.On("SetMasterPasswordHash", ctx, mock.AnythingOfType("string")).Return(nil).Once()

		err := manager.SetMasterPassword(ctx, password)

		require.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Repository Error", func(t *testing.T) {
		mockRepo.On("SetMasterPasswordHash", ctx, mock.AnythingOfType("string")).Return(testError).Once()

		err := manager.SetMasterPassword(ctx, password)

		require.Error(t, err)
		assert.Equal(t, testError, err)
		mockRepo.AssertExpectations(t)
	})
}
