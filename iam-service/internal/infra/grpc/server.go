package grpc

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	iamv1 "motrava/proto/gen/iam/v1"

	"motrava/iam-service/internal/app/middleware"
	"motrava/iam-service/internal/core/models"
	"motrava/iam-service/internal/core/repository"
)

type Server struct {
	iamv1.UnimplementedIAMServiceServer
	userRepo  repository.UserRepository
	jwtSecret string
	jwtIssuer string
	log       *slog.Logger
}

func NewServer(userRepo repository.UserRepository, jwtSecret, jwtIssuer string, logger *slog.Logger) *Server {
	return &Server{
		userRepo:  userRepo,
		jwtSecret: jwtSecret,
		jwtIssuer: jwtIssuer,
		log:       logger,
	}
}

func (s *Server) Run(port string) error {
	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		return fmt.Errorf("grpc listen: %w", err)
	}

	grpcServer := grpc.NewServer()
	iamv1.RegisterIAMServiceServer(grpcServer, s)

	s.log.Info("grpc server starting", "module", "grpc_server", "port", port)
	return grpcServer.Serve(lis)
}

func (s *Server) ValidateToken(ctx context.Context, req *iamv1.ValidateTokenRequest) (*iamv1.ValidateTokenResponse, error) {
	claims := &middleware.AuthClaims{}
	token, err := jwt.ParseWithClaims(req.Token, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return []byte(s.jwtSecret), nil
	})
	if err != nil || !token.Valid {
		return &iamv1.ValidateTokenResponse{Valid: false}, nil
	}

	if claims.TokenType != "access" || claims.Subject == "" {
		return &iamv1.ValidateTokenResponse{Valid: false}, nil
	}

	if s.jwtIssuer != "" && claims.Issuer != s.jwtIssuer {
		return &iamv1.ValidateTokenResponse{Valid: false}, nil
	}

	userID, err := uuid.Parse(claims.Subject)
	if err != nil {
		return &iamv1.ValidateTokenResponse{Valid: false}, nil
	}

	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &iamv1.ValidateTokenResponse{Valid: false}, nil
		}
		s.log.Error("grpc find user failed", "module", "grpc_server", "error", err)
		return nil, status.Error(codes.Internal, "failed to find user")
	}

	fullName := ""
	if user.FullName != nil {
		fullName = *user.FullName
	}
	avatarURL := ""
	if user.AvatarURL != nil {
		avatarURL = *user.AvatarURL
	}

	return &iamv1.ValidateTokenResponse{
		Valid:        user.IsActive,
		UserId:       user.ID.String(),
		Email:        user.Email,
		FullName:     fullName,
		AuthProvider: user.AuthProvider,
		AvatarUrl:    avatarURL,
		IsActive:     user.IsActive,
		IsVerified:   user.IsVerified,
	}, nil
}

func (s *Server) GetUser(ctx context.Context, req *iamv1.GetUserRequest) (*iamv1.GetUserResponse, error) {
	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user id")
	}

	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, status.Error(codes.NotFound, "user not found")
		}
		s.log.Error("grpc get user failed", "module", "grpc_server", "error", err)
		return nil, status.Error(codes.Internal, "failed to get user")
	}

	return toUserResponse(user), nil
}

func (s *Server) GetUserByEmail(ctx context.Context, req *iamv1.GetUserByEmailRequest) (*iamv1.GetUserResponse, error) {
	user, err := s.userRepo.FindByEmail(req.Email)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, status.Error(codes.NotFound, "user not found")
		}
		s.log.Error("grpc get user by email failed", "module", "grpc_server", "error", err)
		return nil, status.Error(codes.Internal, "failed to get user")
	}

	return toUserResponse(user), nil
}

func (s *Server) GetUserByGoogleSub(ctx context.Context, req *iamv1.GetUserByGoogleSubRequest) (*iamv1.GetUserResponse, error) {
	user, err := s.userRepo.FindByGoogleSub(req.GoogleSub)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, status.Error(codes.NotFound, "user not found")
		}
		s.log.Error("grpc get user by google sub failed", "module", "grpc_server", "error", err)
		return nil, status.Error(codes.Internal, "failed to get user")
	}

	return toUserResponse(user), nil
}

func toUserResponse(user *models.User) *iamv1.GetUserResponse {
	fullName := ""
	if user.FullName != nil {
		fullName = *user.FullName
	}
	googleSub := ""
	if user.GoogleSub != nil {
		googleSub = *user.GoogleSub
	}
	avatarURL := ""
	if user.AvatarURL != nil {
		avatarURL = *user.AvatarURL
	}
	phoneNumber := ""
	if user.PhoneNumber != nil {
		phoneNumber = *user.PhoneNumber
	}

	return &iamv1.GetUserResponse{
		UserId:       user.ID.String(),
		Email:        user.Email,
		FullName:     fullName,
		AuthProvider: user.AuthProvider,
		GoogleSub:    googleSub,
		AvatarUrl:    avatarURL,
		PhoneNumber:  phoneNumber,
		IsActive:     user.IsActive,
		IsVerified:   user.IsVerified,
		CreatedAt:    user.CreatedAt.Format(time.RFC3339),
		UpdatedAt:    user.UpdatedAt.Format(time.RFC3339),
	}
}
