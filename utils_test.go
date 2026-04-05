package prompter

import (
	"testing"
)

func TestSelect_EmptyChoices(t *testing.T) {
	_, err := Select("test", []string{})
	if err == nil || err.Error() != "no choices provided" {
		t.Fatalf("expected error for empty choices, got: %v", err)
	}
}

func TestSelectValue(t *testing.T) {
	choices := []string{"apple", "banana", "cherry"}
	_, err := SelectValue("Pick fruit", choices)
	if err == nil {
		return
	}
}
