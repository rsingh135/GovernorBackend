package handlers

import (
	"net/http"

	"agentpay/internal/httpjson"
)

// MCPHandler handles MCP (Model Context Protocol) discovery endpoints.
type MCPHandler struct{}

// NewMCPHandler creates a new MCP handler.
func NewMCPHandler() *MCPHandler {
	return &MCPHandler{}
}

// MCPTool represents a tool in MCP format.
type MCPTool struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	InputSchema interface{} `json:"inputSchema"`
}

// MCPDiscovery represents the MCP discovery document.
type MCPDiscovery struct {
	ToolsURL    string `json:"tools_url"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

// Tools handles GET /mcp/tools - returns available tools in MCP format.
func (h *MCPHandler) Tools(w http.ResponseWriter, r *http.Request) {
	tools := []MCPTool{
		{
			Name:        "governor_spend",
			Description: "Request approval to spend money on behalf of a user. Governor evaluates spending policies and returns approve/deny/escalate.",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"request_id": map[string]interface{}{
						"type":        "string",
						"description": "Unique UUID for idempotency",
					},
					"amount": map[string]interface{}{
						"type":        "integer",
						"description": "Amount in cents",
					},
					"vendor": map[string]interface{}{
						"type":        "string",
						"description": "Vendor name or domain",
					},
					"meta": map[string]interface{}{
						"type":        "object",
						"description": "Additional context about the spend",
					},
				},
				"required": []string{"request_id", "amount", "vendor"},
			},
		},
		{
			Name:        "governor_check_budget",
			Description: "Check remaining daily budget for the authenticated agent.",
			InputSchema: map[string]interface{}{
				"type":       "object",
				"properties": map[string]interface{}{},
			},
		},
	}

	httpjson.Write(w, http.StatusOK, tools)
}

// Discovery handles GET /.well-known/mcp.json - returns MCP discovery document.
func (h *MCPHandler) Discovery(w http.ResponseWriter, r *http.Request) {
	discovery := MCPDiscovery{
		ToolsURL:    "/mcp/tools",
		Name:        "Governor",
		Description: "AI agent financial governance API",
	}

	httpjson.Write(w, http.StatusOK, discovery)
}
