package prompter

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
)

const maxConfirmLineLength = 1024

// Confirm displays a [y/N] prompt and returns true if the user answers y/yes.
// Input is capped to prevent memory exhaustion from piped input without newlines.
func Confirm(prompt string) (bool, error) {
	fmt.Fprintf(os.Stderr, "%s [y/N] ", sanitize(prompt))

	var buf []byte
	tmp := make([]byte, 1)
	for {
		_, err := os.Stdin.Read(tmp)
		if err != nil {
			if err == io.EOF {
				break
			}
			return false, err
		}
		if tmp[0] == '\n' {
			break
		}
		if len(buf) >= maxConfirmLineLength {
			return false, errors.New("input too long")
		}
		buf = append(buf, tmp[0])
	}

	line := strings.ToLower(strings.TrimSpace(string(buf)))
	return line == "y" || line == "yes", nil
}

// SelectValue is a convenience wrapper around Select that returns the chosen string.
func SelectValue(prompt string, choices []string) (string, error) {
	idx, err := Select(prompt, choices)
	if err != nil {
		return "", err
	}
	return choices[idx], nil
}

func selectNumeric(prompt string, choices []string) (int, error) {
	fmt.Fprintln(os.Stderr, prompt+":")
	for i, c := range choices {
		fmt.Fprintf(os.Stderr, "  [%d] %s\n", i+1, c)
	}
	fmt.Fprintf(os.Stderr, "Select [1-%d]: ", len(choices))

	var n int
	_, err := fmt.Scanln(&n)
	if err != nil {
		return 0, errors.New("invalid selection")
	}

	if n < 1 || n > len(choices) {
		return 0, fmt.Errorf("selection %d out of range", n)
	}

	return n - 1, nil
}
