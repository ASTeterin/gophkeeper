package grpc

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb "github.com/ASTeterin/gophkeeper/api"
	"github.com/ASTeterin/gophkeeper/internal/contracts"
	"github.com/ASTeterin/gophkeeper/internal/mock"
	"github.com/ASTeterin/gophkeeper/internal/model"
)

func setupGRPCTest(t *testing.T) (*privateDataGRPCServer, *mock.MockPrivateDataService) {
	ctrl := gomock.NewController(t)
	mockSvc := mock.NewMockPrivateDataService(ctrl)
	server := NewPrivateDataGRPCServer(mockSvc).(*privateDataGRPCServer)
	return server, mockSvc
}

func Test_privateDataGRPCServer_Store(t *testing.T) {
	s, mockSvc := setupGRPCTest(t)
	ctx := context.Background()
	userID := uuid.New()

	tests := []struct {
		name     string
		req      *pb.StoreRequest
		setup    func()
		wantCode codes.Code
	}{
		{
			name: "success",
			req: &pb.StoreRequest{
				UserId:      userID.String(),
				DataKey:     "key1",
				Description: "desc",
				Data:        []byte("data"),
			},
			setup: func() {
				mockSvc.EXPECT().
					AddData(gomock.Any(), userID, "key1", "desc", []byte("data")).
					Return(&model.PrivateData{ID: uuid.New(), DataKey: "key1"}, nil).
					Times(1)
			},
			wantCode: codes.OK,
		},
		{
			name: "key exists",
			req: &pb.StoreRequest{
				UserId:  userID.String(),
				DataKey: "key1",
				Data:    []byte("data"),
			},
			setup: func() {
				mockSvc.EXPECT().
					AddData(gomock.Any(), userID, "key1", gomock.Any(), gomock.Any()).
					Return(nil, contracts.ErrKeyExists).
					Times(1)
			},
			wantCode: codes.AlreadyExists,
		},
		{
			name: "invalid user id",
			req: &pb.StoreRequest{
				UserId:  "invalid-uuid",
				DataKey: "key1",
			},
			setup:    func() {},
			wantCode: codes.InvalidArgument,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			_, err := s.Store(ctx, tt.req)

			if tt.wantCode != codes.OK {
				st, ok := status.FromError(err)
				assert.True(t, ok)
				assert.Equal(t, tt.wantCode, st.Code())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func Test_privateDataGRPCServer_GetByKey(t *testing.T) {
	s, mockSvc := setupGRPCTest(t)
	ctx := context.Background()
	userID := uuid.New()

	tests := []struct {
		name     string
		req      *pb.GetByKeyRequest
		setup    func()
		wantCode codes.Code
	}{
		{
			name: "success",
			req: &pb.GetByKeyRequest{
				UserId: userID.String(),
				Key:    "key1",
			},
			setup: func() {
				mockSvc.EXPECT().
					GetDataByKey(gomock.Any(), userID, "key1").
					Return(&model.PrivateData{ID: uuid.New(), DataKey: "key1"}, nil).
					Times(1)
			},
			wantCode: codes.OK,
		},
		{
			name: "not found",
			req: &pb.GetByKeyRequest{
				UserId: userID.String(),
				Key:    "missing",
			},
			setup: func() {
				mockSvc.EXPECT().
					GetDataByKey(gomock.Any(), userID, "missing").
					Return(nil, contracts.ErrNotFound).
					Times(1)
			},
			wantCode: codes.NotFound,
		},
		{
			name: "invalid user id",
			req: &pb.GetByKeyRequest{
				UserId: "invalid-uuid",
				Key:    "key1",
			},
			setup:    func() {},
			wantCode: codes.InvalidArgument,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			_, err := s.GetByKey(ctx, tt.req)

			if tt.wantCode != codes.OK {
				st, ok := status.FromError(err)
				assert.True(t, ok)
				assert.Equal(t, tt.wantCode, st.Code())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func Test_privateDataGRPCServer_GetAll(t *testing.T) {
	s, mockSvc := setupGRPCTest(t)
	ctx := context.Background()
	userID := uuid.New()

	tests := []struct {
		name     string
		req      *pb.GetAllRequest
		setup    func()
		wantCode codes.Code
	}{
		{
			name: "success",
			req: &pb.GetAllRequest{
				UserId: userID.String(),
			},
			setup: func() {
				mockSvc.EXPECT().
					GetAllData(gomock.Any(), userID).
					Return([]*model.PrivateData{}, nil).
					Times(1)
			},
			wantCode: codes.OK,
		},
		{
			name: "invalid user id",
			req: &pb.GetAllRequest{
				UserId: "invalid-uuid",
			},
			setup:    func() {},
			wantCode: codes.InvalidArgument,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			_, err := s.GetAll(ctx, tt.req)

			if tt.wantCode != codes.OK {
				st, ok := status.FromError(err)
				assert.True(t, ok)
				assert.Equal(t, tt.wantCode, st.Code())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func Test_privateDataGRPCServer_Delete(t *testing.T) {
	s, mockSvc := setupGRPCTest(t)
	ctx := context.Background()
	userID := uuid.New()

	tests := []struct {
		name     string
		req      *pb.DeleteRequest
		setup    func()
		wantCode codes.Code
	}{
		{
			name: "success",
			req: &pb.DeleteRequest{
				UserId: userID.String(),
				Key:    "key1",
			},
			setup: func() {
				mockSvc.EXPECT().
					DeleteData(gomock.Any(), userID, "key1").
					Return(nil).
					Times(1)
			},
			wantCode: codes.OK,
		},
		{
			name: "not found",
			req: &pb.DeleteRequest{
				UserId: userID.String(),
				Key:    "missing",
			},
			setup: func() {
				mockSvc.EXPECT().
					DeleteData(gomock.Any(), userID, "missing").
					Return(contracts.ErrNotFound).
					Times(1)
			},
			wantCode: codes.NotFound,
		},
		{
			name: "invalid user id",
			req: &pb.DeleteRequest{
				UserId: "invalid-uuid",
				Key:    "key1",
			},
			setup:    func() {},
			wantCode: codes.InvalidArgument,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			_, err := s.Delete(ctx, tt.req)

			if tt.wantCode != codes.OK {
				st, ok := status.FromError(err)
				assert.True(t, ok)
				assert.Equal(t, tt.wantCode, st.Code())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func Test_privateDataGRPCServer_ReplaceAll(t *testing.T) {
	s, mockSvc := setupGRPCTest(t)
	ctx := context.Background()
	userID := uuid.New()

	tests := []struct {
		name     string
		req      *pb.ReplaceAllRequest
		setup    func()
		wantCode codes.Code
	}{
		{
			name: "success",
			req: &pb.ReplaceAllRequest{
				UserId: userID.String(),
				Items: []*pb.DataItem{
					{DataKey: "k1", Data: []byte("d1")},
				},
			},
			setup: func() {
				mockSvc.EXPECT().
					ReplaceAllData(gomock.Any(), userID, gomock.Any()).
					Return(nil).
					Times(1)
			},
			wantCode: codes.OK,
		},
		{
			name: "invalid user id",
			req: &pb.ReplaceAllRequest{
				UserId: "invalid-uuid",
			},
			setup:    func() {},
			wantCode: codes.InvalidArgument,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			_, err := s.ReplaceAll(ctx, tt.req)

			if tt.wantCode != codes.OK {
				st, ok := status.FromError(err)
				assert.True(t, ok)
				assert.Equal(t, tt.wantCode, st.Code())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
