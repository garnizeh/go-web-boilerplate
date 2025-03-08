package validator_test

import (
	"testing"

	"github.com/garnizeh/go-web-boilerplate/pkg/validator"
)

func TestSanitizeString(t *testing.T) {
	tests := []struct {
		name string
		str  string
		want string
	}{
		{
			name: "empty string",
			str:  "",
			want: "",
		},
		{
			name: "only spaces",
			str:  "   ",
			want: "",
		},
		{
			name: "script injection",
			str:  "<script>alert('XSS')</script)",
			want: "",
		},
		{
			name: "word with spaces",
			str:  " word ",
			want: "word",
		},
		{
			name: "word without spaces",
			str:  "word",
			want: "word",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := validator.SanitizeString(tt.str)
			if got != tt.want {
				t.Errorf("%q got %q, want %q", tt.name, got, tt.want)
			}
		})
	}
}

func TestIsValidEmail(t *testing.T) {
	tests := []struct {
		name  string
		email string
		want  bool
	}{
		{
			name:  "empty string",
			email: "",
			want:  false,
		},
		{
			name:  "invalid format 1",
			email: "a",
			want:  false,
		},
		{
			name:  "invalid format 2",
			email: "a@",
			want:  false,
		},
		{
			name:  "invalid format 3",
			email: "a@a",
			want:  false,
		},
		{
			name:  "invalid format 4",
			email: "a@.com",
			want:  false,
		},
		{
			name:  "invalid format 5",
			email: "a.a",
			want:  false,
		},
		{
			name:  "invalid format 6",
			email: "a.com",
			want:  false,
		},
		{
			name:  "invalid format 7",
			email: "@a.com",
			want:  false,
		},
		{
			name:  "valid",
			email: "a@a.com",
			want:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := validator.IsValidEmail(tt.email)
			if got != tt.want {
				t.Errorf("%q got %v, want %v", tt.name, got, tt.want)
			}
		})
	}
}
