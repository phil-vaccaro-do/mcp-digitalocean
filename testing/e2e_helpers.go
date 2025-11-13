//go:build integration

package testing

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/stretchr/testify/require"
)

// getSSHKeys retrieves SSH key IDs from the account for use in droplet creation
func getSSHKeys(ctx context.Context, c *client.Client, t *testing.T) []interface{} {
	resp, err := c.CallTool(ctx, mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name:      "key-list",
			Arguments: map[string]interface{}{},
		},
	})
	require.NoError(t, err)
	if resp.IsError {
		t.Logf("SSH key list error response: %+v", resp.Content)
	}
	require.False(t, resp.IsError, "Failed to list SSH keys")

	var keys []map[string]interface{}
	keysJSON := resp.Content[0].(mcp.TextContent).Text
	err = json.Unmarshal([]byte(keysJSON), &keys)
	require.NoError(t, err)
	require.NotEmpty(t, keys, "No SSH keys found in account. Please add at least one SSH key.")

	// Extract key IDs
	var keyIDs []interface{}
	for _, key := range keys {
		if id, ok := key["id"].(float64); ok {
			keyIDs = append(keyIDs, id)
		}
	}
	require.NotEmpty(t, keyIDs, "No valid SSH key IDs found")

	t.Logf("Found %d SSH key(s) in account", len(keyIDs))
	return keyIDs
}

// getTestImage retrieves a suitable Ubuntu LTS image ID for testing
func getTestImage(ctx context.Context, c *client.Client, t *testing.T) float64 {
	// List images and find Ubuntu 22.04 LTS
	imagesResp, err := c.CallTool(ctx, mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "image-list",
			Arguments: map[string]interface{}{
				"Type": "distribution",
			},
		},
	})
	require.NoError(t, err)
	require.False(t, imagesResp.IsError, "Failed to list images")

	var images []map[string]interface{}
	imagesJSON := imagesResp.Content[0].(mcp.TextContent).Text
	err = json.Unmarshal([]byte(imagesJSON), &images)
	require.NoError(t, err)
	require.NotEmpty(t, images, "No images found")

	// Prefer Ubuntu 22.04 LTS for consistency
	for _, img := range images {
		if slug, ok := img["slug"].(string); ok {
			if slug == "ubuntu-22-04-x64" {
				imageID := img["id"].(float64)
				t.Logf("Using Ubuntu 22.04 LTS image (ID: %.0f)", imageID)
				return imageID
			}
		}
	}

	// Fallback to first available image
	imageID := images[0]["id"].(float64)
	t.Logf("Using fallback image (ID: %.0f)", imageID)
	return imageID
}
