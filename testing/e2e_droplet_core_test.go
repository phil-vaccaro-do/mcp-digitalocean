//go:build integration

package testing

import (
	"fmt"
	"mcp-digitalocean/internal/testhelpers"
	"testing"
	"time"

	"github.com/digitalocean/godo"
	"github.com/stretchr/testify/require"
)

func TestDropletLifecycle(t *testing.T) {
	ctx, c, gclient, teardown := setupTest(t)
	defer teardown()

	// 1. Create (Logs "new" and "active" automatically)
	droplet := CreateTestDroplet(ctx, c, t, "mcp-e2e-test")

	// 2. List & Verify
	type dropletShort struct {
		ID int `json:"id"`
	}
	droplets := callTool[[]dropletShort](ctx, c, t, "droplet-list", map[string]interface{}{"Page": 1, "PerPage": 50})

	found := false
	for _, d := range droplets {
		if d.ID == droplet.ID {
			found = true
			break
		}
	}
	require.True(t, found, "Created droplet not found in list")

	// 3. Delete
	DeleteResource(ctx, c, t, "droplet", droplet.ID)

	// 4. Confirm Deletion (Direct API)
	err := testhelpers.WaitForDropletDeleted(ctx, gclient, droplet.ID, 3*time.Second, 2*time.Minute)
	if err != nil {
		t.Logf("Warning: direct WaitForDropletDeleted failed: %v", err)
	} else {
		t.Logf("Confirmed droplet deletion via direct API")
	}
}

func TestDropletSnapshot(t *testing.T) {
	ctx, c, gclient, teardown := setupTest(t)
	defer teardown()

	droplet := CreateTestDroplet(ctx, c, t, "mcp-e2e-snapshot")
	defer deferCleanupDroplet(ctx, c, t, droplet.ID)()

	// 1. Trigger Snapshot
	snapshotName := fmt.Sprintf("snapshot-%d", time.Now().Unix())
	action := callTool[godo.Action](ctx, c, t, "snapshot-droplet", map[string]interface{}{
		"ID":   droplet.ID,
		"Name": snapshotName,
	})

	t.Logf("Snapshot initiated: %s", snapshotName)

	// 2. Wait
	WaitForActionComplete(ctx, c, t, droplet.ID, action.ID, 2*time.Minute)

	// 3. Verify Snapshot Exists
	d, err := testhelpers.WaitForDroplet(ctx, gclient, droplet.ID, func(d *godo.Droplet) bool {
		return d != nil && len(d.SnapshotIDs) > 0
	}, 3*time.Second, 2*time.Minute)

	require.NoError(t, err)
	require.NotEmpty(t, d.SnapshotIDs)

	t.Logf("Snapshot verified. Image ID: %d", d.SnapshotIDs[0])
	defer deferCleanupImage(ctx, c, t, float64(d.SnapshotIDs[0]))()
}

func TestDropletRebuildBySlug(t *testing.T) {
	ctx, c, _, teardown := setupTest(t)
	defer teardown()

	// 1. Find Ubuntu Image Slug
	type imageShort struct {
		ID   int    `json:"id"`
		Slug string `json:"slug"`
	}
	images := callTool[[]imageShort](ctx, c, t, "image-list", map[string]interface{}{"Page": 1, "PerPage": 20})

	var imageSlug string
	for _, img := range images {
		if img.Slug == "ubuntu-22-04-x64" {
			imageSlug = img.Slug
			break
		}
	}
	if imageSlug == "" && len(images) > 0 {
		imageSlug = images[0].Slug
	}
	if imageSlug == "" {
		t.Skip("No suitable image slug found")
	}

	// 2. Create Droplet
	droplet := CreateTestDroplet(ctx, c, t, "mcp-e2e-rebuild")
	defer deferCleanupDroplet(ctx, c, t, droplet.ID)()

	// 3. Rebuild
	action := callTool[godo.Action](ctx, c, t, "rebuild-droplet-by-slug", map[string]interface{}{
		"ID":        droplet.ID,
		"ImageSlug": imageSlug,
	})
	t.Logf("Rebuild initiated with slug: %s", imageSlug)

	WaitForActionComplete(ctx, c, t, droplet.ID, action.ID, 5*time.Minute)
	LogActionCompleted(t, "Rebuild", action)
}

func TestDropletRestore(t *testing.T) {
	ctx, c, _, teardown := setupTest(t)
	defer teardown()

	droplet := CreateTestDroplet(ctx, c, t, "mcp-e2e-restore")
	defer deferCleanupDroplet(ctx, c, t, droplet.ID)()

	// 1. Create Snapshot
	snapName := fmt.Sprintf("restore-snap-%d", time.Now().Unix())
	snapAction := callTool[godo.Action](ctx, c, t, "snapshot-droplet", map[string]interface{}{
		"ID":   droplet.ID,
		"Name": snapName,
	})
	WaitForActionComplete(ctx, c, t, droplet.ID, snapAction.ID, 2*time.Minute)

	// 2. Get Snapshot ID from Droplet
	refreshed := callTool[godo.Droplet](ctx, c, t, "droplet-get", map[string]interface{}{"ID": droplet.ID})
	require.NotEmpty(t, refreshed.SnapshotIDs, "Droplet should have a snapshot")

	imageID := float64(refreshed.SnapshotIDs[0])
	defer deferCleanupImage(ctx, c, t, imageID)()
	t.Logf("Restoring from Image ID: %.0f", imageID)

	// 3. Restore
	restoreAction := callTool[godo.Action](ctx, c, t, "restore-droplet", map[string]interface{}{
		"ID":      droplet.ID,
		"ImageID": imageID,
	})

	WaitForActionComplete(ctx, c, t, droplet.ID, restoreAction.ID, 2*time.Minute)
	LogActionCompleted(t, "Restore", restoreAction)
}
