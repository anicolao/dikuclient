package tui

import (
	"strings"
	"testing"
)

// Test that auto-login prompt detection works correctly
func TestAutoLoginPromptDetection(t *testing.T) {
	tests := []struct {
		name           string
		prompt         string
		shouldMatch    bool
		promptType     string // "username" or "password"
	}{
		// Username prompts
		{"Name prompt", "What is your name: ", true, "username"},
		{"Login prompt", "Login: ", true, "username"},
		{"Account prompt", "Account: ", true, "username"},
		{"Character prompt", "Character name: ", true, "username"},
		{"Uppercase name", "NAME: ", true, "username"},
		
		// Password prompts
		{"Password prompt", "Password: ", true, "password"},
		{"Pass prompt", "Pass: ", true, "password"},
		{"Uppercase password", "PASSWORD: ", true, "password"},
		
		// Non-matching prompts
		{"Command prompt", "> ", false, ""},
		{"Generic prompt", "What would you like to do? ", false, ""},
		{"Exit prompt", "Are you sure you want to quit? ", false, ""},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lastLine := strings.ToLower(strings.TrimSpace(tt.prompt))
			
			isUsernamePrompt := strings.Contains(lastLine, "name:") || 
				strings.Contains(lastLine, "login:") || 
				strings.Contains(lastLine, "account:") ||
				strings.Contains(lastLine, "character:")
			
			isPasswordPrompt := strings.Contains(lastLine, "password:") || 
				strings.Contains(lastLine, "pass:")
			
			if tt.promptType == "username" {
				if !isUsernamePrompt && tt.shouldMatch {
					t.Errorf("Expected to match username prompt, but didn't")
				}
				if isUsernamePrompt && !tt.shouldMatch {
					t.Errorf("Matched username prompt when it shouldn't")
				}
			} else if tt.promptType == "password" {
				if !isPasswordPrompt && tt.shouldMatch {
					t.Errorf("Expected to match password prompt, but didn't")
				}
				if isPasswordPrompt && !tt.shouldMatch {
					t.Errorf("Matched password prompt when it shouldn't")
				}
			} else {
				if (isUsernamePrompt || isPasswordPrompt) && !tt.shouldMatch {
					t.Errorf("Matched a prompt when it shouldn't")
				}
			}
		})
	}
}
