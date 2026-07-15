package engine

import "regexp"

var xdebugVersionPattern = regexp.MustCompile(`^[A-Za-z0-9]+(\.[A-Za-z0-9]+)*$`)

// ValidateXdebugVersion reports whether version is a well-formed PECL Xdebug
// version specifier (e.g. "3.5.3"). The value is embedded into a Docker image
// tag and a build-arg passed to a shell command, so it is restricted to
// alphanumerics and dots to rule out anything that could break either.
func ValidateXdebugVersion(version string) bool {
	return xdebugVersionPattern.MatchString(version)
}
