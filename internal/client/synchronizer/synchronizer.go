package synchronizer

import (
	"context"
	"sync"
	"time"

	"github.com/g123udini/gophkeeper/internal/client/model"
	"github.com/g123udini/gophkeeper/internal/common/logger"
	"github.com/g123udini/gophkeeper/internal/common/proto"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

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
	wg          sync.WaitGroup
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

func (s *Synchronizer) Start(ctx context.Context) {
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()

		ticker := time.NewTicker(s.interval)
		defer ticker.Stop()

		s.syncOnce(ctx)

		for {
			select {
			case <-ticker.C:
				s.syncOnce(ctx)
			case <-s.stopCh:
				return
			case <-ctx.Done():
				return
			}
		}
	}()
	s.wg.Wait()
}

func (s *Synchronizer) Stop() {
	close(s.stopCh)
	s.wg.Wait()
}

func (s *Synchronizer) syncOnce(ctx context.Context) {
	hasToken, err := s.metaManager.HasToken(ctx)
	if err != nil {
		logger.Logger.Fatal("sync: get token state error", zap.Error(err))
	}

	if !hasToken {
		return
	}

	lastSync, err := s.metaManager.GetLastSync(ctx)
	if err != nil {
		logger.Logger.Fatal("sync: get lastSync error", zap.Error(err))
	}

	if s.pushLocalUpdates(ctx, lastSync) && s.fetchRemoteUpdates(ctx, lastSync) {
		if err := s.metaManager.SetLastSync(ctx, time.Now().UTC()); err != nil {
			logger.Logger.Fatal("sync: set lastSync error", zap.Error(err))
		}
	}
}

func (s *Synchronizer) pushLocalUpdates(ctx context.Context, lastSync time.Time) bool {
	localUpdates, err := s.userDataMgr.GetUpdates(ctx, lastSync)
	if err != nil {
		logger.Logger.Fatal("sync: can't get local updates", zap.Error(err))
	}

	for _, data := range localUpdates {
		_, err := s.client.Upsert(ctx, data)
		if err != nil {
			if status.Code(err) == codes.Unauthenticated {
				return false
			}

			logger.Logger.Warn("sync: can't push local update to server", zap.Error(err))
			return false
		}
	}

	return true
}

func (s *Synchronizer) fetchRemoteUpdates(ctx context.Context, lastSync time.Time) bool {
	resp, err := s.client.GetUpdates(ctx, lastSync)
	if err != nil {
		if status.Code(err) == codes.Unauthenticated {
			return false
		}

		logger.Logger.Warn("sync: can't get updates", zap.Error(err))
		return false
	}

	for _, item := range resp.Items {
		data := &model.UserData{
			DataKey:   item.DataKey,
			DataValue: item.DataValue,
			UpdatedAt: item.UpdatedAt.AsTime(),
			DeletedAt: item.DeletedAt.AsTime(),
		}

		if err := s.userDataMgr.Upsert(ctx, data); err != nil {
			logger.Logger.Fatal("sync: can't update local data", zap.Error(err))
		}
	}

	return true
}
