package liveservice

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"github.com/dgrijalva/jwt-go"
	"net/http"
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
	TeacherName  string                      // 直播间发起人的用户名
	StreamURL    string                      // 推流地址
	MessageQueue chan *pb.Message            // 消息队列
	Subscribers  []chan *pb.Message          // 订阅者列表
	Questions    map[string]*pb.Question     // 题目列表
	Answers      map[string]map[string]int32 // 答案统计
}

// NewLiveClassServiceServer 初始化服务
func NewLiveClassServiceServer() *LiveClassServiceServer {
	return &LiveClassServiceServer{
		streams: make(map[string]*LiveClass),
	}
}

// authenticateToken 验证 JWT Token 并返回用户名
func (s *LiveClassServiceServer) authenticateToken(ctx context.Context) (string, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", status.Errorf(codes.Unauthenticated, "missing metadata")
	}

	authHeader, ok := md["authorization"]
	if !ok || len(authHeader) == 0 {
		return "", status.Errorf(codes.Unauthenticated, "missing authorization header")
	}

	token := authHeader[0][len("Bearer "):]
	claims := &utils.Claims{}

	// 解析 JWT Token
	_, err := jwt.ParseWithClaims(token, claims, func(token *jwt.Token) (interface{}, error) {
		return utils.JwtSecret, nil
	})
	if err != nil {
		return "", status.Errorf(codes.Unauthenticated, "invalid token: %v", err)
	}

	return claims.Username, nil
}

// CreateLiveClass 创建直播课
func (s *LiveClassServiceServer) CreateLiveClass(ctx context.Context, req *pb.CreateLiveClassRequest) (*pb.CreateLiveClassResponse, error) {
	// 检查是否存在该直播课
	classID := req.ClassName
	roomName := req.RoomName

	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.streams[classID]; ok {
		return nil, errors.New("live class already exists")
	}

	// 调用 LiveGo 服务器获取推流密钥
	livegoURL := "http://localhost:8080/api/rtmp/publish" // 假设 LiveGo 服务器的 API 地址
	livegoResp, err := http.Post(livegoURL, "application/json", bytes.NewBuffer([]byte(`{"roomname": "`+roomName+`"}`)))
	if err != nil {
		return nil, err
	}
	defer livegoResp.Body.Close()

	var livegoData map[string]interface{}
	if err := json.NewDecoder(livegoResp.Body).Decode(&livegoData); err != nil {
		return nil, err
	}

	streamKey := livegoData["stream_key"].(string) // 获取推流密钥

	// 初始化直播课
	liveClass := &LiveClass{
		TeacherName:  req.TeacherName, // 记录发起人的用户名
		StreamURL:    streamKey,
		MessageQueue: make(chan *pb.Message, 100),
		Subscribers:  []chan *pb.Message{},
		Questions:    make(map[string]*pb.Question),
		Answers:      make(map[string]map[string]int32),
	}

	s.streams[classID] = liveClass

	return &pb.CreateLiveClassResponse{
		ClassId:   classID,
		Status:    "success",
		StreamKey: streamKey, // 返回推流密钥
	}, nil
}

// JoinLiveClass 加入直播课（双向流）
func (s *LiveClassServiceServer) JoinLiveClass(stream pb.LiveClassService_JoinLiveClassServer) error {
	// 从流中读取请求
	req, err := stream.Recv()
	if err != nil {
		return err
	}

	classID := req.ClassId

	s.mu.Lock()
	liveClass, ok := s.streams[classID]
	s.mu.Unlock()
	if !ok {
		return errors.New("live class not found")
	}

	// 创建一个订阅者通道
	subscriber := make(chan *pb.Message, 100)
	s.mu.Lock()
	liveClass.Subscribers = append(liveClass.Subscribers, subscriber)
	s.mu.Unlock()

	defer func() {
		s.mu.Lock()
		defer s.mu.Unlock()
		for i, sub := range liveClass.Subscribers {
			if sub == subscriber {
				liveClass.Subscribers = append(liveClass.Subscribers[:i], liveClass.Subscribers[i+1:]...)
				break
			}
		}
		close(subscriber)
	}()

	// 发送初始响应
	initialResponse := &pb.JoinLiveClassResponse{
		Status:    "success",
		StreamUrl: liveClass.StreamURL,
	}
	if err := stream.Send(initialResponse); err != nil {
		return err
	}

	// 从消息队列中读取消息并发送给订阅者
	for {
		select {
		case msg, ok := <-liveClass.MessageQueue:
			if !ok {
				return nil
			}
			response := &pb.JoinLiveClassResponse{
				Message: msg,
			}
			if err := stream.Send(response); err != nil {
				return err
			}
		case <-stream.Context().Done():
			return stream.Context().Err()
		}
	}
}

// SendMessage 发送消息
func (s *LiveClassServiceServer) SendMessage(ctx context.Context, req *pb.SendMessageRequest) (*pb.SendMessageResponse, error) {
	username, err := s.authenticateToken(ctx)
	if err != nil {
		return nil, err
	}

	classID := req.ClassId
	message := req.Message

	s.mu.Lock()
	defer s.mu.Unlock()

	// 检查是否存在该直播课
	liveClass, ok := s.streams[classID]
	if !ok {
		return nil, errors.New("live class not found")
	}

	// 设置消息的发送者为用户名
	message.SenderName = username

	// 将消息发送到消息队列
	select {
	case liveClass.MessageQueue <- message:
		return &pb.SendMessageResponse{
			Status: "success",
		}, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

// EndLiveClass 结束直播课
func (s *LiveClassServiceServer) EndLiveClass(ctx context.Context, req *pb.EndLiveClassRequest) (*pb.EndLiveClassResponse, error) {
	classID := req.ClassId
	username := req.Username

	s.mu.Lock()
	defer s.mu.Unlock()

	// 检查是否存在该直播课
	liveClass, ok := s.streams[classID]
	if !ok {
		return nil, errors.New("live class not found")
	}

	// 检查请求用户是否是直播间的发起人
	if liveClass.TeacherName != username {
		return nil, status.Errorf(codes.PermissionDenied, "only the class initiator can end the live class")
	}

	// 关闭消息队列
	close(liveClass.MessageQueue)

	// 清理订阅者
	for _, sub := range liveClass.Subscribers {
		close(sub)
	}

	// 从 streams 中移除直播课
	delete(s.streams, classID)

	return &pb.EndLiveClassResponse{
		Status: "success",
	}, nil
}

// PublishQuestion 发布题目
func (s *LiveClassServiceServer) PublishQuestion(ctx context.Context, req *pb.PublishQuestionRequest) (*pb.PublishQuestionResponse, error) {
	classID := req.ClassId
	question := req.Question

	s.mu.Lock()
	defer s.mu.Unlock()

	// 检查是否存在该直播课
	liveClass, ok := s.streams[classID]
	if !ok {
		return nil, errors.New("live class not found")
	}

	// 生成题目ID
	questionID := time.Now().Format("20060102150405")

	// 保存题目
	liveClass.Questions[questionID] = &pb.Question{
		QuestionId:   questionID,
		QuestionText: question,
	}

	// 初始化答案统计
	liveClass.Answers[questionID] = make(map[string]int32)

	return &pb.PublishQuestionResponse{
		Status:     "success",
		QuestionId: questionID,
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

// GetAnswerStatistics 获取答题结果统计（流式接口）
func (s *LiveClassServiceServer) GetAnswerStatistics(req *pb.GetAnswerStatisticsRequest, stream pb.LiveClassService_GetAnswerStatisticsServer) error {
	classID := req.ClassId
	questionID := req.QuestionId

	s.mu.Lock()
	defer s.mu.Unlock()

	// 检查是否存在该直播课
	liveClass, ok := s.streams[classID]
	if !ok {
		return errors.New("live class not found")
	}

	// 检查题目是否存在
	_, ok = liveClass.Questions[questionID]
	if !ok {
		return errors.New("question not found")
	}

	// 获取答案统计
	answerCounts := liveClass.Answers[questionID]

	// 发送统计结果
	stat := &pb.AnswerStatistics{
		QuestionId:   questionID,
		AnswerCounts: answerCounts,
	}
	if err := stream.Send(stat); err != nil {
		return err
	}

	// 检测客户端是否关闭连接
	select {
	case <-stream.Context().Done():
		return stream.Context().Err()
	}
}
