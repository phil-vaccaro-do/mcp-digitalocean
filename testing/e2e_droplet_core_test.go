//go:build integration

package testing

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/digitalocean/godo"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/stretchr/testify/require"
)

// TestDropletLifecycle tests the complete droplet lifecycle: create, get, list, delete
func TestDropletLifecycle(t *testing.T) {
	ctx := context.Background()
	c := initializeClient(ctx, t)
	defer c.Close()

	// Get SSH keys for droplet creation
	sshKeys := getSSHKeys(ctx, c, t)

	// Get a suitable test image (Ubuntu LTS)
	imageID := getTestImage(ctx, c, t)

	// Create a test droplet
	dropletName := fmt.Sprintf("mcp-e2e-test-%d", time.Now().Unix())
	createResp, err := c.CallTool(ctx, mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "droplet-create",
			Arguments: map[string]interface{}{
				"Name":       dropletName,
				"Size":       "s-1vcpu-1gb",
				"ImageID":    imageID,
				"Region":     "nyc3",
				"Backup":     false,
				"Monitoring": true,
				"SSHKeys":    sshKeys,
			},
		},
	})
	require.NoError(t, err)
	require.False(t, createResp.IsError, "Failed to create droplet")

	var droplet godo.Droplet
	dropletJSON := createResp.Content[0].(mcp.TextContent).Text
	err = json.Unmarshal([]byte(dropletJSON), &droplet)
	require.NoError(t, err)
	require.NotEmpty(t, droplet.ID)
	require.Equal(t, dropletName, droplet.Name)
	t.Logf("Created droplet: ID=%d, Name=%s", droplet.ID, droplet.Name)

	// Test get droplet by ID
	getResp, err := c.CallTool(ctx, mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "droplet-get",
			Arguments: map[string]interface{}{
				"ID": float64(droplet.ID),
			},
		},
	})
	require.NoError(t, err)
	require.False(t, getResp.IsError)

	var retrievedDroplet godo.Droplet
	retrievedJSON := getResp.Content[0].(mcp.TextContent).Text
	err = json.Unmarshal([]byte(retrievedJSON), &retrievedDroplet)
	require.NoError(t, err)
	require.Equal(t, droplet.ID, retrievedDroplet.ID)
	require.Equal(t, dropletName, retrievedDroplet.Name)
	t.Logf("Retrieved droplet successfully")

	// Test list droplets (should include our new droplet)
	listResp, err := c.CallTool(ctx, mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "droplet-list",
			Arguments: map[string]interface{}{
				"Page":    1,
				"PerPage": 50,
			},
		},
	})
	require.NoError(t, err)
	require.False(t, listResp.IsError)

	var droplets []map[string]interface{}
	dropletsJSON := listResp.Content[0].(mcp.TextContent).Text
	err = json.Unmarshal([]byte(dropletsJSON), &droplets)
	require.NoError(t, err)
	require.NotEmpty(t, droplets)

	// Verify our droplet is in the list
	found := false
	for _, d := range droplets {
		if int(d["id"].(float64)) == droplet.ID {
			found = true
			break
		}
	}
	require.True(t, found, "Created droplet not found in list")
	t.Logf("Found %d droplets in list", len(droplets))

	// Cleanup - delete the droplet
	deleteResp, err := c.CallTool(ctx, mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "droplet-delete",
			Arguments: map[string]interface{}{
				"ID": float64(droplet.ID),
			},
		},
	})
	require.NoError(t, err)
	require.False(t, deleteResp.IsError)
	t.Logf("Droplet %d deleted successfully", droplet.ID)
}

// TestDropletSnapshot tests snapshot functionality
func TestDropletSnapshot(t *testing.T) {
	ctx := context.Background()
	c := initializeClient(ctx, t)
	defer c.Close()

	// Get SSH keys and test image
	sshKeys := getSSHKeys(ctx, c, t)
	imageID := getTestImage(ctx, c, t)

	// Create droplet
	dropletName := fmt.Sprintf("mcp-e2e-snapshot-%d", time.Now().Unix())
	createResp, err := c.CallTool(ctx, mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "droplet-create",
			Arguments: map[string]interface{}{
				"Name":    dropletName,
				"Size":    "s-1vcpu-1gb",
				"ImageID": imageID,
				"Region":  "nyc3",
				"SSHKeys": sshKeys,
			},
		},
	})
	require.NoError(t, err)
	require.False(t, createResp.IsError)

	var droplet godo.Droplet
	err = json.Unmarshal([]byte(createResp.Content[0].(mcp.TextContent).Text), &droplet)
	require.NoError(t, err)

	defer func() {
		_, _ = c.CallTool(ctx, mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name:      "droplet-delete",
				Arguments: map[string]interface{}{"ID": float64(droplet.ID)},
			},
		})
	}()

	// Wait for droplet to be active
	time.Sleep(30 * time.Second)

	// Take a snapshot
	snapshotName := fmt.Sprintf("snapshot-%d", time.Now().Unix())
	snapshotResp, err := c.CallTool(ctx, mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "snapshot-droplet",
			Arguments: map[string]interface{}{
				"ID":   float64(droplet.ID),
				"Name": snapshotName,
			},
		},
	})
	require.NoError(t, err)
	require.False(t, snapshotResp.IsError)

	var action godo.Action
	err = json.Unmarshal([]byte(snapshotResp.Content[0].(mcp.TextContent).Text), &action)
	require.NoError(t, err)
	require.NotEmpty(t, action.ID)
	t.Logf("Snapshot action initiated: ID=%d, Name=%s", action.ID, snapshotName)
}

// TestDropletKernels tests getting available kernels for a droplet
func TestDropletKernels(t *testing.T) {
	ctx := context.Background()
	c := initializeClient(ctx, t)
	defer c.Close()

	// Get SSH keys and test image
	sshKeys := getSSHKeys(ctx, c, t)
	imageID := getTestImage(ctx, c, t)

	// Create droplet
	dropletName := fmt.Sprintf("mcp-e2e-kernels-%d", time.Now().Unix())
	createResp, err := c.CallTool(ctx, mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "droplet-create",
			Arguments: map[string]interface{}{
				"Name":    dropletName,
				"Size":    "s-1vcpu-1gb",
				"ImageID": imageID,
				"Region":  "nyc3",
				"SSHKeys": sshKeys,
			},
		},
	})
	require.NoError(t, err)
	require.False(t, createResp.IsError)

	var droplet godo.Droplet
	err = json.Unmarshal([]byte(createResp.Content[0].(mcp.TextContent).Text), &droplet)
	require.NoError(t, err)

	defer func() {
		_, _ = c.CallTool(ctx, mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name:      "droplet-delete",
				Arguments: map[string]interface{}{"ID": float64(droplet.ID)},
			},
		})
	}()

	// Wait for droplet to be active
	time.Sleep(30 * time.Second)

	// Get kernels
	kernelsResp, err := c.CallTool(ctx, mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "droplet-kernels",
			Arguments: map[string]interface{}{
				"ID": float64(droplet.ID),
			},
		},
	})
	require.NoError(t, err)
	require.False(t, kernelsResp.IsError)

	var kernels []godo.Kernel
	err = json.Unmarshal([]byte(kernelsResp.Content[0].(mcp.TextContent).Text), &kernels)
	require.NoError(t, err)
	t.Logf("Found %d kernels for droplet", len(kernels))
}

// TestDropletRebuildBySlug tests rebuilding a droplet using an image slug
func TestDropletRebuildBySlug(t *testing.T) {
	ctx := context.Background()
	c := initializeClient(ctx, t)
	defer c.Close()

	// Get SSH keys for droplet creation
	sshKeys := getSSHKeys(ctx, c, t)

	// Get images with slug
	// Get images to find a valid slug
	imagesResp, err := c.CallTool(ctx, mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name:      "image-list",
			Arguments: map[string]interface{}{"Page": 1, "PerPage": 10},
		},
	})
	require.NoError(t, err)
	require.False(t, imagesResp.IsError)

	var images []map[string]interface{}
	err = json.Unmarshal([]byte(imagesResp.Content[0].(mcp.TextContent).Text), &images)
	require.NoError(t, err)
	require.NotEmpty(t, images)

	// Prefer the mainstream Ubuntu 22.04 LTS image slug if available
	var imageSlug string
	var imageID float64
	for _, img := range images {
		if slug, ok := img["slug"].(string); ok && slug == "ubuntu-22-04-x64" {
			imageSlug = slug
			imageID = img["id"].(float64)
			break
		}
	}
	// Fallback: use first available image with a slug
	if imageSlug == "" {
		for _, img := range images {
			if slug, ok := img["slug"].(string); ok && slug != "" {
				imageSlug = slug
				imageID = img["id"].(float64)
				break
			}
		}
	}
	if imageSlug == "" {
		t.Skip("No image with slug found; skipping rebuild-by-slug test")
	}
	t.Logf("Using image slug: %s (ID: %v)", imageSlug, imageID)

	// Create droplet
	dropletName := fmt.Sprintf("mcp-e2e-rebuild-slug-%d", time.Now().Unix())
	createResp, err := c.CallTool(ctx, mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "droplet-create",
			Arguments: map[string]interface{}{
				"Name":    dropletName,
				"Size":    "s-1vcpu-1gb",
				"ImageID": imageID,
				"Region":  "nyc3",
				"SSHKeys": sshKeys,
			},
		},
	})
	require.NoError(t, err)
	require.False(t, createResp.IsError)

	var droplet godo.Droplet
	err = json.Unmarshal([]byte(createResp.Content[0].(mcp.TextContent).Text), &droplet)
	require.NoError(t, err)

	defer func() {
		_, _ = c.CallTool(ctx, mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name:      "droplet-delete",
				Arguments: map[string]interface{}{"ID": float64(droplet.ID)},
			},
		})
	}()

	// Wait for droplet to be active
	time.Sleep(30 * time.Second)

	// Rebuild using image slug
	resp, err := c.CallTool(ctx, mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "rebuild-droplet-by-slug",
			Arguments: map[string]interface{}{
				"ID":        float64(droplet.ID),
				"ImageSlug": imageSlug,
			},
		},
	})
	require.NoError(t, err)
	require.False(t, resp.IsError)

	var action godo.Action
	err = json.Unmarshal([]byte(resp.Content[0].(mcp.TextContent).Text), &action)
	require.NoError(t, err)
	require.NotEmpty(t, action.ID)
	t.Logf("Rebuild by slug action initiated: ID=%d, ImageSlug=%s", action.ID, imageSlug)
}

// TestDropletRestore tests restoring a droplet from a snapshot
func TestDropletRestore(t *testing.T) {
	ctx := context.Background()
	c := initializeClient(ctx, t)
	defer c.Close()

	// Get SSH keys and test image
	sshKeys := getSSHKeys(ctx, c, t)
	imageID := getTestImage(ctx, c, t)

	// Create droplet
	dropletName := fmt.Sprintf("mcp-e2e-restore-%d", time.Now().Unix())
	createResp, err := c.CallTool(ctx, mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "droplet-create",
			Arguments: map[string]interface{}{
				"Name":    dropletName,
				"Size":    "s-1vcpu-1gb",
				"ImageID": imageID,
				"Region":  "nyc3",
				"SSHKeys": sshKeys,
			},
		},
	})
	require.NoError(t, err)
	require.False(t, createResp.IsError)

	var droplet godo.Droplet
	err = json.Unmarshal([]byte(createResp.Content[0].(mcp.TextContent).Text), &droplet)
	require.NoError(t, err)

	defer func() {
		_, _ = c.CallTool(ctx, mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name:      "droplet-delete",
				Arguments: map[string]interface{}{"ID": float64(droplet.ID)},
			},
		})
	}()

	// Wait for droplet to be active
	time.Sleep(30 * time.Second)

	// Take a snapshot
	snapshotName := fmt.Sprintf("restore-snapshot-%d", time.Now().Unix())
	snapshotResp, err := c.CallTool(ctx, mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "snapshot-droplet",
			Arguments: map[string]interface{}{
				"ID":   float64(droplet.ID),
				"Name": snapshotName,
			},
		},
	})
	require.NoError(t, err)
	require.False(t, snapshotResp.IsError)

	var snapshotAction godo.Action
	err = json.Unmarshal([]byte(snapshotResp.Content[0].(mcp.TextContent).Text), &snapshotAction)
	require.NoError(t, err)
	t.Logf("Snapshot created: %s", snapshotName)

	// Wait for snapshot to complete
	time.Sleep(60 * time.Second)

	// Get the snapshot image ID - refresh droplet to see snapshots
	getResp, err := c.CallTool(ctx, mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name:      "droplet-get",
			Arguments: map[string]interface{}{"ID": float64(droplet.ID)},
		},
	})
	require.NoError(t, err)
	require.False(t, getResp.IsError)

	var refreshedDroplet godo.Droplet
	err = json.Unmarshal([]byte(getResp.Content[0].(mcp.TextContent).Text), &refreshedDroplet)
	require.NoError(t, err)
	require.NotEmpty(t, refreshedDroplet.SnapshotIDs, "Droplet should have at least one snapshot")

	snapshotImageID := float64(refreshedDroplet.SnapshotIDs[0])
	t.Logf("Using snapshot image ID: %v", snapshotImageID)

	// Restore from snapshot
	restoreResp, err := c.CallTool(ctx, mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "restore-droplet",
			Arguments: map[string]interface{}{
				"ID":      float64(droplet.ID),
				"ImageID": snapshotImageID,
			},
		},
	})
	require.NoError(t, err)
	require.False(t, restoreResp.IsError)

	var restoreAction godo.Action
	err = json.Unmarshal([]byte(restoreResp.Content[0].(mcp.TextContent).Text), &restoreAction)
	require.NoError(t, err)
	require.NotEmpty(t, restoreAction.ID)
	t.Logf("Restore action initiated: ID=%d, ImageID=%v", restoreAction.ID, snapshotImageID)
}
