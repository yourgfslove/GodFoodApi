package getToken

import (
	"errors"
	"net/http"
	"strings"
)

func GetTokenFromHeader(headers http.Header, prefix string) (string, error) {
	authHeader := headers.Get("Authorization")
	if authHeader == "" {
		return "", errors.New("no authorization header found")
	}
	if !strings.HasPrefix(authHeader, prefix) {
		return "", errors.New("invalid authorization header")
	}
	token := strings.TrimSpace(authHeader[len(prefix):])
	if token == "" {
		return "", errors.New("empty token")
	}
	return token, nil
}
