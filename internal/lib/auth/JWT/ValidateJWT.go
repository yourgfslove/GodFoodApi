package JWT

import (
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"strconv"
	"time"
)

func ValidateJWT(tokenString, tokenSecret string) (int32, error) {
	token, err := jwt.ParseWithClaims(tokenString, &jwt.RegisteredClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(tokenSecret), nil
	})
	if err != nil {
		return 0, err
	}
	if claims, ok := token.Claims.(*jwt.RegisteredClaims); ok && token.Valid {
		if claims.ExpiresAt != nil && time.Now().After(claims.ExpiresAt.Time) {
			return 0, fmt.Errorf("token expired")
		}
		userId, err := strconv.Atoi(claims.Subject)
		if err != nil {
			return 0, err
		}
		return int32(userId), nil
	}
	return 0, fmt.Errorf("invalid token")
}
