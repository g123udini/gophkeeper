package synchronizer

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/g123udini/gophkeeper/internal/client/model"
	"github.com/g123udini/gophkeeper/internal/common/proto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type MockGRPCClient struct {
	mock.Mock
}

func (m *MockGRPCClient) Upsert(ctx context.Context, data *model.UserData) (*proto.DataResponse, error) {
	args := m.Called(ctx, data)

	var resp *proto.DataResponse
	if v := args.Get(0); v != nil {
		resp = v.(*proto.DataResponse)
	}

	return resp, args.Error(1)
}

func (m *MockGRPCClient) GetUpdates(ctx context.Context, lastSync time.Time) (*proto.DataListResponse, error) {
	args := m.Called(ctx, lastSync)

	var resp *proto.DataListResponse
	if v := args.Get(0); v != nil {
		resp = v.(*proto.DataListResponse)
	}

	return resp, args.Error(1)
}

type MockUserDataManager struct {
	mock.Mock
}

func (m *MockUserDataManager) GetUpdates(ctx context.Context, lastSync time.Time) ([]*model.UserData, error) {
	args := m.Called(ctx, lastSync)
	return args.Get(0).([]*model.UserData), args.Error(1)
}

func (m *MockUserDataManager) Upsert(ctx context.Context, data *model.UserData) error {
	args := m.Called(ctx, data)
	return args.Error(0)
}

type MockMetaManager struct {
	mock.Mock
}

func (m *MockMetaManager) GetLastSync(ctx context.Context) (time.Time, error) {
	args := m.Called(ctx)
	return args.Get(0).(time.Time), args.Error(1)
}

func (m *MockMetaManager) SetLastSync(ctx context.Context, lastSync time.Time) error {
	args := m.Called(ctx, lastSync)
	return args.Error(0)
}

func (m *MockMetaManager) HasToken(ctx context.Context) (bool, error) {
	args := m.Called(ctx)
	return args.Bool(0), args.Error(1)
}

func TestNew(t *testing.T) {
	client := &MockGRPCClient{}
	userDataMgr := &MockUserDataManager{}
	metaManager := &MockMetaManager{}
	interval := 30 * time.Second

	s := New(client, userDataMgr, metaManager, interval)

	assert.Equal(t, client, s.client)
	assert.Equal(t, userDataMgr, s.userDataMgr)
	assert.Equal(t, metaManager, s.metaManager)
	assert.Equal(t, interval, s.interval)
	assert.NotNil(t, s.stopCh)
}

func TestSynchronizer_StartStop(t *testing.T) {
	client := &MockGRPCClient{}
	userDataMgr := &MockUserDataManager{}
	metaManager := &MockMetaManager{}
	interval := 50 * time.Millisecond

	lastSyncTime := time.Now().Add(-time.Hour).UTC()

	metaManager.On("HasToken", mock.Anything).Return(true, nil)
	metaManager.On("GetLastSync", mock.Anything).Return(lastSyncTime, nil)
	userDataMgr.On("GetUpdates", mock.Anything, lastSyncTime).Return([]*model.UserData{}, nil)
	client.On("GetUpdates", mock.Anything, lastSyncTime).
		Return(&proto.DataListResponse{Items: []*proto.DataResponse{}}, nil)
	metaManager.On("SetLastSync", mock.Anything, mock.AnythingOfType("time.Time")).Return(nil)

	s := New(client, userDataMgr, metaManager, interval)

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	s.Start(ctx)
	time.Sleep(75 * time.Millisecond)
	s.Stop()

	metaManager.AssertCalled(t, "HasToken", mock.Anything)
	metaManager.AssertCalled(t, "GetLastSync", mock.Anything)
	userDataMgr.AssertCalled(t, "GetUpdates", mock.Anything, lastSyncTime)
	client.AssertCalled(t, "GetUpdates", mock.Anything, lastSyncTime)
	metaManager.AssertCalled(t, "SetLastSync", mock.Anything, mock.AnythingOfType("time.Time"))
}

func TestSynchronizer_syncOnce(t *testing.T) {
	client := &MockGRPCClient{}
	userDataMgr := &MockUserDataManager{}
	metaManager := &MockMetaManager{}
	interval := 30 * time.Second

	lastSyncTime := time.Now().Add(-time.Hour).UTC()
	localUpdate := &model.UserData{
		DataKey:   "test-key",
		DataValue: []byte("test-value"),
		UpdatedAt: time.Now(),
	}

	remoteItem := &proto.DataResponse{
		DataKey:   "remote-key",
		DataValue: []byte("remote-value"),
		UpdatedAt: timestamppb.Now(),
		DeletedAt: timestamppb.New(time.Time{}),
	}

	metaManager.On("HasToken", mock.Anything).Return(true, nil)
	metaManager.On("GetLastSync", mock.Anything).Return(lastSyncTime, nil)
	userDataMgr.On("GetUpdates", mock.Anything, lastSyncTime).Return([]*model.UserData{localUpdate}, nil)
	client.On("Upsert", mock.Anything, localUpdate).Return(&proto.DataResponse{}, nil)
	client.On("GetUpdates", mock.Anything, lastSyncTime).Return(&proto.DataListResponse{
		Items: []*proto.DataResponse{remoteItem},
	}, nil)
	userDataMgr.On("Upsert", mock.Anything, mock.AnythingOfType("*model.UserData")).Return(nil)
	metaManager.On("SetLastSync", mock.Anything, mock.AnythingOfType("time.Time")).Return(nil)

	s := New(client, userDataMgr, metaManager, interval)

	ctx := context.Background()
	s.syncOnce(ctx)

	metaManager.AssertCalled(t, "HasToken", ctx)
	metaManager.AssertCalled(t, "GetLastSync", ctx)
	userDataMgr.AssertCalled(t, "GetUpdates", ctx, lastSyncTime)
	client.AssertCalled(t, "Upsert", ctx, localUpdate)
	client.AssertCalled(t, "GetUpdates", ctx, lastSyncTime)
	userDataMgr.AssertCalled(t, "Upsert", ctx, mock.AnythingOfType("*model.UserData"))
	metaManager.AssertCalled(t, "SetLastSync", ctx, mock.AnythingOfType("time.Time"))
}

func TestSynchronizer_syncOnce_NoToken(t *testing.T) {
	client := &MockGRPCClient{}
	userDataMgr := &MockUserDataManager{}
	metaManager := &MockMetaManager{}
	interval := 30 * time.Second

	metaManager.On("HasToken", mock.Anything).Return(false, nil)

	s := New(client, userDataMgr, metaManager, interval)
	ctx := context.Background()

	s.syncOnce(ctx)

	metaManager.AssertCalled(t, "HasToken", ctx)
	metaManager.AssertNotCalled(t, "GetLastSync", mock.Anything)
	userDataMgr.AssertNotCalled(t, "GetUpdates", mock.Anything, mock.Anything)
	client.AssertNotCalled(t, "GetUpdates", mock.Anything, mock.Anything)
}

func TestSynchronizer_pushLocalUpdates(t *testing.T) {
	tests := []struct {
		name           string
		setupMocks     func(*MockUserDataManager, *MockGRPCClient)
		expectedResult bool
	}{
		{
			name: "success_no_updates",
			setupMocks: func(userDataMgr *MockUserDataManager, client *MockGRPCClient) {
				userDataMgr.On("GetUpdates", mock.Anything, mock.AnythingOfType("time.Time")).
					Return([]*model.UserData{}, nil)
			},
			expectedResult: true,
		},
		{
			name: "success_with_updates",
			setupMocks: func(userDataMgr *MockUserDataManager, client *MockGRPCClient) {
				updates := []*model.UserData{
					{DataKey: "key1", DataValue: []byte("value1")},
					{DataKey: "key2", DataValue: []byte("value2")},
				}
				userDataMgr.On("GetUpdates", mock.Anything, mock.AnythingOfType("time.Time")).
					Return(updates, nil)
				for _, update := range updates {
					client.On("Upsert", mock.Anything, update).Return(&proto.DataResponse{}, nil)
				}
			},
			expectedResult: true,
		},
		{
			name: "failure_upsert_error",
			setupMocks: func(userDataMgr *MockUserDataManager, client *MockGRPCClient) {
				updates := []*model.UserData{
					{DataKey: "key1", DataValue: []byte("value1")},
				}
				userDataMgr.On("GetUpdates", mock.Anything, mock.AnythingOfType("time.Time")).
					Return(updates, nil)
				client.On("Upsert", mock.Anything, updates[0]).Return((*proto.DataResponse)(nil), errors.New("upsert error"))
			},
			expectedResult: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := &MockGRPCClient{}
			userDataMgr := &MockUserDataManager{}
			metaManager := &MockMetaManager{}
			interval := 30 * time.Second

			tt.setupMocks(userDataMgr, client)

			s := New(client, userDataMgr, metaManager, interval)
			result := s.pushLocalUpdates(context.Background(), time.Now())

			assert.Equal(t, tt.expectedResult, result)
			userDataMgr.AssertExpectations(t)
			client.AssertExpectations(t)
		})
	}
}

func TestSynchronizer_fetchRemoteUpdates(t *testing.T) {
	tests := []struct {
		name           string
		setupMocks     func(*MockGRPCClient, *MockUserDataManager)
		expectedResult bool
	}{
		{
			name: "success_no_updates",
			setupMocks: func(client *MockGRPCClient, userDataMgr *MockUserDataManager) {
				client.On("GetUpdates", mock.Anything, mock.AnythingOfType("time.Time")).
					Return(&proto.DataListResponse{Items: []*proto.DataResponse{}}, nil)
			},
			expectedResult: true,
		},
		{
			name: "success_with_updates",
			setupMocks: func(client *MockGRPCClient, userDataMgr *MockUserDataManager) {
				updates := []*proto.DataResponse{
					{
						DataKey:   "key1",
						DataValue: []byte("value1"),
						UpdatedAt: timestamppb.Now(),
						DeletedAt: timestamppb.New(time.Time{}),
					},
					{
						DataKey:   "key2",
						DataValue: []byte("value2"),
						UpdatedAt: timestamppb.Now(),
						DeletedAt: timestamppb.New(time.Time{}),
					},
				}
				client.On("GetUpdates", mock.Anything, mock.AnythingOfType("time.Time")).
					Return(&proto.DataListResponse{Items: updates}, nil)
				for range updates {
					userDataMgr.On("Upsert", mock.Anything, mock.AnythingOfType("*model.UserData")).Return(nil).Once()
				}
			},
			expectedResult: true,
		},
		{
			name: "failure_get_updates_error",
			setupMocks: func(client *MockGRPCClient, userDataMgr *MockUserDataManager) {
				client.On("GetUpdates", mock.Anything, mock.AnythingOfType("time.Time")).
					Return((*proto.DataListResponse)(nil), errors.New("get updates error"))
			},
			expectedResult: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := &MockGRPCClient{}
			userDataMgr := &MockUserDataManager{}
			metaManager := &MockMetaManager{}
			interval := 30 * time.Second

			tt.setupMocks(client, userDataMgr)

			s := New(client, userDataMgr, metaManager, interval)
			result := s.fetchRemoteUpdates(context.Background(), time.Now())

			assert.Equal(t, tt.expectedResult, result)
			client.AssertExpectations(t)
			userDataMgr.AssertExpectations(t)
		})
	}
}
