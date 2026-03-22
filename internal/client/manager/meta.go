package manager

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

type MetaManager struct {
	repo MetaRepository
}

func NewMetaManager(repo MetaRepository) *MetaManager {
	return &MetaManager{repo: repo}
}

func (m *MetaManager) GetLastSync(ctx context.Context) (time.Time, error) {
	return m.repo.GetLastSync(ctx)
}

func (m *MetaManager) SetLastSync(ctx context.Context, t time.Time) error {
	return m.repo.SetLastSync(ctx, t)
}

func (m *MetaManager) GetToken(ctx context.Context) (string, error) {
	return m.repo.GetToken(ctx)
}

func (m *MetaManager) SetToken(ctx context.Context, token string) error {
	return m.repo.SetToken(ctx, token)
}

func (m *MetaManager) HasToken(ctx context.Context) (bool, error) {
	token, err := m.repo.GetToken(ctx)
	if err != nil {
		return false, err
	}

	return token != "", nil
}

func (m *MetaManager) MasterPasswordHashDefined(ctx context.Context) (bool, error) {
	h, err := m.repo.GetMasterPasswordHash(ctx)
	if err != nil {
		return false, err
	}

	return h != "", nil
}

func (m *MetaManager) ValidateMasterPassword(ctx context.Context, password string) error {
	h, err := m.repo.GetMasterPasswordHash(ctx)
	if err != nil {
		return err
	}

	return bcrypt.CompareHashAndPassword([]byte(h), []byte(password))
}

func (m *MetaManager) SetMasterPassword(ctx context.Context, password string) error {
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	return m.repo.SetMasterPasswordHash(ctx, string(passwordHash))
}
