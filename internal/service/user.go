package service

import (
	"context"
	"time"

	"github.com/zhang/microservice/internal/database"
	"github.com/zhang/microservice/internal/logger"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// User 用户模型
type User struct {
	ID        int64     `gorm:"primaryKey" json:"id"`
	Name      string    `gorm:"type:varchar(100);not null" json:"name"`
	Email     string    `gorm:"type:varchar(100);uniqueIndex;not null" json:"email"`
	Phone     string    `gorm:"type:varchar(20)" json:"phone"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// TableName 指定表名
func (User) TableName() string {
	return "users"
}

// UserService 用户服务
type UserService struct{}

// NewUserService 创建用户服务实例
// 返回:
//
//	*UserService: 用户服务实例
func NewUserService() *UserService {
	return &UserService{}
}

// GetUser 获取用户
// 参数:
//
//	ctx: 上下文
//	id: 用户 ID
//
// 返回:
//
//	*User: 用户信息
//	error: 错误信息
func (s *UserService) GetUser(ctx context.Context, id int64) (*User, error) {
	var user User

	if err := database.DB.WithContext(ctx).First(&user, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		logger.Error("查询用户失败", zap.Int64("id", id), zap.Error(err))
		return nil, err
	}

	return &user, nil
}

// CreateUser 创建用户
// 参数:
//
//	ctx: 上下文
//	user: 用户信息
//
// 返回:
//
//	*User: 创建的用户
//	error: 错误信息
func (s *UserService) CreateUser(ctx context.Context, user *User) (*User, error) {
	if err := database.DB.WithContext(ctx).Create(user).Error; err != nil {
		logger.Error("创建用户失败", zap.Error(err))
		return nil, err
	}

	logger.Info("用户创建成功", zap.Int64("id", user.ID), zap.String("name", user.Name))
	return user, nil
}

// UpdateUser 更新用户
// 参数:
//
//	ctx: 上下文
//	user: 用户信息
//
// 返回:
//
//	*User: 更新后的用户
//	error: 错误信息
func (s *UserService) UpdateUser(ctx context.Context, user *User) (*User, error) {
	if err := database.DB.WithContext(ctx).Save(user).Error; err != nil {
		logger.Error("更新用户失败", zap.Int64("id", user.ID), zap.Error(err))
		return nil, err
	}

	logger.Info("用户更新成功", zap.Int64("id", user.ID))
	return user, nil
}

// DeleteUser 删除用户
// 参数:
//
//	ctx: 上下文
//	id: 用户 ID
//
// 返回:
//
//	error: 错误信息
func (s *UserService) DeleteUser(ctx context.Context, id int64) error {
	if err := database.DB.WithContext(ctx).Delete(&User{}, id).Error; err != nil {
		logger.Error("删除用户失败", zap.Int64("id", id), zap.Error(err))
		return err
	}

	logger.Info("用户删除成功", zap.Int64("id", id))
	return nil
}

// ListUsers 获取用户列表
// 参数:
//
//	ctx: 上下文
//	offset: 偏移量
//	limit: 限制数量
//
// 返回:
//
//	[]*User: 用户列表
//	int64: 总数
//	error: 错误信息
func (s *UserService) ListUsers(ctx context.Context, offset, limit int) ([]*User, int64, error) {
	var users []*User
	var total int64

	db := database.DB.WithContext(ctx).Model(&User{})

	// 获取总数
	if err := db.Count(&total).Error; err != nil {
		logger.Error("查询用户总数失败", zap.Error(err))
		return nil, 0, err
	}

	// 获取列表
	if err := db.Offset(offset).Limit(limit).Find(&users).Error; err != nil {
		logger.Error("查询用户列表失败", zap.Error(err))
		return nil, 0, err
	}

	return users, total, nil
}
