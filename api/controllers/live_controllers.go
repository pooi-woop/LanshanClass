// FilePath: C:/LanshanClass1.3/api/controllers/live_controllers.go
// FilePath: C:/LanshanClass1.3/api/controllers/live_controllers.go
package controllers

import (
	"LanshanClass1.3/utils"
	"context"
	"errors"
	"github.com/golang-jwt/jwt/v4"
	"io"
	"net/http"

	"LanshanClass1.3/proto"
	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

var (
	// 假设 gRPC 服务地址
	grpcServerAddress = "127.0.0.1:50052"
)

// dialGRPC 连接到 gRPC 服务
func dialGRPC() (*grpc.ClientConn, error) {
	return grpc.Dial(grpcServerAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
}

// createGRPCClient 创建 gRPC 客户端
func createGRPCClient(c *gin.Context) (proto.LiveClassServiceClient, *grpc.ClientConn, error) {
	conn, err := dialGRPC()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to connect to gRPC server"})
		return nil, nil, err
	}
	return proto.NewLiveClassServiceClient(conn), conn, nil
}

// extractToken 从 HTTP 请求中提取 JWT Token
func extractToken(c *gin.Context) (string, error) {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		return "", errors.New("authorization header is required")
	}
	token := authHeader[len("Bearer "):]
	if len(token) == 0 {
		return "", errors.New("token is required")
	}
	return token, nil
}

// CreateLiveClass 创建直播课
func CreateLiveClass(c *gin.Context) {
	var req proto.CreateLiveClassRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	client, conn, err := createGRPCClient(c)
	if err != nil {
		return
	}
	defer conn.Close()

	// 获取 JWT Token
	token, err := extractToken(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	// 将 Token 传递给 gRPC 客户端
	ctx := metadata.AppendToOutgoingContext(context.Background(), "authorization", token)

	// 调用 gRPC 服务创建直播课
	resp, err := client.CreateLiveClass(ctx, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to call gRPC service"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"classID":   resp.ClassId,
		"status":    resp.Status,
		"streamKey": resp.StreamKey, // 返回推流密钥给用户
	})
}

// JoinLiveClass 加入直播课
func JoinLiveClass(c *gin.Context) {
	var req proto.JoinLiveClassRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	client, conn, err := createGRPCClient(c)
	if err != nil {
		return
	}
	defer conn.Close()

	// 获取 JWT Token
	token, err := extractToken(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	// 将 Token 传递给 gRPC 客户端
	ctx := metadata.AppendToOutgoingContext(context.Background(), "authorization", token)

	// 创建双向流
	stream, err := client.JoinLiveClass(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to call gRPC service"})
		return
	}
	defer stream.CloseSend()

	// 发送请求
	if err := stream.Send(&req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to send request"})
		return
	}
	stream.CloseSend()

	// 接收响应
	go func() {
		for {
			in, err := stream.Recv()
			if err == io.EOF {
				break
			}
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to receive response"})
				return
			}
			// 处理响应
			c.JSON(http.StatusOK, gin.H{"status": in.Status, "streamUrl": in.StreamUrl})
		}
	}()
}

// SendMessage 发送消息
func SendMessage(c *gin.Context) {
	var req proto.SendMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	client, conn, err := createGRPCClient(c)
	if err != nil {
		return
	}
	defer conn.Close()

	// 获取 JWT Token
	token, err := extractToken(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	// 解析 JWT Token 获取用户名
	claims := &utils.Claims{}
	_, err = jwt.ParseWithClaims(token, claims, func(token *jwt.Token) (interface{}, error) {
		return utils.JwtSecret, nil
	})
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
		return
	}
	req.SenderName = claims.Username

	// 将 Token 传递给 gRPC 客户端
	ctx := metadata.AppendToOutgoingContext(context.Background(), "authorization", token)

	resp, err := client.SendMessage(ctx, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to call gRPC service"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": resp.Status})
}

// EndLiveClass 结束直播课
func EndLiveClass(c *gin.Context) {
	classID := c.Query("class_id")
	if classID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "class_id is required"})
		return
	}

	client, conn, err := createGRPCClient(c)
	if err != nil {
		return
	}
	defer conn.Close()

	// 获取 JWT Token
	token, err := extractToken(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	// 解析 JWT Token 获取用户名
	claims := &utils.Claims{}
	_, err = jwt.ParseWithClaims(token, claims, func(token *jwt.Token) (interface{}, error) {
		return utils.JwtSecret, nil
	})
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
		return
	}
	username := claims.Username

	// 将 Token 传递给 gRPC 客户端
	ctx := metadata.AppendToOutgoingContext(context.Background(), "authorization", token)

	// 调用 gRPC 服务结束直播课
	resp, err := client.EndLiveClass(ctx, &proto.EndLiveClassRequest{
		ClassId:  classID,
		Username: username, // 传递用户名
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to call gRPC service"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": resp.Status})
}

// PublishQuestion 发布题目
func PublishQuestion(c *gin.Context) {
	var req proto.PublishQuestionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	client, conn, err := createGRPCClient(c)
	if err != nil {
		return
	}
	defer conn.Close()

	// 获取 JWT Token
	token, err := extractToken(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	// 将 Token 传递给 gRPC 客户端
	ctx := metadata.AppendToOutgoingContext(context.Background(), "authorization", token)

	resp, err := client.PublishQuestion(ctx, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to call gRPC service"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": resp.Status})
}

// SubmitAnswer 提交答案
func SubmitAnswer(c *gin.Context) {
	var req proto.SubmitAnswerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	client, conn, err := createGRPCClient(c)
	if err != nil {
		return
	}
	defer conn.Close()

	// 获取 JWT Token
	token, err := extractToken(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	// 解析 JWT Token 获取用户名
	claims := &utils.Claims{}
	_, err = jwt.ParseWithClaims(token, claims, func(token *jwt.Token) (interface{}, error) {
		return utils.JwtSecret, nil
	})
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
		return
	}
	req.StudentName = claims.Username

	// 将 Token 传递给 gRPC 客户端
	ctx := metadata.AppendToOutgoingContext(context.Background(), "authorization", token)

	resp, err := client.SubmitAnswer(ctx, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to call gRPC service"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": resp.Status})
}

// GetAnswerStatistics 获取答题结果统计（流式接口）
func GetAnswerStatistics(c *gin.Context) {
	classID := c.Query("class_id")
	questionID := c.Query("question_id")
	if classID == "" || questionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "class_id and question_id are required"})
		return
	}

	client, conn, err := createGRPCClient(c)
	if err != nil {
		return
	}
	defer conn.Close()

	// 获取 JWT Token
	token, err := extractToken(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	// 将 Token 传递给 gRPC 客户端
	ctx := metadata.AppendToOutgoingContext(context.Background(), "authorization", token)

	stream, err := client.GetAnswerStatistics(ctx, &proto.GetAnswerStatisticsRequest{ClassId: classID, QuestionId: questionID})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to call gRPC service"})
		return
	}

	// 返回流式响应
	c.Stream(func(w io.Writer) bool {
		in, err := stream.Recv()
		if err == io.EOF {
			return false
		}
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to receive answer statistics"})
			return false
		}

		// 将统计结果写入 HTTP 响应流
		_, err = w.Write([]byte(in.String()))
		return err == nil
	})
}
