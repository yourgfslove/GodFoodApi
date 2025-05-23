package register

import (
	"context"
	_ "database/sql"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"github.com/go-playground/validator/v10"
	"github.com/yourgfslove/GodFoodApi/internal/database"
	"github.com/yourgfslove/GodFoodApi/internal/lib/api/response"
	"github.com/yourgfslove/GodFoodApi/internal/lib/auth/hashPassword"
	"github.com/yourgfslove/GodFoodApi/internal/lib/logger/sl"
	"github.com/yourgfslove/GodFoodApi/internal/lib/validation/phoneValidation"
	"log/slog"
	"net/http"
)

type Request struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password"`
	Role     string `json:"role"`
	Phone    string `json:"phone"`
}

type Response struct {
	response.Response
	Email string `json:"email"`
}

type UserSaver interface {
	CreateUser(ctx context.Context, arg database.CreateUserParams) (database.User, error)
}

func New(log *slog.Logger, saver UserSaver) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "http.auth.register.new"
		log = log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())))
		var req Request

		err := render.DecodeJSON(r.Body, &req)
		if err != nil {
			log.Error("failed to decode request", sl.Err(err))
			render.JSON(w, r, response.Error("failed to decode JSON"))
			return
		}
		log.Info("request body decoded", slog.Any("request", req))

		if err := validator.New().Struct(req); err != nil {
			log.Error("failed to validate email", sl.Err(err))
			render.JSON(w, r, response.Error("failed to validate Email"))
			return
		}
		if !phoneValidation.IsValidRuPhoneNumber(req.Phone) {
			log.Error("invalid phone", sl.Err(err))
			render.JSON(w, r, response.Error("failed to validate phone"))
			return
		}
		hashedPassword, err := hashPassword.HashPassword(req.Password)
		if err != nil {
			log.Error("failed to hash password", sl.Err(err))
			render.JSON(w, r, response.Error("failed to set password"))
			return
		}
		savedUser, err := saver.CreateUser(context.Background(), database.CreateUserParams{
			Email:        req.Email,
			HashPassword: hashedPassword,
			UserRole:     req.Role,
			Phone:        req.Phone,
		})
		if err != nil {
			log.Error("failed to create user", sl.Err(err))
			render.JSON(w, r, response.Error("failed to create user"))
			return
		}
		render.JSON(w, r, Response{
			Response: response.OK(),
			Email:    savedUser.Email,
		})
	}
}
