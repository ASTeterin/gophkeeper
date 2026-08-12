package grpc

import (
	"context"
	"errors"
	"fmt"

	pb "github.com/ASTeterin/gophkeeper/api"
	"github.com/ASTeterin/gophkeeper/internal/service"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type UserGRPCServer struct {
	pb.UnimplementedUserServiceServer
	service service.UserService
}

func NewUserGRPCServer(svc service.UserService) *UserGRPCServer {
	return &UserGRPCServer{service: svc}
}

func (s *UserGRPCServer) Register(ctx context.Context, req *pb.RegisterRequest) (*pb.RegisterResponse, error) {
	userID, err := s.service.Register(ctx, req.Login, req.Password)
	if err != nil {
		if errors.Is(err, service.ErrUserExists) {
			return nil, status.Error(codes.AlreadyExists, "user already exists")
		}
		return nil, status.Error(codes.Internal, fmt.Sprintf("registration failed: %v", err))
	}

	if userID == nil {
		return nil, status.Error(codes.Internal, "failed to generate user ID")
	}

	return &pb.RegisterResponse{
		UserId: *userID,
	}, nil
}

func (s *UserGRPCServer) Authenticate(ctx context.Context, req *pb.AuthenticateRequest) (*pb.AuthenticateResponse, error) {
	userID, err := s.service.Authenticate(ctx, req.Login, req.Password)
	if err != nil {
		if errors.Is(err, service.ErrUserNotAuthenticate) {
			return nil, status.Error(codes.Unauthenticated, "invalid login or password")
		}
		return nil, status.Error(codes.Internal, fmt.Sprintf("authentication failed: %v", err))
	}

	if userID == nil {
		return nil, status.Error(codes.Unauthenticated, "user not found")
	}

	return &pb.AuthenticateResponse{
		UserId: *userID,
	}, nil
}
