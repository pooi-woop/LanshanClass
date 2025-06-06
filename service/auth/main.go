// FilePath: C:/LanshanClass1.3/service/auth\main.go
package main

import (
	"LanshanClass1.3/global/database"
	"LanshanClass1.3/service/auth/authservice"
	"log"
	"net"

	"LanshanClass1.3/proto"
	"google.golang.org/grpc"
)

func main() {
	database.Init()
	// 定义 gRPC 服务监听的地址
	lis, err := net.Listen("tcp", "localhost:50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	// 创建 gRPC 服务器
	s := grpc.NewServer()
	// 注册 AuthService 服务
	proto.RegisterAuthServiceServer(s, &authservice.AuthService{})
	log.Println("gRPC server started at :50051")
	// 启动 gRPC 服务
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
