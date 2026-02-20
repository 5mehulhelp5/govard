package engine

import "strings"

// CompareNumericDotVersions compares two dot-separated numeric versions.
// It returns (comparison, true) when both values are comparable:
// comparison is 1 when left > right, -1 when left < right, and 0 when equal.
func CompareNumericDotVersions(left, right string) (int, bool) {
	leftParts, ok := parseNumericDotVersion(left)
	if !ok {
		return 0, false
	}
	rightParts, ok := parseNumericDotVersion(right)
	if !ok {
		return 0, false
	}

	maxLen := len(leftParts)
	if len(rightParts) > maxLen {
		maxLen = len(rightParts)
	}
	for i := 0; i < maxLen; i++ {
		lv := 0
		if i < len(leftParts) {
			lv = leftParts[i]
		}
		rv := 0
		if i < len(rightParts) {
			rv = rightParts[i]
		}
		if lv > rv {
			return 1, true
		}
		if lv < rv {
			return -1, true
		}
	}
	return 0, true
}

func isNumericDotVersionAtLeast(raw string, minimum string) bool {
	comparison, comparable := CompareNumericDotVersions(raw, minimum)
	return comparable && comparison >= 0
}

func parseNumericDotVersion(raw string) ([]int, bool) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, false
	}
	segments := strings.Split(raw, ".")
	parts := make([]int, 0, len(segments))
	for _, segment := range segments {
		segment = strings.TrimSpace(segment)
		if segment == "" {
			return nil, false
		}
		value, ok := parseLeadingDigits(segment)
		if !ok {
			return nil, false
		}
		parts = append(parts, value)
	}
	return parts, true
}

func parseLeadingDigits(segment string) (int, bool) {
	value := 0
	seenDigit := false
	for _, r := range segment {
		if r >= '0' && r <= '9' {
			seenDigit = true
			value = value*10 + int(r-'0')
			continue
		}
		if !seenDigit {
			return 0, false
		}
		return value, true
	}
	if !seenDigit {
		return 0, false
	}
	return value, true
}
