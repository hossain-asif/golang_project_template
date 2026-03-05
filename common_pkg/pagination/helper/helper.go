package pagination

import "strconv"

func ParseInt(s string, fallback int) int {
	if s == "" {
		return fallback
	}
	v, err := strconv.Atoi(s)
	if err != nil {
		return fallback
	}
	return v
}

func ParseString(s string, fallback string) string {
	if s == "" {
		return fallback
	}
	return s
}
