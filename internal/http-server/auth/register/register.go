package register

import (
	"context"
	"database/sql"
	_ "database/sql"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"github.com/go-playground/validator/v10"
	"github.com/yourgfslove/GodFoodApi/internal/database"
	"github.com/yourgfslove/GodFoodApi/internal/lib/api/response"
	"github.com/yourgfslove/GodFoodApi/internal/lib/auth/JWT"
	"github.com/yourgfslove/GodFoodApi/internal/lib/auth/hashPassword"
	"github.com/yourgfslove/GodFoodApi/internal/lib/auth/refreshToken"
	"github.com/yourgfslove/GodFoodApi/internal/lib/logger/sl"
	"github.com/yourgfslove/GodFoodApi/internal/lib/validation/phoneValidation"
	"log/slog"
	"net/http"
	"time"
)

type Request struct {
	Email    string `json:"email" validate:"required,email" example:"user@example.com"`
	Password string `json:"password" validate:"required" example:"password123"`
	Role     string `json:"role" validate:"required,oneof=courier restaurant customer" example:"customer"`
	Phone    string `json:"phone" example:"89035433434"`
	Address  string `json:"address,omitempty" example:"123 street 1"`
	Name     string `json:"name" example:"Bill"`
}

type Response struct {
	Email        string `json:"email" example:"user@example.com"`
	RefreshToken string `json:"refresh_token" example:"7027102e5ddecf9dfaa1fa602851f7e77a212c486a37f014a5c016d3f3a2cdce"`
	JWT          string `json:"jwt" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJHb2RGb29kIiwic3ViIjoiMTUiLCJleHAiOjE3NTAwOTE3MDIsImlhdCI6MTc1MDA4ODEwMn0.NUKzisW-QLalMwaADr5dwb9VnfYb3W-pivD5f4hVZ5A"`
	Address      string `json:"address,omitempty" example:"123 street 1"`
	Name         string `json:"name" example:"Bill"`
}

type UserSaver interface {
	CreateUser(ctx context.Context, arg database.CreateUserParams) (database.User, error)
}

type RefreshTokenSaver interface {
	CreateToken(ctx context.Context, arg database.CreateTokenParams) (database.CreateTokenRow, error)
}

// RegisterUser godoc
// @Summary Регистрация
// @Description Создает нового пользователя с ролью(courier, restaurant, customer), email, телефоном и паролем. Возвращает JWT и refresh-token
// @Tags auth
// @Accept json
// @Produce json
// @Param request body register.Request true "Данные для Регистрации"
// @Success 200 {object} register.Response "Пользователь успешно зарегистрирован"
// @Failure 400 {object} response.Response "Некорректные данные"
// @Failure 500 {object} response.Response "Серверная Ошибка"
// @Router /register [post]
func New(log *slog.Logger, saver UserSaver, tokenSaver RefreshTokenSaver, tokenSecret string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "http.auth.register.new"
		log = log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())))

		var req Request
		err := render.DecodeJSON(r.Body, &req)
		if err != nil {
			response.Error(log, w, r, "failed to decode request", "Failed to decode JSON", http.StatusInternalServerError)
			return
		}
		log.Info("request body decoded", slog.Any("request", req))

		if err := validator.New().Struct(req); err != nil {
			if validationErrors, ok := err.(validator.ValidationErrors); ok {
				response.ValidationError(log, w, r, validationErrors)
			} else {
				response.Error(log, w, r, "failed to validate", "failed to validate JSON", http.StatusBadRequest)
			}
			return
		}

		if !phoneValidation.IsValidRuPhoneNumber(req.Phone) {
			response.Error(log, w, r, "failed to validate phone", "invalid phone", http.StatusBadRequest)
			return
		}

		hashedPassword, err := hashPassword.HashPassword(req.Password)
		if err != nil {
			response.Error(log, w, r, "failed to set password", "failed to hash password", http.StatusInternalServerError)
			return
		}

		savedUser, err := saver.CreateUser(r.Context(), database.CreateUserParams{
			Email:        req.Email,
			HashPassword: hashedPassword,
			UserRole:     req.Role,
			Phone:        req.Phone,
			Address: sql.NullString{
				String: req.Address,
				Valid:  req.Address != ""},
			UserName: sql.NullString{
				String: req.Name,
				Valid:  req.Name != "",
			},
		})
		if err != nil {
			response.Error(log, w, r, "failed to create user", "failed to create user", http.StatusInternalServerError)
			return
		}

		newRefreshToken, err := refreshToken.MakeRefreshToken()
		if err != nil {
			response.Error(log, w, r, "something went wrong", sl.Err(err).String(), http.StatusInternalServerError)
			return
		}
		savedToken, err := tokenSaver.CreateToken(r.Context(), database.CreateTokenParams{
			UserID: savedUser.ID,
			Token:  newRefreshToken})
		if err != nil {
			response.Error(log, w, r, "something went wrong", sl.Err(err).String(), http.StatusInternalServerError)
			return
		}
		newJWT, err := JWT.MakeJWT(savedUser.ID, tokenSecret, time.Hour)
		if err != nil {
			response.Error(log, w, r, "something went wrong", sl.Err(err).String(), http.StatusInternalServerError)
			return
		}

		render.Status(r, http.StatusCreated)
		render.JSON(w, r, Response{
			Email:        savedUser.Email,
			RefreshToken: savedToken.Token,
			JWT:          newJWT,
			Address:      savedUser.Address.String,
			Name:         savedUser.UserName.String,
		})
	}
}
