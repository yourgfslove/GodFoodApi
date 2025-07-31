package hashPassword

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHashPassword(t *testing.T) {
	testcases := []struct {
		name   string
		pass   string
		expErr bool
	}{
		{
			name:   "valid pass",
			pass:   "1234",
			expErr: false,
		},
		{
			name:   "empty pass",
			pass:   "",
			expErr: true,
		},
		{
			name:   "too long pass",
			pass:   string(make([]byte, 73)),
			expErr: true,
		}}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			hash, err := HashPassword(tc.pass)
			if tc.expErr {
				assert.Error(t, err, "expect err, but not got")
				assert.Empty(t, hash, "expect empty hash? but not one")
			} else {
				assert.NoError(t, err, "expect no err, but got one")
				assert.NotEmpty(t, hash, "expect hash, but not got one")
			}
		})
	}
}

func TestVertifyPass(t *testing.T) {
	pass := "sec123"
	hash, err := HashPassword(pass)
	assert.NoError(t, err, "expect no err, but got one")

	err = VerifyPassword(pass, hash)
	assert.NoError(t, err, "expect no err, but got one")
	err = VerifyPassword("wrongone", hash)
	assert.Error(t, err, "expect err, but not got one")
}
