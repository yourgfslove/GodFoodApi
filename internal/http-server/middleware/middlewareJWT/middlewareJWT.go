package middlewareJWT

import (
	"context"
	"github.com/yourgfslove/GodFoodApi/internal/lib/api/response"
	"github.com/yourgfslove/GodFoodApi/internal/lib/auth/JWT"
	"github.com/yourgfslove/GodFoodApi/internal/lib/auth/getToken"
	"log/slog"
	"net/http"
)

func AuthJWTMiddleware(log *slog.Logger, tokenSecret string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token, err := getToken.GetTokenFromHeader(r.Header, "Bearer")
			if err != nil {
				response.Error(log, w, r,
					"failed to get the token",
					"failed to get the token",
					http.StatusUnauthorized)
				return
			}

			userID, err := JWT.ValidateJWT(token, tokenSecret)
			if err != nil {
				response.Error(log, w, r,
					"failed to validate the token",
					"failed to validate the token",
					http.StatusUnauthorized)
				return
			}
			ctx := context.WithValue(r.Context(), "userID", userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
