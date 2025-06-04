package register

import (
	"context"
	"database/sql"
	_ "database/sql"
	"errors"
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
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
	Role     string `json:"role" validate:"required,oneof=courier restaurant customer"`
	Phone    string `json:"phone"`
	Address  string `json:"address,omitempty"`
	Name     string `json:"name"`
}

type Response struct {
	response.Response
	Email        string `json:"email"`
	RefreshToken string `json:"refresh_token"`
	JWT          string `json:"jwt"`
	Address      string `json:"address,omitempty"`
	Name         string `json:"name"`
}

type UserSaver interface {
	CreateUser(ctx context.Context, arg database.CreateUserParams) (database.User, error)
}

type RefreshTokenSaver interface {
	CreateToken(ctx context.Context, arg database.CreateTokenParams) (database.CreateTokenRow, error)
}

func New(log *slog.Logger, saver UserSaver, tokenSaver RefreshTokenSaver, tokenSecret string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "http.auth.register.new"
		log = log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())))
		var req Request

		err := render.DecodeJSON(r.Body, &req)
		if err != nil {
			log.Error("failed to decode request", sl.Err(err))
			w.WriteHeader(http.StatusInternalServerError)
			render.JSON(w, r, response.Error("failed to decode JSON"))
			return
		}
		log.Info("request body decoded", slog.Any("request", req))

		if err := validator.New().Struct(req); err != nil {
			log.Error("failed to validate request", sl.Err(err))
			w.WriteHeader(http.StatusBadRequest)
			render.JSON(w, r, response.ValidationError(err.(validator.ValidationErrors)))
			return
		}
		if !phoneValidation.IsValidRuPhoneNumber(req.Phone) {
			log.Error("invalid phone", sl.Err(errors.New("invalid phone")))
			w.WriteHeader(http.StatusBadRequest)
			render.JSON(w, r, response.Error("failed to validate phone"))
			return
		}
		hashedPassword, err := hashPassword.HashPassword(req.Password)
		if err != nil {
			log.Error("failed to hash password", sl.Err(err))
			w.WriteHeader(http.StatusInternalServerError)
			render.JSON(w, r, response.Error("failed to set password"))
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
			log.Error("failed to create user", sl.Err(err))
			w.WriteHeader(http.StatusInternalServerError)
			render.JSON(w, r, response.Error("failed to create user"))
			return
		}
		newRefreshToken, err := refreshToken.MakeRefreshToken()
		if err != nil {
			log.Error("failed to make refresh token", sl.Err(err))
			w.WriteHeader(http.StatusInternalServerError)
			render.JSON(w, r, response.Error("something went wrong"))
			return
		}
		savedToken, err := tokenSaver.CreateToken(r.Context(), database.CreateTokenParams{
			UserID: savedUser.ID,
			Token:  newRefreshToken})
		if err != nil {
			log.Error("failed to save token", sl.Err(err))
			w.WriteHeader(http.StatusInternalServerError)
			render.JSON(w, r, response.Error("something went wrong"))
			return
		}
		newJWT, err := JWT.MakeJWT(savedUser.ID, tokenSecret, time.Hour)
		if err != nil {
			log.Error("failed to make JWT", sl.Err(err))
			w.WriteHeader(http.StatusInternalServerError)
			render.JSON(w, r, response.Error("something went wrong"))
			return
		}
		render.JSON(w, r, Response{
			Response:     response.OK(),
			Email:        savedUser.Email,
			RefreshToken: savedToken.Token,
			JWT:          newJWT,
			Address:      savedUser.Address.String,
			Name:         savedUser.UserName.String,
		})
	}
}
