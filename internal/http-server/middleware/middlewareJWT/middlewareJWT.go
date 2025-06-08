package middlewareJWT

import (
	"context"
	"github.com/go-chi/render"
	"github.com/yourgfslove/GodFoodApi/internal/lib/api/response"
	"github.com/yourgfslove/GodFoodApi/internal/lib/auth/JWT"
	"github.com/yourgfslove/GodFoodApi/internal/lib/auth/getToken"
	"net/http"
)

func AuthJWTMiddleware(tokenSecret string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token, err := getToken.GetTokenFromHeader(r.Header, "Bearer")
			if err != nil {
				w.WriteHeader(http.StatusUnauthorized)
				render.JSON(w, r, response.Error("failed to get the token"))
				return
			}

			userID, err := JWT.ValidateJWT(token, tokenSecret)
			if err != nil {
				w.WriteHeader(http.StatusForbidden)
				render.JSON(w, r, response.Error("failed to validate the token"))
				return
			}
			ctx := context.WithValue(r.Context(), "userID", userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
