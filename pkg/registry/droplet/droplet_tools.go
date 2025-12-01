package droplet

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/digitalocean/godo"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// DropletTool provides droplet management tools
type DropletTool struct {
	client func(ctx context.Context) (*godo.Client, error)
}

// NewDropletTool creates a new droplet tool
func NewDropletTool(client func(ctx context.Context) (*godo.Client, error)) *DropletTool {
	return &DropletTool{
		client: client,
	}
}

// enablePrivateNetworking enables private networking on a droplet
func (d *DropletTool) enablePrivateNetworking(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	dropletID := req.GetArguments()["ID"].(float64)

	client, err := d.client(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get DigitalOcean client: %w", err)
	}

	action, _, err := client.DropletActions.EnablePrivateNetworking(ctx, int(dropletID))
	if err != nil {
		return mcp.NewToolResultErrorFromErr("api error", err), nil
	}

	jsonAction, err := json.MarshalIndent(action, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("marshal error: %w", err)
	}

	return mcp.NewToolResultText(string(jsonAction)), nil
}

// getDropletKernels gets available kernels for a droplet
func (d *DropletTool) getDropletKernels(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	dropletID := req.GetArguments()["ID"].(float64)

	// Use list options to get all kernels
	opt := &godo.ListOptions{
		Page:    1,
		PerPage: 100,
	}

	client, err := d.client(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get DigitalOcean client: %w", err)
	}

	kernels, _, err := client.Droplets.Kernels(ctx, int(dropletID), opt)
	if err != nil {
		return mcp.NewToolResultErrorFromErr("api error", err), nil
	}

	jsonKernels, err := json.MarshalIndent(kernels, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("marshal error: %w", err)
	}

	return mcp.NewToolResultText(string(jsonKernels)), nil
}

// getDropletBackupPolicy returns the backup policy for a droplet.
func (d *DropletTool) getDropletBackupPolicy(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	id, ok := req.GetArguments()["ID"].(float64)
	if !ok {
		return mcp.NewToolResultError("Droplet ID is required"), nil
	}

	client, err := d.client(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get DigitalOcean client: %w", err)
	}

	policy, _, err := client.Droplets.GetBackupPolicy(ctx, int(id))
	if err != nil {
		return mcp.NewToolResultErrorFromErr("api error", err), nil
	}

	jsonData, err := json.MarshalIndent(policy, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("marshal error: %w", err)
	}
	return mcp.NewToolResultText(string(jsonData)), nil
}

func (d *DropletTool) getDropletActionByID(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	dropletID, ok := req.GetArguments()["DropletID"].(float64)
	if !ok {
		return mcp.NewToolResultError("DropletID is required"), nil
	}
	actionID, ok := req.GetArguments()["ActionID"].(float64)
	if !ok {
		return mcp.NewToolResultError("ActionID is required"), nil
	}

	client, err := d.client(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get DigitalOcean client: %w", err)
	}

	action, _, err := client.DropletActions.Get(ctx, int(dropletID), int(actionID))
	if err != nil {
		return mcp.NewToolResultErrorFromErr("api error", err), nil
	}
	jsonData, err := json.MarshalIndent(action, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("marshal error: %w", err)
	}
	return mcp.NewToolResultText(string(jsonData)), nil
}

func (d *DropletTool) Tools() []server.ServerTool {
	tools := []server.ServerTool{
		// Basic droplet tools using new ToolConfig system
		BuildServerTool(dropletListConfig(), d.client),
		BuildServerTool(dropletGetConfig(), d.client),
		BuildServerTool(dropletCreateConfig(), d.client),
		BuildServerTool(dropletDeleteConfig(), d.client),
		BuildServerTool(dropletNeighborsConfig(), d.client),

		// Legacy tools (to be migrated in future PRs)
		{
			Handler: d.enablePrivateNetworking,
			Tool: mcp.NewTool("droplet-enable-private-net",
				mcp.WithDescription("Enable private networking on a droplet"),
				mcp.WithNumber("ID", mcp.Required(), mcp.Description("ID of the droplet")),
			),
		},
		{
			Handler: d.getDropletKernels,
			Tool: mcp.NewTool("droplet-kernels",
				mcp.WithDescription("Get available kernels for a droplet"),
				mcp.WithNumber("ID", mcp.Required(), mcp.Description("ID of the droplet")),
			),
		},

		{
			Handler: d.getDropletBackupPolicy,
			Tool: mcp.NewTool("droplet-backup-policy",
				mcp.WithDescription("Get a droplet's backup policy"),
				mcp.WithNumber("ID", mcp.Required(), mcp.Description("Droplet ID")),
			),
		},
		{
			Handler: d.getDropletActionByID,
			Tool: mcp.NewTool("droplet-action",
				mcp.WithDescription("Get a droplet action by droplet ID and action ID"),
				mcp.WithNumber("DropletID", mcp.Required(), mcp.Description("Droplet ID")),
				mcp.WithNumber("ActionID", mcp.Required(), mcp.Description("Action ID")),
			),
		},
	}
	return tools
}
