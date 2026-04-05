//go:build !windows

package prompter

import (
	"errors"
	"fmt"
	"os"
	"regexp"

	"golang.org/x/term"
)

var ansiEscRe = regexp.MustCompile(`\x1b[\x5b\x5d][\x20-\x7e]*[\x40-\x7e]?|\x1b[^\x5b\x5d]`)

// sanitize strips ANSI/OSC escape sequences to prevent terminal injection.
func sanitize(s string) string { return ansiEscRe.ReplaceAllString(s, "") }

// Select presents an interactive arrow-key menu and returns the chosen index.
// Falls back to numeric input when stdin is not a terminal.
func Select(prompt string, choices []string) (int, error) {
	if len(choices) == 0 {
		return 0, errors.New("no choices provided")
	}

	prompt = sanitize(prompt)
	sanitized := make([]string, len(choices))
	for i, c := range choices {
		sanitized[i] = sanitize(c)
	}

	fd := int(os.Stdin.Fd())
	if !term.IsTerminal(fd) {
		return selectNumeric(prompt, sanitized)
	}

	oldState, err := term.MakeRaw(fd)
	if err != nil {
		return selectNumeric(prompt, sanitized)
	}

	// Restore terminal first so it is always recovered, even on panic.
	defer term.Restore(fd, oldState)

	selected := 0
	fmt.Fprint(os.Stderr, "\033[?25l")
	defer func() {
		fmt.Fprint(os.Stderr, "\033[?25h")
		if r := recover(); r != nil {
			panic(r)
		}
	}()

	renderSelect(prompt, sanitized, selected)

	buf := make([]byte, 3)
	for {
		n, err := os.Stdin.Read(buf)
		if err != nil {
			return 0, err
		}

		if n == 1 {
			switch buf[0] {
			case 3:
				return 0, ErrCancelled
			case 13, 10:
				fmt.Fprintln(os.Stderr)
				return selected, nil
			case 'q', 'Q':
				return 0, ErrCancelled
			case 'j', 'J':
				if selected < len(sanitized)-1 {
					selected++
					renderSelect(prompt, sanitized, selected)
				}
			case 'k', 'K':
				if selected > 0 {
					selected--
					renderSelect(prompt, sanitized, selected)
				}
			case '1', '2', '3', '4', '5', '6', '7', '8', '9':
				idx := int(buf[0]-'0') - 1
				if idx >= 0 && idx < len(sanitized) {
					fmt.Fprintln(os.Stderr)
					return idx, nil
				}
			}
		} else if n == 3 && buf[0] == 27 && buf[1] == 91 {
			switch buf[2] {
			case 65:
				if selected > 0 {
					selected--
					renderSelect(prompt, sanitized, selected)
				}
			case 66:
				if selected < len(sanitized)-1 {
					selected++
					renderSelect(prompt, sanitized, selected)
				}
			}
		}
	}
}

func renderSelect(prompt string, choices []string, selected int) {
	fmt.Fprintf(os.Stderr, "\033[%dA", len(choices)+2)
	fmt.Fprintf(os.Stderr, "%s:\n", prompt)
	for i, choice := range choices {
		if i == selected {
			fmt.Fprintf(os.Stderr, "  \033[7m> %s\033[0m\n", choice)
		} else {
			fmt.Fprintf(os.Stderr, "    %s\n", choice)
		}
	}
	fmt.Fprintln(os.Stderr, "  [←↑↓→/j/k/num] Move • [Enter] Select • [q] Quit")
}
