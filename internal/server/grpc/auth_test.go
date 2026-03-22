package grpc

import (
	"context"
	"errors"
	"testing"

	"github.com/g123udini/gophkeeper/internal/server/jwt"
	"github.com/g123udini/gophkeeper/internal/server/service"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type mockAuthUserManager struct {
	decodeTokenFunc func(token string) (*jwt.Claims, error)
}

func (m *mockAuthUserManager) DecodeToken(token string) (*jwt.Claims, error) {
	return m.decodeTokenFunc(token)
}

// Реализация интерфейса UserHandlerInterface
func (m *mockAuthUserManager) Register(login, password, masterPassword string) (string, error) {
	return "", errors.New("not implemented")
}

func (m *mockAuthUserManager) Login(login, password, masterPassword string) (string, error) {
	return "", errors.New("not implemented")
}

func TestNewAuthInterceptor(t *testing.T) {
	um := &service.UserService{}
	ai := NewAuthInterceptor(um)

	if ai.userManager != um {
		t.Errorf("NewAuthInterceptor() userManager = %v, want %v", ai.userManager, um)
	}
}

func TestAuthInterceptor_Unary(t *testing.T) {
	tests := []struct {
		name          string
		setupMock     func() UserHandlerInterface
		setupContext  func() context.Context
		wantErrCode   codes.Code
		wantErrString string
	}{
		{
			name: "no metadata",
			setupMock: func() UserHandlerInterface {
				return &mockAuthUserManager{}
			},
			setupContext: func() context.Context {
				return context.Background()
			},
			wantErrCode:   codes.Unauthenticated,
			wantErrString: "metadata is not provided",
		},
		{
			name: "no auth token",
			setupMock: func() UserHandlerInterface {
				return &mockAuthUserManager{}
			},
			setupContext: func() context.Context {
				md := metadata.New(map[string]string{})
				return metadata.NewIncomingContext(context.Background(), md)
			},
			wantErrCode:   codes.Unauthenticated,
			wantErrString: "authorization token is missing",
		},
		{
			name: "invalid token",
			setupMock: func() UserHandlerInterface {
				return &mockAuthUserManager{
					decodeTokenFunc: func(token string) (*jwt.Claims, error) {
						return nil, errors.New("invalid token")
					},
				}
			},
			setupContext: func() context.Context {
				md := metadata.New(map[string]string{
					"authorization": "Bearer token123",
				})
				return metadata.NewIncomingContext(context.Background(), md)
			},
			wantErrCode:   codes.Unauthenticated,
			wantErrString: "invalid token",
		},
		{
			name: "valid token",
			setupMock: func() UserHandlerInterface {
				return &mockAuthUserManager{
					decodeTokenFunc: func(token string) (*jwt.Claims, error) {
						return &jwt.Claims{SubjectID: uint32(123)}, nil
					},
				}
			},
			setupContext: func() context.Context {
				md := metadata.New(map[string]string{
					"authorization": "Bearer token123",
				})
				return metadata.NewIncomingContext(context.Background(), md)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUM := tt.setupMock()
			ai := &AuthInterceptor{
				userManager: mockUM,
			}

			interceptor := ai.Unary()
			ctx := tt.setupContext()

			handler := func(ctx context.Context, req interface{}) (interface{}, error) {
				claims, err := GetClaimsFromContext(ctx)
				if err != nil {
					return nil, err
				}
				return claims, nil
			}

			resp, err := interceptor(ctx, "request", &grpc.UnaryServerInfo{}, handler)

			if tt.wantErrString != "" {
				if err == nil {
					t.Fatalf("Unary() error = nil, wantErr %v", tt.wantErrString)
				}

				statusErr, ok := status.FromError(err)
				if !ok {
					t.Fatalf("Unary() error is not a status error")
				}

				if statusErr.Code() != tt.wantErrCode {
					t.Errorf("Unary() error code = %v, want %v", statusErr.Code(), tt.wantErrCode)
				}

				if statusErr.Message() != tt.wantErrString {
					t.Errorf("Unary() error message = %v, want %v", statusErr.Message(), tt.wantErrString)
				}
			} else {
				if err != nil {
					t.Fatalf("Unary() error = %v, wantErr nil", err)
				}

				claims, ok := resp.(*jwt.Claims)
				if !ok {
					t.Fatalf("Unary() response is not Claims type")
				}

				if claims.SubjectID != uint32(123) {
					t.Errorf("Unary() claims.SubjectID = %v, want %v", claims.SubjectID, uint32(123))
				}
			}
		})
	}
}

func TestGetClaimsFromContext(t *testing.T) {
	tests := []struct {
		name    string
		ctx     context.Context
		want    *jwt.Claims
		wantErr bool
	}{
		{
			name:    "no claims in context",
			ctx:     context.Background(),
			want:    nil,
			wantErr: true,
		},
		{
			name: "claims in context",
			ctx: context.WithValue(
				context.Background(),
				userClaimsKey{},
				&jwt.Claims{SubjectID: uint32(123)},
			),
			want:    &jwt.Claims{SubjectID: uint32(123)},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetClaimsFromContext(tt.ctx)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetClaimsFromContext() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got.SubjectID != tt.want.SubjectID {
				t.Errorf("GetClaimsFromContext() = %v, want %v", got, tt.want)
			}
		})
	}
}
