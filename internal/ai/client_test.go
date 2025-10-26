package ai

import (
	"testing"
)

// TestNewClient tests creating a new AI client
func TestNewClient(t *testing.T) {
	tests := []struct {
		name     string
		aiType   string
		url      string
		apiKey   string
		expected string
	}{
		{
			name:     "OpenAI client",
			aiType:   "openai",
			url:      "https://api.openai.com/v1/chat/completions",
			apiKey:   "test-key",
			expected: "openai",
		},
		{
			name:     "Ollama client",
			aiType:   "Ollama",
			url:      "http://localhost:11434/api/generate",
			apiKey:   "",
			expected: "ollama",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := NewClient(tt.aiType, tt.url, tt.apiKey)
			if client == nil {
				t.Fatal("Expected client, got nil")
			}
			if client.Type != tt.expected {
				t.Errorf("Expected type %s, got %s", tt.expected, client.Type)
			}
			if client.URL != tt.url {
				t.Errorf("Expected URL %s, got %s", tt.url, client.URL)
			}
			if client.APIKey != tt.apiKey {
				t.Errorf("Expected API key %s, got %s", tt.apiKey, client.APIKey)
			}
		})
	}
}

// TestUnsupportedType tests that unsupported AI types return an error
func TestUnsupportedType(t *testing.T) {
	client := NewClient("unsupported", "http://example.com", "")
	_, err := client.GenerateResponse("test prompt")
	
	if err == nil {
		t.Error("Expected error for unsupported type, got nil")
	}
}
