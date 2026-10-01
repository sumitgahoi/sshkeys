package main

import "testing"

func TestValidateGitHubUser(t *testing.T) {
	tests := []struct {
		user  string
		valid bool
	}{
		{"sumitgahoi", true},
		{"sumit-gahoi", true},
		{"Sumit123", true},
		{"", false},
		{"-sumit", false},
		{"sumit-", false},
		{"sumit--gahoi", false},
		{"sumit/gahoi", false},
		{"sumit?x", false},
		{"sumit#x", false},
		{"sumit gahoi", false},
	}

	for _, tt := range tests {
		t.Run(tt.user, func(t *testing.T) {
			err := validateGitHubUser(tt.user)

			if tt.valid && err != nil {
				t.Fatalf("validateGitHubUser(%q) error = %v", tt.user, err)
			}

			if !tt.valid && err == nil {
				t.Fatalf("validateGitHubUser(%q) error = nil, want error", tt.user)
			}
		})
	}
}
