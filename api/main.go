// FilePath: C:/LanshanClass1.3/api\main.go
package main

import (
	"LanshanClass1.3/api/controllers"
	"LanshanClass1.3/global/database"
	"log"

	"LanshanClass1.3/api/routers"
	"LanshanClass1.3/proto"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
)

func main() {
	// 初始化 gRPC 客户端
	conn, err := grpc.Dial("localhost:50051", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("failed to connect: %v", err)
	}
	defer conn.Close()
	controllers.AuthServiceClient = proto.NewAuthServiceClient(conn)
	/*con, err := grpc.Dial("localhost:50052", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("failed to connect: %v", err)
	}
	defer con.Close()
	controllers.LiveServiceClient = proto.NewAuthServiceClient(con)*/
	database.Init()
	r := gin.Default()
	routers.AuthRouter(r)
	routers.LiveRouter(r)

	r.Run(":8080")
}
