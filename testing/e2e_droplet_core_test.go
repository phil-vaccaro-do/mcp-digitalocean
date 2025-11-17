//go:build integration

package testing

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/digitalocean/godo"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/stretchr/testify/require"
)

func verboseDump(t *testing.T, prefix string, v interface{}) {
	if os.Getenv("E2E_VERBOSE") == "true" {
		t.Logf("%s: %+v", prefix, v)
	} else {
		t.Logf("%s: (suppressed). Set E2E_VERBOSE=true for details", prefix)
	}
}

func TestDropletLifecycle(t *testing.T) {
	ctx := context.Background()
	c := initializeClient(ctx, t)
	defer c.Close()

	droplet := CreateTestDroplet(ctx, c, t, "mcp-e2e-test")
	LogResourceCreated(t, "droplet", droplet.ID, droplet.Name, droplet.Status, droplet.Region.Slug)

	require.Equal(t, "active", droplet.Status)
	t.Logf("Retrieved droplet successfully")

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
	if listResp.IsError {
		if len(listResp.Content) > 0 {
			verboseDump(t, "droplet-list error content", listResp.Content)
		} else {
			t.Logf("droplet-list returned error: %v", listResp)
		}
	}
	require.False(t, listResp.IsError)

	var droplets []map[string]interface{}
	dropletsJSON := listResp.Content[0].(mcp.TextContent).Text
	err = json.Unmarshal([]byte(dropletsJSON), &droplets)
	require.NoError(t, err)
	require.NotEmpty(t, droplets)

	found := false
	for _, d := range droplets {
		if int(d["id"].(float64)) == droplet.ID {
			found = true
			break
		}
	}
	require.True(t, found, "Created droplet not found in list")
	LogResourceList(t, "droplet", "in list", droplets)

	listResp, err = c.CallTool(ctx, mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "droplet-list",
			Arguments: map[string]interface{}{
				"Page":    1,
				"PerPage": 50,
			},
		},
	})
	require.NoError(t, err)
	if listResp.IsError {
		if len(listResp.Content) > 0 {
			verboseDump(t, "droplet-list pre-delete content", listResp.Content)
		} else {
			t.Logf("droplet-list returned error before deletion: %v", listResp)
		}
	}
	require.False(t, listResp.IsError)
	resources := ListResources(ctx, c, t, "droplet", "before deletion", 1, 50)
	LogResourceList(t, "droplet", "before deletion", resources)

	deleteResp, err := c.CallTool(ctx, mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "droplet-delete",
			Arguments: map[string]interface{}{
				"ID": float64(droplet.ID),
			},
		},
	})
	require.NoError(t, err)
	LogResourceDeleted(t, "droplet", droplet.ID, err, deleteResp)
}

func TestDropletSnapshot(t *testing.T) {
	ctx := context.Background()
	c := initializeClient(ctx, t)
	defer c.Close()

	droplet := CreateTestDroplet(ctx, c, t, "mcp-e2e-snapshot")
	defer func() {
		resources := ListResources(ctx, c, t, "droplet", "before deletion", 1, 50)
		LogResourceList(t, "droplet", "before deletion", resources)
		DeleteResource(ctx, c, t, "droplet", float64(droplet.ID))
	}()
	LogResourceCreated(t, "droplet", droplet.ID, droplet.Name, droplet.Status, droplet.Region.Slug)

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
	if snapshotResp.IsError {
		if len(snapshotResp.Content) > 0 {
			verboseDump(t, "snapshot-droplet error content", snapshotResp.Content)
		} else {
			t.Logf("snapshot-droplet returned error: %v", snapshotResp)
		}
	}
	require.False(t, snapshotResp.IsError)

	var action godo.Action
	err = json.Unmarshal([]byte(snapshotResp.Content[0].(mcp.TextContent).Text), &action)
	require.NoError(t, err)
	require.NotEmpty(t, action.ID)
	t.Logf("Snapshot action initiated: ID=%s, Name=%s", formatID(action.ID), snapshotName)
	completedAction := WaitForActionComplete(ctx, c, t, action.ID, 2*time.Minute)
	LogActionCompleted(t, "Snapshot", completedAction)
}

func TestDropletRebuildBySlug(t *testing.T) {
	ctx := context.Background()
	c := initializeClient(ctx, t)
	defer c.Close()

	imagesResp, err := c.CallTool(ctx, mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name:      "image-list",
			Arguments: map[string]interface{}{"Page": 1, "PerPage": 10},
		},
	})
	require.NoError(t, err)
	if imagesResp.IsError {
		if len(imagesResp.Content) > 0 {
			verboseDump(t, "image-list error content", imagesResp.Content)
		} else {
			t.Logf("image-list returned error: %v", imagesResp)
		}
	}
	require.False(t, imagesResp.IsError)

	var images []map[string]interface{}
	err = json.Unmarshal([]byte(imagesResp.Content[0].(mcp.TextContent).Text), &images)
	require.NoError(t, err)
	require.NotEmpty(t, images)

	var imageSlug string
	var imageID float64
	for _, img := range images {
		if slug, ok := img["slug"].(string); ok && slug == "ubuntu-22-04-x64" {
			imageSlug = slug
			imageID = img["id"].(float64)
			break
		}
	}
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
	t.Logf("Using image slug: %s (ID: %s)", imageSlug, formatID(imageID))

	droplet := CreateTestDroplet(ctx, c, t, "mcp-e2e-rebuild-slug")
	LogResourceCreated(t, "droplet", droplet.ID, droplet.Name, droplet.Status, droplet.Region.Slug)
	defer func() {
		resources := ListResources(ctx, c, t, "droplet", "before deletion", 1, 50)
		LogResourceList(t, "droplet", "before deletion", resources)
		DeleteResource(ctx, c, t, "droplet", float64(droplet.ID))
	}()

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
	if resp.IsError {
		if len(resp.Content) > 0 {
			verboseDump(t, "rebuild-droplet-by-slug error content", resp.Content)
		} else {
			t.Logf("rebuild-droplet-by-slug returned error: %v", resp)
		}
	}
	require.False(t, resp.IsError)

	var action godo.Action
	err = json.Unmarshal([]byte(resp.Content[0].(mcp.TextContent).Text), &action)
	require.NoError(t, err)
	require.NotEmpty(t, action.ID)
	t.Logf("Rebuild by slug action initiated: ID=%s, ImageSlug=%s", formatID(action.ID), imageSlug)
	completedAction := WaitForActionComplete(ctx, c, t, action.ID, 2*time.Minute)
	LogActionCompleted(t, "Rebuild", completedAction)
}

func TestDropletRestore(t *testing.T) {
	ctx := context.Background()
	c := initializeClient(ctx, t)
	defer c.Close()

	droplet := CreateTestDroplet(ctx, c, t, "mcp-e2e-restore")
	LogResourceCreated(t, "droplet", droplet.ID, droplet.Name, droplet.Status, droplet.Region.Slug)
	defer func() {
		resources := ListResources(ctx, c, t, "droplet", "before deletion", 1, 50)
		LogResourceList(t, "droplet", "before deletion", resources)
		DeleteResource(ctx, c, t, "droplet", float64(droplet.ID))
	}()

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
	if snapshotResp.IsError {
		if len(snapshotResp.Content) > 0 {
			verboseDump(t, "snapshot-droplet error content", snapshotResp.Content)
		} else {
			t.Logf("snapshot-droplet returned error: %v", snapshotResp)
		}
	}
	require.False(t, snapshotResp.IsError)

	var snapshotAction godo.Action
	err = json.Unmarshal([]byte(snapshotResp.Content[0].(mcp.TextContent).Text), &snapshotAction)
	require.NoError(t, err)
	t.Logf("Snapshot created: %s", snapshotName)

	_ = WaitForActionComplete(ctx, c, t, snapshotAction.ID, 2*time.Minute)

	getResp, err := c.CallTool(ctx, mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name:      "droplet-get",
			Arguments: map[string]interface{}{"ID": float64(droplet.ID)},
		},
	})
	require.NoError(t, err)
	if getResp.IsError {
		if len(getResp.Content) > 0 {
			verboseDump(t, "droplet-get error content", getResp.Content)
		} else {
			t.Logf("droplet-get returned error: %v", getResp)
		}
	}
	require.False(t, getResp.IsError)

	var refreshedDroplet godo.Droplet
	err = json.Unmarshal([]byte(getResp.Content[0].(mcp.TextContent).Text), &refreshedDroplet)
	require.NoError(t, err)
	require.NotEmpty(t, refreshedDroplet.SnapshotIDs, "Droplet should have at least one snapshot")

	snapshotImageID := float64(refreshedDroplet.SnapshotIDs[0])
	t.Logf("Refreshed droplet: ID=%s, Status=%s, Name=%s, Snapshots=%v", formatID(refreshedDroplet.ID), refreshedDroplet.Status, refreshedDroplet.Name, refreshedDroplet.SnapshotIDs)
	t.Logf("Using snapshot image ID: %s", formatID(snapshotImageID))

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
	if restoreResp.IsError {
		if len(restoreResp.Content) > 0 {
			verboseDump(t, "restore-droplet error content", restoreResp.Content)
		} else {
			t.Logf("restore-droplet returned error: %v", restoreResp)
		}
	}
	require.False(t, restoreResp.IsError)

	var restoreAction godo.Action
	err = json.Unmarshal([]byte(restoreResp.Content[0].(mcp.TextContent).Text), &restoreAction)
	require.NoError(t, err)
	require.NotEmpty(t, restoreAction.ID)
	t.Logf("Restore action initiated: ID=%s, ImageID=%s", formatID(restoreAction.ID), formatID(snapshotImageID))
	completedAction := WaitForActionComplete(ctx, c, t, restoreAction.ID, 2*time.Minute)
	LogActionCompleted(t, "Restore", completedAction)
}
