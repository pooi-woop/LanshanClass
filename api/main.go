// FilePath: C:/LanshanClass1.3/api\main.go

package main

import (
	"LanshanClass1.3/api/controllers"
	"LanshanClass1.3/api/routers"
	"LanshanClass1.3/global/database"
	"LanshanClass1.3/proto"
	"LanshanClass1.3/utils"
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-contrib/cors"
	"log"
	"time"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	database.Init()
	// 初始化 gRPC 客户端
	conn, err := grpc.Dial("127.0.0.1:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("failed to connect: %v", err)
	}
	defer conn.Close()
	controllers.AuthServiceClient = proto.NewAuthServiceClient(conn)

	r := gin.Default()

	// 注册鉴权中间件
	r.Use(func(c *gin.Context) {
		// 获取当前请求的路径
		path := c.Request.URL.Path
		// 获取请求方法
		method := c.Request.Method

		// 如果是注册或登录请求，则跳过认证
		if (path == "/register" || path == "/login") && method == "POST" {
			c.Next()
			return
		}

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
	r.Use(cors.New(cors.Config{
		AllowAllOrigins:  true, // 允许所有来源
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"},
		AllowHeaders:     []string{"*"}, // 允许所有头
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))
	routers.AuthRouter(r)
	routers.LiveRouter(r)

	r.Run(":8080")
}
