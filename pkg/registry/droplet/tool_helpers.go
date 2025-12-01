package droplet

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/digitalocean/godo"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// GenericToolHandler creates a generic handler that wraps a ToolConfig handler
// It handles client creation, argument validation, and response formatting
func GenericToolHandler(config *ToolConfig, clientFactory func(ctx context.Context) (*godo.Client, error)) func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		// Validate arguments
		args := req.GetArguments()
		if err := config.ValidateArguments(args); err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		// Get client
		client, err := clientFactory(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to get DigitalOcean client: %w", err)
		}

		// Call the handler
		result, err := config.Handler(ctx, client, args)
		if err != nil {
			return mcp.NewToolResultErrorFromErr("api error", err), nil
		}

		// Handle string results directly
		if str, ok := result.(string); ok {
			return mcp.NewToolResultText(str), nil
		}

		// Marshal result to JSON for non-string results
		jsonData, err := json.MarshalIndent(result, "", "  ")
		if err != nil {
			return nil, fmt.Errorf("json marshal error: %w", err)
		}

		return mcp.NewToolResultText(string(jsonData)), nil
	}
}

// BuildServerTool converts a ToolConfig into a server.ServerTool
func BuildServerTool(config *ToolConfig, clientFactory func(ctx context.Context) (*godo.Client, error)) server.ServerTool {
	return server.ServerTool{
		Handler: GenericToolHandler(config, clientFactory),
		Tool:    config.BuildMCPTool(),
	}
}
