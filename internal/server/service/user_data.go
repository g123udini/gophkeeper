package service

import (
	"context"
	"time"

	"github.com/g123udini/gophkeeper/internal/server/model"
	"github.com/g123udini/gophkeeper/internal/server/repository"
)

type UserDataRepository interface {
	Upsert(ctx context.Context, data *model.UserData) error
	GetUpdates(ctx context.Context, userID uint32, since time.Time) ([]*model.UserData, error)
}

type UserDataService struct {
	dataRepo UserDataRepository
}

func NewUserDataManager(dataRepo *repository.UserDataRepository) *UserDataService {
	return &UserDataService{
		dataRepo: dataRepo,
	}
}

func (m *UserDataService) Upsert(ctx context.Context, data *model.UserData) error {
	return m.dataRepo.Upsert(ctx, data)
}

func (m *UserDataService) GetUpdates(ctx context.Context, userID uint32, since time.Time) ([]*model.UserData, error) {
	return m.dataRepo.GetUpdates(ctx, userID, since)
}
