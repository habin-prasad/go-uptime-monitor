package checker

import "testing"

func TestValidateURL(t *testing.T) {
	tests := []struct {
		name    string
		url     string
		wantErr bool
	}{
		{"HTTP URL", "http://example.com", false},
		{"HTTPS URL", "https://example.com/health", false},
		{"Missing scheme", "example.com", true},
		{"Unsupported scheme", "ftp://example.com", true},
		{"Missing hostname", "https:///health", true},
		{"Empty URL", "", true},
		{"Whitespace", " https://example.com ", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateURL(tt.url)

			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateURL(%q) error = %v, wantErr %t",
					tt.url, err, tt.wantErr)
			}
		})
	}
}
