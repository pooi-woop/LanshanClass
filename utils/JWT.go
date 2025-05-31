// FilePath: C:/LanshanClass1.3/utils\JWT.go

package utils

import (
	"github.com/golang-jwt/jwt/v4"
	"time"
)

var JwtSecret = []byte("114514") // JWT 密钥

// Claims 自定义 JWT Claims
type Claims struct {
	Username string `json:"username"`
	jwt.StandardClaims
}

func GenerateToken(username string) string {
	claims := Claims{
		Username: username,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(24 * time.Hour).Unix(),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, _ := token.SignedString(JwtSecret)
	return tokenString
}
