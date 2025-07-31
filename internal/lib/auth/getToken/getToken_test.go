package getToken

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetToken(t *testing.T) {
	testcases := []struct {
		name      string
		headerVal string
		token     string
		err       string
	}{
		{
			name:      "Valid token",
			headerVal: "Bearer qqq",
			token:     "qqq",
		},
		{
			name:      "Valid with a lot of space",
			headerVal: "Bearer 			   qqq  ",
			token:     "qqq",
		},
		{
			name:      "Wrong prefix",
			headerVal: "Wrong prefix qqq",
			err:       "invalid authorization header",
		},
		{
			name:      "No header",
			headerVal: "",
			err:       "no authorization header found",
		},
		{
			name:      "No Token",
			headerVal: "Bearer",
			err:       "empty token",
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			h := http.Header{}
			if tc.headerVal != "" {
				h.Set("Authorization", tc.headerVal)
			}
			RetToken, err := GetTokenFromHeader(h, "Bearer")
			if tc.err != "" {
				assert.EqualError(t, err, tc.err, "expected error")
			} else {
				assert.NoError(t, err, "Expected no err, but got one")
				assert.Equal(t, tc.token, RetToken, "expected tokens does not match")
			}

		})
	}
}
