package main

import (
	"regexp"
	"testing"
)

func TestExtractSteamAppID(t *testing.T) {
	re := regexp.MustCompile(`^(?:https?://)?store\.steampowered\.com/app/(\d+)`)

	tests := []struct {
		name      string
		input     string
		expected  string
		wantMatch bool
	}{
		{
			name:      "Full HTTPS URL with trailing slash",
			input:     "https://store.steampowered.com/app/1030300/Hollow_Knight_Silksong/",
			expected:  "1030300",
			wantMatch: true,
		},
		{
			name:      "HTTP URL without trailing slash",
			input:     "http://store.steampowered.com/app/1030300/Hollow_Knight_Silksong",
			expected:  "1030300",
			wantMatch: true,
		},
		{
			name:      "No protocol, no slug, with trailing slash",
			input:     "store.steampowered.com/app/1030300/",
			expected:  "1030300",
			wantMatch: true,
		},
		{
			name:      "Bare domain and app ID",
			input:     "store.steampowered.com/app/1030300",
			expected:  "1030300",
			wantMatch: true,
		},
		{
			name:      "Invalid URL domain",
			input:     "https://example.com/app/1030300/",
			expected:  "",
			wantMatch: false,
		},
		{
			name:      "Missing App ID",
			input:     "https://store.steampowered.com/app/",
			expected:  "",
			wantMatch: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matches := re.FindStringSubmatch(tt.input)

			if !tt.wantMatch {
				if len(matches) > 0 {
					t.Errorf("expected no match for %q, but got matches: %v", tt.input, matches)
				}
				return
			}

			if len(matches) < 2 {
				t.Fatalf("expected at least 1 capture group for %q, but got none", tt.input)
			}

			got := matches[1]
			if got != tt.expected {
				t.Errorf("ExtractAppID(%q) = %q; want %q", tt.input, got, tt.expected)
			}
		})
	}
}
