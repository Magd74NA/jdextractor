package jdextract

import "testing"

func TestSanitize(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		// Emails
		{
			name:  "plain email",
			input: "Contact me at john.doe@example.com for details.",
			want:  "Contact me at [email redacted] for details.",
		},
		{
			name:  "subdomain email",
			input: "Send to alice@mail.company.org please.",
			want:  "Send to [email redacted] please.",
		},
		{
			name:  "plus-addressed email",
			input: "user+tag@example.com",
			want:  "[email redacted]",
		},
		{
			name:  "multiple emails",
			input: "From: a@b.com To: c@d.org",
			want:  "From: [email redacted] To: [email redacted]",
		},

		// Phones
		{
			name:  "dashes",
			input: "Call 555-123-4567 anytime.",
			want:  "Call [phone redacted] anytime.",
		},
		{
			name:  "parens and space",
			input: "Phone: (555) 123-4567",
			want:  "Phone: [phone redacted]",
		},
		{
			name:  "dots",
			input: "555.123.4567",
			want:  "[phone redacted]",
		},
		{
			name:  "country code with space",
			input: "+1 555 123 4567",
			want:  "[phone redacted]",
		},
		{
			name:  "country code with dashes",
			input: "+1-555-123-4567",
			want:  "[phone redacted]",
		},

		// Mixed
		{
			name:  "email and phone on same line",
			input: "jane@example.com | (555) 987-6543",
			want:  "[email redacted] | [phone redacted]",
		},
		{
			name:  "multiline resume header",
			input: "John Smith\njohn@smith.io\n555-000-1234\nNew York, NY",
			want:  "John Smith\n[email redacted]\n[phone redacted]\nNew York, NY",
		},

		// No-op cases
		{
			name:  "clean string unchanged",
			input: "Senior Software Engineer with 5 years experience.",
			want:  "Senior Software Engineer with 5 years experience.",
		},
		{
			name:  "empty string",
			input: "",
			want:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Sanitize(tt.input)
			if got != tt.want {
				t.Errorf("Sanitize(%q)\n got: %q\nwant: %q", tt.input, got, tt.want)
			}
		})
	}
}
