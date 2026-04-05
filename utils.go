package prompter

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strings"

	"golang.org/x/term"
)

// Confirm asks a yes/no question.
func Confirm(prompt string) (bool, error) {
	reader := bufio.NewReader(os.Stdin)
	fmt.Fprintf(os.Stderr, "%s [y/N] ", prompt)

	line, err := reader.ReadString('\n')
	if err != nil {
		return false, err
	}

	line = strings.ToLower(strings.TrimSpace(line))
	return line == "y" || line == "yes", nil
}

// Select displays an interactive list with cursor navigation.
// Supports: ↑/↓ arrows, j/k (vim), number keys 1-9, Enter to select, q to quit.
func Select(prompt string, choices []string) (int, error) {
	if len(choices) == 0 {
		return 0, errors.New("no choices provided")
	}

	fd := int(os.Stdin.Fd())
	if !term.IsTerminal(fd) {
		// Fallback to numeric input for non-terminal (pipes, CI)
		return selectNumeric(prompt, choices)
	}

	// Save terminal state and switch to raw mode
	oldState, err := term.MakeRaw(fd)
	if err != nil {
		return selectNumeric(prompt, choices)
	}
	defer term.Restore(fd, oldState)

	selected := 0

	// Hide cursor
	fmt.Fprint(os.Stderr, "\033[?25l")
	defer fmt.Fprint(os.Stderr, "\033[?25h") // Show cursor on exit

	// Initial render
	renderSelect(prompt, choices, selected)

	buf := make([]byte, 3)
	for {
		n, err := os.Stdin.Read(buf)
		if err != nil {
			return 0, err
		}

		if n == 1 {
			switch buf[0] {
			case 3: // Ctrl+C
				return 0, errors.New("selection cancelled")
			case 13, 10: // Enter
				fmt.Fprintln(os.Stderr) // Move to new line
				return selected, nil
			case 'q', 'Q':
				return 0, errors.New("selection cancelled")
			case 'j', 'J':
				if selected < len(choices)-1 {
					selected++
					renderSelect(prompt, choices, selected)
				}
			case 'k', 'K':
				if selected > 0 {
					selected--
					renderSelect(prompt, choices, selected)
				}
			case '1', '2', '3', '4', '5', '6', '7', '8', '9':
				idx := int(buf[0]-'0') - 1
				if idx >= 0 && idx < len(choices) {
					fmt.Fprintln(os.Stderr)
					return idx, nil
				}
			}
		} else if n == 3 && buf[0] == 27 && buf[1] == 91 {
			// Arrow keys: ESC [ A (up), ESC [ B (down)
			switch buf[2] {
			case 65: // Up
				if selected > 0 {
					selected--
					renderSelect(prompt, choices, selected)
				}
			case 66: // Down
				if selected < len(choices)-1 {
					selected++
					renderSelect(prompt, choices, selected)
				}
			}
		}
	}
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

func renderSelect(prompt string, choices []string, selected int) {
	// Clear screen from cursor to end, then move cursor up to redraw
	// Move cursor up len(choices)+1 lines to overwrite
	if selected > 0 || true { // Always redraw full list
		// Move up to overwrite previous render
		fmt.Fprintf(os.Stderr, "\033[%dA", len(choices)+2) // +2 for prompt line and empty line
	}

	fmt.Fprintf(os.Stderr, "%s:\n", prompt)
	for i, choice := range choices {
		if i == selected {
			fmt.Fprintf(os.Stderr, "  \033[7m> %s\033[0m\n", choice) // Reverse video for selected
		} else {
			fmt.Fprintf(os.Stderr, "    %s\n", choice)
		}
	}
	fmt.Fprintln(os.Stderr, "  [←↑↓→/j/k/num] Move • [Enter] Select • [q] Quit")
}
