package testutil

import "regexp"

var ansiRegex = regexp.MustCompile(`\x1b\[[0-9;]*m`)

func StripAnsi(s string) string {
	return ansiRegex.ReplaceAllString(s, "")
}
