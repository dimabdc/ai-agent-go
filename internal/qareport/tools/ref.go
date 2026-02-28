package tools

import (
	"errors"
	"strings"
)

func parseRef(ref string) (string, string, string, error) {
	parts := strings.Split(ref, ":")
	if len(parts) != 3 {
		return "", "", "", errors.New("invalid ref")
	}

	return parts[0], parts[1], parts[2], nil
}
