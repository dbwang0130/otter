package database

import (
	"crypto/rand"
	"fmt"
	"log/slog"
	"math/big"

	"github.com/galilio/otter/internal/auth"
	"github.com/galilio/otter/internal/calendar"
	"github.com/galilio/otter/internal/common/config"
	"github.com/galilio/otter/internal/common/utils"
	"github.com/galilio/otter/internal/user"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func NewDB(cfg *config.DatabaseConfig) (*gorm.DB, error) {
	if cfg.URL == "" {
		return nil, fmt.Errorf("数据库 URL 不能为空")
	}

	db, err := gorm.Open(postgres.Open(cfg.URL), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return nil, fmt.Errorf("连接数据库失败: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("获取数据库实例失败: %w", err)
	}

	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(cfg.ConnMaxLifetime)

	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("数据库连接测试失败: %w", err)
	}

	return db, nil
}

func Migrate(db *gorm.DB) error {
	// 执行自动迁移
	if err := db.AutoMigrate(
		&user.User{},
		&auth.RefreshToken{},
		&calendar.CalendarItem{},
		&calendar.Valarm{},
	); err != nil {
		return fmt.Errorf("自动迁移失败: %w", err)
	}

	// 创建部分唯一索引（仅对未删除的记录建立唯一约束）
	// 这样软删除后的用户名和邮箱可以被重用
	if err := createPartialUniqueIndexes(db); err != nil {
		return fmt.Errorf("创建部分唯一索引失败: %w", err)
	}

	return nil
}

// createPartialUniqueIndexes 创建部分唯一索引
// 只对 deleted_at IS NULL 的记录建立唯一约束，允许软删除后重用用户名和邮箱
func createPartialUniqueIndexes(db *gorm.DB) error {
	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("获取数据库连接失败: %w", err)
	}

	// 删除可能存在的旧索引（如果存在）
	// GORM 的 AutoMigrate 可能会创建名为 idx_users_username 和 idx_users_email 的索引
	dropIndexes := []string{
		"DROP INDEX IF EXISTS idx_users_username",
		"DROP INDEX IF EXISTS idx_users_email",
		"DROP INDEX IF EXISTS idx_users_username_unique",
		"DROP INDEX IF EXISTS idx_users_email_unique",
	}

	for _, sql := range dropIndexes {
		if _, err := sqlDB.Exec(sql); err != nil {
			slog.Debug("警告: 删除索引失败（可能不存在）", "error", err)
		}
	}

	// 创建部分唯一索引
	indexes := []string{
		`CREATE UNIQUE INDEX IF NOT EXISTS idx_users_username_unique ON users(username) WHERE deleted_at IS NULL`,
		`CREATE UNIQUE INDEX IF NOT EXISTS idx_users_email_unique ON users(email) WHERE deleted_at IS NULL`,
	}

	for _, sql := range indexes {
		if _, err := sqlDB.Exec(sql); err != nil {
			return fmt.Errorf("创建索引失败: %w, SQL: %s", err, sql)
		}
	}

	slog.Debug("部分唯一索引创建成功（允许软删除后重用用户名和邮箱）")
	return nil
}

// generateRandomPassword 生成随机密码
// 长度16个字符，包含大小写字母、数字
func generateRandomPassword(length int) (string, error) {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	password := make([]byte, length)

	for i := range password {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			return "", fmt.Errorf("生成随机数失败: %w", err)
		}
		password[i] = charset[num.Int64()]
	}

	return string(password), nil
}

// SeedAdminUser 创建默认管理员用户（仅用于开发环境）
// 如果已存在管理员用户，则跳过创建
// 密码会随机生成并在日志中输出
func SeedAdminUser(db *gorm.DB) error {
	// 检查是否已存在管理员用户
	var count int64
	if err := db.Model(&user.User{}).Where("is_admin = ?", true).Count(&count).Error; err != nil {
		return fmt.Errorf("检查管理员用户失败: %w", err)
	}

	if count > 0 {
		slog.Debug("管理员用户已存在，跳过初始化")
		return nil
	}

	// 默认管理员用户名和邮箱
	adminUsername := "admin"
	adminEmail := "admin@example.com"

	// 生成随机密码
	adminPassword, err := generateRandomPassword(16)
	if err != nil {
		return fmt.Errorf("生成随机密码失败: %w", err)
	}

	// 加密密码
	hashedPassword, err := utils.HashPassword(adminPassword)
	if err != nil {
		return fmt.Errorf("密码加密失败: %w", err)
	}

	// 创建管理员用户
	adminUser := &user.User{
		Username:  adminUsername,
		Email:     adminEmail,
		Password:  hashedPassword,
		FirstName: "Admin",
		LastName:  "User",
		Status:    "active",
		IsAdmin:   true,
	}

	if err := db.Create(adminUser).Error; err != nil {
		return fmt.Errorf("创建默认管理员用户失败: %w", err)
	}

	slog.Debug("========================================")
	slog.Debug("默认管理员用户创建成功！")
	slog.Debug("管理员信息", "username", adminUsername, "password", adminPassword, "email", adminEmail)
	slog.Debug("========================================")
	slog.Debug("⚠️  请妥善保管密码，此密码仅显示一次！")

	return nil
}
