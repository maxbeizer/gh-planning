package cmd

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

func parseDuration(value string) (time.Duration, error) {
	if value == "" {
		return 0, nil
	}
	if strings.HasSuffix(value, "d") {
		days := strings.TrimSuffix(value, "d")
		count, err := strconv.Atoi(days)
		if err != nil {
			return 0, err
		}
		return time.Duration(count) * 24 * time.Hour, nil
	}
	return time.ParseDuration(value)
}

func humanizeDuration(d time.Duration) string {
	if d < 0 {
		d = -d
	}
	minutes := int(d.Minutes())
	if minutes < 60 {
		return fmt.Sprintf("%dm ago", minutes)
	}
	hours := int(d.Hours())
	if hours < 24 {
		return fmt.Sprintf("%dh ago", hours)
	}
	days := int(d.Hours() / 24)
	if days < 7 {
		return fmt.Sprintf("%dd ago", days)
	}
	weeks := days / 7
	return fmt.Sprintf("%dw ago", weeks)
}
