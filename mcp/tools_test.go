package mcp

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func TestToolsCount(t *testing.T) {
	tools := Tools()
	if len(tools) < 30 {
		t.Errorf("expected at least 30 tools, got %d", len(tools))
	}
}

func TestToolsSorted(t *testing.T) {
	tools := Tools()
	for i := 1; i < len(tools); i++ {
		if tools[i].Name < tools[i-1].Name {
			t.Errorf("tools not sorted: %q comes after %q", tools[i].Name, tools[i-1].Name)
		}
	}
}

func TestToolByName_Found(t *testing.T) {
	tool, ok := ToolByName("planning-status")
	if !ok {
		t.Fatal("planning-status not found")
	}
	if tool.Description == "" {
		t.Error("expected non-empty description")
	}
}

func TestToolByName_NotFound(t *testing.T) {
	_, ok := ToolByName("planning-nonexistent")
	if ok {
		t.Error("expected not found for nonexistent tool")
	}
}

func TestToolNames(t *testing.T) {
	expected := []string{
		"planning-agentContext",
		"planning-blocked",
		"planning-board",
		"planning-breakdown",
		"planning-catchup",
		"planning-cheatsheet",
		"planning-claim",
		"planning-complete",
		"planning-criticalPath",
		"planning-estimate",
		"planning-focus",
		"planning-guide",
		"planning-handoff",
		"planning-log",
		"planning-logs",
		"planning-prioritize",
		"planning-profile-create",
		"planning-profile-detect",
		"planning-profile-list",
		"planning-profile-show",
		"planning-pulse",
		"planning-queue",
		"planning-review",
		"planning-roadmap",
		"planning-sprint",
		"planning-standup",
		"planning-status",
		"planning-team",
		"planning-track",
	}
	tools := Tools()
	names := make(map[string]bool)
	for _, tool := range tools {
		names[tool.Name] = true
	}
	for _, name := range expected {
		if !names[name] {
			t.Errorf("expected tool %q not found", name)
		}
	}
}

func TestBuildFlags_StatusTool(t *testing.T) {
	tool, ok := ToolByName("planning-status")
	if !ok {
		t.Fatal("planning-status not found")
	}
	args, err := tool.Build(map[string]interface{}{
		"project": float64(25),
		"owner":   "maxbeizer",
		"board":   true,
	})
	if err != nil {
		t.Fatalf("Build error: %v", err)
	}
	joined := strings.Join(args, " ")
	if !strings.Contains(joined, "--project 25") {
		t.Errorf("expected --project 25, got: %s", joined)
	}
	if !strings.Contains(joined, "--owner maxbeizer") {
		t.Errorf("expected --owner maxbeizer, got: %s", joined)
	}
	if !strings.Contains(joined, "--board") {
		t.Errorf("expected --board, got: %s", joined)
	}
}

func TestBuildFlags_TrackTool(t *testing.T) {
	tool, ok := ToolByName("planning-track")
	if !ok {
		t.Fatal("planning-track not found")
	}
	args, err := tool.Build(map[string]interface{}{
		"title": "Fix auth bug",
		"repo":  "maxbeizer/app",
	})
	if err != nil {
		t.Fatalf("Build error: %v", err)
	}
	if args[2] != "Fix auth bug" {
		t.Errorf("expected title as positional arg, got: %v", args)
	}
}

func TestBuildFlags_TrackRequiresTitle(t *testing.T) {
	tool, ok := ToolByName("planning-track")
	if !ok {
		t.Fatal("planning-track not found")
	}
	_, err := tool.Build(map[string]interface{}{
		"repo": "maxbeizer/app",
	})
	if err == nil {
		t.Error("expected error when title is missing")
	}
}

func TestBuildFlags_CheatsheetUsesPlain(t *testing.T) {
	tool, ok := ToolByName("planning-cheatsheet")
	if !ok {
		t.Fatal("planning-cheatsheet not found")
	}
	args, err := tool.Build(map[string]interface{}{})
	if err != nil {
		t.Fatalf("Build error: %v", err)
	}
	joined := strings.Join(args, " ")
	if !strings.Contains(joined, "--plain") {
		t.Errorf("expected --plain for non-interactive MCP use, got: %s", joined)
	}
}

func TestBuildFlags_GuideWithWorkflow(t *testing.T) {
	tool, ok := ToolByName("planning-guide")
	if !ok {
		t.Fatal("planning-guide not found")
	}
	args, err := tool.Build(map[string]interface{}{
		"workflow": "morning",
	})
	if err != nil {
		t.Fatalf("Build error: %v", err)
	}
	joined := strings.Join(args, " ")
	if !strings.Contains(joined, "morning") {
		t.Errorf("expected morning workflow, got: %s", joined)
	}
	if !strings.Contains(joined, "--plain") {
		t.Errorf("expected --plain, got: %s", joined)
	}
}

func TestServerInitialize(t *testing.T) {
	req := `{"jsonrpc":"2.0","id":1,"method":"initialize"}` + "\n"
	in := strings.NewReader(req)
	var out bytes.Buffer

	server := &Server{In: in, Out: &out, Err: &bytes.Buffer{}}
	if err := server.Run(); err != nil {
		t.Fatalf("Run error: %v", err)
	}

	var resp rpcResponse
	if err := json.Unmarshal(out.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal error: %v\nraw: %s", err, out.String())
	}
	if resp.Error != nil {
		t.Fatalf("unexpected error: %v", resp.Error)
	}
	result, ok := resp.Result.(map[string]interface{})
	if !ok {
		t.Fatal("result is not a map")
	}
	serverInfo, ok := result["serverInfo"].(map[string]interface{})
	if !ok {
		t.Fatal("serverInfo missing")
	}
	if serverInfo["name"] != "gh-planning" {
		t.Errorf("server name = %v, want gh-planning", serverInfo["name"])
	}
}

func TestServerToolsList(t *testing.T) {
	req := `{"jsonrpc":"2.0","id":2,"method":"tools/list"}` + "\n"
	in := strings.NewReader(req)
	var out bytes.Buffer

	server := &Server{In: in, Out: &out, Err: &bytes.Buffer{}}
	if err := server.Run(); err != nil {
		t.Fatalf("Run error: %v", err)
	}

	var resp rpcResponse
	if err := json.Unmarshal(out.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}
	if resp.Error != nil {
		t.Fatalf("unexpected error: %v", resp.Error)
	}
	result, ok := resp.Result.(map[string]interface{})
	if !ok {
		t.Fatal("result is not a map")
	}
	toolsList, ok := result["tools"].([]interface{})
	if !ok {
		t.Fatal("tools is not an array")
	}
	if len(toolsList) < 29 {
		t.Errorf("expected at least 30 tools, got %d", len(toolsList))
	}
}

func TestServerToolNotFound(t *testing.T) {
	req := `{"jsonrpc":"2.0","id":3,"method":"tools/call","params":{"name":"planning-fake","arguments":{}}}` + "\n"
	in := strings.NewReader(req)
	var out bytes.Buffer

	server := &Server{In: in, Out: &out, Err: &bytes.Buffer{}}
	if err := server.Run(); err != nil {
		t.Fatalf("Run error: %v", err)
	}

	var resp rpcResponse
	if err := json.Unmarshal(out.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}
	if resp.Error == nil {
		t.Fatal("expected error for unknown tool")
	}
	if resp.Error.Code != -32601 {
		t.Errorf("error code = %d, want -32601", resp.Error.Code)
	}
}

func TestServerMethodNotFound(t *testing.T) {
	req := `{"jsonrpc":"2.0","id":4,"method":"unknown/method"}` + "\n"
	in := strings.NewReader(req)
	var out bytes.Buffer

	server := &Server{In: in, Out: &out, Err: &bytes.Buffer{}}
	if err := server.Run(); err != nil {
		t.Fatalf("Run error: %v", err)
	}

	var resp rpcResponse
	if err := json.Unmarshal(out.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}
	if resp.Error == nil {
		t.Fatal("expected error for unknown method")
	}
}

func TestParseOutput_JSON(t *testing.T) {
	result := parseOutput([]byte(`{"items": [1, 2, 3]}`))
	m, ok := result.(map[string]interface{})
	if !ok {
		t.Fatalf("expected map, got %T", result)
	}
	if _, ok := m["items"]; !ok {
		t.Error("expected items key in parsed output")
	}
}

func TestParseOutput_PlainText(t *testing.T) {
	result := parseOutput([]byte("just some text"))
	m, ok := result.(map[string]interface{})
	if !ok {
		t.Fatalf("expected map, got %T", result)
	}
	if m["output"] != "just some text" {
		t.Errorf("output = %v, want 'just some text'", m["output"])
	}
}

func TestParseOutput_Empty(t *testing.T) {
	result := parseOutput([]byte(""))
	m, ok := result.(map[string]interface{})
	if !ok {
		t.Fatalf("expected map, got %T", result)
	}
	if m["output"] != "" {
		t.Errorf("output = %v, want empty string", m["output"])
	}
}

func TestHelperFunctions(t *testing.T) {
	// boolValue
	if v, ok := boolValue(true); !ok || !v {
		t.Error("boolValue(true) failed")
	}
	if v, ok := boolValue("true"); !ok || !v {
		t.Error(`boolValue("true") failed`)
	}
	if _, ok := boolValue("maybe"); ok {
		t.Error(`boolValue("maybe") should not be ok`)
	}

	// intValue
	if v, ok := intValue(42); !ok || v != 42 {
		t.Error("intValue(42) failed")
	}
	if v, ok := intValue(float64(99)); !ok || v != 99 {
		t.Error("intValue(float64(99)) failed")
	}
	if v, ok := intValue("7"); !ok || v != 7 {
		t.Error(`intValue("7") failed`)
	}

	// sliceValue
	if items := sliceValue([]interface{}{"a", "b"}); len(items) != 2 {
		t.Errorf("sliceValue([]interface{}) = %v", items)
	}
	if items := sliceValue("single"); len(items) != 1 || items[0] != "single" {
		t.Errorf("sliceValue(single) = %v", items)
	}

	// firstString
	args := map[string]interface{}{"a": "", "b": "found"}
	if v := firstString(args, "a", "b"); v != "found" {
		t.Errorf("firstString = %q, want 'found'", v)
	}
}
