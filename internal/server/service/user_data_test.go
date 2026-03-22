package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/g123udini/gophkeeper/internal/server/model"
	"github.com/g123udini/gophkeeper/internal/server/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockUserDataRepository struct {
	mock.Mock
}

func (m *MockUserDataRepository) Upsert(ctx context.Context, data *model.UserData) error {
	args := m.Called(ctx, data)
	return args.Error(0)
}

func (m *MockUserDataRepository) GetUpdates(ctx context.Context, userID uint32, since time.Time) ([]*model.UserData, error) {
	args := m.Called(ctx, userID, since)
	return args.Get(0).([]*model.UserData), args.Error(1)
}

func TestUserDataManager_Upsert(t *testing.T) {
	mockRepo := new(MockUserDataRepository)
	manager := NewUserDataManager((*repository.UserDataRepository)(nil))
	manager.dataRepo = mockRepo

	ctx := context.Background()
	now := time.Now()
	data := &model.UserData{
		ID:        1,
		UserID:    1,
		DataKey:   "example.com",
		DataValue: []byte("secretpassword"),
		UpdatedAt: now,
	}

	mockRepo.On("Upsert", ctx, data).Return(nil)

	err := manager.Upsert(ctx, data)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestUserDataManager_Upsert_Error(t *testing.T) {
	mockRepo := new(MockUserDataRepository)
	manager := NewUserDataManager((*repository.UserDataRepository)(nil))
	manager.dataRepo = mockRepo

	ctx := context.Background()
	now := time.Now()
	data := &model.UserData{
		ID:        1,
		UserID:    1,
		DataKey:   "example.com",
		DataValue: []byte("secretpassword"),
		UpdatedAt: now,
	}

	expectedError := errors.New("db error")
	mockRepo.On("Upsert", ctx, data).Return(expectedError)

	err := manager.Upsert(ctx, data)

	assert.Error(t, err)
	assert.Equal(t, expectedError, err)
	mockRepo.AssertExpectations(t)
}

func TestUserDataManager_GetUpdates(t *testing.T) {
	mockRepo := new(MockUserDataRepository)
	manager := NewUserDataManager((*repository.UserDataRepository)(nil))
	manager.dataRepo = mockRepo

	ctx := context.Background()
	userID := uint32(1)
	since := time.Now().Add(-24 * time.Hour)
	now := time.Now()

	expectedData := []*model.UserData{
		{
			ID:        1,
			UserID:    userID,
			DataKey:   "example.com",
			DataValue: []byte("secretpassword"),
			UpdatedAt: now,
		},
		{
			ID:        2,
			UserID:    userID,
			DataKey:   "visa",
			DataValue: []byte("4111111111111111"),
			UpdatedAt: now,
		},
	}

	mockRepo.On("GetUpdates", ctx, userID, since).Return(expectedData, nil)

	result, err := manager.GetUpdates(ctx, userID, since)

	assert.NoError(t, err)
	assert.Equal(t, expectedData, result)
	assert.Len(t, result, 2)
	mockRepo.AssertExpectations(t)
}

func TestUserDataManager_GetUpdates_Error(t *testing.T) {
	mockRepo := new(MockUserDataRepository)
	manager := NewUserDataManager((*repository.UserDataRepository)(nil))
	manager.dataRepo = mockRepo

	ctx := context.Background()
	userID := uint32(1)
	since := time.Now().Add(-24 * time.Hour)

	expectedError := errors.New("db error")
	mockRepo.On("GetUpdates", ctx, userID, since).Return([]*model.UserData{}, expectedError)

	result, err := manager.GetUpdates(ctx, userID, since)

	assert.Error(t, err)
	assert.Equal(t, expectedError, err)
	assert.Empty(t, result)
	mockRepo.AssertExpectations(t)
}

func TestUserDataManager_GetUpdates_EmptyResult(t *testing.T) {
	mockRepo := new(MockUserDataRepository)
	manager := NewUserDataManager((*repository.UserDataRepository)(nil))
	manager.dataRepo = mockRepo

	ctx := context.Background()
	userID := uint32(1)
	since := time.Now().Add(-24 * time.Hour)

	emptyData := []*model.UserData{}
	mockRepo.On("GetUpdates", ctx, userID, since).Return(emptyData, nil)

	result, err := manager.GetUpdates(ctx, userID, since)

	assert.NoError(t, err)
	assert.Empty(t, result)
	mockRepo.AssertExpectations(t)
}
