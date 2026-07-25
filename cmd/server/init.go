package main

import (
	"fmt"
	"os"
	"regexp"
	"strings"
	"unicode/utf8"

	"consul-journey/internal"
	"consul-journey/internal/utils"
)

func init() {
	visibleStringLength := func(s string) int {
		return utf8.RuneCountInString(regexp.MustCompile(`\x1b\[[0-?]*[ -\/]*[@-~]`).ReplaceAllString(s, ""))
	}
	lines := []string{
		"",
		"\x1b[1;92m" + internal.AppName() + " Server\x1b[0m",
		"\x1b[34mVersion:\x1b[0m  " + internal.Version(),
		"\x1b[34mRevision:\x1b[0m " + internal.Revision(),
		"",
	}
	maxLen := utils.Max(visibleStringLength, lines...)
	lines[0] = "\x1b[34m" + strings.Repeat("-", maxLen) + "\x1b[0m"
	lines[len(lines)-1] = "\x1b[34m" + strings.Repeat("-", maxLen) + "\x1b[0m"
	for _, line := range lines {
		fmt.Fprintln(os.Stderr, line)
	}
}
