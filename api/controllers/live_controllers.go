// FilePath: C:/LanshanClass1.3/api/controllers/live_controllers.go
// FilePath: C:/LanshanClass1.3/api/controllers/live_controllers.go
package controllers

import (
	"LanshanClass1.3/proto"
	"LanshanClass1.3/utils"
	"context"
	"encoding/json"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
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
	conn, err := grpc.Dial(grpcServerAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Printf("gRPC dial failed: %v", err)
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

	// 更健壮的Token提取方式
	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		return "", errors.New("invalid authorization header format")
	}

	return parts[1], nil
}

// GetMessages 获取消息
func GetMessages(c *gin.Context) {
	classID := c.Query("class_id")
	lastTimestampStr := c.Query("last_timestamp")

	if classID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "class_id is required"})
		return
	}

	var lastTimestamp int64
	if lastTimestampStr != "" {
		var err error
		lastTimestamp, err = strconv.ParseInt(lastTimestampStr, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid last_timestamp"})
			return
		}
	}

	client, conn, err := createGRPCClient(c)
	if err != nil {
		log.Printf("Failed to create gRPC client: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}
	defer conn.Close()

	token, err := extractToken(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	ctx := metadata.AppendToOutgoingContext(context.Background(), "authorization", token)

	// 调用 gRPC 服务获取消息
	resp, err := client.GetMessages(ctx, &proto.GetMessagesRequest{
		ClassId:       classID,
		LastTimestamp: lastTimestamp,
	})
	if err != nil {
		log.Printf("GetMessages gRPC call failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get messages"})
		return
	}

	// 将消息转换为更友好的结构
	var messages []map[string]interface{}
	for _, msg := range resp.Messages {
		messages = append(messages, map[string]interface{}{
			"sender":    msg.SenderName,
			"content":   msg.MessageContent, // 确保包含消息内容
			"timestamp": msg.Timestamp,
		})
	}

	log.Printf("Returning %d messages for class %s", len(messages), classID)

	c.JSON(http.StatusOK, gin.H{
		"messages":       messages,
		"last_timestamp": getLastTimestamp(messages), // 返回最新的时间戳供客户端使用
	})
}

// 获取最新消息的时间戳
func getLastTimestamp(messages []map[string]interface{}) int64 {
	if len(messages) == 0 {
		return 0
	}
	lastMsg := messages[len(messages)-1]
	return lastMsg["timestamp"].(int64)
}

// CreateLiveClass 创建直播课
// CreateLiveClass 创建直播课
func CreateLiveClass(c *gin.Context) {
	var req proto.CreateLiveClassRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.ClassName == "" || req.RoomName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "class_name and room_name are required"})
		return
	}

	client, conn, err := createGRPCClient(c)
	if err != nil {
		return
	}
	defer conn.Close()

	// 获取JWT Token
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header is required"})
		return
	}

	// 确保Authorization头格式正确
	if !strings.HasPrefix(authHeader, "Bearer ") {
		authHeader = "Bearer " + authHeader
	}

	// 创建带认证信息的上下文
	md := metadata.New(map[string]string{
		"authorization": authHeader,
	})
	ctx := metadata.NewOutgoingContext(context.Background(), md)

	// 调用gRPC服务
	resp, err := client.CreateLiveClass(ctx, &req)
	if err != nil {
		log.Printf("CreateLiveClass failed: %v", err)

		// 添加详细错误信息
		var errorMsg string
		if status, ok := status.FromError(err); ok {
			errorMsg = status.Message()
		} else {
			errorMsg = err.Error()
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "failed to create live class",
			"details": errorMsg,
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"classID":   resp.ClassId,
		"status":    resp.Status,
		"streamKey": resp.StreamKey,
	})
}

// JoinLiveClass 加入直播课
func JoinLiveClass(c *gin.Context) {
	var req proto.JoinLiveClassRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("BindJSON error: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	log.Printf("Join request received: %+v", req)

	client, conn, err := createGRPCClient(c)
	if err != nil {
		log.Printf("Failed to create gRPC client: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}
	defer conn.Close()

	token, err := extractToken(c)
	if err != nil {
		log.Printf("Failed to extract token: %v", err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	log.Printf("Using token: %s", token)

	ctx := metadata.AppendToOutgoingContext(context.Background(), "authorization", "Bearer "+token)

	resp, err := client.JoinLiveClass(ctx, &req)
	if err != nil {
		log.Printf("gRPC call failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "failed to join live class",
			"details": err.Error(), // 添加详细错误信息
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":     resp.Status,
		"stream_url": resp.StreamUrl,
		"message":    resp.Message,
	})
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
	ctx := metadata.AppendToOutgoingContext(context.Background(), "authorization", c.GetHeader("Authorization"))
	// 将 Token 传递给 gRPC 客户端

	resp, err := client.SendMessage(ctx, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to call gRPC service"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": resp.Status})
}

// EndLiveClass 结束直播课
func EndLiveClass(c *gin.Context) {
	// 定义请求结构体
	type Request struct {
		ClassID string `json:"class_id"`
	}

	var req Request
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body", "details": err.Error()})
		return
	}

	if req.ClassID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "class_id is required"})
		return
	}

	client, conn, err := createGRPCClient(c)
	if err != nil {
		log.Printf("创建gRPC客户端失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}
	defer conn.Close()

	token, err := extractToken(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	claims := &utils.Claims{}
	_, err = jwt.ParseWithClaims(token, claims, func(token *jwt.Token) (interface{}, error) {
		return utils.JwtSecret, nil
	})
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
		return
	}
	username := claims.Username

	// 创建带有超时的上下文
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 创建带认证信息的上下文
	md := metadata.New(map[string]string{"authorization": "Bearer " + token})
	ctx = metadata.NewOutgoingContext(ctx, md)

	resp, err := client.EndLiveClass(ctx, &proto.EndLiveClassRequest{
		ClassId:  req.ClassID,
		Username: username,
	})
	if err != nil {
		log.Printf("EndLiveClass gRPC调用失败: %v", err)
		// 尝试解析gRPC错误
		if status, ok := status.FromError(err); ok {
			c.JSON(http.StatusInternalServerError, gin.H{"error": status.Message()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to end live class"})
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

	if req.Question == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "question is required"})
		return
	}

	client, conn, err := createGRPCClient(c)
	if err != nil {
		return
	}
	defer conn.Close()

	// 从HTTP请求头获取Authorization
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authorization header is required"})
		return
	}

	// 直接传递整个Authorization头
	ctx := metadata.AppendToOutgoingContext(context.Background(), "authorization", authHeader)

	resp, err := client.PublishQuestion(ctx, &req)
	if err != nil {
		log.Printf("PublishQuestion failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "failed to publish question",
			"details": err.Error(), // 返回具体错误信息
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":      resp.Status,
		"question_id": resp.QuestionId,
	})
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
		log.Printf("创建gRPC客户端失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}
	defer conn.Close()

	token, err := extractToken(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	// 创建带超时的上下文
	ctx, cancel := context.WithTimeout(c.Request.Context(), 30*time.Second)
	defer cancel()

	// 创建带认证信息的上下文
	md := metadata.New(map[string]string{"authorization": "Bearer " + token})
	ctx = metadata.NewOutgoingContext(ctx, md)

	// 调用gRPC流式方法
	stream, err := client.GetAnswerStatistics(ctx, &proto.GetAnswerStatisticsRequest{
		ClassId:    classID,
		QuestionId: questionID,
	})
	if err != nil {
		log.Printf("gRPC流创建失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get statistics"})
		return
	}

	// 设置流式响应头
	c.Header("Content-Type", "application/x-ndjson")
	c.Status(http.StatusOK)

	// 流式传输
	for {
		// 从gRPC流接收
		stat, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Printf("接收统计错误: %v", err)
			break
		}

		// 将统计结果转换为JSON
		jsonData, err := json.Marshal(stat)
		if err != nil {
			log.Printf("JSON编码错误: %v", err)
			break
		}

		// 写入HTTP响应流（每条记录后加换行符）
		if _, err := c.Writer.Write(append(jsonData, '\n')); err != nil {
			log.Printf("写入响应流失败: %v", err)
			break
		}

		// 刷新响应缓冲区
		c.Writer.Flush()
	}
}
