package utils

import (
	"golang.org/x/crypto/bcrypt"
)

const (
	// BcryptCost 密码加密成本因子
	// 值越高，加密越慢但越安全
	// 推荐值：
	// - 10: 默认值，适合2010年代的硬件
	// - 12: 适合2020年代的硬件（推荐）
	// - 14: 适合高安全要求场景
	// 注意：每增加1，计算时间大约翻倍
	BcryptCost = 12
)

// HashPassword 加密密码
// 安全性分析：
//  1. ✅ 抗彩虹表攻击：bcrypt每次加密都会自动生成随机salt（22字节），
//     即使相同密码也会产生不同的hash，彩虹表攻击无效
//  2. ✅ 抗撞库攻击：bcrypt是慢速哈希算法，增加暴力破解成本
//     Cost=12时，单次加密约需100-300ms，大幅降低撞库效率
//  3. ✅ Salt内置：salt自动包含在hash字符串中，无需单独存储
//  4. ⚠️ 建议：配合应用层措施（登录失败次数限制、密码强度要求等）
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), BcryptCost)
	return string(bytes), err
}

// CheckPassword 验证密码
// 安全性：使用bcrypt的CompareHashAndPassword，会从hash中提取salt进行验证
func CheckPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}













