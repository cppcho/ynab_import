package main

import (
	"os"
	"testing"
)

func TestFlipSign(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"positive number", "1000", "-1000"},
		{"negative number", "-500", "500"},
		{"zero", "0", "0"},
		{"number with comma", "1,000", "-1000"},
		{"number with multiple commas", "1,000,000", "-1000000"},
		{"invalid input", "abc", "abc"}, // Returns original on error
		{"empty string", "", ""},        // Returns empty on error
		{"large number", "999999999", "-999999999"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := flipSign(tt.input)
			if got != tt.expected {
				t.Errorf("flipSign(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestConvertDate(t *testing.T) {
	tests := []struct {
		name       string
		fromLayout string
		toLayout   string
		value      string
		expected   string
		wantErr    bool
	}{
		{
			name:       "standard format with single digits",
			fromLayout: "2006/1/2",
			toLayout:   "2006-01-02",
			value:      "2024/1/15",
			expected:   "2024-01-15",
			wantErr:    false,
		},
		{
			name:       "compact format",
			fromLayout: "20060102",
			toLayout:   "2006-01-02",
			value:      "20240115",
			expected:   "2024-01-15",
			wantErr:    false,
		},
		{
			name:       "Japanese format",
			fromLayout: "2006年01月02日",
			toLayout:   "2006-01-02",
			value:      "2024年01月15日",
			expected:   "2024-01-15",
			wantErr:    false,
		},
		{
			name:       "format with padded digits",
			fromLayout: "2006/01/02",
			toLayout:   "2006-01-02",
			value:      "2024/01/15",
			expected:   "2024-01-15",
			wantErr:    false,
		},
		{
			name:       "leap year date",
			fromLayout: "2006/1/2",
			toLayout:   "2006-01-02",
			value:      "2024/2/29",
			expected:   "2024-02-29",
			wantErr:    false,
		},
		{
			name:       "invalid date format",
			fromLayout: "2006-01-02",
			toLayout:   "2006-01-02",
			value:      "invalid",
			expected:   "",
			wantErr:    true,
		},
		{
			name:       "mismatched layout",
			fromLayout: "2006/01/02",
			toLayout:   "2006-01-02",
			value:      "20240115", // compact format doesn't match
			expected:   "",
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := convertDate(tt.fromLayout, tt.toLayout, tt.value)
			if tt.wantErr {
				if err == nil {
					t.Errorf("convertDate(%q, %q, %q) expected error, got nil", tt.fromLayout, tt.toLayout, tt.value)
				}
			} else {
				if err != nil {
					t.Errorf("convertDate(%q, %q, %q) unexpected error: %v", tt.fromLayout, tt.toLayout, tt.value, err)
				}
				if got != tt.expected {
					t.Errorf("convertDate(%q, %q, %q) = %q, want %q", tt.fromLayout, tt.toLayout, tt.value, got, tt.expected)
				}
			}
		})
	}
}

func TestExpandHomeDir(t *testing.T) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("Failed to get home directory: %v", err)
	}

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"path with tilde", "~/foo", homeDir + "/foo"},
		{"just tilde", "~", homeDir},
		{"no tilde", "/absolute/path", "/absolute/path"},
		{"empty string", "", ""},
		{"tilde in middle (unchanged)", "/foo/~/bar", "/foo/~/bar"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := expandHomeDir(tt.input)
			if got != tt.expected {
				t.Errorf("expandHomeDir(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestGetEnvOrDefault(t *testing.T) {
	tests := []struct {
		name         string
		key          string
		defaultValue string
		envValue     string
		setEnv       bool
		expected     string
	}{
		{
			name:         "env var exists",
			key:          "TEST_VAR_EXISTS",
			defaultValue: "default",
			envValue:     "custom",
			setEnv:       true,
			expected:     "custom",
		},
		{
			name:         "env var doesn't exist",
			key:          "TEST_VAR_NOT_EXISTS",
			defaultValue: "default",
			envValue:     "",
			setEnv:       false,
			expected:     "default",
		},
		{
			name:         "env var is empty string",
			key:          "TEST_VAR_EMPTY",
			defaultValue: "default",
			envValue:     "",
			setEnv:       true,
			expected:     "default",
		},
		{
			name:         "both env and default are non-empty",
			key:          "TEST_VAR_BOTH",
			defaultValue: "fallback",
			envValue:     "environment",
			setEnv:       true,
			expected:     "environment",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clean up before test
			os.Unsetenv(tt.key)

			// Set environment variable if needed
			if tt.setEnv {
				os.Setenv(tt.key, tt.envValue)
				defer os.Unsetenv(tt.key) // Clean up after test
			}

			got := getEnvOrDefault(tt.key, tt.defaultValue)
			if got != tt.expected {
				t.Errorf("getEnvOrDefault(%q, %q) = %q, want %q", tt.key, tt.defaultValue, got, tt.expected)
			}
		})
	}
}
