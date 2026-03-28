package service

import (
	"context"
	"errors"
	"time"

	"github.com/g123udini/gophkeeper/internal/client/model"
	"github.com/g123udini/gophkeeper/internal/client/repository"
)

var (
	ErrNotFound = errors.New("data not found")
)

type UserDataRepository interface {
	Upsert(ctx context.Context, data *model.UserData) error
	Get(ctx context.Context, key string) (*model.UserData, error)
	GetUpdates(ctx context.Context, lastSync time.Time) ([]*model.UserData, error)
}

type UserDataService struct {
	dataRepo UserDataRepository
}

func NewUserDataManager(repo *repository.UserDataRepository) *UserDataService {
	return &UserDataService{
		dataRepo: repo,
	}
}

func (m *UserDataService) Upsert(ctx context.Context, data *model.UserData) error {
	return m.dataRepo.Upsert(ctx, data)
}

func (m *UserDataService) Get(ctx context.Context, key string) (*model.UserData, error) {
	data, err := m.dataRepo.Get(ctx, key)
	if err != nil {
		return nil, err
	}
	if data == nil {
		return nil, ErrNotFound
	}
	return data, nil
}

func (m *UserDataService) GetUpdates(ctx context.Context, lastSync time.Time) ([]*model.UserData, error) {
	return m.dataRepo.GetUpdates(ctx, lastSync)
}
