package cmd

import (
	"govard/internal/engine"
)

// FixWordPressCompatibility ensures WordPress compatibility (WP-CLI) is set up in the container.
func FixWordPressCompatibility(config engine.Config) error {
	return engine.FixWordPressCompatibility(config)
}
