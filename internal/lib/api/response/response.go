package response

import (
	"fmt"
	"github.com/go-playground/validator/v10"
	"strings"
)

type Response struct {
	Status string `json:"status"`
	Error  string `json:"error,omitempty"`
}

const (
	StatusOK    = "OK"
	StatusError = "ERROR"
)

func OK() Response {
	return Response{
		Status: StatusOK,
	}
}

func Error(msg string) Response {
	return Response{
		Status: StatusError,
		Error:  msg,
	}
}

func ValidationError(errs validator.ValidationErrors) Response {
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
	return Response{
		Status: StatusError,
		Error:  strings.Join(errList, "; "),
	}
}
