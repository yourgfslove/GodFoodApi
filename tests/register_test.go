package tests

import (
	"github.com/brianvoe/gofakeit/v6"
	"github.com/gavv/httpexpect/v2"
	"github.com/yourgfslove/GodFoodApi/internal/http-server/auth/register"
	"net/http"
	"net/url"
	"testing"
)

const (
	host = "localhost:8082"
)

func Test_register(t *testing.T) {
	testcases := []struct {
		name     string
		email    string
		username string
		password string
		role     string
		phone    string
		address  string
		err      string
	}{
		{
			name:     "success",
			email:    gofakeit.Email(),
			username: gofakeit.Username(),
			password: "qwerty1234",
			role:     "customer",
			phone:    "+79035433434",
			address:  gofakeit.Address().Address,
		},
		{
			name:     "not valid role",
			email:    gofakeit.Email(),
			username: gofakeit.Username(),
			password: "qwerty1234",
			role:     "cust",
			phone:    "+79035433434",
			address:  gofakeit.Address().Address,
			err:      "Role is not a valid field",
		},
		{
			name:     "not valid email",
			email:    "WrongEmail",
			username: gofakeit.Username(),
			password: "qwerty1234",
			role:     "customer",
			phone:    "+79035433434",
			address:  gofakeit.Address().Address,
			err:      "Email is not a valid email",
		},
		{
			name:     "Missing Pass",
			email:    gofakeit.Email(),
			username: gofakeit.Username(),
			role:     "cust",
			phone:    "+79035433434",
			address:  gofakeit.Address().Address,
			err:      "Password is required",
		},
	}
	for _, testcase := range testcases {
		t.Run(testcase.name, func(t *testing.T) {
			u := url.URL{
				Scheme: "http",
				Host:   host,
			}
			e := httpexpect.Default(t, u.String())
			resp := e.POST("/register").WithJSON(register.Request{
				Email:    testcase.email,
				Password: testcase.password,
				Role:     testcase.role,
				Phone:    testcase.phone,
				Address:  testcase.address,
				Name:     testcase.username,
			}).Expect()
			if testcase.err != "" {
				resp.Status(http.StatusBadRequest)
				resp.JSON().Object().Value("error").String().Contains(testcase.err)
			} else {
				resp.Status(http.StatusCreated)
				resp.JSON().Object().Value("email").String().IsEqual(testcase.email)
			}
		})
	}
}
