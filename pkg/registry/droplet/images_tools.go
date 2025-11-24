package droplet

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/digitalocean/godo"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

const (
	defaultImagesPageSize = 50
	defaultImagesPage     = 1
)

// ImageTool provides tool-based handlers for DigitalOcean images.
type ImageTool struct {
	client func(ctx context.Context) (*godo.Client, error)
}

// NewImageTool creates a new ImageTool instance.
func NewImageTool(client func(ctx context.Context) (*godo.Client, error)) *ImageTool {
	return &ImageTool{client: client}
}

// listImages lists images with pagination and optional type filtering.
func (i *ImageTool) listImages(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	page, ok := req.GetArguments()["Page"].(float64)
	if !ok {
		page = defaultImagesPage
	}
	perPage, ok := req.GetArguments()["PerPage"].(float64)
	if !ok {
		perPage = defaultImagesPageSize
	}
	imageType, _ := req.GetArguments()["Type"].(string)

	opt := &godo.ListOptions{
		Page:    int(page),
		PerPage: int(perPage),
	}

	client, err := i.client(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get DigitalOcean client: %w", err)
	}

	var images []godo.Image
	var apiErr error

	// Dispatch based on requested image type
	switch imageType {
	case "distribution":
		images, _, apiErr = client.Images.ListDistribution(ctx, opt)
	case "application":
		images, _, apiErr = client.Images.ListApplication(ctx, opt)
	case "user":
		images, _, apiErr = client.Images.ListUser(ctx, opt)
	default:
		// Default to listing all if unspecified, or distribution if that fits your default use-case
		// Using List() to get everything matches standard "list" expectations best
		images, _, apiErr = client.Images.List(ctx, opt)
	}

	if apiErr != nil {
		return mcp.NewToolResultErrorFromErr("api error", apiErr), nil
	}

	// Create a simplified view or return full object.
	// Returning mapped structure to match other tools' verbosity.
	filteredImages := make([]map[string]any, len(images))
	for idx, image := range images {
		filteredImages[idx] = map[string]any{
			"id":            image.ID,
			"name":          image.Name,
			"slug":          image.Slug,
			"distribution":  image.Distribution,
			"type":          image.Type,
			"public":        image.Public,
			"regions":       image.Regions,
			"created_at":    image.Created,
			"min_disk_size": image.MinDiskSize,
		}
	}

	jsonData, err := json.MarshalIndent(filteredImages, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("marshal error: %w", err)
	}

	return mcp.NewToolResultText(string(jsonData)), nil
}

func (i *ImageTool) getImageByID(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	id, ok := req.GetArguments()["ID"].(float64)
	if !ok {
		return mcp.NewToolResultError("Image ID is required"), nil
	}

	client, err := i.client(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get DigitalOcean client: %w", err)
	}

	image, _, err := client.Images.GetByID(ctx, int(id))
	if err != nil {
		return mcp.NewToolResultErrorFromErr("api error", err), nil
	}

	jsonData, err := json.MarshalIndent(image, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("marshal error: %w", err)
	}

	return mcp.NewToolResultText(string(jsonData)), nil
}

// updateImage updates an image's name.
func (i *ImageTool) updateImage(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	id, ok := req.GetArguments()["ID"].(float64)
	if !ok {
		return mcp.NewToolResultError("Image ID is required"), nil
	}
	name, ok := req.GetArguments()["Name"].(string)
	if !ok || name == "" {
		return mcp.NewToolResultError("Name is required"), nil
	}

	client, err := i.client(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get DigitalOcean client: %w", err)
	}

	updateReq := &godo.ImageUpdateRequest{
		Name: name,
	}

	image, _, err := client.Images.Update(ctx, int(id), updateReq)
	if err != nil {
		return mcp.NewToolResultErrorFromErr("api error", err), nil
	}

	jsonData, err := json.MarshalIndent(image, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("marshal error: %w", err)
	}

	return mcp.NewToolResultText(string(jsonData)), nil
}

// deleteImage deletes an image/snapshot by its numeric ID.
func (i *ImageTool) deleteImage(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	id, ok := req.GetArguments()["ID"].(float64)
	if !ok {
		return mcp.NewToolResultError("ID is required"), nil
	}

	client, err := i.client(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get DigitalOcean client: %w", err)
	}

	_, err = client.Images.Delete(ctx, int(id))
	if err != nil {
		return mcp.NewToolResultErrorFromErr("api error", err), nil
	}

	return mcp.NewToolResultText("Image deleted successfully"), nil
}

// Tools returns the list of server tools for images.
func (i *ImageTool) Tools() []server.ServerTool {
	return []server.ServerTool{
		{
			Handler: i.listImages,
			Tool: mcp.NewTool(
				"image-list",
				mcp.WithDescription("List available images (snapshots, backups, distributions, applications)."),
				mcp.WithNumber("Page", mcp.DefaultNumber(defaultImagesPage), mcp.Description("Page number")),
				mcp.WithNumber("PerPage", mcp.DefaultNumber(defaultImagesPageSize), mcp.Description("Items per page")),
				mcp.WithString("Type", mcp.Description("Filter by type: 'distribution', 'application', 'user' (snapshots/backups). If omitted, lists all.")),
			),
		},
		{
			Handler: i.getImageByID,
			Tool: mcp.NewTool(
				"image-get",
				mcp.WithDescription("Get a specific image by its numeric ID."),
				mcp.WithNumber("ID", mcp.Required(), mcp.Description("Image ID")),
			),
		},
		{
			Handler: i.updateImage,
			Tool: mcp.NewTool(
				"image-update",
				mcp.WithDescription("Update an image's name."),
				mcp.WithNumber("ID", mcp.Required(), mcp.Description("Image ID")),
				mcp.WithString("Name", mcp.Required(), mcp.Description("New name for the image")),
			),
		},
		{
			Handler: i.deleteImage,
			Tool: mcp.NewTool(
				"image-delete",
				mcp.WithDestructiveHintAnnotation(true),
				mcp.WithDescription("Delete an image or snapshot."),
				mcp.WithNumber("ID", mcp.Required(), mcp.Description("ID of the image to delete")),
			),
		},
	}
}
