package mongo

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetDownloadURL(t *testing.T) {
	tests := []struct {
		name        string
		osType      string
		expectedExt string
		expectEmpty bool
	}{
		{"Windows", "windows", ".zip", false},
		{"Darwin", "darwin", ".zip", false},
		{"Linux fallback", "linux", "", true},
		{"Unknown", "unknownOS", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url, ext := getDownloadURL(tt.osType)
			if tt.expectEmpty {
				assert.Empty(t, url)
				assert.Empty(t, ext)
			} else {
				assert.NotEmpty(t, url)
				assert.Equal(t, tt.expectedExt, ext)
			}
		})
	}
}
