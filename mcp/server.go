package mcp

import (
	"bufio"
	"encoding/json"
	"errors"
	"io"
	"os"
	"os/exec"
	"strings"
)

type Server struct {
	In  io.Reader
	Out io.Writer
	Err io.Writer
}

type rpcRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      interface{}     `json:"id"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

type rpcResponse struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      interface{} `json:"id"`
	Result  interface{} `json:"result,omitempty"`
	Error   *rpcError   `json:"error,omitempty"`
}

type rpcError struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

type toolCallParams struct {
	Name      string                 `json:"name"`
	Arguments map[string]interface{} `json:"arguments"`
}

func Serve(in io.Reader, out io.Writer, errOut io.Writer) error {
	server := &Server{In: in, Out: out, Err: errOut}
	return server.Run()
}

func (s *Server) Run() error {
	dec := json.NewDecoder(bufio.NewReader(s.In))
	enc := json.NewEncoder(s.Out)
	for {
		var req rpcRequest
		if err := dec.Decode(&req); err != nil {
			if errors.Is(err, io.EOF) {
				return nil
			}
			return err
		}
		if req.JSONRPC == "" {
			req.JSONRPC = "2.0"
		}
		if req.Method == "" {
			continue
		}
		resp := s.handleRequest(&req)
		if resp == nil {
			continue
		}
		// JSON-RPC notifications (no id) don't get responses
		if req.ID == nil {
			continue
		}
		if err := enc.Encode(resp); err != nil {
			return err
		}
	}
}

func (s *Server) handleRequest(req *rpcRequest) *rpcResponse {
	switch req.Method {
	case "initialize":
		return &rpcResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Result: map[string]interface{}{
				"protocolVersion": "2024-11-05",
				"serverInfo": map[string]interface{}{
					"name":    "gh-planning",
					"version": "0.1.0",
				},
				"capabilities": map[string]interface{}{
					"tools": map[string]interface{}{},
				},
			},
		}
	case "tools/list":
		toolList := []ToolDefinition{}
		for _, tool := range Tools() {
			toolList = append(toolList, ToolDefinition{
				Name:        tool.Name,
				Description: tool.Description,
				InputSchema: tool.InputSchema,
			})
		}
		return &rpcResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Result: map[string]interface{}{
				"tools": toolList,
			},
		}
	case "tools/call":
		var params toolCallParams
		if len(req.Params) > 0 {
			if err := json.Unmarshal(req.Params, &params); err != nil {
				return rpcErrorResponse(req.ID, -32602, "Invalid params", err.Error())
			}
		}
		if params.Name == "" {
			return rpcErrorResponse(req.ID, -32602, "Tool name required", nil)
		}
		tool, ok := ToolByName(params.Name)
		if !ok {
			return rpcErrorResponse(req.ID, -32601, "Tool not found", params.Name)
		}
		args := params.Arguments
		if args == nil {
			args = map[string]interface{}{}
		}
		cmdArgs, err := tool.Build(args)
		if err != nil {
			return rpcErrorResponse(req.ID, -32602, "Invalid params", err.Error())
		}
		cmdArgs = append(cmdArgs, "--json")
		output, execErr := runCommand(cmdArgs)
		if execErr != nil {
			// Return error as content so the client can display it
			return &rpcResponse{
				JSONRPC: "2.0",
				ID:      req.ID,
				Result: map[string]interface{}{
					"content": []map[string]interface{}{
						{"type": "text", "text": string(output)},
					},
					"isError": true,
				},
			}
		}
		text := strings.TrimSpace(string(output))
		if text == "" {
			text = "(no output)"
		}
		return &rpcResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Result: map[string]interface{}{
				"content": []map[string]interface{}{
					{"type": "text", "text": text},
				},
			},
		}
	default:
		// Silently ignore notifications (e.g., notifications/initialized)
		if strings.HasPrefix(req.Method, "notifications/") {
			return nil
		}
		return rpcErrorResponse(req.ID, -32601, "Method not found", req.Method)
	}
}

func runCommand(args []string) ([]byte, error) {
	cmd := exec.Command("gh", args...)
	cmd.Env = os.Environ()
	return cmd.CombinedOutput()
}

func parseOutput(output []byte) interface{} {
	trimmed := strings.TrimSpace(string(output))
	if trimmed == "" {
		return map[string]interface{}{"output": ""}
	}
	var payload interface{}
	if err := json.Unmarshal([]byte(trimmed), &payload); err == nil {
		return payload
	}
	return map[string]interface{}{"output": trimmed}
}

func rpcErrorResponse(id interface{}, code int, message string, data interface{}) *rpcResponse {
	return &rpcResponse{
		JSONRPC: "2.0",
		ID:      id,
		Error: &rpcError{
			Code:    code,
			Message: message,
			Data:    data,
		},
	}
}

func ExampleRequest() string {
	req := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      1,
		"method":  "tools/call",
		"params": map[string]interface{}{
			"name": "planning-status",
			"arguments": map[string]interface{}{
				"project": 25,
				"owner":   "maxbeizer",
			},
		},
	}
	payload, _ := json.MarshalIndent(req, "", "  ")
	return string(payload)
}

func ExampleResponse() string {
	resp := rpcResponse{
		JSONRPC: "2.0",
		ID:      1,
		Result: map[string]interface{}{
			"status": "ok",
		},
	}
	payload, _ := json.MarshalIndent(resp, "", "  ")
	return string(payload)
}
