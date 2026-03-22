package service

import (
	"context"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type MetaRepository interface {
	GetLastSync(ctx context.Context) (time.Time, error)
	SetLastSync(ctx context.Context, t time.Time) error

	GetMasterPasswordHash(ctx context.Context) (string, error)
	SetMasterPasswordHash(ctx context.Context, h string) error

	GetToken(ctx context.Context) (string, error)
	SetToken(ctx context.Context, token string) error
}

type MetaService struct {
	repo MetaRepository
}

func NewMetaManager(repo MetaRepository) *MetaService {
	return &MetaService{repo: repo}
}

func (m *MetaService) GetLastSync(ctx context.Context) (time.Time, error) {
	return m.repo.GetLastSync(ctx)
}

func (m *MetaService) SetLastSync(ctx context.Context, t time.Time) error {
	return m.repo.SetLastSync(ctx, t)
}

func (m *MetaService) GetToken(ctx context.Context) (string, error) {
	return m.repo.GetToken(ctx)
}

func (m *MetaService) SetToken(ctx context.Context, token string) error {
	return m.repo.SetToken(ctx, token)
}

func (m *MetaService) HasToken(ctx context.Context) (bool, error) {
	token, err := m.repo.GetToken(ctx)
	if err != nil {
		return false, err
	}

	return token != "", nil
}

func (m *MetaService) MasterPasswordHashDefined(ctx context.Context) (bool, error) {
	h, err := m.repo.GetMasterPasswordHash(ctx)
	if err != nil {
		return false, err
	}

	return h != "", nil
}

func (m *MetaService) ValidateMasterPassword(ctx context.Context, password string) error {
	h, err := m.repo.GetMasterPasswordHash(ctx)
	if err != nil {
		return err
	}

	return bcrypt.CompareHashAndPassword([]byte(h), []byte(password))
}

func (m *MetaService) SetMasterPassword(ctx context.Context, password string) error {
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	return m.repo.SetMasterPasswordHash(ctx, string(passwordHash))
}
