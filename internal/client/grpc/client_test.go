package grpc

import (
	"context"
	"errors"
	"net"
	"testing"
	"time"

	"github.com/g123udini/gophkeeper/internal/client/model"
	"github.com/g123udini/gophkeeper/internal/common/proto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/test/bufconn"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type mockAuthServiceClient struct {
	loginFunc    func(ctx context.Context, in *proto.LoginRequest, opts ...grpc.CallOption) (*proto.TokenResponse, error)
	registerFunc func(ctx context.Context, in *proto.RegisterRequest, opts ...grpc.CallOption) (*proto.TokenResponse, error)
}

func (m *mockAuthServiceClient) Login(ctx context.Context, in *proto.LoginRequest, opts ...grpc.CallOption) (*proto.TokenResponse, error) {
	return m.loginFunc(ctx, in, opts...)
}

func (m *mockAuthServiceClient) Register(ctx context.Context, in *proto.RegisterRequest, opts ...grpc.CallOption) (*proto.TokenResponse, error) {
	return m.registerFunc(ctx, in, opts...)
}

type mockDataServiceClient struct {
	upsertFunc     func(ctx context.Context, in *proto.UpsertRequest, opts ...grpc.CallOption) (*proto.DataResponse, error)
	getUpdatesFunc func(ctx context.Context, in *proto.GetUpdatesRequest, opts ...grpc.CallOption) (*proto.DataListResponse, error)
}

func (m *mockDataServiceClient) Upsert(ctx context.Context, in *proto.UpsertRequest, opts ...grpc.CallOption) (*proto.DataResponse, error) {
	return m.upsertFunc(ctx, in, opts...)
}

func (m *mockDataServiceClient) GetUpdates(ctx context.Context, in *proto.GetUpdatesRequest, opts ...grpc.CallOption) (*proto.DataListResponse, error) {
	return m.getUpdatesFunc(ctx, in, opts...)
}

func TestNewClient(t *testing.T) {
	lis := bufconn.Listen(1024 * 1024)
	s := grpc.NewServer()
	defer s.Stop()

	go func() {
		_ = s.Serve(lis)
	}()

	conn, err := grpc.NewClient(
		"bufnet",
		grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) {
			return lis.Dial()
		}),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	require.NoError(t, err)
	defer conn.Close()

	client := &Client{
		conn:       conn,
		AuthClient: proto.NewAuthServiceClient(conn),
		DataClient: proto.NewDataServiceClient(conn),
	}

	assert.NotNil(t, client)
	assert.NotNil(t, client.AuthClient)
	assert.NotNil(t, client.DataClient)
}

func TestClient_Close(t *testing.T) {
	conn, err := grpc.NewClient("bufnet", grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(t, err)

	client := &Client{
		conn: conn,
	}

	err = client.Close()
	assert.NoError(t, err)
}

func TestClient_Login(t *testing.T) {
	expectedToken := "test-token"
	mockAuth := &mockAuthServiceClient{
		loginFunc: func(ctx context.Context, in *proto.LoginRequest, opts ...grpc.CallOption) (*proto.TokenResponse, error) {
			assert.Equal(t, "testuser", in.Login)
			assert.Equal(t, "testpass", in.Password)
			assert.Equal(t, "masterpass", in.MasterPassword)
			return &proto.TokenResponse{Token: expectedToken}, nil
		},
	}

	client := &Client{
		AuthClient: mockAuth,
	}

	token, err := client.Login(context.Background(), "testuser", "testpass", []byte("masterpass"))
	assert.NoError(t, err)
	assert.Equal(t, expectedToken, token)
	assert.Equal(t, expectedToken, client.authToken)
}

func TestClient_Login_Error(t *testing.T) {
	expectedErr := errors.New("login error")
	mockAuth := &mockAuthServiceClient{
		loginFunc: func(ctx context.Context, in *proto.LoginRequest, opts ...grpc.CallOption) (*proto.TokenResponse, error) {
			return nil, expectedErr
		},
	}

	client := &Client{
		AuthClient: mockAuth,
	}

	token, err := client.Login(context.Background(), "testuser", "testpass", []byte("masterpass"))
	assert.Error(t, err)
	assert.Equal(t, expectedErr, err)
	assert.Empty(t, token)
	assert.Empty(t, client.authToken)
}

func TestClient_Register(t *testing.T) {
	expectedToken := "test-token"
	mockAuth := &mockAuthServiceClient{
		registerFunc: func(ctx context.Context, in *proto.RegisterRequest, opts ...grpc.CallOption) (*proto.TokenResponse, error) {
			assert.Equal(t, "testuser", in.Login)
			assert.Equal(t, "testpass", in.Password)
			assert.Equal(t, "masterpass", in.MasterPassword)
			return &proto.TokenResponse{Token: expectedToken}, nil
		},
	}

	client := &Client{
		AuthClient: mockAuth,
	}

	token, err := client.Register(context.Background(), "testuser", "testpass", []byte("masterpass"))
	assert.NoError(t, err)
	assert.Equal(t, expectedToken, token)
	assert.Equal(t, expectedToken, client.authToken)
}

func TestClient_Register_Error(t *testing.T) {
	expectedErr := errors.New("register error")
	mockAuth := &mockAuthServiceClient{
		registerFunc: func(ctx context.Context, in *proto.RegisterRequest, opts ...grpc.CallOption) (*proto.TokenResponse, error) {
			return nil, expectedErr
		},
	}

	client := &Client{
		AuthClient: mockAuth,
	}

	token, err := client.Register(context.Background(), "testuser", "testpass", []byte("masterpass"))
	assert.Error(t, err)
	assert.Equal(t, expectedErr, err)
	assert.Empty(t, token)
	assert.Empty(t, client.authToken)
}

func TestClient_Upsert(t *testing.T) {
	now := time.Now()
	deletedAt := time.Now().Add(time.Hour)
	userData := &model.UserData{
		DataKey:   "test-key",
		DataValue: []byte("test-value"),
		UpdatedAt: now,
		DeletedAt: deletedAt,
	}

	expectedResponse := &proto.DataResponse{}

	mockData := &mockDataServiceClient{
		upsertFunc: func(ctx context.Context, in *proto.UpsertRequest, opts ...grpc.CallOption) (*proto.DataResponse, error) {
			md, ok := metadata.FromOutgoingContext(ctx)
			assert.True(t, ok)
			assert.Contains(t, md["authorization"], "Bearer test-token")

			assert.Equal(t, "test-key", in.DataKey)
			assert.Equal(t, []byte("test-value"), in.DataValue)
			assert.Equal(t, timestamppb.New(now), in.UpdatedAt)
			assert.Equal(t, timestamppb.New(deletedAt), in.DeletedAt)
			return expectedResponse, nil
		},
	}

	client := &Client{
		DataClient: mockData,
		authToken:  "test-token",
	}

	response, err := client.Upsert(context.Background(), userData)
	assert.NoError(t, err)
	assert.Equal(t, expectedResponse, response)
}

func TestClient_Upsert_Error(t *testing.T) {
	expectedErr := errors.New("upsert error")
	mockData := &mockDataServiceClient{
		upsertFunc: func(ctx context.Context, in *proto.UpsertRequest, opts ...grpc.CallOption) (*proto.DataResponse, error) {
			return nil, expectedErr
		},
	}

	client := &Client{
		DataClient: mockData,
		authToken:  "test-token",
	}

	userData := &model.UserData{
		DataKey:   "test-key",
		DataValue: []byte("test-value"),
	}

	response, err := client.Upsert(context.Background(), userData)
	assert.Error(t, err)
	assert.Equal(t, expectedErr, err)
	assert.Nil(t, response)
}

func TestClient_GetUpdates(t *testing.T) {
	timestamp := time.Now().Add(-time.Hour)
	expectedResponse := &proto.DataListResponse{
		Items: []*proto.DataResponse{
			{
				DataKey:   "key1",
				DataValue: []byte("value1"),
			},
		},
	}

	mockData := &mockDataServiceClient{
		getUpdatesFunc: func(ctx context.Context, in *proto.GetUpdatesRequest, opts ...grpc.CallOption) (*proto.DataListResponse, error) {
			md, ok := metadata.FromOutgoingContext(ctx)
			assert.True(t, ok)
			assert.Contains(t, md["authorization"], "Bearer test-token")

			assert.Equal(t, timestamppb.New(timestamp), in.UpdatedAfter)
			return expectedResponse, nil
		},
	}

	client := &Client{
		DataClient: mockData,
		authToken:  "test-token",
	}

	response, err := client.GetUpdates(context.Background(), timestamp)
	assert.NoError(t, err)
	assert.Equal(t, expectedResponse, response)
}

func TestClient_GetUpdates_Error(t *testing.T) {
	expectedErr := errors.New("get updates error")
	mockData := &mockDataServiceClient{
		getUpdatesFunc: func(ctx context.Context, in *proto.GetUpdatesRequest, opts ...grpc.CallOption) (*proto.DataListResponse, error) {
			return nil, expectedErr
		},
	}

	client := &Client{
		DataClient: mockData,
		authToken:  "test-token",
	}

	timestamp := time.Now().Add(-time.Hour)
	response, err := client.GetUpdates(context.Background(), timestamp)
	assert.Error(t, err)
	assert.Equal(t, expectedErr, err)
	assert.Nil(t, response)
}

func TestClient_withAuth(t *testing.T) {
	client := &Client{
		authToken: "test-token",
	}

	ctx := context.Background()
	newCtx := client.withAuth(ctx)

	md, ok := metadata.FromOutgoingContext(newCtx)
	assert.True(t, ok)
	assert.Contains(t, md["authorization"], "Bearer test-token")
}

func TestClient_withAuth_EmptyToken(t *testing.T) {
	client := &Client{
		authToken: "",
	}

	ctx := context.Background()
	newCtx := client.withAuth(ctx)

	md, ok := metadata.FromOutgoingContext(newCtx)
	assert.False(t, ok)
	assert.Empty(t, md)
	assert.Equal(t, ctx, newCtx)
}
