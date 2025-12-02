package auth

import (
	"database/sql"
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

// TestRefreshTokenRepository_Create 测试创建refresh token
func TestRefreshTokenRepository_Create(t *testing.T) {
	db, mock := setupTestDB(t)
	repo := NewRefreshTokenRepository(db)

	token := &RefreshToken{
		Token:     "test-refresh-token",
		UserID:    1,
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
		IsRevoked: false,
	}

	// 设置期望：INSERT 语句（GORM会自动添加created_at, updated_at, deleted_at等字段）
	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO "refresh_tokens"`).
		WithArgs(
			sqlmock.AnyArg(), // CreatedAt
			sqlmock.AnyArg(), // UpdatedAt
			sqlmock.AnyArg(), // DeletedAt (NULL)
			token.Token,
			token.UserID,
			token.ExpiresAt,
			token.IsRevoked,
		).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	mock.ExpectCommit()

	err := repo.Create(token)

	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// TestRefreshTokenRepository_GetByToken_Success 测试根据token获取refresh token（成功）
func TestRefreshTokenRepository_GetByToken_Success(t *testing.T) {
	db, mock := setupTestDB(t)
	repo := NewRefreshTokenRepository(db)

	tokenStr := "valid-token"
	expectedToken := &RefreshToken{
		ID:        1,
		Token:     tokenStr,
		UserID:    1,
		ExpiresAt: time.Now().Add(24 * time.Hour),
		IsRevoked: false,
	}

	// 设置期望：SELECT 语句（GORM会自动添加deleted_at IS NULL、ORDER BY和LIMIT等条件）
	rows := sqlmock.NewRows([]string{"id", "created_at", "updated_at", "deleted_at", "token", "user_id", "expires_at", "is_revoked"}).
		AddRow(
			expectedToken.ID,
			time.Now(),
			time.Now(),
			nil,
			expectedToken.Token,
			expectedToken.UserID,
			expectedToken.ExpiresAt,
			expectedToken.IsRevoked,
		)

	mock.ExpectQuery(`SELECT \* FROM "refresh_tokens"`).
		WithArgs(tokenStr, false, 1). // GORM的First()会添加LIMIT 1作为第3个参数
		WillReturnRows(rows)

	token, err := repo.GetByToken(tokenStr)

	assert.NoError(t, err)
	assert.NotNil(t, token)
	assert.Equal(t, expectedToken.Token, token.Token)
	assert.Equal(t, expectedToken.UserID, token.UserID)
	assert.Equal(t, expectedToken.IsRevoked, token.IsRevoked)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// TestRefreshTokenRepository_GetByToken_NotFound 测试token不存在
func TestRefreshTokenRepository_GetByToken_NotFound(t *testing.T) {
	db, mock := setupTestDB(t)
	repo := NewRefreshTokenRepository(db)

	tokenStr := "non-existent-token"

	// 设置期望：返回 sql.ErrNoRows
	mock.ExpectQuery(`SELECT \* FROM "refresh_tokens"`).
		WithArgs(tokenStr, false, 1). // GORM的First()会添加LIMIT 1
		WillReturnError(sql.ErrNoRows)

	token, err := repo.GetByToken(tokenStr)

	assert.Error(t, err)
	assert.Nil(t, token)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// TestRefreshTokenRepository_GetByToken_Revoked 测试已撤销的token不会被返回
func TestRefreshTokenRepository_GetByToken_Revoked(t *testing.T) {
	db, mock := setupTestDB(t)
	repo := NewRefreshTokenRepository(db)

	tokenStr := "revoked-token"

	// 设置期望：查询 is_revoked = false，所以已撤销的token不会返回
	mock.ExpectQuery(`SELECT \* FROM "refresh_tokens"`).
		WithArgs(tokenStr, false, 1). // GORM的First()会添加LIMIT 1
		WillReturnError(sql.ErrNoRows)

	token, err := repo.GetByToken(tokenStr)

	assert.Error(t, err)
	assert.Nil(t, token)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// TestRefreshTokenRepository_RevokeByToken 测试撤销指定token
func TestRefreshTokenRepository_RevokeByToken(t *testing.T) {
	db, mock := setupTestDB(t)
	repo := NewRefreshTokenRepository(db)

	tokenStr := "token-to-revoke"

	// 设置期望：GORM的Update会自动开启事务，并添加updated_at和deleted_at条件
	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE "refresh_tokens" SET`).
		WithArgs(true, sqlmock.AnyArg(), tokenStr). // is_revoked, updated_at, token
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	err := repo.RevokeByToken(tokenStr)

	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// TestRefreshTokenRepository_RevokeByUserID 测试撤销用户的所有token
func TestRefreshTokenRepository_RevokeByUserID(t *testing.T) {
	db, mock := setupTestDB(t)
	repo := NewRefreshTokenRepository(db)

	userID := uint(1)

	// 设置期望：GORM的Update会自动开启事务，并添加updated_at和deleted_at条件
	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE "refresh_tokens" SET`).
		WithArgs(true, sqlmock.AnyArg(), userID). // is_revoked, updated_at, user_id
		WillReturnResult(sqlmock.NewResult(0, 3)) // 假设撤销了3个token
	mock.ExpectCommit()

	err := repo.RevokeByUserID(userID)

	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// TestRefreshTokenRepository_DeleteExpired 测试删除过期token
func TestRefreshTokenRepository_DeleteExpired(t *testing.T) {
	db, mock := setupTestDB(t)
	repo := NewRefreshTokenRepository(db)

	// 设置期望：GORM的Delete是软删除（UPDATE deleted_at），会自动开启事务
	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE "refresh_tokens" SET`).
		WithArgs(sqlmock.AnyArg(), "now()").      // deleted_at, expires_at条件
		WillReturnResult(sqlmock.NewResult(0, 5)) // 假设删除了5个过期token
	mock.ExpectCommit()

	err := repo.DeleteExpired()

	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// TestRefreshTokenRepository_DeleteExpired_NoExpiredTokens 测试没有过期token的情况
func TestRefreshTokenRepository_DeleteExpired_NoExpiredTokens(t *testing.T) {
	db, mock := setupTestDB(t)
	repo := NewRefreshTokenRepository(db)

	// 设置期望：GORM的Delete是软删除（UPDATE deleted_at），会自动开启事务
	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE "refresh_tokens" SET`).
		WithArgs(sqlmock.AnyArg(), "now()"). // deleted_at, expires_at条件
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectCommit()

	err := repo.DeleteExpired()

	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// TestRefreshTokenRepository_Create_DuplicateToken 测试创建重复token（唯一约束）
func TestRefreshTokenRepository_Create_DuplicateToken(t *testing.T) {
	db, mock := setupTestDB(t)
	repo := NewRefreshTokenRepository(db)

	token := &RefreshToken{
		Token:     "duplicate-token",
		UserID:    1,
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
		IsRevoked: false,
	}

	// 设置期望：INSERT 失败（唯一约束冲突）
	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO "refresh_tokens"`).
		WithArgs(
			sqlmock.AnyArg(), // CreatedAt
			sqlmock.AnyArg(), // UpdatedAt
			sqlmock.AnyArg(), // DeletedAt
			token.Token,
			token.UserID,
			token.ExpiresAt,
			token.IsRevoked,
		).
		WillReturnError(sql.ErrNoRows) // 模拟唯一约束错误
	mock.ExpectRollback()

	err := repo.Create(token)

	assert.Error(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// TestRefreshTokenRepository_GetByToken_DatabaseError 测试数据库错误
func TestRefreshTokenRepository_GetByToken_DatabaseError(t *testing.T) {
	db, mock := setupTestDB(t)
	repo := NewRefreshTokenRepository(db)

	tokenStr := "test-token"

	// 设置期望：数据库错误
	mock.ExpectQuery(`SELECT \* FROM "refresh_tokens"`).
		WithArgs(tokenStr, false, 1). // GORM的First()会添加LIMIT 1
		WillReturnError(sql.ErrConnDone)

	token, err := repo.GetByToken(tokenStr)

	assert.Error(t, err)
	assert.Nil(t, token)
	assert.NoError(t, mock.ExpectationsWereMet())
}
