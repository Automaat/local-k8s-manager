package namegen

import (
	"strings"
	"testing"
)

func TestGenerate(t *testing.T) {
	name := Generate()

	if name == "" {
		t.Error("expected non-empty name")
	}

	parts := strings.Split(name, "-")
	if len(parts) != 2 {
		t.Errorf("expected name in format 'adjective-noun', got %s", name)
	}

	adjective := parts[0]
	noun := parts[1]

	if adjective == "" {
		t.Error("expected non-empty adjective")
	}

	if noun == "" {
		t.Error("expected non-empty noun")
	}

	// Verify adjective is from the list
	found := false
	for _, a := range adjectives {
		if a == adjective {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("adjective %s not found in adjectives list", adjective)
	}

	// Verify noun is from the list
	found = false
	for _, n := range nouns {
		if n == noun {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("noun %s not found in nouns list", noun)
	}
}

func TestGenerateUniqueness(t *testing.T) {
	// Generate multiple names and check they can be different
	names := make(map[string]bool)
	iterations := 100

	for i := 0; i < iterations; i++ {
		name := Generate()
		names[name] = true
	}

	// With the current lists, we should get some variety
	// adjectives: 93, nouns: 237, total combinations: 22,041
	// So 100 iterations should give us more than 1 unique name
	if len(names) < 2 {
		t.Errorf("expected multiple unique names in %d iterations, got %d unique names", iterations, len(names))
	}
}

func TestGenerateFormat(t *testing.T) {
	// Generate several names and verify they all follow the format
	for i := 0; i < 10; i++ {
		name := Generate()

		// Should contain exactly one hyphen
		hyphens := strings.Count(name, "-")
		if hyphens != 1 {
			t.Errorf("expected exactly 1 hyphen in name, got %d in %s", hyphens, name)
		}

		// Should not start or end with hyphen
		if strings.HasPrefix(name, "-") || strings.HasSuffix(name, "-") {
			t.Errorf("name should not start or end with hyphen: %s", name)
		}

		// Should be lowercase
		if name != strings.ToLower(name) {
			t.Errorf("expected lowercase name, got %s", name)
		}
	}
}
