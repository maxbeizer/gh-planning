package github

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
)

func Run(ctx context.Context, args ...string) ([]byte, error) {
	return runGH(ctx, args...)
}

func runGH(ctx context.Context, args ...string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, "gh", args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("gh %v failed: %w: %s", args, err, stderr.String())
	}
	return stdout.Bytes(), nil
}

func GraphQL(ctx context.Context, query string, variables map[string]interface{}) ([]byte, error) {
	args := []string{"api", "graphql", "-f", fmt.Sprintf("query=%s", query)}
	for key, val := range variables {
		switch v := val.(type) {
		case int:
			args = append(args, "-F", fmt.Sprintf("%s=%d", key, v))
		case string:
			args = append(args, "-f", fmt.Sprintf("%s=%s", key, v))
		default:
			args = append(args, "-f", fmt.Sprintf("%s=%v", key, v))
		}
	}
	return runGH(ctx, args...)
}

func API(ctx context.Context, method string, path string, fields map[string]string) ([]byte, error) {
	args := []string{"api"}
	if method != "" {
		args = append(args, "-X", method)
	}
	args = append(args, path)
	for key, val := range fields {
		args = append(args, "-f", fmt.Sprintf("%s=%s", key, val))
	}
	return runGH(ctx, args...)
}
