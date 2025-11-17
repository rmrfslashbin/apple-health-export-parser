package main

import (
	"testing"
)

func TestSanitizeFilename(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "Simple name",
			input: "Heart Rate",
			want:  "Heart_Rate",
		},
		{
			name:  "Special characters",
			input: "Test/File:Name*With?Chars",
			want:  "TestFileNameWithChars",
		},
		{
			name:  "Multiple spaces",
			input: "Multiple   Spaces   Here",
			want:  "Multiple___Spaces___Here",
		},
		{
			name:  "Alphanumeric with underscores",
			input: "Valid_Name_123",
			want:  "Valid_Name_123",
		},
		{
			name:  "Empty string",
			input: "",
			want:  "",
		},
		{
			name:  "Only special characters",
			input: "!@#$%^&*()",
			want:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := sanitizeFilename(tt.input)
			if got != tt.want {
				t.Errorf("sanitizeFilename() = %v, want %v", got, tt.want)
			}
		})
	}
}
