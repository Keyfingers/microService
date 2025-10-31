package service

import (
	"context"
	"testing"
	"time"
)

// TestUserModel 测试用户模型
func TestUserModel(t *testing.T) {
	user := User{
		ID:        1,
		Name:      "测试用户",
		Email:     "test@example.com",
		Phone:     "13800138000",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if user.TableName() != "users" {
		t.Errorf("期望表名为 users, 实际为 %s", user.TableName())
	}
}

// TestNewUserService 测试创建用户服务
func TestNewUserService(t *testing.T) {
	service := NewUserService()
	if service == nil {
		t.Error("用户服务创建失败")
	}
}

// 注意：以下测试需要数据库连接，实际测试时需要先初始化数据库
// 这里仅作为示例展示如何编写测试

// TestUserService_CRUD 测试用户 CRUD 操作（需要数据库）
func TestUserService_CRUD(t *testing.T) {
	t.Skip("跳过需要数据库的测试")

	ctx := context.Background()
	service := NewUserService()

	// 测试创建用户
	user := &User{
		Name:  "测试用户",
		Email: "test@example.com",
		Phone: "13800138000",
	}

	// createdUser, err := service.CreateUser(ctx, user)
	// if err != nil {
	// 	t.Fatalf("创建用户失败: %v", err)
	// }

	// 测试获取用户
	// gotUser, err := service.GetUser(ctx, createdUser.ID)
	// if err != nil {
	// 	t.Fatalf("获取用户失败: %v", err)
	// }
	// if gotUser.Email != user.Email {
	// 	t.Errorf("期望邮箱为 %s, 实际为 %s", user.Email, gotUser.Email)
	// }

	// 测试更新用户
	// gotUser.Name = "更新后的用户"
	// updatedUser, err := service.UpdateUser(ctx, gotUser)
	// if err != nil {
	// 	t.Fatalf("更新用户失败: %v", err)
	// }
	// if updatedUser.Name != "更新后的用户" {
	// 	t.Errorf("期望名称为 '更新后的用户', 实际为 %s", updatedUser.Name)
	// }

	// 测试删除用户
	// err = service.DeleteUser(ctx, createdUser.ID)
	// if err != nil {
	// 	t.Fatalf("删除用户失败: %v", err)
	// }

	_ = ctx
	_ = service
}
