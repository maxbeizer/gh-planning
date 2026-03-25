package config

import "strings"

// ParseSectionFilter parses a filter string like "status:In Progress" into
// field/value pairs. If no colon is present, field is empty and value is the
// entire input.
func ParseSectionFilter(filter string) (field, value string) {
	parts := strings.SplitN(filter, ":", 2)
	if len(parts) == 2 {
		return strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1])
	}
	return "", filter
}
