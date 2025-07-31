package phoneValidation

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPhoneValid(t *testing.T) {
	testcases := []struct {
		name  string
		phone string
		ans   bool
	}{{
		name:  "valid phone",
		phone: "8(999)-999-99-99",
		ans:   true,
	},
		{
			name:  "empty phone",
			phone: "",
			ans:   false,
		},
		{
			name:  "not valid phone",
			phone: "242341234123412",
			ans:   false,
		}}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			res := IsValidRuPhoneNumber(tc.phone)
			assert.Equal(t, tc.ans, res, "expected equal, but no")
		})
	}
}
