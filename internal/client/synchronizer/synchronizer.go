package synchronizer

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/g123udini/gophkeeper/internal/client/model"
	"github.com/g123udini/gophkeeper/internal/common/logger"
	"github.com/g123udini/gophkeeper/internal/common/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var ErrUnauthenticated = errors.New("unauthenticated")

type GRPCClient interface {
	Upsert(ctx context.Context, data *model.UserData) (*proto.DataResponse, error)
	GetUpdates(ctx context.Context, lastSync time.Time) (*proto.DataListResponse, error)
}

type UserDataManager interface {
	GetUpdates(ctx context.Context, lastSync time.Time) ([]*model.UserData, error)
	Upsert(ctx context.Context, data *model.UserData) error
}

type MetaManager interface {
	GetLastSync(ctx context.Context) (time.Time, error)
	SetLastSync(ctx context.Context, lastSync time.Time) error
	HasToken(ctx context.Context) (bool, error)
}

type Synchronizer struct {
	client      GRPCClient
	userDataMgr UserDataManager
	metaManager MetaManager
	interval    time.Duration
	stopCh      chan struct{}
	onceStop    sync.Once
}

func New(
	client GRPCClient,
	userDataMgr UserDataManager,
	metaManager MetaManager,
	interval time.Duration,
) *Synchronizer {
	return &Synchronizer{
		client:      client,
		userDataMgr: userDataMgr,
		metaManager: metaManager,
		interval:    interval,
		stopCh:      make(chan struct{}),
	}
}

func (s *Synchronizer) Start(ctx context.Context) error {
	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()

	if err := s.syncOnce(ctx); err != nil {
		if errors.Is(err, ErrUnauthenticated) {
			logger.Logger.Warn("sync skipped: unauthenticated")
		} else {
			return err
		}
	}

	for {
		select {
		case <-ticker.C:
			if err := s.syncOnce(ctx); err != nil {
				if errors.Is(err, ErrUnauthenticated) {
					logger.Logger.Warn("sync skipped: unauthenticated")
					continue
				}

				return err
			}
		case <-s.stopCh:
			return nil
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func (s *Synchronizer) Stop() {
	s.onceStop.Do(func() {
		close(s.stopCh)
	})
}

func (s *Synchronizer) syncOnce(ctx context.Context) error {
	hasToken, err := s.metaManager.HasToken(ctx)
	if err != nil {
		return fmt.Errorf("sync: get token state: %w", err)
	}

	if !hasToken {
		return nil
	}

	lastSync, err := s.metaManager.GetLastSync(ctx)
	if err != nil {
		return fmt.Errorf("sync: get last sync: %w", err)
	}

	if err := s.pushLocalUpdates(ctx, lastSync); err != nil {
		return err
	}

	if err := s.fetchRemoteUpdates(ctx, lastSync); err != nil {
		return err
	}

	if err := s.metaManager.SetLastSync(ctx, time.Now().UTC()); err != nil {
		return fmt.Errorf("sync: set last sync: %w", err)
	}

	return nil
}

func (s *Synchronizer) pushLocalUpdates(ctx context.Context, lastSync time.Time) error {
	localUpdates, err := s.userDataMgr.GetUpdates(ctx, lastSync)
	if err != nil {
		return fmt.Errorf("sync: get local updates: %w", err)
	}

	for _, data := range localUpdates {
		_, err := s.client.Upsert(ctx, data)
		if err != nil {
			if status.Code(err) == codes.Unauthenticated {
				return ErrUnauthenticated
			}

			return fmt.Errorf("sync: push local update: %w", err)
		}
	}

	return nil
}

func (s *Synchronizer) fetchRemoteUpdates(ctx context.Context, lastSync time.Time) error {
	resp, err := s.client.GetUpdates(ctx, lastSync)
	if err != nil {
		if status.Code(err) == codes.Unauthenticated {
			return ErrUnauthenticated
		}

		return fmt.Errorf("sync: get remote updates: %w", err)
	}

	for _, item := range resp.Items {
		data := &model.UserData{
			DataKey:   item.DataKey,
			DataValue: item.DataValue,
			UpdatedAt: item.UpdatedAt.AsTime(),
			DeletedAt: item.DeletedAt.AsTime(),
		}

		if err := s.userDataMgr.Upsert(ctx, data); err != nil {
			return fmt.Errorf("sync: update local data: %w", err)
		}
	}

	return nil
}
