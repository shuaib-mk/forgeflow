package validation

import (
	"net/mail"
	"regexp"
	"strings"
)

var slugPattern = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)

func Required(value string, max int) bool {
	value = strings.TrimSpace(value)
	return value != "" && len(value) <= max
}

func Email(value string) bool {
	address, err := mail.ParseAddress(strings.TrimSpace(value))
	return err == nil && address.Address == strings.TrimSpace(value)
}

func Slug(value string) bool { return len(value) <= 63 && slugPattern.MatchString(value) }

