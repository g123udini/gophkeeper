package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/g123udini/gophkeeper/internal/client/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type MockUserDataRepository struct {
	mock.Mock
}

func (m *MockUserDataRepository) Upsert(ctx context.Context, data *model.UserData) error {
	args := m.Called(ctx, data)
	return args.Error(0)
}

func (m *MockUserDataRepository) Get(ctx context.Context, key string) (*model.UserData, error) {
	args := m.Called(ctx, key)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.UserData), args.Error(1)
}

func (m *MockUserDataRepository) GetUpdates(ctx context.Context, lastSync time.Time) ([]*model.UserData, error) {
	args := m.Called(ctx, lastSync)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*model.UserData), args.Error(1)
}

func TestNewUserDataManager(t *testing.T) {
	mockRepo := new(MockUserDataRepository)
	manager := &UserDataService{dataRepo: mockRepo}

	assert.NotNil(t, manager)
}

func TestUpsert(t *testing.T) {
	mockRepo := new(MockUserDataRepository)
	manager := &UserDataService{dataRepo: mockRepo}

	ctx := context.Background()
	userData := &model.UserData{
		DataKey:   "test-key",
		DataValue: []byte("test-value"),
		UpdatedAt: time.Now(),
		DeletedAt: time.Time{},
	}
	testError := errors.New("test error")

	t.Run("Success", func(t *testing.T) {
		mockRepo.On("Upsert", ctx, userData).Return(nil).Once()

		err := manager.Upsert(ctx, userData)

		require.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Error", func(t *testing.T) {
		mockRepo.On("Upsert", ctx, userData).Return(testError).Once()

		err := manager.Upsert(ctx, userData)

		require.Error(t, err)
		assert.Equal(t, testError, err)
		mockRepo.AssertExpectations(t)
	})
}

func TestGet(t *testing.T) {
	mockRepo := new(MockUserDataRepository)
	manager := &UserDataService{dataRepo: mockRepo}

	ctx := context.Background()
	key := "test-key"
	expectedData := &model.UserData{
		DataKey:   key,
		DataValue: []byte("test-value"),
		UpdatedAt: time.Now(),
		DeletedAt: time.Time{},
	}
	testError := errors.New("test error")

	t.Run("Success", func(t *testing.T) {
		mockRepo.On("Get", ctx, key).Return(expectedData, nil).Once()

		data, err := manager.Get(ctx, key)

		require.NoError(t, err)
		assert.Equal(t, expectedData, data)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Not Found", func(t *testing.T) {
		mockRepo.On("Get", ctx, key).Return(nil, nil).Once()

		data, err := manager.Get(ctx, key)

		require.Error(t, err)
		assert.Equal(t, ErrNotFound, err)
		assert.Nil(t, data)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Repository Error", func(t *testing.T) {
		mockRepo.On("Get", ctx, key).Return(nil, testError).Once()

		data, err := manager.Get(ctx, key)

		require.Error(t, err)
		assert.Equal(t, testError, err)
		assert.Nil(t, data)
		mockRepo.AssertExpectations(t)
	})
}

func TestGetUpdates(t *testing.T) {
	mockRepo := new(MockUserDataRepository)
	manager := &UserDataService{dataRepo: mockRepo}

	ctx := context.Background()
	lastSync := time.Now().Add(-24 * time.Hour)
	expectedUpdates := []*model.UserData{
		{
			DataKey:   "key1",
			DataValue: []byte("value1"),
			UpdatedAt: time.Now(),
			DeletedAt: time.Time{},
		},
		{
			DataKey:   "key2",
			DataValue: []byte("value2"),
			UpdatedAt: time.Now(),
			DeletedAt: time.Time{},
		},
	}
	testError := errors.New("test error")

	t.Run("Success", func(t *testing.T) {
		mockRepo.On("GetUpdates", ctx, lastSync).Return(expectedUpdates, nil).Once()

		updates, err := manager.GetUpdates(ctx, lastSync)

		require.NoError(t, err)
		assert.Equal(t, expectedUpdates, updates)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Empty Result", func(t *testing.T) {
		mockRepo.On("GetUpdates", ctx, lastSync).Return([]*model.UserData{}, nil).Once()

		updates, err := manager.GetUpdates(ctx, lastSync)

		require.NoError(t, err)
		assert.Empty(t, updates)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Repository Error", func(t *testing.T) {
		mockRepo.On("GetUpdates", ctx, lastSync).Return(nil, testError).Once()

		updates, err := manager.GetUpdates(ctx, lastSync)

		require.Error(t, err)
		assert.Equal(t, testError, err)
		assert.Nil(t, updates)
		mockRepo.AssertExpectations(t)
	})
}
