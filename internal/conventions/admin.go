package conventions

import "strings"

func AdminEmailForDomain(domain string) string {
	trimmedDomain := strings.TrimSpace(domain)
	if trimmedDomain == "" {
		return DefaultAdminEmail
	}
	return DefaultAdminUser + "@" + trimmedDomain
}
