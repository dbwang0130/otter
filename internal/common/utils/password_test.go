package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"
)

// TestHashPassword_Success 测试密码哈希成功
func TestHashPassword_Success(t *testing.T) {
	password := "testPassword123"

	hashedPassword, err := HashPassword(password)

	assert.NoError(t, err)
	assert.NotEmpty(t, hashedPassword)
	assert.NotEqual(t, password, hashedPassword) // 哈希后的密码应该和原密码不同
	assert.Greater(t, len(hashedPassword), 50)   // bcrypt哈希通常比较长
}

// TestHashPassword_DifferentPasswords 测试不同密码生成不同的哈希
func TestHashPassword_DifferentPasswords(t *testing.T) {
	password1 := "password1"
	password2 := "password2"

	hash1, err1 := HashPassword(password1)
	hash2, err2 := HashPassword(password2)

	assert.NoError(t, err1)
	assert.NoError(t, err2)
	assert.NotEqual(t, hash1, hash2) // 不同密码应该生成不同的哈希
}

// TestHashPassword_SamePasswordDifferentHash 测试相同密码每次生成不同的哈希（由于salt）
func TestHashPassword_SamePasswordDifferentHash(t *testing.T) {
	password := "samePassword123"

	hash1, err1 := HashPassword(password)
	hash2, err2 := HashPassword(password)

	assert.NoError(t, err1)
	assert.NoError(t, err2)
	assert.NotEqual(t, hash1, hash2) // 由于salt，相同密码每次生成的哈希都不同
}

// TestHashPassword_EmptyPassword 测试空密码
func TestHashPassword_EmptyPassword(t *testing.T) {
	hashedPassword, err := HashPassword("")

	assert.NoError(t, err)
	assert.NotEmpty(t, hashedPassword)
}

// TestHashPassword_LongPassword 测试长密码（接近但不超过bcrypt的72字节限制）
func TestHashPassword_LongPassword(t *testing.T) {
	// bcrypt限制密码长度为72字节，测试接近限制的密码
	longPassword := ""
	for i := 0; i < 70; i++ { // 70个字符，确保不超过72字节
		longPassword += "a"
	}

	hashedPassword, err := HashPassword(longPassword)

	assert.NoError(t, err)
	assert.NotEmpty(t, hashedPassword)
}

// TestHashPassword_ExceedsMaxLength 测试超过bcrypt最大长度的密码（应该返回错误）
func TestHashPassword_ExceedsMaxLength(t *testing.T) {
	// bcrypt限制密码长度为72字节，创建一个超过限制的密码
	longPassword := ""
	for i := 0; i < 80; i++ { // 80个字符，超过72字节限制
		longPassword += "a"
	}

	hashedPassword, err := HashPassword(longPassword)

	assert.Error(t, err)
	assert.Empty(t, hashedPassword)
	assert.Contains(t, err.Error(), "password length exceeds 72 bytes")
}

// TestHashPassword_SpecialCharacters 测试特殊字符密码
func TestHashPassword_SpecialCharacters(t *testing.T) {
	password := "!@#$%^&*()_+-=[]{}|;:,.<>?"

	hashedPassword, err := HashPassword(password)

	assert.NoError(t, err)
	assert.NotEmpty(t, hashedPassword)
}

// TestCheckPassword_Success 测试密码验证成功
func TestCheckPassword_Success(t *testing.T) {
	password := "testPassword123"

	hashedPassword, err := HashPassword(password)
	assert.NoError(t, err)

	isValid := CheckPassword(password, hashedPassword)

	assert.True(t, isValid)
}

// TestCheckPassword_WrongPassword 测试错误密码
func TestCheckPassword_WrongPassword(t *testing.T) {
	password := "testPassword123"
	wrongPassword := "wrongPassword"

	hashedPassword, err := HashPassword(password)
	assert.NoError(t, err)

	isValid := CheckPassword(wrongPassword, hashedPassword)

	assert.False(t, isValid)
}

// TestCheckPassword_EmptyPassword 测试空密码验证
func TestCheckPassword_EmptyPassword(t *testing.T) {
	hashedPassword, err := HashPassword("")
	assert.NoError(t, err)

	isValid := CheckPassword("", hashedPassword)

	assert.True(t, isValid)
}

// TestCheckPassword_EmptyHash 测试空哈希
func TestCheckPassword_EmptyHash(t *testing.T) {
	isValid := CheckPassword("password", "")

	assert.False(t, isValid)
}

// TestCheckPassword_InvalidHash 测试无效哈希格式
func TestCheckPassword_InvalidHash(t *testing.T) {
	isValid := CheckPassword("password", "invalid-hash-format")

	assert.False(t, isValid)
}

// TestCheckPassword_CaseSensitive 测试密码大小写敏感
func TestCheckPassword_CaseSensitive(t *testing.T) {
	password := "TestPassword123"

	hashedPassword, err := HashPassword(password)
	assert.NoError(t, err)

	// 测试不同大小写
	testCases := []struct {
		name     string
		password string
		expect   bool
	}{
		{"正确密码", "TestPassword123", true},
		{"小写开头", "testPassword123", false},
		{"全小写", "testpassword123", false},
		{"全大写", "TESTPASSWORD123", false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			isValid := CheckPassword(tc.password, hashedPassword)
			assert.Equal(t, tc.expect, isValid)
		})
	}
}

// TestCheckPassword_Unicode 测试Unicode字符密码
func TestCheckPassword_Unicode(t *testing.T) {
	password := "密码123🔒"

	hashedPassword, err := HashPassword(password)
	assert.NoError(t, err)

	isValid := CheckPassword(password, hashedPassword)

	assert.True(t, isValid)
}

// TestHashPassword_VerifyBcryptCost 测试验证bcrypt成本因子
func TestHashPassword_VerifyBcryptCost(t *testing.T) {
	password := "testPassword123"

	hashedPassword, err := HashPassword(password)
	assert.NoError(t, err)

	// 解析哈希以验证成本因子
	cost, err := bcrypt.Cost([]byte(hashedPassword))
	assert.NoError(t, err)
	assert.Equal(t, BcryptCost, cost) // 验证使用的成本因子
}

// TestCheckPassword_WithDifferentHash 测试使用不同哈希验证
func TestCheckPassword_WithDifferentHash(t *testing.T) {
	password := "testPassword123"

	// 生成两个不同的哈希
	hash1, err1 := HashPassword(password)
	hash2, err2 := HashPassword(password)

	assert.NoError(t, err1)
	assert.NoError(t, err2)
	assert.NotEqual(t, hash1, hash2)

	// 两个哈希都应该能验证同一个密码
	assert.True(t, CheckPassword(password, hash1))
	assert.True(t, CheckPassword(password, hash2))
}

// TestHashPassword_Performance 测试哈希性能（确保不会太慢）
func TestHashPassword_Performance(t *testing.T) {
	password := "testPassword123"

	// 多次哈希，确保性能可接受
	for i := 0; i < 10; i++ {
		_, err := HashPassword(password)
		assert.NoError(t, err)
	}
}
