package login

import (
	"context"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"github.com/go-playground/validator/v10"
	"github.com/yourgfslove/GodFoodApi/internal/database"
	"github.com/yourgfslove/GodFoodApi/internal/lib/api/response"
	"github.com/yourgfslove/GodFoodApi/internal/lib/auth/JWT"
	"github.com/yourgfslove/GodFoodApi/internal/lib/auth/hashPassword"
	"github.com/yourgfslove/GodFoodApi/internal/lib/auth/refreshToken"
	"github.com/yourgfslove/GodFoodApi/internal/lib/logger/sl"
	"log/slog"
	"net/http"
	"time"
)

type loginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password"`
}

type loginResponse struct {
	response.Response
	Jwt          string `json:"jwt"`
	RefreshToken string `json:"refresh_token"`
	Email        string `json:"email"`
}

type RefreshTokenSaverGetter interface {
	CreateToken(ctx context.Context, arg database.CreateTokenParams) (database.CreateTokenRow, error)
	GetTokensByUser(ctx context.Context, userID int32) ([]database.Refreshtoken, error)
}

type UserGetter interface {
	GetUserByEmail(ctx context.Context, email string) (database.User, error)
}

func New(log *slog.Logger, saver RefreshTokenSaverGetter, userGetter UserGetter, tokenSecret string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "http-server.auth.login"
		log = log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())))
		var req loginRequest
		if err := render.DecodeJSON(r.Body, &req); err != nil {
			log.Error("failed to decode body", sl.Err(err))
			w.WriteHeader(http.StatusBadRequest)
			render.JSON(w, r, response.Error("failed to decode JSON"))
			return
		}
		log.Info("JSON body decoded")
		if err := validator.New().Struct(req); err != nil {
			log.Error("failed to validate email", sl.Err(err))
			w.WriteHeader(http.StatusBadRequest)
			render.JSON(w, r, response.Error("failed to validate email"))
			return
		}
		user, err := userGetter.GetUserByEmail(r.Context(), req.Email)
		if err != nil {
			log.Error("failed to get user by email", sl.Err(err))
			w.WriteHeader(http.StatusInternalServerError)
			render.JSON(w, r, response.Error("no user on email"))
			return
		}
		if err := hashPassword.VerifyPassword(req.Password, user.HashPassword); err != nil {
			log.Error("failed to verify password", sl.Err(err))
			w.WriteHeader(http.StatusUnauthorized)
			render.JSON(w, r, response.Error("wrong password"))
			return
		}
		refreshTokens, err := saver.GetTokensByUser(r.Context(), user.ID)
		if err != nil {
			log.Error("failed to get tokens by user", sl.Err(err))
			w.WriteHeader(http.StatusInternalServerError)
			render.JSON(w, r, response.Error("failed to get tokens by user"))
			return
		}
		var userRefreshToken string
		for _, t := range refreshTokens {
			if !t.RevokedAt.Valid && t.ExpiresAt.Time.After(time.Now()) {
				userRefreshToken = t.Token
				break
			}
		}
		if userRefreshToken == "" {
			userRefreshToken, err = refreshToken.MakeRefreshToken()
			if err != nil {
				log.Error("failed to make refresh token", sl.Err(err))
				w.WriteHeader(http.StatusInternalServerError)
				render.JSON(w, r, response.Error("something went wrong"))
				return
			}
			_, err = saver.CreateToken(r.Context(), database.CreateTokenParams{
				Token:  userRefreshToken,
				UserID: user.ID,
			})
			if err != nil {
				log.Error("failed to save refresh token", sl.Err(err))
				w.WriteHeader(http.StatusInternalServerError)
				render.JSON(w, r, response.Error("something went wrong"))
				return
			}
			log.Info("refresh token saved")
		}
		jwt, err := JWT.MakeJWT(user.ID, tokenSecret, time.Hour)
		if err != nil {
			log.Error("failed to make jwt", sl.Err(err))
			w.WriteHeader(http.StatusInternalServerError)
			render.JSON(w, r, response.Error("something went wrong"))
			return
		}
		render.JSON(w, r, loginResponse{
			Response:     response.OK(),
			Jwt:          jwt,
			RefreshToken: userRefreshToken,
			Email:        user.Email})
	}
}
