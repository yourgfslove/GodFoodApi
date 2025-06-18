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
	"log/slog"
	"net/http"
	"time"
)

type loginRequest struct {
	Email    string `json:"email" validate:"required,email" example:"user@example.com"`
	Password string `json:"password" example:"password123"`
}

type loginResponse struct {
	Jwt          string `json:"jwt" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJHb2RGb29kIiwic3ViIjoiMTUiLCJleHAiOjE3NTAwOTE3MDIsImlhdCI6MTc1MDA4ODEwMn0.NUKzisW-QLalMwaADr5dwb9VnfYb3W-pivD5f4hVZ5A"`
	RefreshToken string `json:"refresh_token" exmaple:"7027102e5ddecf9dfaa1fa602851f7e77a212c486a37f014a5c016d3f3a2cdce"`
	Email        string `json:"email" example:"user@example.com"`
}

type RefreshTokenSaverGetter interface {
	CreateToken(ctx context.Context, arg database.CreateTokenParams) (database.CreateTokenRow, error)
	GetTokensByUser(ctx context.Context, userID int32) ([]database.Refreshtoken, error)
}

type UserGetter interface {
	GetUserByEmail(ctx context.Context, email string) (database.User, error)
}

// Login godoc
// @Summary Авторизация
// @Description Принимает email и пароль, возвращает JWT и refresh-token
// @Tags auth
// @Accept json
// @Produce json
// @Param loginRequest body login.loginRequest true "Данные для входа"
// @Success 200 {object} login.loginResponse
// @Failure 400 {object} response.Response
// @Fastringilure 401 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /login [post]
func New(log *slog.Logger, saver RefreshTokenSaverGetter, userGetter UserGetter, tokenSecret string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "http-server.auth.login"
		log = log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())))
		var req loginRequest
		if err := render.DecodeJSON(r.Body, &req); err != nil {
			response.Error(log, w, r, "something went wrong", "failed to decode JSON", http.StatusInternalServerError)
			return
		}
		log.Info("JSON body decoded")
		if err := validator.New().Struct(req); err != nil {
			if validationErrors, ok := err.(validator.ValidationErrors); ok {
				response.ValidationError(log, w, r, validationErrors)
			} else {
				response.Error(log, w, r, "failed to validate", "failed to validate JSON", http.StatusBadRequest)
			}
			return
		}

		if req.Email == "" || req.Password == "" {
			response.Error(log, w, r, "empty request", "No email or pass", http.StatusBadRequest)
			return
		}
		user, err := userGetter.GetUserByEmail(r.Context(), req.Email)

		if err != nil {
			response.Error(log, w, r, "wrong email", "no User on email", http.StatusUnauthorized)
			return
		}

		if err := hashPassword.VerifyPassword(req.Password, user.HashPassword); err != nil {
			response.Error(log, w, r, "wrong password", "failed to verify pass", http.StatusUnauthorized)
			return
		}

		refreshTokens, err := saver.GetTokensByUser(r.Context(), user.ID)
		if err != nil {
			response.Error(log, w, r, "something went wrong", "failed to get tokens by user", http.StatusInternalServerError)
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
				response.Error(log, w, r, "something went wrong", "failed to make new RefreshToken", http.StatusInternalServerError)
				return
			}
			_, err = saver.CreateToken(r.Context(), database.CreateTokenParams{
				Token:  userRefreshToken,
				UserID: user.ID,
			})
			if err != nil {
				response.Error(log, w, r, "something went wrong", "failed to save refresh token", http.StatusInternalServerError)
				return
			}
			log.Info("refresh token saved")
		}

		jwt, err := JWT.MakeJWT(user.ID, tokenSecret, time.Hour)
		if err != nil {
			response.Error(log, w, r, "something went wrong", "failed to make jwt", http.StatusInternalServerError)
			return
		}

		render.Status(r, http.StatusOK)
		render.JSON(w, r, loginResponse{
			Jwt:          jwt,
			RefreshToken: userRefreshToken,
			Email:        user.Email})
	}
}
