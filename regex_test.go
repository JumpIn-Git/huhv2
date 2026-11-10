package main

import (
	"testing"
)

func TestSteamURLRegex(t *testing.T) {
	re := steamURLRegex

	tests := []struct {
		name      string
		url       string
		wantMatch bool
		wantAppID string
	}{
		{
			name:      "https with game name and trailing slash",
			url:       "https://store.steampowered.com/app/1030300/Hollow_Knight_Silksong/",
			wantMatch: true,
			wantAppID: "1030300",
		},
		{
			name:      "http without trailing slash",
			url:       "http://store.steampowered.com/app/1030300/Hollow_Knight_Silksong",
			wantMatch: true,
			wantAppID: "1030300",
		},
		{
			name:      "no protocol with trailing slash",
			url:       "store.steampowered.com/app/1030300/Hollow_Knight_Silksong/",
			wantMatch: true,
			wantAppID: "1030300",
		},
		{
			name:      "no protocol and no game name",
			url:       "store.steampowered.com/app/1030300",
			wantMatch: true,
			wantAppID: "1030300",
		},
		{
			name:      "https with app ID and trailing slash only",
			url:       "https://store.steampowered.com/app/1030300/",
			wantMatch: true,
			wantAppID: "1030300",
		},
		{
			name:      "invalid domain",
			url:       "https://fake.steampowered.com/app/1030300/Hollow_Knight_Silksong/",
			wantMatch: false,
		},
		{
			name:      "non-numeric app ID",
			url:       "https://store.steampowered.com/app/abc/Hollow_Knight_Silksong/",
			wantMatch: false,
		},
		{
			name:      "extra path segments after game name",
			url:       "https://store.steampowered.com/app/1030300/Hollow_Knight_Silksong/extra",
			wantMatch: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matches := re.FindStringSubmatch(tt.url)

			if !tt.wantMatch {
				if len(matches) > 0 {
					t.Errorf("expected no match for URL %q, but got %v", tt.url, matches)
				}
				return
			}

			if len(matches) == 0 {
				t.Fatalf("expected match for URL %q, but got none", tt.url)
			}

			// Group 1: App ID
			gotAppID := matches[1]
			if gotAppID != tt.wantAppID {
				t.Errorf("App ID = %q, want %q", gotAppID, tt.wantAppID)
			}
		})
	}
}
