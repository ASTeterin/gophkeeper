package grpc

import (
	"context"
	"errors"

	"github.com/ASTeterin/gophkeeper/internal/model"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb "github.com/ASTeterin/gophkeeper/api"
	"github.com/ASTeterin/gophkeeper/internal/service"
)

type privateDataGRPCServer struct {
	pb.UnimplementedPrivateDataServiceServer
	service service.PrivateDataService
}

func NewPrivateDataGRPCServer(svc service.PrivateDataService) pb.PrivateDataServiceServer {
	return &privateDataGRPCServer{service: svc}
}

func (s *privateDataGRPCServer) Store(ctx context.Context, req *pb.StoreRequest) (*pb.StoreResponse, error) {
	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user id format")
	}

	_, err = s.service.AddData(ctx, userID, req.DataKey, req.Description, req.Data)
	if err != nil {
		if errors.Is(err, service.ErrKeyExists) {
			return nil, status.Error(codes.AlreadyExists, err.Error())
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.StoreResponse{}, nil
}

func (s *privateDataGRPCServer) GetByKey(ctx context.Context, req *pb.GetByKeyRequest) (*pb.PrivateData, error) {
	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user id format")
	}

	result, err := s.service.GetDataByKey(ctx, userID, req.Key)
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			return nil, status.Error(codes.NotFound, err.Error())
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	return toProtoPrivateData(result), nil
}

func (s *privateDataGRPCServer) GetAll(ctx context.Context, req *pb.GetAllRequest) (*pb.PrivateDataList, error) {
	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user id format")
	}

	results, err := s.service.GetAllData(ctx, userID)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	items := make([]*pb.PrivateData, 0, len(results))
	for _, item := range results {
		items = append(items, toProtoPrivateData(item))
	}

	return &pb.PrivateDataList{Items: items}, nil
}

func (s *privateDataGRPCServer) Delete(ctx context.Context, req *pb.DeleteRequest) (*pb.DeleteResponse, error) {
	id, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid id format")
	}

	if err := s.service.DeleteData(ctx, id); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.DeleteResponse{}, nil
}

func (s *privateDataGRPCServer) ReplaceAll(ctx context.Context, req *pb.ReplaceAllRequest) (*pb.ReplaceAllResponse, error) {
	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user id format")
	}

	items := make([]service.DataItem, 0, len(req.Items))
	for _, item := range req.Items {
		items = append(items, service.DataItem{
			DataKey:     item.DataKey,
			Description: item.Description,
			Data:        item.Data,
		})
	}

	if err := s.service.ReplaceAllData(ctx, userID, items); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.ReplaceAllResponse{}, nil
}

func toProtoPrivateData(data *model.PrivateData) *pb.PrivateData {
	return &pb.PrivateData{
		Id:          data.ID.String(),
		UserId:      data.UserID.String(),
		DataKey:     data.DataKey,
		Description: data.Description,
		Data:        data.Data,
	}
}
