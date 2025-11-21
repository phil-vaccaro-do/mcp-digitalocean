//go:build integration

package testing

import (
	"context"
	"encoding/json"
	"fmt"
	"mcp-digitalocean/internal/testhelpers"
	"os"
	"testing"
	"time"

	"github.com/digitalocean/godo"
	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/stretchr/testify/require"
)

// setupTest initializes context, MCP client (via Docker/HTTP), and Godo client.
func setupTest(t *testing.T) (context.Context, *client.Client, *godo.Client, func()) {
	ctx := context.Background()
	// Calls initializeClient from e2e_test.go
	c := initializeClient(ctx, t)
	gclient := testhelpers.MustGodoClient()

	return ctx, c, gclient, func() {
		c.Close()
	}
}

// triggerActionAndWait is a helper to execute an action tool and wait for the result.
func triggerActionAndWait(t *testing.T, ctx context.Context, c *client.Client, gclient *godo.Client, tool string, args map[string]interface{}, resourceID int) {
	action := callTool[godo.Action](ctx, c, t, tool, args)
	require.NotZero(t, action.ID, "Action ID should not be zero")

	t.Logf("Action %s initiated (ID: %d)", tool, action.ID)

	final, err := testhelpers.WaitForAction(ctx, gclient, resourceID, action.ID, 2*time.Second, 2*time.Minute)
	require.NoError(t, err, "Action failed to complete")
	LogActionCompleted(t, tool, *final)
}

// callTool calls an MCP tool and returns the unmarshaled result T.
func callTool[T any](ctx context.Context, c *client.Client, t *testing.T, name string, args map[string]interface{}) T {
	var result T
	resp, err := c.CallTool(ctx, mcp.CallToolRequest{
		Params: mcp.CallToolParams{Name: name, Arguments: args},
	})
	require.NoError(t, err)

	if resp.IsError {
		var errText string
		if len(resp.Content) > 0 {
			if tc, ok := resp.Content[0].(mcp.TextContent); ok {
				errText = tc.Text
			} else {
				errText = fmt.Sprintf("%v", resp.Content)
			}
		}
		t.Fatalf("Tool %s failed: %s", name, errText)
	}

	require.NotEmpty(t, resp.Content, "Tool %s returned empty content", name)
	tc, ok := resp.Content[0].(mcp.TextContent)
	require.True(t, ok, "Unexpected content type for %s", name)

	err = json.Unmarshal([]byte(tc.Text), &result)
	require.NoError(t, err, "Failed to unmarshal %s response", name)

	return result
}

// --- Resource Lifecycle Helpers ---

func CreateTestDroplet(ctx context.Context, c *client.Client, t *testing.T, namePrefix string) godo.Droplet {
	sshKeys := getSSHKeys(ctx, c, t)
	region := selectRegion(ctx, c, t)
	imageID := getTestImage(ctx, c, t)
	dropletName := fmt.Sprintf("%s-%d", namePrefix, time.Now().Unix())

	t.Logf("Creating Droplet: %s (Image: %.0f, Region: %s)", dropletName, imageID, region)

	droplet := callTool[godo.Droplet](ctx, c, t, "droplet-create", map[string]interface{}{
		"Name":       dropletName,
		"Size":       "s-1vcpu-1gb",
		"ImageID":    imageID,
		"Region":     region,
		"Backup":     false,
		"Monitoring": true,
		"SSHKeys":    sshKeys,
	})

	return WaitForDropletActive(ctx, c, t, droplet.ID, 2*time.Minute)
}

func DeleteResource(ctx context.Context, c *client.Client, t *testing.T, resourceType string, id interface{}) {
	resp, err := c.CallTool(ctx, mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name:      fmt.Sprintf("%s-delete", resourceType),
			Arguments: map[string]interface{}{"ID": id, "ImageID": id},
		},
	})
	LogResourceDeleted(t, resourceType, id, err, resp)
}

func ListResources(ctx context.Context, c *client.Client, t *testing.T, resourceType string, page, perPage int) []map[string]interface{} {
	return callTool[[]map[string]interface{}](ctx, c, t, fmt.Sprintf("%s-list", resourceType), map[string]interface{}{
		"Page":    page,
		"PerPage": perPage,
	})
}

// --- Prerequisite Helpers ---

func getSSHKeys(ctx context.Context, c *client.Client, t *testing.T) []interface{} {
	keys := callTool[[]map[string]interface{}](ctx, c, t, "key-list", map[string]interface{}{})
	require.NotEmpty(t, keys, "No SSH keys found in account.")

	var keyIDs []interface{}
	for _, key := range keys {
		if id, ok := key["id"].(float64); ok {
			keyIDs = append(keyIDs, id)
		}
	}
	return keyIDs
}

func getTestImage(ctx context.Context, c *client.Client, t *testing.T) float64 {
	images := callTool[[]map[string]interface{}](ctx, c, t, "image-list", map[string]interface{}{"Type": "distribution"})

	for _, img := range images {
		if slug, ok := img["slug"].(string); ok && slug == "ubuntu-22-04-x64" {
			return img["id"].(float64)
		}
	}
	require.NotEmpty(t, images, "No images found")
	return images[0]["id"].(float64)
}

func selectRegion(ctx context.Context, c *client.Client, t *testing.T) string {
	if rg := os.Getenv("TEST_REGION"); rg != "" {
		return rg
	}

	regions := callTool[[]map[string]interface{}](ctx, c, t, "region-list", map[string]interface{}{"Page": 1, "PerPage": 100})

	for _, r := range regions {
		slug, _ := r["slug"].(string)
		avail, _ := r["available"].(bool)
		if slug != "" && avail {
			return slug
		}
	}
	t.Fatal("No available region found")
	return ""
}

// --- Wait Wrappers ---

func WaitForDropletActive(ctx context.Context, _ *client.Client, t *testing.T, dropletID int, timeout time.Duration) godo.Droplet {
	gclient := testhelpers.MustGodoClient()
	d, err := testhelpers.WaitForDroplet(ctx, gclient, dropletID, testhelpers.IsDropletActive, 3*time.Second, timeout)
	require.NoError(t, err, "WaitForDropletActive failed")
	return *d
}

func WaitForActionComplete(ctx context.Context, c *client.Client, t *testing.T, actionID int, timeout time.Duration) godo.Action {
	gclient := testhelpers.MustGodoClient()

	act := callTool[godo.Action](ctx, c, t, "action-get", map[string]interface{}{"ID": float64(actionID)})
	require.NotZero(t, act.ResourceID, "Action ResourceID is 0")

	final, err := testhelpers.WaitForAction(ctx, gclient, act.ResourceID, actionID, 2*time.Second, timeout)
	require.NoError(t, err, "WaitForActionComplete failed")
	return *final
}

// --- Cleanup & Logging ---

func deferCleanupDroplet(ctx context.Context, c *client.Client, t *testing.T, dropletID int) func() {
	return func() {
		t.Logf("Cleaning up droplet %d...", dropletID)
		DeleteResource(ctx, c, t, "droplet", float64(dropletID))
	}
}

func deferCleanupImage(ctx context.Context, c *client.Client, t *testing.T, imageID float64) func() {
	return func() {
		t.Logf("Cleaning up snapshot image %.0f...", imageID)
		DeleteResource(ctx, c, t, "snapshot", imageID)
	}
}

func LogResourceCreated(t *testing.T, resourceType string, id interface{}, name, status, region string) {
	t.Logf("[Created] %s %s: Name=%s, Status=%s, Region=%s", resourceType, formatID(id), name, status, region)
}

func LogResourceDeleted(t *testing.T, resourceType string, id interface{}, err error, resp *mcp.CallToolResult) {
	if err != nil || (resp != nil && resp.IsError) {
		t.Logf("[Delete] Failed %s %s: %v", resourceType, formatID(id), err)
	} else {
		t.Logf("[Delete] Success %s %s", resourceType, formatID(id))
	}
}

func LogActionCompleted(t *testing.T, actionType string, action godo.Action) {
	t.Logf("[Action] %s Completed: ID=%d, Status=%s", actionType, action.ID, action.Status)
}

func formatID(id interface{}) string {
	switch v := id.(type) {
	case float64:
		return fmt.Sprintf("%.0f", v)
	case float32:
		return fmt.Sprintf("%.0f", v)
	case int, int32, int64, uint, uint32, uint64:
		return fmt.Sprintf("%d", v)
	default:
		return fmt.Sprintf("%v", v)
	}
}
