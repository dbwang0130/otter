package user

import (
	"database/sql"
	"fmt"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// setupTestDB 创建测试用的数据库连接（使用sqlmock）
func setupTestDB(t *testing.T) (*gorm.DB, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("创建sqlmock失败: %v", err)
	}

	gormDB, err := gorm.Open(postgres.New(postgres.Config{
		Conn: db,
	}), &gorm.Config{})
	if err != nil {
		t.Fatalf("创建GORM连接失败: %v", err)
	}

	return gormDB, mock
}

// TestRepository_Create 测试创建用户
func TestRepository_Create(t *testing.T) {
	db, mock := setupTestDB(t)
	repo := NewRepository(db)

	user := &User{
		Username:  "testuser",
		Email:     "test@example.com",
		Password:  "hashedpassword",
		FirstName: "Test",
		LastName:  "User",
		Status:    "active",
		IsAdmin:   false,
	}

	// 设置期望：INSERT 语句
	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO "users"`).
		WithArgs(
			sqlmock.AnyArg(), // CreatedAt
			sqlmock.AnyArg(), // UpdatedAt
			sqlmock.AnyArg(), // DeletedAt
			user.Username,
			user.Email,
			user.Password,
			user.FirstName,
			user.LastName,
			user.Phone,
			user.Status,
			user.IsAdmin,
		).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	mock.ExpectCommit()

	err := repo.Create(user)

	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// TestRepository_GetByID_Success 测试根据ID获取用户成功
func TestRepository_GetByID_Success(t *testing.T) {
	db, mock := setupTestDB(t)
	repo := NewRepository(db)

	userID := uint(1)
	expectedUser := &User{
		ID:        userID,
		Username:  "testuser",
		Email:     "test@example.com",
		FirstName: "Test",
		LastName:  "User",
		Status:    "active",
		IsAdmin:   false,
	}

	// 设置期望：SELECT 语句
	rows := sqlmock.NewRows([]string{"id", "created_at", "updated_at", "deleted_at", "username", "email", "password", "first_name", "last_name", "phone", "status", "is_admin"}).
		AddRow(
			expectedUser.ID,
			time.Now(),
			time.Now(),
			nil,
			expectedUser.Username,
			expectedUser.Email,
			"hashedpassword",
			expectedUser.FirstName,
			expectedUser.LastName,
			"",
			expectedUser.Status,
			expectedUser.IsAdmin,
		)

	mock.ExpectQuery(`SELECT \* FROM "users"`).
		WithArgs(userID, 1). // id, LIMIT
		WillReturnRows(rows)

	user, err := repo.GetByID(userID)

	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, expectedUser.ID, user.ID)
	assert.Equal(t, expectedUser.Username, user.Username)
	assert.Equal(t, expectedUser.Email, user.Email)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// TestRepository_GetByID_NotFound 测试用户不存在
func TestRepository_GetByID_NotFound(t *testing.T) {
	db, mock := setupTestDB(t)
	repo := NewRepository(db)

	userID := uint(999)

	// 设置期望：返回 sql.ErrNoRows
	mock.ExpectQuery(`SELECT \* FROM "users"`).
		WithArgs(userID, 1).
		WillReturnError(sql.ErrNoRows)

	user, err := repo.GetByID(userID)

	assert.Error(t, err)
	assert.Nil(t, user)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// TestRepository_GetByEmail_Success 测试根据邮箱获取用户成功
func TestRepository_GetByEmail_Success(t *testing.T) {
	db, mock := setupTestDB(t)
	repo := NewRepository(db)

	email := "test@example.com"
	expectedUser := &User{
		ID:       1,
		Username: "testuser",
		Email:    email,
	}

	rows := sqlmock.NewRows([]string{"id", "created_at", "updated_at", "deleted_at", "username", "email", "password", "first_name", "last_name", "phone", "status", "is_admin"}).
		AddRow(
			expectedUser.ID,
			time.Now(),
			time.Now(),
			nil,
			expectedUser.Username,
			expectedUser.Email,
			"hashedpassword",
			"",
			"",
			"",
			"active",
			false,
		)

	mock.ExpectQuery(`SELECT \* FROM "users"`).
		WithArgs(email, 1).
		WillReturnRows(rows)

	user, err := repo.GetByEmail(email)

	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, email, user.Email)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// TestRepository_GetByEmail_NotFound 测试邮箱不存在
func TestRepository_GetByEmail_NotFound(t *testing.T) {
	db, mock := setupTestDB(t)
	repo := NewRepository(db)

	email := "notfound@example.com"

	mock.ExpectQuery(`SELECT \* FROM "users"`).
		WithArgs(email, 1).
		WillReturnError(sql.ErrNoRows)

	user, err := repo.GetByEmail(email)

	assert.Error(t, err)
	assert.Nil(t, user)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// TestRepository_GetByUsername_Success 测试根据用户名获取用户成功
func TestRepository_GetByUsername_Success(t *testing.T) {
	db, mock := setupTestDB(t)
	repo := NewRepository(db)

	username := "testuser"
	expectedUser := &User{
		ID:       1,
		Username: username,
		Email:    "test@example.com",
	}

	rows := sqlmock.NewRows([]string{"id", "created_at", "updated_at", "deleted_at", "username", "email", "password", "first_name", "last_name", "phone", "status", "is_admin"}).
		AddRow(
			expectedUser.ID,
			time.Now(),
			time.Now(),
			nil,
			expectedUser.Username,
			expectedUser.Email,
			"hashedpassword",
			"",
			"",
			"",
			"active",
			false,
		)

	mock.ExpectQuery(`SELECT \* FROM "users"`).
		WithArgs(username, 1).
		WillReturnRows(rows)

	user, err := repo.GetByUsername(username)

	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, username, user.Username)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// TestRepository_GetByUsername_NotFound 测试用户名不存在
func TestRepository_GetByUsername_NotFound(t *testing.T) {
	db, mock := setupTestDB(t)
	repo := NewRepository(db)

	username := "nonexistent"

	mock.ExpectQuery(`SELECT \* FROM "users"`).
		WithArgs(username, 1).
		WillReturnError(sql.ErrNoRows)

	user, err := repo.GetByUsername(username)

	assert.Error(t, err)
	assert.Nil(t, user)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// TestRepository_Update 测试更新用户
func TestRepository_Update(t *testing.T) {
	db, mock := setupTestDB(t)
	repo := NewRepository(db)

	user := &User{
		ID:        1,
		Username:  "testuser",
		Email:     "updated@example.com",
		FirstName: "Updated",
		LastName:  "Name",
	}

	// 设置期望：GORM的Save会自动开启事务，并更新所有字段
	// GORM的Save会更新所有字段，包括created_at, updated_at, deleted_at
	// 还会自动添加软删除条件 WHERE "users"."deleted_at" IS NULL
	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE "users" SET`).
		WithArgs(
			sqlmock.AnyArg(), // CreatedAt
			sqlmock.AnyArg(), // UpdatedAt
			sqlmock.AnyArg(), // DeletedAt (NULL)
			user.Username,
			user.Email,
			user.Password,
			user.FirstName,
			user.LastName,
			user.Phone,
			user.Status,
			user.IsAdmin,
			user.ID, // WHERE条件中的ID
		).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	err := repo.Update(user)

	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// TestRepository_Delete 测试删除用户
func TestRepository_Delete(t *testing.T) {
	db, mock := setupTestDB(t)
	repo := NewRepository(db)

	userID := uint(1)

	// 设置期望：GORM的Delete是软删除（UPDATE deleted_at），会自动开启事务
	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE "users" SET`).
		WithArgs(sqlmock.AnyArg(), userID). // deleted_at, id
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	err := repo.Delete(userID)

	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// TestRepository_List_Success 测试获取用户列表成功
func TestRepository_List_Success(t *testing.T) {
	db, mock := setupTestDB(t)
	repo := NewRepository(db)

	offset := 0
	limit := 10
	total := int64(25)

	// 设置期望：COUNT查询（GORM会自动添加软删除条件）
	countRows := sqlmock.NewRows([]string{"count"}).AddRow(total)
	mock.ExpectQuery(`SELECT count\(\*\) FROM "users" WHERE "users"\."deleted_at" IS NULL`).
		WillReturnRows(countRows)

	// 设置期望：SELECT查询（GORM会自动添加软删除条件）
	// 当offset=0时，GORM不会生成OFFSET子句，只有LIMIT，所以只有1个参数
	rows := sqlmock.NewRows([]string{"id", "created_at", "updated_at", "deleted_at", "username", "email", "password", "first_name", "last_name", "phone", "status", "is_admin"})
	for i := 1; i <= 10; i++ {
		rows.AddRow(
			uint(i),
			time.Now(),
			time.Now(),
			nil,
			"user"+fmt.Sprintf("%d", i),
			"user"+fmt.Sprintf("%d", i)+"@example.com",
			"hashedpassword",
			"First",
			"Last",
			"",
			"active",
			false,
		)
	}

	mock.ExpectQuery(`SELECT \* FROM "users" WHERE "users"\."deleted_at" IS NULL LIMIT`).
		WithArgs(limit).
		WillReturnRows(rows)

	users, totalCount, err := repo.List(offset, limit)

	assert.NoError(t, err)
	assert.Equal(t, total, totalCount)
	assert.Len(t, users, 10)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// TestRepository_List_Empty 测试空列表
func TestRepository_List_Empty(t *testing.T) {
	db, mock := setupTestDB(t)
	repo := NewRepository(db)

	offset := 0
	limit := 10
	total := int64(0)

	// 设置期望：COUNT查询（GORM会自动添加软删除条件）
	countRows := sqlmock.NewRows([]string{"count"}).AddRow(total)
	mock.ExpectQuery(`SELECT count\(\*\) FROM "users" WHERE "users"\."deleted_at" IS NULL`).
		WillReturnRows(countRows)

	// 设置期望：SELECT查询（空结果，GORM会自动添加软删除条件）
	// 当offset=0时，GORM不会生成OFFSET子句，只有LIMIT，所以只有1个参数
	rows := sqlmock.NewRows([]string{"id", "created_at", "updated_at", "deleted_at", "username", "email", "password", "first_name", "last_name", "phone", "status", "is_admin"})
	mock.ExpectQuery(`SELECT \* FROM "users" WHERE "users"\."deleted_at" IS NULL LIMIT`).
		WithArgs(limit).
		WillReturnRows(rows)

	users, totalCount, err := repo.List(offset, limit)

	assert.NoError(t, err)
	assert.Equal(t, total, totalCount)
	assert.Len(t, users, 0)
	assert.NoError(t, mock.ExpectationsWereMet())
}
