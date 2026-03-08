package github

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
)

const breakdownSystemPrompt = "You are a senior engineering manager breaking down a GitHub issue into smaller, independently shippable sub-issues. Each sub-issue should be completable in 1-2 days. Output JSON array."

type BreakdownItem struct {
	Title     string   `json:"title"`
	Body      string   `json:"body"`
	Labels    []string `json:"labels,omitempty"`
	DependsOn []int    `json:"dependsOn,omitempty"`
}

type modelsRequest struct {
	Model    string `json:"model"`
	Messages []struct {
		Role    string `json:"role"`
		Content string `json:"content"`
	} `json:"messages"`
}

type modelsResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

func ModelsBreakdown(ctx context.Context, model string, prompt string) ([]BreakdownItem, error) {
	tokenPayload, err := runGH(ctx, "auth", "token")
	if err != nil {
		return nil, err
	}
	token := strings.TrimSpace(string(tokenPayload))
	if token == "" {
		return nil, errors.New("empty GitHub auth token")
	}
	reqBody := modelsRequest{Model: model}
	reqBody.Messages = append(reqBody.Messages, struct {
		Role    string `json:"role"`
		Content string `json:"content"`
	}{Role: "system", Content: breakdownSystemPrompt})
	reqBody.Messages = append(reqBody.Messages, struct {
		Role    string `json:"role"`
		Content string `json:"content"`
	}{Role: "user", Content: prompt})
	payload, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://models.github.ai/inference/chat/completions", bytes.NewReader(payload))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("models API failed: %s", resp.Status)
	}
	var response modelsResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, err
	}
	if len(response.Choices) == 0 {
		return nil, errors.New("no model choices returned")
	}
	content := strings.TrimSpace(response.Choices[0].Message.Content)
	content = stripCodeFence(content)
	var items []BreakdownItem
	if err := json.Unmarshal([]byte(content), &items); err != nil {
		return nil, err
	}
	return items, nil
}

const estimateSystemPrompt = "You are a senior engineer estimating the effort required for a GitHub issue. Compare it against similar recently closed issues and suggest a t-shirt size. Output a single JSON object."

type EstimateResult struct {
	Size       string           `json:"size"`
	Confidence string           `json:"confidence"`
	Reasoning  string           `json:"reasoning"`
	Similar    []SimilarIssue   `json:"similar"`
}

type SimilarIssue struct {
	Number int    `json:"number"`
	Title  string `json:"title"`
	Size   string `json:"size"`
}

func ModelsEstimate(ctx context.Context, model string, prompt string) (*EstimateResult, error) {
	tokenPayload, err := runGH(ctx, "auth", "token")
	if err != nil {
		return nil, err
	}
	token := strings.TrimSpace(string(tokenPayload))
	if token == "" {
		return nil, errors.New("empty GitHub auth token")
	}
	reqBody := modelsRequest{Model: model}
	reqBody.Messages = append(reqBody.Messages, struct {
		Role    string `json:"role"`
		Content string `json:"content"`
	}{Role: "system", Content: estimateSystemPrompt})
	reqBody.Messages = append(reqBody.Messages, struct {
		Role    string `json:"role"`
		Content string `json:"content"`
	}{Role: "user", Content: prompt})
	payload, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://models.github.ai/inference/chat/completions", bytes.NewReader(payload))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("models API failed: %s", resp.Status)
	}
	var response modelsResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, err
	}
	if len(response.Choices) == 0 {
		return nil, errors.New("no model choices returned")
	}
	content := strings.TrimSpace(response.Choices[0].Message.Content)
	content = stripCodeFence(content)
	var result EstimateResult
	if err := json.Unmarshal([]byte(content), &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func stripCodeFence(value string) string {
	value = strings.TrimSpace(value)
	if strings.HasPrefix(value, "```") {
		value = strings.TrimPrefix(value, "```")
		value = strings.TrimPrefix(value, "json")
		value = strings.TrimSpace(value)
		value = strings.TrimSuffix(value, "```")
	}
	return strings.TrimSpace(value)
}
