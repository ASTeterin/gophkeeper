package service

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"

	"github.com/ASTeterin/gophkeeper/internal/mock"
	"github.com/ASTeterin/gophkeeper/internal/model"
)

const (
	userLogin    = "user_login"
	userPassword = "user_password"
	hash         = "$2a$10$F1.UPmsh2bwaHH427GJ0A.Qmr3dWdDdLFhut/LiPaqAazvatOLyiK"
)

func Test_userService_Register(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	tests := []struct {
		name     string
		login    string
		password string
		setup    func(m *mock.MockUserRepository)
		expected error
	}{
		{
			name:     "positive test",
			login:    "new_login",
			password: "password",
			setup: func(m *mock.MockUserRepository) {
				m.EXPECT().
					GetByLogin(gomock.Any(), "new_login").
					Return(nil, model.ErrUserNotFound).
					Times(1)
				m.EXPECT().
					Store(gomock.Any(), gomock.Any()).
					Return(nil).
					Times(1)
			},
			expected: nil,
		},
		{
			name:     "user already exists",
			login:    "existing_login",
			password: "password",
			setup: func(m *mock.MockUserRepository) {
				existingUser := &model.User{
					UUID:     uuid.New(),
					Login:    "existing_login",
					PassHash: "",
				}
				m.EXPECT().
					GetByLogin(gomock.Any(), "existing_login").
					Return(existingUser, nil).
					Times(1)
			},
			expected: ErrUserExists,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := mock.NewMockUserRepository(ctrl)
			tt.setup(mockRepo)

			s := NewUserService(mockRepo)
			_, err := s.Register(context.TODO(), tt.login, tt.password)

			if tt.expected != nil {
				assert.ErrorIs(t, err, tt.expected)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func Test_userService_Authenticate(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	tests := []struct {
		name     string
		login    string
		password string
		setup    func(m *mock.MockUserRepository)
		expected error
	}{
		{
			name:     "positive test",
			login:    userLogin,
			password: userPassword,
			setup: func(m *mock.MockUserRepository) {
				user := &model.User{
					UUID:     uuid.New(),
					Login:    userLogin,
					PassHash: hash,
				}
				m.EXPECT().
					GetByLogin(gomock.Any(), userLogin).
					Return(user, nil).
					Times(1)
			},
			expected: nil,
		},
		{
			name:     "user not found",
			login:    "unknown_user",
			password: "password",
			setup: func(m *mock.MockUserRepository) {
				m.EXPECT().
					GetByLogin(gomock.Any(), "unknown_user").
					Return(nil, model.ErrUserNotFound).
					Times(1)
			},
			expected: ErrUserNotAuthenticate,
		},
		{
			name:     "wrong password",
			login:    userLogin,
			password: "wrong_password",
			setup: func(m *mock.MockUserRepository) {
				user := &model.User{
					UUID:     uuid.New(),
					Login:    userLogin,
					PassHash: hash,
				}
				m.EXPECT().
					GetByLogin(gomock.Any(), userLogin).
					Return(user, nil).
					Times(1)
			},
			expected: ErrUserNotAuthenticate,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := mock.NewMockUserRepository(ctrl)
			tt.setup(mockRepo)

			s := NewUserService(mockRepo)
			_, err := s.Authenticate(context.TODO(), tt.login, tt.password)

			if tt.expected != nil {
				assert.ErrorIs(t, err, tt.expected)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
