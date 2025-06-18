package response

import (
	"fmt"
	"github.com/go-chi/render"
	"github.com/go-playground/validator/v10"
	"log/slog"
	"net/http"
	"strings"
)

type Response struct {
	Error string `json:"error,omitempty" example:"error message"`
}

func ValidationError(log *slog.Logger, w http.ResponseWriter, r *http.Request, errs validator.ValidationErrors) {
	log.Info("Validation Error")
	var errList []string
	for _, err := range errs {
		switch err.ActualTag() {
		case "required":
			errList = append(errList, fmt.Sprintf("%s is required", err.Field()))
		case "email":
			errList = append(errList, fmt.Sprintf("%s is not a valid email", err.Field()))
		case "password":
			errList = append(errList, fmt.Sprintf("%s is not a valid password", err.Field()))
		case "role":
			errList = append(errList, fmt.Sprintf("%s is not a valid role", err.Field()))
		default:
			errList = append(errList, fmt.Sprintf("%s is not a valid field", err.Field()))
		}
	}

	render.Status(r, http.StatusBadRequest)
	render.JSON(w, r, Response{
		Error: strings.Join(errList, "; "),
	})
}

func Error(log *slog.Logger, w http.ResponseWriter, r *http.Request, msg string, logMsg string, statusCode int) {
	log.Info(logMsg)
	render.Status(r, statusCode)
	render.JSON(w, r, Response{
		Error: msg,
	})
}
