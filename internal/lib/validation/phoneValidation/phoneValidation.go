package phoneValidation

import (
	"regexp"
	"strings"
)

// Validating russian phone numbers
func IsValidRuPhoneNumber(phone string) bool {
	cleaned := strings.NewReplacer(" ", "", "-", "", "(", "", ")", "").Replace(phone)
	re := regexp.MustCompile(`^(?:\+7|8)\d{10}$`)
	return re.MatchString(cleaned)
}
