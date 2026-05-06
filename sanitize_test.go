package md2img

import "testing"

func TestSanitize(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"plain ascii", "hello world", "hello world"},
		{"em dash", "em\u2014dash", "em-dash"},
		{"en dash", "en\u2013dash", "en-dash"},
		{"ellipsis", "ellipsis\u2026", "ellipsis..."},
		{"right arrow", "right\u2192arrow", "right->arrow"},
		{"left arrow", "left\u2190arrow", "left<-arrow"},
		{"double arrow", "a\u21d2b", "a=>b"},
		{"not equal", "not\u2260equal", "not!=equal"},
		{"lte", "a\u2264b", "a<=b"},
		{"gte", "a\u2265b", "a>=b"},
		{"multiply", "3\u00d74", "3x4"},
		{"divide", "6\u00f72", "6/2"},
		{"bullet", "bullet\u2022item", "bullet*item"},
		{"star", "star\u2605", "star*"},
		{"check", "check\u2713", "check[OK]"},
		{"cross", "cross\u2717", "cross[X]"},
		{"cross mark", "\u274c", "[X]"},
		{"check mark", "\u2705", "[OK]"},
		{"left double quote", "\u201cquoted\u201d", "\"quoted\""},
		{"left single quote", "\u2018single\u2019", "'single'"},
		{"german flag", "\U0001f1e9\U0001f1ea", "[DE]"},
		{"french flag", "\U0001f1eb\U0001f1f7", "[FR]"},
		{"uk flag", "\U0001f1ec\U0001f1e7", "[UK]"},
		{"us flag", "\U0001f1fa\U0001f1f8", "[US]"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := sanitize(tt.input)
			if got != tt.expected {
				t.Errorf("sanitize(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestSanitizePassthrough(t *testing.T) {
	input := "Hello World 123 !@#$%^&*()"
	got := sanitize(input)
	if got != input {
		t.Errorf("sanitize modified ASCII: got %q, want %q", got, input)
	}
}

func TestSanitizeMixed(t *testing.T) {
	input := "The \u201cbest\u201d talk \u2014 \u2713 recommended"
	got := sanitize(input)

	if expected := `"best"`; !contains(got, expected) {
		t.Errorf("quotes not sanitized: got %q, want substring %q", got, expected)
	}
	if !contains(got, "talk -") {
		t.Errorf("em dash not sanitized: got %q", got)
	}
	if !contains(got, "[OK]") {
		t.Errorf("check mark not sanitized: got %q", got)
	}
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && searchSubstring(s, sub)
}

func searchSubstring(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
