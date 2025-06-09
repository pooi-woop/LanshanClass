package liveservice

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"log"
	"net/http"
	"reflect"
	"strings"
	"sync"
	"time"

	pb "LanshanClass1.3/proto"
	"LanshanClass1.3/utils"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// LiveClassServiceServer 定义服务
type LiveClassServiceServer struct {
	streams map[string]*LiveClass
	mu      sync.Mutex
	pb.UnimplementedLiveClassServiceServer
}

// LiveClass 定义直播间结构
type LiveClass struct {
	TeacherName string                      // 直播间发起人的用户名
	StreamURL   string                      // 推流地址
	Messages    []*pb.Message               // 存储的消息列表
	mu          sync.Mutex                  // 保护消息列表的互斥锁
	Questions   map[string]*pb.Question     // 题目列表
	Answers     map[string]map[string]int32 // 答案统计
}

// NewLiveClassServiceServer 初始化服务
func NewLiveClassServiceServer() *LiveClassServiceServer {
	return &LiveClassServiceServer{
		streams: make(map[string]*LiveClass),
	}
}

// authenticateToken 验证 JWT Token 并返回用户名
func (s *LiveClassServiceServer) authenticateToken(ctx context.Context) (string, error) {
	// 获取元数据
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		log.Println("Error: Missing metadata in request")
		return "", status.Errorf(codes.Unauthenticated, "missing metadata")
	}

	// 详细记录元数据
	log.Printf("Received metadata: %+v", md)

	// 获取授权头
	authHeaders := md.Get("authorization")
	if len(authHeaders) == 0 {
		log.Println("Error: No authorization headers found")
		return "", status.Errorf(codes.Unauthenticated, "missing authorization header")
	}

	authHeader := authHeaders[0]
	log.Printf("Authorization header: %s", authHeader)

	// 解析授权头
	parts := strings.Fields(authHeader) // 使用Fields而不是Split处理多个空格
	if len(parts) < 2 || strings.ToLower(parts[0]) != "bearer" {
		log.Printf("Invalid authorization header format: %s", authHeader)
		return "", status.Errorf(codes.Unauthenticated, "invalid authorization header format")
	}

	// 获取token（处理多个空格情况）
	token := strings.Join(parts[1:], " ")

	claims := &utils.Claims{}
	_, err := jwt.ParseWithClaims(token, claims, func(token *jwt.Token) (interface{}, error) {
		return utils.JwtSecret, nil
	})

	if err != nil {
		log.Printf("Token validation failed: %v", err)
		return "", status.Errorf(codes.Unauthenticated, "invalid token: %v", err)
	}

	log.Printf("User authenticated: %s", claims.Username)
	return claims.Username, nil
}

// CreateLiveClass 创建直播课
func (s *LiveClassServiceServer) CreateLiveClass(ctx context.Context, req *pb.CreateLiveClassRequest) (*pb.CreateLiveClassResponse, error) {
	// 从认证信息中获取教师用户名
	teacherName, err := s.authenticateToken(ctx)
	if err != nil {
		log.Printf("Authentication failed in CreateLiveClass: %v", err)
		return nil, err
	}

	log.Printf("Creating live class for teacher: %s", teacherName)

	// 调用 LiveGo 服务器获取推流密钥
	livegoURL := "http://localhost:8090/control/get?room=" + req.RoomName
	livegoResp, err := http.Get(livegoURL)
	if err != nil {
		log.Printf("Failed to call LiveGo server: %v", err)
		return nil, err
	}
	defer livegoResp.Body.Close()

	var result map[string]interface{}
	err = json.NewDecoder(livegoResp.Body).Decode(&result)
	if err != nil {
		log.Printf("Failed to decode LiveGo response: %v", err)
		return nil, err
	}
	log.Printf("LiveGo response: %+v", result)

	streamKey, ok := result["data"].(string)
	if !ok {
		log.Printf("Failed to extract stream key from response: %+v", result)
		return nil, errors.New("failed to extract stream key from response")
	}

	// 拼接完整的推流地址
	streamURL := "rtmp://localhost:8090/live/" + streamKey

	// 初始化直播课 - 使用认证的用户名作为教师名
	liveClass := &LiveClass{
		TeacherName: teacherName,
		StreamURL:   streamURL,
		Messages:    make([]*pb.Message, 0),
		Questions:   make(map[string]*pb.Question),
		Answers:     make(map[string]map[string]int32),
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if s.streams == nil {
		s.streams = make(map[string]*LiveClass)
	}

	s.streams[req.ClassName] = liveClass

	log.Printf("Live class created: %s by %s", req.ClassName, teacherName)

	return &pb.CreateLiveClassResponse{
		ClassId:   req.ClassName,
		Status:    "success",
		StreamKey: streamKey,
	}, nil
}
func (s *LiveClassServiceServer) JoinLiveClass(ctx context.Context, req *pb.JoinLiveClassRequest) (*pb.JoinLiveClassResponse, error) {
	log.Printf("Received JoinLiveClass request: %+v", req)

	// 认证用户
	username, err := s.authenticateToken(ctx)
	if err != nil {
		log.Printf("Authentication failed: %v", err)
		return nil, err
	}

	log.Printf("User %s authenticated", username)

	s.mu.Lock()
	defer s.mu.Unlock()

	// 检查是否存在该直播课
	liveClass, ok := s.streams[req.ClassId]
	if !ok {
		log.Printf("Live class not found: %s", req.ClassId)
		return nil, errors.New("live class not found")
	}

	log.Printf("Returning stream URL for class %s: %s", req.ClassId, liveClass.StreamURL)

	return &pb.JoinLiveClassResponse{
		Status:    "success",
		StreamUrl: liveClass.StreamURL,
		Message:   fmt.Sprintf("%s 加入了直播课", username),
	}, nil
}

// SendMessage 发送消息
func (s *LiveClassServiceServer) SendMessage(ctx context.Context, req *pb.SendMessageRequest) (*pb.SendMessageResponse, error) {
	username, err := s.authenticateToken(ctx)
	if err != nil {
		return nil, err
	}

	classID := req.ClassId

	s.mu.Lock()
	liveClass, ok := s.streams[classID]
	s.mu.Unlock()
	if !ok {
		return nil, errors.New("live class not found")
	}

	// 创建新消息
	message := &pb.Message{
		SenderName:     username,
		MessageContent: req.MessageContent,
		Timestamp:      time.Now().Unix(),
	}

	// 将消息添加到列表
	liveClass.mu.Lock()
	liveClass.Messages = append(liveClass.Messages, message)
	liveClass.mu.Unlock()

	return &pb.SendMessageResponse{
		Status: "success",
	}, nil
}

func (s *LiveClassServiceServer) EndLiveClass(ctx context.Context, req *pb.EndLiveClassRequest) (*pb.EndLiveClassResponse, error) {
	classID := req.ClassId
	username := req.Username

	log.Printf("结束直播课请求: 教室ID=%s, 用户名=%s", classID, username)

	s.mu.Lock()
	defer s.mu.Unlock()

	// 检查是否存在该直播课
	liveClass, ok := s.streams[classID]
	if !ok {
		log.Printf("直播课不存在: %s", classID)
		return nil, errors.New("live class not found")
	}

	// 检查请求用户是否是直播间的发起人
	if liveClass.TeacherName != username {
		log.Printf("权限不足: 创建者=%s, 请求者=%s", liveClass.TeacherName, username)
		return nil, status.Errorf(codes.PermissionDenied, "only the class initiator can end the live class")
	}

	// 清理消息队列（如果有的话，现在改为切片，不需要关闭）
	// 清理消息列表
	liveClass.Messages = nil

	// 清理题目和答案数据
	liveClass.Questions = nil
	liveClass.Answers = nil

	// 从 streams 中移除直播课
	delete(s.streams, classID)

	log.Printf("直播课已结束并清理: %s", classID)

	return &pb.EndLiveClassResponse{
		Status: "success",
	}, nil
}

// PublishQuestion 发布题目
func (s *LiveClassServiceServer) PublishQuestion(ctx context.Context, req *pb.PublishQuestionRequest) (*pb.PublishQuestionResponse, error) {
	// 认证用户（确保是教师）
	username, err := s.authenticateToken(ctx)
	if err != nil {
		return nil, err
	}

	classID := req.ClassId

	s.mu.Lock()
	defer s.mu.Unlock()

	// 检查是否存在该直播课
	liveClass, ok := s.streams[classID]
	if !ok {
		return nil, errors.New("live class not found")
	}

	// 检查请求用户是否是直播间的发起人
	if liveClass.TeacherName != username {
		return nil, status.Errorf(codes.PermissionDenied, "only the class initiator can publish questions")
	}

	// 生成题目ID（使用更精确的纳秒级时间戳）
	questionID := time.Now().Format("20060102150405.999999999")

	// 保存题目
	liveClass.Questions[questionID] = &pb.Question{
		QuestionId:   questionID,
		QuestionText: req.Question,
	}

	// 初始化答案统计
	liveClass.Answers[questionID] = make(map[string]int32)

	log.Printf("题目已发布: ID=%s, 内容=%s", questionID, req.Question)

	return &pb.PublishQuestionResponse{
		Status:     "success",
		QuestionId: questionID, // 确保返回question_id
	}, nil
}
func (s *LiveClassServiceServer) GetMessages(ctx context.Context, req *pb.GetMessagesRequest) (*pb.GetMessagesResponse, error) {
	s.mu.Lock()
	liveClass, ok := s.streams[req.ClassId]
	s.mu.Unlock()

	if !ok {
		return nil, errors.New("live class not found")
	}

	var messages []*pb.Message

	liveClass.mu.Lock()
	for _, msg := range liveClass.Messages {
		if msg.Timestamp > req.LastTimestamp {
			// 复制消息对象以确保完整返回
			messageCopy := &pb.Message{
				SenderName:     msg.SenderName,
				MessageContent: msg.MessageContent,
				Timestamp:      msg.Timestamp,
			}
			messages = append(messages, messageCopy)
		}
	}
	liveClass.mu.Unlock()

	log.Printf("Returning %d messages for class %s", len(messages), req.ClassId)

	return &pb.GetMessagesResponse{
		Messages: messages,
	}, nil
}

// SubmitAnswer 提交答案
func (s *LiveClassServiceServer) SubmitAnswer(ctx context.Context, req *pb.SubmitAnswerRequest) (*pb.SubmitAnswerResponse, error) {
	classID := req.ClassId
	answer := req.Answer
	questionID := req.QuestionId

	s.mu.Lock()
	defer s.mu.Unlock()

	// 检查是否存在该直播课
	liveClass, ok := s.streams[classID]
	if !ok {
		return nil, errors.New("live class not found")
	}

	// 检查题目是否存在
	_, ok = liveClass.Questions[questionID]
	if !ok {
		return nil, errors.New("question not found")
	}

	// 更新答案统计
	liveClass.Answers[questionID][answer]++

	return &pb.SubmitAnswerResponse{
		Status: "success",
	}, nil
}

// 修改方法签名以匹配接口
func (s *LiveClassServiceServer) GetAnswerStatistics(
	req *pb.GetAnswerStatisticsRequest,
	stream pb.LiveClassService_GetAnswerStatisticsServer,
) error {
	// 从流中获取上下文
	ctx := stream.Context()

	classID := req.ClassId
	questionID := req.QuestionId

	log.Printf("获取答题统计: 教室=%s, 问题=%s", classID, questionID)

	s.mu.Lock()
	liveClass, ok := s.streams[classID]
	s.mu.Unlock()

	if !ok {
		return status.Errorf(codes.NotFound, "直播课不存在")
	}

	// 检查题目是否存在
	liveClass.mu.Lock()
	_, ok = liveClass.Questions[questionID]
	answerCounts := liveClass.Answers[questionID] // 获取当前统计的副本
	liveClass.mu.Unlock()

	if !ok {
		return status.Errorf(codes.NotFound, "题目不存在")
	}

	// 创建超时上下文
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	// 发送初始统计
	if err := stream.Send(&pb.AnswerStatistics{
		QuestionId:   questionID,
		AnswerCounts: answerCounts,
	}); err != nil {
		return err
	}

	// 定期更新
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Println("上下文结束:", ctx.Err())
			return nil
		case <-ticker.C:
			// 获取最新统计
			liveClass.mu.Lock()
			updatedCounts := liveClass.Answers[questionID]
			liveClass.mu.Unlock()

			// 检查是否有变化
			if !reflect.DeepEqual(answerCounts, updatedCounts) {
				// 发送更新
				if err := stream.Send(&pb.AnswerStatistics{
					QuestionId:   questionID,
					AnswerCounts: updatedCounts,
				}); err != nil {
					log.Printf("发送失败: %v", err)
					return err
				}
				answerCounts = updatedCounts
			}
		}
	}
}
