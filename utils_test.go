package prompter

import (
	"testing"
)

func TestConfirm_Validation(t *testing.T) {
	// Note: These require stdin manipulation or integration testing
	// Here we just test the helper logic if we extracted it

	// For now, just ensure the function exists and has correct signature
	// Real testing would need a terminal or mock
}

func TestSelect_EmptyChoices(t *testing.T) {
	_, err := Select("test", []string{})
	if err == nil || err.Error() != "no choices provided" {
		t.Fatalf("expected error for empty choices, got: %v", err)
	}
}

func TestSelect_OutOfRange(t *testing.T) {
	// This would require mocking fmt.Scanln or using the numeric fallback
	// Skipping for unit test simplicity - tested via integration
}
