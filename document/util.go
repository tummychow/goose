package document

import (
	"regexp"
)

var nameRegex = regexp.MustCompile(`^(?:/[\x20-\x2E\x30-\x7E]+)+$`)
var dotsRegex = regexp.MustCompile(`/\.\.?(/|$)`)

// ValidateName determines if a given string is a valid Document Name.
func ValidateName(name string) bool {
	return nameRegex.MatchString(name) && !dotsRegex.MatchString(name)
}
