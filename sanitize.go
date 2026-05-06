package md2img

import "strings"

// sanitize replaces Unicode characters that core PDF fonts (Helvetica, Courier)
// cannot render with ASCII equivalents.
func sanitize(s string) string {
	replacer := strings.NewReplacer(
		// Dashes
		"\u2014", "-", // em dash
		"\u2013", "-", // en dash
		// Punctuation
		"\u2026", "...",  // ellipsis
		"\u2018", "'",    // left single quote
		"\u2019", "'",    // right single quote
		"\u201c", "\"",   // left double quote
		"\u201d", "\"",   // right double quote
		// Arrows
		"\u2192", "->",   // right arrow
		"\u2190", "<-",   // left arrow
		"\u21d2", "=>",   // right double arrow
		// Math
		"\u2260", "!=",   // not equal
		"\u2264", "<=",   // less than or equal
		"\u2265", ">=",   // greater than or equal
		"\u00d7", "x",    // multiplication
		"\u00f7", "/",    // division
		// Bullets / symbols
		"\u2022", "*",    // bullet
		"\u2605", "*",    // star
		// Check marks
		"\u2713", "[OK]", // check
		"\u2717", "[X]",  // cross
		"\u274c", "[X]",  // cross mark
		"\u2705", "[OK]", // check mark
		// Flag emoji (regional indicators)
		"\U0001f1e9\U0001f1ea", "[DE]", // German flag
		"\U0001f1eb\U0001f1f7", "[FR]", // French flag
		"\U0001f1ec\U0001f1e7", "[UK]", // UK flag
		"\U0001f1fa\U0001f1f8", "[US]", // US flag
	)
	return replacer.Replace(s)
}
