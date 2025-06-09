// auth.service.go
package authservice

import (
	"LanshanClass1.3/global/database"
	"LanshanClass1.3/proto"
	"LanshanClass1.3/utils"
	"context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// AuthService 实现了 proto.AuthServiceServer 接口
type AuthService struct {
	proto.UnimplementedAuthServiceServer
}

// Register 注册方法
func (s *AuthService) Register(ctx context.Context, req *proto.RegisterRequest) (*proto.RegisterResponse, error) {
	err := database.CreateUser(req.Username, req.Password)
	if err != nil {
		return nil, err
	}

	token := utils.GenerateToken(req.Username)
	return &proto.RegisterResponse{
		Token:   token,
		Message: "注册成功",
	}, nil
}

// Login 登录方法
func (s *AuthService) Login(ctx context.Context, req *proto.LoginRequest) (*proto.LoginResponse, error) {
	rightornot, err := database.VerifyPassword(req.Username, req.Password)
	if err != nil {
		return nil, err
	}

	if !rightornot {
		return nil, status.Errorf(codes.InvalidArgument, "用户名或密码错误")
	}
	token := utils.GenerateToken(req.Username)
	return &proto.LoginResponse{
		Token:   token,
		Message: "登录成功",
	}, nil
}
