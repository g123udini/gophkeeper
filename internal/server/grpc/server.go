package grpc

import (
	"context"
	"errors"

	"github.com/g123udini/gophkeeper/internal/common/logger"
	"github.com/g123udini/gophkeeper/internal/common/proto"
	"github.com/g123udini/gophkeeper/internal/server/model"
	"github.com/g123udini/gophkeeper/internal/server/service"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type Server struct {
	proto.UnimplementedAuthServiceServer
	proto.UnimplementedDataServiceServer

	userManager UserHandlerInterface
	dataManager UserDataManagerInterface
}

func NewServer(
	userManager *service.UserService,
	dataManager *service.UserDataService,
) *Server {
	return &Server{
		userManager: userManager,
		dataManager: dataManager,
	}
}

func (s *Server) Register(ctx context.Context, req *proto.RegisterRequest) (*proto.TokenResponse, error) {
	token, err := s.userManager.Register(req.Login, req.Password, req.MasterPassword)
	if err != nil {
		return nil, convertError(err)
	}

	return &proto.TokenResponse{Token: token}, nil
}

func (s *Server) Login(ctx context.Context, req *proto.LoginRequest) (*proto.TokenResponse, error) {
	token, err := s.userManager.Login(req.Login, req.Password, req.MasterPassword)
	if err != nil {
		return nil, convertError(err)
	}

	return &proto.TokenResponse{Token: token}, nil
}

func (s *Server) Upsert(ctx context.Context, req *proto.UpsertRequest) (*proto.DataResponse, error) {
	claims, err := GetClaimsFromContext(ctx)
	if err != nil {
		return nil, err
	}

	data := &model.UserData{
		UserID:    claims.SubjectID,
		DataKey:   req.DataKey,
		DataValue: req.DataValue,
		UpdatedAt: req.UpdatedAt.AsTime(),
		DeletedAt: req.DeletedAt.AsTime(),
	}
	err = s.dataManager.Upsert(ctx, data)

	if err != nil {
		return nil, convertError(err)
	}

	return &proto.DataResponse{
		DataKey:   data.DataKey,
		DataValue: data.DataValue,
		UpdatedAt: timestamppb.New(data.UpdatedAt),
		DeletedAt: timestamppb.New(data.DeletedAt),
	}, nil
}

func (s *Server) GetUpdates(ctx context.Context, req *proto.GetUpdatesRequest) (*proto.DataListResponse, error) {
	claims, err := GetClaimsFromContext(ctx)
	if err != nil {
		return nil, err
	}

	updates, err := s.dataManager.GetUpdates(ctx, claims.SubjectID, req.UpdatedAfter.AsTime())
	if err != nil {
		return nil, convertError(err)
	}

	pbUpdates := make([]*proto.DataResponse, 0, len(updates))
	for _, data := range updates {
		pbUpdates = append(pbUpdates, &proto.DataResponse{
			DataKey:   data.DataKey,
			DataValue: data.DataValue,
			UpdatedAt: timestamppb.New(data.UpdatedAt),
			DeletedAt: timestamppb.New(data.DeletedAt),
		})
	}

	return &proto.DataListResponse{Items: pbUpdates}, nil
}

func convertError(err error) error {
	switch {
	case errors.Is(err, service.ErrUserExists):
		return status.Error(codes.AlreadyExists, err.Error())
	case errors.Is(err, service.ErrInvalidCredentials):
		return status.Error(codes.Unauthenticated, err.Error())
	default:
		logger.Logger.Error("error occurred", zap.Error(err))

		return status.Error(codes.Internal, "internal server error")
	}
}
