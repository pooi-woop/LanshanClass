// FilePath: C:/LanshanClass1.3/api\main.go

package main

import (
	"LanshanClass1.3/api/controllers"
	"LanshanClass1.3/global/database"
	"LanshanClass1.3/utils"
	"github.com/dgrijalva/jwt-go"
	"log"

	"LanshanClass1.3/api/routers"
	"LanshanClass1.3/proto"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	// 初始化 gRPC 客户端
	conn, err := grpc.Dial("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("failed to connect: %v", err)
	}
	defer conn.Close()
	controllers.AuthServiceClient = proto.NewAuthServiceClient(conn)

	database.Init()
	r := gin.Default()

	// 注册鉴权中间件
	r.Use(func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(401, gin.H{"error": "Authorization header is required"})
			c.Abort()
			return
		}
		token := authHeader[len("Bearer "):]
		claims := &utils.Claims{}

		// 解析 JWT Token
		_, err := jwt.ParseWithClaims(token, claims, func(token *jwt.Token) (interface{}, error) {
			return utils.JwtSecret, nil
		})
		if err != nil {
			c.JSON(401, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}
		c.Next()
	})

	routers.AuthRouter(r)
	routers.LiveRouter(r)

	r.Run(":8080")
}
