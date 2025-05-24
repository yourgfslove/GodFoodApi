package JWT

import (
	"github.com/golang-jwt/jwt/v5"
	"strconv"
	"time"
)

func MakeJWT(userID int32, tokenSecret string, expiresIN time.Duration) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiresIN)),
		Issuer:    "GodFood",
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		Subject:   strconv.Itoa(int(userID)),
	})
	tokenString, err := token.SignedString([]byte(tokenSecret))
	if err != nil {
		return "", err
	}
	return tokenString, nil
}
