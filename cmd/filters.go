package cmd

import (
	"context"
	"fmt"
	"strings"

	"github.com/maxbeizer/gh-planning/internal/github"
)

type fieldFilter struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

func resolveFieldFilters(ctx context.Context, defaults map[string]string, values []string) ([]fieldFilter, error) {
	filters, needsUser, err := parseFieldFilters(defaults, values)
	if err != nil {
		return nil, err
	}
	if !needsUser {
		return filters, nil
	}
	user, err := github.CurrentUser(ctx)
	if err != nil {
		return nil, err
	}
	return substituteMe(filters, user), nil
}

func resolveFieldFiltersWithUser(defaults map[string]string, values []string, currentUser string) ([]fieldFilter, error) {
	filters, _, err := parseFieldFilters(defaults, values)
	if err != nil {
		return nil, err
	}
	return substituteMe(filters, currentUser), nil
}

func parseFieldFilters(defaults map[string]string, values []string) ([]fieldFilter, bool, error) {
	result := make([]fieldFilter, 0, len(defaults)+len(values))
	for name, value := range defaults {
		name = strings.TrimSpace(name)
		value = strings.TrimSpace(value)
		if name == "" || value == "" {
			continue
		}
		result = append(result, fieldFilter{Name: name, Value: value})
	}
	for _, raw := range values {
		parts := strings.SplitN(raw, "=", 2)
		if len(parts) != 2 || strings.TrimSpace(parts[0]) == "" {
			return nil, false, fmt.Errorf("invalid field filter %q (expected Field=Value)", raw)
		}
		result = append(result, fieldFilter{
			Name:  strings.TrimSpace(parts[0]),
			Value: strings.TrimSpace(parts[1]),
		})
	}
	needsUser := false
	for _, filter := range result {
		if strings.EqualFold(filter.Value, "me") {
			needsUser = true
			break
		}
	}
	return result, needsUser, nil
}

func substituteMe(filters []fieldFilter, currentUser string) []fieldFilter {
	if currentUser == "" {
		return filters
	}
	for i := range filters {
		if strings.EqualFold(filters[i].Value, "me") {
			filters[i].Value = currentUser
		}
	}
	return filters
}

func matchesFieldFilters(item github.ProjectItem, filters []fieldFilter) bool {
	for _, filter := range filters {
		value, ok := fieldValue(item, filter.Name)
		if !ok || !fieldValueMatches(value, filter.Value) {
			return false
		}
	}
	return true
}

func fieldValue(item github.ProjectItem, name string) (string, bool) {
	for fieldName, value := range item.Fields {
		if strings.EqualFold(fieldName, name) {
			return value, true
		}
	}
	if strings.EqualFold(name, "Status") {
		return item.Status, item.Status != ""
	}
	return "", false
}

func fieldValueMatches(actual, expected string) bool {
	for _, part := range strings.Split(actual, ",") {
		if strings.EqualFold(strings.TrimSpace(part), expected) {
			return true
		}
	}
	return strings.EqualFold(strings.TrimSpace(actual), strings.TrimSpace(expected))
}

func projectWithItems(items map[string][]github.ProjectItem) *github.Project {
	return &github.Project{Items: items}
}
