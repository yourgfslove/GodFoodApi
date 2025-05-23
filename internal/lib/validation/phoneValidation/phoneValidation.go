package phoneValidation

import (
	"regexp"
	"strings"
)

func IsValidRuPhoneNumber(phone string) bool {
	cleaned := strings.ReplaceAll(phone, " ", "")
	cleaned = strings.ReplaceAll(phone, "-", "")
	cleaned = strings.ReplaceAll(phone, "(", "")
	cleaned = strings.ReplaceAll(phone, ")", "")

	re := regexp.MustCompile(`^(?:\+7|8)\d{10}$`)
	return re.MatchString(cleaned)
}
