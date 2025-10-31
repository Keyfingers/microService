package main

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/zhang/microservice/internal/cache"
	"github.com/zhang/microservice/internal/config"
	"github.com/zhang/microservice/internal/database"
	"github.com/zhang/microservice/internal/logger"
	"github.com/zhang/microservice/internal/service"
	pb "github.com/zhang/microservice/proto"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

// server gRPC 服务器
type server struct {
	pb.UnimplementedUserServiceServer
	userService *service.UserService
}

// GetUser 获取用户
func (s *server) GetUser(ctx context.Context, req *pb.GetUserRequest) (*pb.GetUserResponse, error) {
	user, err := s.userService.GetUser(ctx, req.Id)
	if err != nil {
		return nil, err
	}

	if user == nil {
		return &pb.GetUserResponse{}, nil
	}

	return &pb.GetUserResponse{
		User: &pb.User{
			Id:        user.ID,
			Name:      user.Name,
			Email:     user.Email,
			Phone:     user.Phone,
			CreatedAt: user.CreatedAt.Format("2006-01-02 15:04:05"),
			UpdatedAt: user.UpdatedAt.Format("2006-01-02 15:04:05"),
		},
	}, nil
}

// CreateUser 创建用户
func (s *server) CreateUser(ctx context.Context, req *pb.CreateUserRequest) (*pb.CreateUserResponse, error) {
	user := &service.User{
		Name:  req.Name,
		Email: req.Email,
		Phone: req.Phone,
	}

	user, err := s.userService.CreateUser(ctx, user)
	if err != nil {
		return nil, err
	}

	return &pb.CreateUserResponse{
		User: &pb.User{
			Id:        user.ID,
			Name:      user.Name,
			Email:     user.Email,
			Phone:     user.Phone,
			CreatedAt: user.CreatedAt.Format("2006-01-02 15:04:05"),
			UpdatedAt: user.UpdatedAt.Format("2006-01-02 15:04:05"),
		},
	}, nil
}

// UpdateUser 更新用户
func (s *server) UpdateUser(ctx context.Context, req *pb.UpdateUserRequest) (*pb.UpdateUserResponse, error) {
	user := &service.User{
		ID:    req.Id,
		Name:  req.Name,
		Email: req.Email,
		Phone: req.Phone,
	}

	user, err := s.userService.UpdateUser(ctx, user)
	if err != nil {
		return nil, err
	}

	return &pb.UpdateUserResponse{
		User: &pb.User{
			Id:        user.ID,
			Name:      user.Name,
			Email:     user.Email,
			Phone:     user.Phone,
			CreatedAt: user.CreatedAt.Format("2006-01-02 15:04:05"),
			UpdatedAt: user.UpdatedAt.Format("2006-01-02 15:04:05"),
		},
	}, nil
}

// DeleteUser 删除用户
func (s *server) DeleteUser(ctx context.Context, req *pb.DeleteUserRequest) (*pb.DeleteUserResponse, error) {
	err := s.userService.DeleteUser(ctx, req.Id)
	if err != nil {
		return &pb.DeleteUserResponse{Success: false}, err
	}

	return &pb.DeleteUserResponse{Success: true}, nil
}

func main() {
	// 加载配置
	if err := config.Load("config/config.yaml"); err != nil {
		fmt.Printf("加载配置失败: %v\n", err)
		os.Exit(1)
	}

	// 初始化日志
	if err := logger.Init(config.GlobalConfig.Logger); err != nil {
		fmt.Printf("初始化日志失败: %v\n", err)
		os.Exit(1)
	}
	defer logger.Sync()

	logger.Info("gRPC 服务启动中...")

	// 初始化数据库
	if err := database.Init(config.GlobalConfig.Database); err != nil {
		logger.Fatal("初始化数据库失败", zap.Error(err))
	}
	defer database.Close()

	// 初始化 Redis
	if err := cache.Init(config.GlobalConfig.Redis); err != nil {
		logger.Fatal("初始化 Redis 失败", zap.Error(err))
	}
	defer cache.Close()

	// 自动迁移数据库表
	if err := database.DB.AutoMigrate(&service.User{}); err != nil {
		logger.Fatal("数据库迁移失败", zap.Error(err))
	}

	// 创建监听器
	addr := fmt.Sprintf(":%d", config.GlobalConfig.Server.GRPCPort)
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		logger.Fatal("创建监听器失败", zap.Error(err))
	}

	// 创建 gRPC 服务器
	s := grpc.NewServer()
	pb.RegisterUserServiceServer(s, &server{
		userService: service.NewUserService(),
	})

	// 启动服务器
	go func() {
		logger.Info("gRPC 服务启动成功",
			zap.String("地址", addr),
		)
		if err := s.Serve(lis); err != nil {
			logger.Fatal("启动 gRPC 服务器失败", zap.Error(err))
		}
	}()

	// 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("正在关闭 gRPC 服务器...")
	s.GracefulStop()
	logger.Info("gRPC 服务器已关闭")
}
