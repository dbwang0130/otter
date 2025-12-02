package user

import (
	"errors"
	"fmt"

	"github.com/galilio/otter/internal/common/utils"
)

var (
	ErrUserNotFound      = errors.New("用户不存在")
	ErrUserAlreadyExists = errors.New("用户已存在")
	ErrInvalidInput      = errors.New("输入参数无效")
	ErrInvalidPassword   = errors.New("密码错误")
)

type Service interface {
	CreateUser(req *CreateUserRequest) (*User, error)
	GetUserByID(id uint) (*User, error)
	UpdateUser(id uint, req *UpdateUserRequest) (*User, error)
	DeleteUser(id uint) error
	ListUsers(page, pageSize int) (*UserListResponse, error)
	Login(req *LoginRequest) (*User, error)
}

type CreateUserRequest struct {
	Username  string `json:"username" binding:"required,min=3,max=100"`
	Email     string `json:"email" binding:"required,email"`
	Password  string `json:"password" binding:"required,min=6"`
	FirstName string `json:"first_name" binding:"max=100"`
	LastName  string `json:"last_name" binding:"max=100"`
	Phone     string `json:"phone" binding:"max=20"`
}

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type UpdateUserRequest struct {
	Email     *string `json:"email" binding:"omitempty,email"`
	FirstName *string `json:"first_name" binding:"omitempty,max=100"`
	LastName  *string `json:"last_name" binding:"omitempty,max=100"`
	Phone     *string `json:"phone" binding:"omitempty,max=20"`
	Status    *string `json:"status" binding:"omitempty,oneof=active inactive"`
}

type UserListResponse struct {
	Users      []*User `json:"users"`
	Total      int64   `json:"total"`
	Page       int     `json:"page"`
	PageSize   int     `json:"page_size"`
	TotalPages int     `json:"total_pages"`
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) CreateUser(req *CreateUserRequest) (*User, error) {
	// 检查用户名是否已存在
	if _, err := s.repo.GetByUsername(req.Username); err == nil {
		return nil, fmt.Errorf("%w: 用户名", ErrUserAlreadyExists)
	}

	// 检查邮箱是否已存在
	if _, err := s.repo.GetByEmail(req.Email); err == nil {
		return nil, fmt.Errorf("%w: 邮箱", ErrUserAlreadyExists)
	}

	// 加密密码
	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		return nil, fmt.Errorf("密码加密失败: %w", err)
	}

	user := &User{
		Username:  req.Username,
		Email:     req.Email,
		Password:  hashedPassword,
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Phone:     req.Phone,
		Status:    "active",
		IsAdmin:   false,
	}

	if err := s.repo.Create(user); err != nil {
		return nil, fmt.Errorf("创建用户失败: %w", err)
	}

	return user, nil
}

func (s *service) GetUserByID(id uint) (*User, error) {
	user, err := s.repo.GetByID(id)
	if err != nil {
		return nil, ErrUserNotFound
	}
	return user, nil
}

func (s *service) UpdateUser(id uint, req *UpdateUserRequest) (*User, error) {
	user, err := s.repo.GetByID(id)
	if err != nil {
		return nil, ErrUserNotFound
	}

	if req.Email != nil {
		// 检查新邮箱是否已被其他用户使用
		existingUser, err := s.repo.GetByEmail(*req.Email)
		if err == nil && existingUser.ID != id {
			return nil, fmt.Errorf("%w: 邮箱", ErrUserAlreadyExists)
		}
		user.Email = *req.Email
	}

	if req.FirstName != nil {
		user.FirstName = *req.FirstName
	}
	if req.LastName != nil {
		user.LastName = *req.LastName
	}
	if req.Phone != nil {
		user.Phone = *req.Phone
	}
	if req.Status != nil {
		user.Status = *req.Status
	}

	if err := s.repo.Update(user); err != nil {
		return nil, fmt.Errorf("更新用户失败: %w", err)
	}

	return user, nil
}

func (s *service) DeleteUser(id uint) error {
	_, err := s.repo.GetByID(id)
	if err != nil {
		return ErrUserNotFound
	}

	if err := s.repo.Delete(id); err != nil {
		return fmt.Errorf("删除用户失败: %w", err)
	}

	return nil
}

func (s *service) ListUsers(page, pageSize int) (*UserListResponse, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	offset := (page - 1) * pageSize
	users, total, err := s.repo.List(offset, pageSize)
	if err != nil {
		return nil, fmt.Errorf("获取用户列表失败: %w", err)
	}

	totalPages := int((total + int64(pageSize) - 1) / int64(pageSize))

	return &UserListResponse{
		Users:      users,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}, nil
}

func (s *service) Login(req *LoginRequest) (*User, error) {
	// 通过用户名或邮箱查找用户
	user, err := s.repo.GetByUsername(req.Username)
	if err != nil {
		// 尝试通过邮箱查找
		user, err = s.repo.GetByEmail(req.Username)
		if err != nil {
			return nil, ErrUserNotFound
		}
	}

	// 验证密码
	if !utils.CheckPassword(req.Password, user.Password) {
		return nil, ErrInvalidPassword
	}

	// 检查用户状态
	if user.Status != "active" {
		return nil, errors.New("用户账户已被禁用")
	}

	return user, nil
}
