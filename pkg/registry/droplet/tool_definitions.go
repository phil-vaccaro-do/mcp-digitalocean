package droplet

import (
	"context"
	"fmt"

	"github.com/digitalocean/godo"
)

// dropletListConfig returns the configuration for listing droplets
func dropletListConfig() *ToolConfig {
	return &ToolConfig{
		Name:        "droplet-list",
		Description: "List all droplets for the user. Supports pagination.",
		Arguments: []ArgumentConfig{
			{
				Name:         "Page",
				Type:         ArgumentTypeNumber,
				Description:  "Page number",
				Required:     false,
				DefaultValue: 1.0,
			},
			{
				Name:         "PerPage",
				Type:         ArgumentTypeNumber,
				Description:  "Items per page",
				Required:     false,
				DefaultValue: 50.0,
			},
		},
		Handler: handleDropletList,
	}
}

// dropletGetConfig returns the configuration for getting a droplet by ID
func dropletGetConfig() *ToolConfig {
	return &ToolConfig{
		Name:        "droplet-get",
		Description: "Get a droplet by its ID",
		Arguments: []ArgumentConfig{
			{
				Name:        "ID",
				Type:        ArgumentTypeNumber,
				Description: "Droplet ID",
				Required:    true,
			},
		},
		Handler: handleDropletGet,
	}
}

// dropletCreateConfig returns the configuration for creating a droplet
func dropletCreateConfig() *ToolConfig {
	return &ToolConfig{
		Name:        "droplet-create",
		Description: "Create a new droplet",
		Arguments: []ArgumentConfig{
			{
				Name:        "Name",
				Type:        ArgumentTypeString,
				Description: "Name of the droplet",
				Required:    true,
			},
			{
				Name:        "Size",
				Type:        ArgumentTypeString,
				Description: "Slug of the droplet size (e.g., s-1vcpu-1gb)",
				Required:    true,
			},
			{
				Name:        "ImageID",
				Type:        ArgumentTypeNumber,
				Description: "ID of the image to use",
				Required:    true,
			},
			{
				Name:        "Region",
				Type:        ArgumentTypeString,
				Description: "Slug of the region (e.g., nyc3)",
				Required:    true,
			},
			{
				Name:         "Backup",
				Type:         ArgumentTypeBoolean,
				Description:  "Whether to enable backups",
				Required:     false,
				DefaultValue: false,
			},
			{
				Name:         "Monitoring",
				Type:         ArgumentTypeBoolean,
				Description:  "Whether to enable monitoring",
				Required:     false,
				DefaultValue: false,
			},
			{
				Name:        "SSHKeys",
				Type:        ArgumentTypeArray,
				Description: "Array of SSH key IDs (numbers) or fingerprints (strings) to add to the droplet",
				Required:    false,
			},
			{
				Name:        "Tags",
				Type:        ArgumentTypeArray,
				Description: "Array of tag names to apply to the droplet",
				Required:    false,
			},
		},
		Handler: handleDropletCreate,
	}
}

// dropletDeleteConfig returns the configuration for deleting a droplet
func dropletDeleteConfig() *ToolConfig {
	return &ToolConfig{
		Name:        "droplet-delete",
		Description: "Delete a droplet",
		Arguments: []ArgumentConfig{
			{
				Name:        "ID",
				Type:        ArgumentTypeNumber,
				Description: "ID of the droplet to delete",
				Required:    true,
			},
		},
		Handler: handleDropletDelete,
	}
}

// dropletNeighborsConfig returns the configuration for getting droplet neighbors
func dropletNeighborsConfig() *ToolConfig {
	return &ToolConfig{
		Name:        "droplet-neighbors",
		Description: "Get a droplet's neighbors",
		Arguments: []ArgumentConfig{
			{
				Name:        "ID",
				Type:        ArgumentTypeNumber,
				Description: "ID of the droplet",
				Required:    true,
			},
		},
		Handler: handleDropletNeighbors,
	}
}

// handleDropletList handles listing droplets
func handleDropletList(ctx context.Context, client *godo.Client, args map[string]interface{}) (interface{}, error) {
	page := GetArgumentNumber(args, "Page")
	if page == 0 {
		page = 1
	}
	perPage := GetArgumentNumber(args, "PerPage")
	if perPage == 0 {
		perPage = 50
	}

	opt := &godo.ListOptions{
		Page:    page,
		PerPage: perPage,
	}

	droplets, _, err := client.Droplets.List(ctx, opt)
	if err != nil {
		return nil, fmt.Errorf("api error: %w", err)
	}

	// Return filtered droplet data
	filteredDroplets := make([]map[string]any, len(droplets))
	for i, droplet := range droplets {
		filteredDroplets[i] = map[string]any{
			"id":                 droplet.ID,
			"name":               droplet.Name,
			"memory":             droplet.Memory,
			"vcpus":              droplet.Vcpus,
			"disk":               droplet.Disk,
			"region":             droplet.Region,
			"image":              droplet.Image,
			"size":               droplet.Size,
			"size_slug":          droplet.SizeSlug,
			"backup_ids":         droplet.BackupIDs,
			"next_backup_window": droplet.NextBackupWindow,
			"snapshot_ids":       droplet.SnapshotIDs,
			"features":           droplet.Features,
			"locked":             droplet.Locked,
			"status":             droplet.Status,
			"networks":           droplet.Networks,
			"created_at":         droplet.Created,
			"kernel":             droplet.Kernel,
			"tags":               droplet.Tags,
			"volume_ids":         droplet.VolumeIDs,
			"vpc_uuid":           droplet.VPCUUID,
		}
	}

	return filteredDroplets, nil
}

// handleDropletGet handles getting a droplet by ID
func handleDropletGet(ctx context.Context, client *godo.Client, args map[string]interface{}) (interface{}, error) {
	id := GetArgumentNumber(args, "ID")
	if id == 0 {
		return nil, fmt.Errorf("Droplet ID is required")
	}

	droplet, _, err := client.Droplets.Get(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("api error: %w", err)
	}

	return droplet, nil
}

// handleDropletCreate handles creating a new droplet
func handleDropletCreate(ctx context.Context, client *godo.Client, args map[string]interface{}) (interface{}, error) {
	dropletName := GetArgumentString(args, "Name")
	size := GetArgumentString(args, "Size")
	imageID := GetArgumentNumber(args, "ImageID")
	region := GetArgumentString(args, "Region")
	backup := GetArgumentBoolean(args, "Backup")
	monitoring := GetArgumentBoolean(args, "Monitoring")

	// Handle SSH keys if provided
	var sshKeys []godo.DropletCreateSSHKey
	if sshKeysRaw := GetArgumentArray(args, "SSHKeys"); sshKeysRaw != nil {
		for _, key := range sshKeysRaw {
			switch v := key.(type) {
			case float64:
				sshKeys = append(sshKeys, godo.DropletCreateSSHKey{ID: int(v)})
			case string:
				sshKeys = append(sshKeys, godo.DropletCreateSSHKey{Fingerprint: v})
			}
		}
	}

	// Handle tags if provided
	var tags []string
	if tagsRaw := GetArgumentArray(args, "Tags"); tagsRaw != nil {
		for _, tag := range tagsRaw {
			if tagStr, ok := tag.(string); ok {
				tags = append(tags, tagStr)
			}
		}
	}

	// Create the droplet
	dropletCreateRequest := &godo.DropletCreateRequest{
		Name:       dropletName,
		Size:       size,
		Image:      godo.DropletCreateImage{ID: imageID},
		Region:     region,
		Backups:    backup,
		Monitoring: monitoring,
		SSHKeys:    sshKeys,
		Tags:       tags,
	}

	droplet, _, err := client.Droplets.Create(ctx, dropletCreateRequest)
	if err != nil {
		return nil, fmt.Errorf("droplet create: %w", err)
	}

	return droplet, nil
}

// handleDropletDelete handles deleting a droplet
func handleDropletDelete(ctx context.Context, client *godo.Client, args map[string]interface{}) (interface{}, error) {
	dropletID := GetArgumentNumber(args, "ID")
	if dropletID == 0 {
		return nil, fmt.Errorf("Droplet ID is required")
	}

	_, err := client.Droplets.Delete(ctx, dropletID)
	if err != nil {
		return nil, fmt.Errorf("api error: %w", err)
	}

	return "Droplet deleted successfully", nil
}

// handleDropletNeighbors handles getting a droplet's neighbors
func handleDropletNeighbors(ctx context.Context, client *godo.Client, args map[string]interface{}) (interface{}, error) {
	dropletID := GetArgumentNumber(args, "ID")
	if dropletID == 0 {
		return nil, fmt.Errorf("Droplet ID is required")
	}

	neighbors, _, err := client.Droplets.Neighbors(ctx, dropletID)
	if err != nil {
		return nil, fmt.Errorf("api error: %w", err)
	}

	return neighbors, nil
}
