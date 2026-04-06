package jdextract

import "regexp"

var (
	emailRe = regexp.MustCompile(`[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}`)
	phoneRe = regexp.MustCompile(`(?:\+?1[-.\s]?)?\(?\d{3}\)?[-.\s]?\d{3}[-.\s]?\d{4}`)
)

// Sanitize replaces email addresses and phone numbers in s with redaction
// placeholders before the text is included in an LLM prompt.
func Sanitize(s string) string {
	s = emailRe.ReplaceAllString(s, "[email redacted]")
	s = phoneRe.ReplaceAllString(s, "[phone redacted]")
	return s
}
