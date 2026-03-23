package tests

import (
	"os"
	"path/filepath"
	"testing"

	"govard/internal/engine"
)

func TestDetectWebRoot(t *testing.T) {
	tests := []struct {
		name       string
		framework  string
		setupFiles []string
		expected   string
	}{
		{
			name:       "Symfony with public",
			framework:  "symfony",
			setupFiles: []string{"public/index.php"},
			expected:   "/public",
		},
		{
			name:       "Laravel with public",
			framework:  "laravel",
			setupFiles: []string{"public/index.php"},
			expected:   "/public",
		},
		{
			name:       "Magento 2 with pub",
			framework:  "magento2",
			setupFiles: []string{"pub/index.php"},
			expected:   "/pub",
		},
		{
			name:       "Drupal with web",
			framework:  "drupal",
			setupFiles: []string{"web/index.php"},
			expected:   "/web",
		},
		{
			name:       "WordPress with wordpress",
			framework:  "wordpress",
			setupFiles: []string{"wordpress/index.php"},
			expected:   "/wordpress",
		},
		{
			name:       "CakePHP with webroot",
			framework:  "cakephp",
			setupFiles: []string{"webroot/index.php"},
			expected:   "/webroot",
		},
		{
			name:       "CakePHP with public",
			framework:  "cakephp",
			setupFiles: []string{"public/index.php"},
			expected:   "/public",
		},
		{
			name:       "Unknown framework",
			framework:  "unknown",
			setupFiles: []string{"public/index.php"},
			expected:   "",
		},
		{
			name:       "Symfony without public folder",
			framework:  "symfony",
			setupFiles: []string{},
			expected:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			for _, file := range tt.setupFiles {
				path := filepath.Join(dir, file)
				err := os.MkdirAll(filepath.Dir(path), 0755)
				if err != nil {
					t.Fatalf("failed to setup test directory: %v", err)
				}
				err = os.WriteFile(path, []byte("test"), 0644)
				if err != nil {
					t.Fatalf("failed to setup test file: %v", err)
				}
			}

			result := engine.DetectWebRoot(dir, tt.framework)
			if result != tt.expected {
				t.Errorf("DetectWebRoot() = %v, want %v", result, tt.expected)
			}
		})
	}
}
