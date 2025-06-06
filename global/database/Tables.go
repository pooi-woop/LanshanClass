// FilePath: C:/LanshanClass1.3/global/database\Tables.go
package database

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
)

// User 表示用户表
type User struct {
	ID       uint   `gorm:"primaryKey;autoIncrement"`
	Username string `gorm:"type:varchar(100);uniqueIndex;not null"`
	Salt     string `gorm:"not_null"`
	Hash     string `gorm:"not_null"`
}

// GenerateSalt 生成随机盐
func GenerateSalt() (string, error) {
	// 生成 16 字节的随机盐
	salt := make([]byte, 16)
	_, err := rand.Read(salt)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(salt), nil
}

// HashPassword 生成哈希密码
func HashPassword(password, salt string) string {
	// 将密码和盐拼接后进行哈希
	hash := sha256.Sum256([]byte(password + salt))
	return hex.EncodeToString(hash[:])
}

// CreateUser 创建新用户
func CreateUser(username, password string) error {
	// 生成盐
	salt, err := GenerateSalt()
	if err != nil {
		return fmt.Errorf("failed to generate salt: %w", err)
	}

	// 生成哈希密码
	hash := HashPassword(password, salt)

	// 创建用户
	user := User{
		Username: username,
		Salt:     salt,
		Hash:     hash,
	}

	// 保存到数据库
	if err := DB.Create(&user).Error; err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	return nil
}

// VerifyPassword 验证密码
func VerifyPassword(username, password string) (bool, error) {
	// 查询用户
	var user User
	if err := DB.Where("username = ?", username).First(&user).Error; err != nil {
		return false, fmt.Errorf("failed to find user: %w", err)
	}

	// 生成哈希密码
	hash := HashPassword(password, user.Salt)

	// 比较哈希值
	return hash == user.Hash, nil
}
