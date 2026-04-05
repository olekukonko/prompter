package prompter

import "regexp"

var ansiEscRe = regexp.MustCompile(`\x1b[\x5b\x5d][\x20-\x7e]*[\x40-\x7e]?|\x1b[^\x5b\x5d]`)

// sanitize strips ANSI/OSC escape sequences to prevent terminal injection.
func sanitize(s string) string { return ansiEscRe.ReplaceAllString(s, "") }
