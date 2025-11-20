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

// CreateDroplet creates a new droplet
func (d *DropletTool) createDroplet(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := req.GetArguments()
	dropletName := args["Name"].(string)
	size := args["Size"].(string)
	imageID := args["ImageID"].(float64)
	region := args["Region"].(string)
	backup, _ := args["Backup"].(bool)         // Defaults to false
	monitoring, _ := args["Monitoring"].(bool) // Defaults to false
	ipv6, _ := args["IPv6"].(bool)
	vpcuuid, _ := args["VPCUUID"].(string)
	userdata, _ := args["UserData"].(string)
	var withAgentPtr *bool
	if raw, ok := args["WithDropletAgent"]; ok {
		if b, ok2 := raw.(bool); ok2 {
			tmp := b
			withAgentPtr = &tmp
		}
	}
	policyJSON, _ := args["BackupPolicy"].(string)

	sshKeys := parseSSHKeys(args["SSHKeys"])
	tags := parseTags(args["Tags"])

	// Volumes is optional (as slice of strings)
	var volumes []string
	if raw, ok := args["Volumes"]; ok && raw != nil {
		if v, err := parseStringArray(raw); err == nil {
			volumes = v
		}
	}

	// Parse BackupPolicy JSON if provided
	var policyReq godo.DropletBackupPolicyRequest
	var policyPtr *godo.DropletBackupPolicyRequest
	if policyJSON != "" {
		if err := json.Unmarshal([]byte(policyJSON), &policyReq); err != nil {
			return mcp.NewToolResultErrorFromErr("invalid backup policy json", err), nil
		}
		policyPtr = &policyReq
	}

	// Create the droplet
	// Convert volumes into expected godo type if available
	var createVolumes []godo.DropletCreateVolume
	for _, vid := range volumes {
		// godo.DropletCreateVolume typically uses VolumeID or similar; try to set ID field if present.
		createVolumes = append(createVolumes, godo.DropletCreateVolume{ID: vid})
	}

	// createDroplet expects pointer for WithDropletAgent
	// withAgentPtr is already prepared above (may be nil)

	dropletCreateRequest := &godo.DropletCreateRequest{
		Name:             dropletName,
		Size:             size,
		Image:            godo.DropletCreateImage{ID: int(imageID)},
		Region:           region,
		Backups:          backup,
		Monitoring:       monitoring,
		IPv6:             ipv6,
		VPCUUID:          vpcuuid,
		UserData:         userdata,
		Volumes:          createVolumes,
		WithDropletAgent: withAgentPtr,
		BackupPolicy:     policyPtr,
		SSHKeys:          sshKeys,
		Tags:             tags,
	}

	client, err := d.client(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get DigitalOcean client: %w", err)
	}

	droplet, _, err := client.Droplets.Create(ctx, dropletCreateRequest)
	if err != nil {
		return mcp.NewToolResultErrorFromErr("droplet create", err), nil
	}
	jsonDroplet, err := json.MarshalIndent(droplet, "", "  ")
	if err != nil {
		return mcp.NewToolResultErrorFromErr("json marshal", err), nil
	}
	return mcp.NewToolResultText(string(jsonDroplet)), nil
}

// deleteDroplet deletes a droplet
func (d *DropletTool) deleteDroplet(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	dropletID := req.GetArguments()["ID"].(float64)

	client, err := d.client(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get DigitalOcean client: %w", err)
	}

	_, err = client.Droplets.Delete(ctx, int(dropletID))
	if err != nil {
		return mcp.NewToolResultErrorFromErr("api error", err), nil
	}
	return mcp.NewToolResultText("Droplet deleted successfully"), nil
}

// getDropletNeighbors gets a droplet's neighbors
func (d *DropletTool) getDropletNeighbors(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	dropletID := req.GetArguments()["ID"].(float64)

	client, err := d.client(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get DigitalOcean client: %w", err)
	}

	neighbors, _, err := client.Droplets.Neighbors(ctx, int(dropletID))
	if err != nil {
		return mcp.NewToolResultErrorFromErr("api error", err), nil
	}

	jsonNeighbors, err := json.MarshalIndent(neighbors, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("marshal error: %w", err)
	}

	return mcp.NewToolResultText(string(jsonNeighbors)), nil
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

// Tools returns a list of tool functions
func (d *DropletTool) getDropletByID(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	id, ok := req.GetArguments()["ID"].(float64)
	if !ok {
		return mcp.NewToolResultError("Droplet ID is required"), nil
	}

	client, err := d.client(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get DigitalOcean client: %w", err)
	}

	droplet, _, err := client.Droplets.Get(ctx, int(id))
	if err != nil {
		return mcp.NewToolResultErrorFromErr("api error", err), nil
	}
	jsonData, err := json.MarshalIndent(droplet, "", "  ")
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

// getDroplets lists all droplets for a user
func (d *DropletTool) getDroplets(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := req.GetArguments()
	page := getNumberArg(args, "Page", 1)
	perPage := getNumberArg(args, "PerPage", 50)

	opt := &godo.ListOptions{
		Page:    int(page),
		PerPage: int(perPage),
	}

	client, err := d.client(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get DigitalOcean client: %w", err)
	}

	droplets, _, err := client.Droplets.List(ctx, opt)
	if err != nil {
		return mcp.NewToolResultErrorFromErr("api error", err), nil
	}

	jsonData, err := formatDropletSummaries(droplets)
	if err != nil {
		return nil, err
	}

	return mcp.NewToolResultText(jsonData), nil
}

func (d *DropletTool) listDropletsWithGPUs(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := req.GetArguments()
	page := getNumberArg(args, "Page", 1)
	perPage := getNumberArg(args, "PerPage", 50)

	opt := &godo.ListOptions{
		Page:    int(page),
		PerPage: int(perPage),
	}

	client, err := d.client(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get DigitalOcean client: %w", err)
	}

	droplets, _, err := client.Droplets.ListWithGPUs(ctx, opt)
	if err != nil {
		return mcp.NewToolResultErrorFromErr("api error", err), nil
	}

	jsonData, err := formatDropletSummaries(droplets)
	if err != nil {
		return nil, err
	}

	return mcp.NewToolResultText(jsonData), nil
}

func (d *DropletTool) listDropletsByName(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := req.GetArguments()
	name, ok := args["Name"].(string)
	if !ok || name == "" {
		return mcp.NewToolResultError("Name is required"), nil
	}
	page := getNumberArg(args, "Page", 1)
	perPage := getNumberArg(args, "PerPage", 50)

	opt := &godo.ListOptions{
		Page:    int(page),
		PerPage: int(perPage),
	}

	client, err := d.client(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get DigitalOcean client: %w", err)
	}

	droplets, _, err := client.Droplets.ListByName(ctx, name, opt)
	if err != nil {
		return mcp.NewToolResultErrorFromErr("api error", err), nil
	}

	jsonData, err := formatDropletSummaries(droplets)
	if err != nil {
		return nil, err
	}

	return mcp.NewToolResultText(jsonData), nil
}

func (d *DropletTool) listDropletsByTag(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := req.GetArguments()
	tag, ok := args["Tag"].(string)
	if !ok || tag == "" {
		return mcp.NewToolResultError("Tag is required"), nil
	}
	page := getNumberArg(args, "Page", 1)
	perPage := getNumberArg(args, "PerPage", 50)

	opt := &godo.ListOptions{
		Page:    int(page),
		PerPage: int(perPage),
	}

	client, err := d.client(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get DigitalOcean client: %w", err)
	}

	droplets, _, err := client.Droplets.ListByTag(ctx, tag, opt)
	if err != nil {
		return mcp.NewToolResultErrorFromErr("api error", err), nil
	}

	jsonData, err := formatDropletSummaries(droplets)
	if err != nil {
		return nil, err
	}

	return mcp.NewToolResultText(jsonData), nil
}

func (d *DropletTool) createMultipleDroplets(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := req.GetArguments()

	namesRaw, ok := args["Names"]
	if !ok {
		return mcp.NewToolResultError("Names is required"), nil
	}
	names, err := parseStringArray(namesRaw)
	if err != nil || len(names) == 0 {
		return mcp.NewToolResultError("Names must be a non-empty array of strings"), nil
	}

	size, ok := args["Size"].(string)
	if !ok || size == "" {
		return mcp.NewToolResultError("Size is required"), nil
	}
	imageID, ok := args["ImageID"].(float64)
	if !ok {
		return mcp.NewToolResultError("ImageID is required"), nil
	}
	region, ok := args["Region"].(string)
	if !ok || region == "" {
		return mcp.NewToolResultError("Region is required"), nil
	}
	backup, _ := args["Backup"].(bool)
	monitoring, _ := args["Monitoring"].(bool)
	ipv6, _ := args["IPv6"].(bool)
	vpcuuid, _ := args["VPCUUID"].(string)
	userdata, _ := args["UserData"].(string)
	var withAgentPtrMulti *bool
	if raw, ok := args["WithDropletAgent"]; ok {
		if b, ok2 := raw.(bool); ok2 {
			tmp := b
			withAgentPtrMulti = &tmp
		}
	}
	policyJSON, _ := args["BackupPolicy"].(string)

	// Parse BackupPolicy JSON if provided
	var policyReq godo.DropletBackupPolicyRequest
	var policyPtr *godo.DropletBackupPolicyRequest
	if policyJSON != "" {
		if err := json.Unmarshal([]byte(policyJSON), &policyReq); err != nil {
			return mcp.NewToolResultErrorFromErr("invalid backup policy json", err), nil
		}
		policyPtr = &policyReq
	}

	// For multi-create: WithDropletAgent pointer is prepared above (may be nil)

	reqBody := &godo.DropletMultiCreateRequest{
		Names:            names,
		Region:           region,
		Size:             size,
		Image:            godo.DropletCreateImage{ID: int(imageID)},
		Backups:          backup,
		Monitoring:       monitoring,
		IPv6:             ipv6,
		VPCUUID:          vpcuuid,
		UserData:         userdata,
		WithDropletAgent: withAgentPtrMulti,
		BackupPolicy:     policyPtr,
		SSHKeys:          parseSSHKeys(args["SSHKeys"]),
		Tags:             parseTags(args["Tags"]),
	}

	client, err := d.client(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get DigitalOcean client: %w", err)
	}

	droplets, _, err := client.Droplets.CreateMultiple(ctx, reqBody)
	if err != nil {
		return mcp.NewToolResultErrorFromErr("api error", err), nil
	}

	jsonData, err := formatDropletSummaries(droplets)
	if err != nil {
		return nil, err
	}

	return mcp.NewToolResultText(jsonData), nil
}

func (d *DropletTool) deleteDropletsByTag(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	tag, ok := req.GetArguments()["Tag"].(string)
	if !ok || tag == "" {
		return mcp.NewToolResultError("Tag is required"), nil
	}

	client, err := d.client(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get DigitalOcean client: %w", err)
	}

	if _, err := client.Droplets.DeleteByTag(ctx, tag); err != nil {
		return mcp.NewToolResultErrorFromErr("api error", err), nil
	}

	return mcp.NewToolResultText(fmt.Sprintf("Droplets with tag %q deleted successfully", tag)), nil
}

func getNumberArg(args map[string]any, key string, defaultValue float64) float64 {
	if value, ok := args[key].(float64); ok {
		return value
	}
	return defaultValue
}

func parseStringArray(raw interface{}) ([]string, error) {
	if raw == nil {
		return nil, fmt.Errorf("value must be an array of strings")
	}
	values, ok := raw.([]interface{})
	if !ok {
		return nil, fmt.Errorf("value must be an array")
	}
	result := make([]string, 0, len(values))
	for _, val := range values {
		str, ok := val.(string)
		if !ok {
			return nil, fmt.Errorf("value must contain only strings")
		}
		if str != "" {
			result = append(result, str)
		}
	}
	return result, nil
}

func parseSSHKeys(raw interface{}) []godo.DropletCreateSSHKey {
	values, ok := raw.([]interface{})
	if !ok || raw == nil {
		return nil
	}
	var sshKeys []godo.DropletCreateSSHKey
	for _, key := range values {
		switch v := key.(type) {
		case float64:
			sshKeys = append(sshKeys, godo.DropletCreateSSHKey{ID: int(v)})
		case string:
			if v != "" {
				sshKeys = append(sshKeys, godo.DropletCreateSSHKey{Fingerprint: v})
			}
		}
	}
	return sshKeys
}

func parseTags(raw interface{}) []string {
	values, ok := raw.([]interface{})
	if !ok || raw == nil {
		return nil
	}
	var tags []string
	for _, tag := range values {
		if tagStr, ok := tag.(string); ok && tagStr != "" {
			tags = append(tags, tagStr)
		}
	}
	return tags
}

func formatDropletSummaries(droplets []godo.Droplet) (string, error) {
	filtered := make([]map[string]any, len(droplets))
	for i, droplet := range droplets {
		filtered[i] = map[string]any{
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
	jsonData, err := json.MarshalIndent(filtered, "", "  ")
	if err != nil {
		return "", fmt.Errorf("marshal error: %w", err)
	}
	return string(jsonData), nil
}

// listDropletSnapshots lists snapshots for a specific droplet
func (d *DropletTool) listDropletSnapshots(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	dropletID, ok := req.GetArguments()["ID"].(float64)
	if !ok {
		return mcp.NewToolResultError("Droplet ID is required"), nil
	}

	page := getNumberArg(req.GetArguments(), "Page", 1)
	perPage := getNumberArg(req.GetArguments(), "PerPage", 50)

	opt := &godo.ListOptions{
		Page:    int(page),
		PerPage: int(perPage),
	}

	client, err := d.client(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get DigitalOcean client: %w", err)
	}

	snapshots, _, err := client.Droplets.Snapshots(ctx, int(dropletID), opt)
	if err != nil {
		return mcp.NewToolResultErrorFromErr("api error", err), nil
	}

	jsonData, err := json.MarshalIndent(snapshots, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("marshal error: %w", err)
	}

	return mcp.NewToolResultText(string(jsonData)), nil
}

// listDropletBackups lists backups for a specific droplet
func (d *DropletTool) listDropletBackups(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	dropletID, ok := req.GetArguments()["ID"].(float64)
	if !ok {
		return mcp.NewToolResultError("Droplet ID is required"), nil
	}

	page := getNumberArg(req.GetArguments(), "Page", 1)
	perPage := getNumberArg(req.GetArguments(), "PerPage", 50)

	opt := &godo.ListOptions{
		Page:    int(page),
		PerPage: int(perPage),
	}

	client, err := d.client(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get DigitalOcean client: %w", err)
	}

	backups, _, err := client.Droplets.Backups(ctx, int(dropletID), opt)
	if err != nil {
		return mcp.NewToolResultErrorFromErr("api error", err), nil
	}

	jsonData, err := json.MarshalIndent(backups, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("marshal error: %w", err)
	}

	return mcp.NewToolResultText(string(jsonData)), nil
}

// listDropletActions lists all actions for a specific droplet
func (d *DropletTool) listDropletActions(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	dropletID, ok := req.GetArguments()["ID"].(float64)
	if !ok {
		return mcp.NewToolResultError("Droplet ID is required"), nil
	}

	page := getNumberArg(req.GetArguments(), "Page", 1)
	perPage := getNumberArg(req.GetArguments(), "PerPage", 50)

	opt := &godo.ListOptions{
		Page:    int(page),
		PerPage: int(perPage),
	}

	client, err := d.client(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get DigitalOcean client: %w", err)
	}

	actions, _, err := client.Droplets.Actions(ctx, int(dropletID), opt)
	if err != nil {
		return mcp.NewToolResultErrorFromErr("api error", err), nil
	}

	jsonData, err := json.MarshalIndent(actions, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("marshal error: %w", err)
	}

	return mcp.NewToolResultText(string(jsonData)), nil
}

// getDropletBackupPolicy gets backup policy for a specific droplet
func (d *DropletTool) getDropletBackupPolicy(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	dropletID, ok := req.GetArguments()["ID"].(float64)
	if !ok {
		return mcp.NewToolResultError("Droplet ID is required"), nil
	}

	client, err := d.client(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get DigitalOcean client: %w", err)
	}

	policy, _, err := client.Droplets.GetBackupPolicy(ctx, int(dropletID))
	if err != nil {
		return mcp.NewToolResultErrorFromErr("api error", err), nil
	}
	jsonData, err := json.MarshalIndent(policy, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("marshal error: %w", err)
	}
	return mcp.NewToolResultText(string(jsonData)), nil
}

// listBackupPolicies lists all backup policies (paginated)
func (d *DropletTool) listBackupPolicies(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := req.GetArguments()
	page := getNumberArg(args, "Page", 1)
	perPage := getNumberArg(args, "PerPage", 50)

	opt := &godo.ListOptions{
		Page:    int(page),
		PerPage: int(perPage),
	}

	client, err := d.client(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get DigitalOcean client: %w", err)
	}

	policies, _, err := client.Droplets.ListBackupPolicies(ctx, opt)
	if err != nil {
		return mcp.NewToolResultErrorFromErr("api error", err), nil
	}

	jsonData, err := json.MarshalIndent(policies, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("marshal error: %w", err)
	}

	return mcp.NewToolResultText(string(jsonData)), nil
}

// listSupportedBackupPolicies lists available backup policy options
func (d *DropletTool) listSupportedBackupPolicies(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := d.client(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get DigitalOcean client: %w", err)
	}

	supported, _, err := client.Droplets.ListSupportedBackupPolicies(ctx)
	if err != nil {
		return mcp.NewToolResultErrorFromErr("api error", err), nil
	}

	jsonData, err := json.MarshalIndent(supported, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("marshal error: %w", err)
	}

	return mcp.NewToolResultText(string(jsonData)), nil
}

// listAssociatedResourcesForDeletion lists resources associated with a droplet that would be removed when deleting the droplet.
func (d *DropletTool) listAssociatedResourcesForDeletion(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	dropletID, ok := req.GetArguments()["ID"].(float64)
	if !ok {
		return mcp.NewToolResultError("Droplet ID is required"), nil
	}

	client, err := d.client(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get DigitalOcean client: %w", err)
	}

	resources, _, err := client.Droplets.ListAssociatedResourcesForDeletion(ctx, int(dropletID))
	if err != nil {
		return mcp.NewToolResultErrorFromErr("api error", err), nil
	}

	jsonData, err := json.MarshalIndent(resources, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("marshal error: %w", err)
	}

	return mcp.NewToolResultText(string(jsonData)), nil
}

func (d *DropletTool) Tools() []server.ServerTool {
	tools := []server.ServerTool{
		{
			Handler: d.createDroplet,
			Tool: mcp.NewTool("droplet-create",
				mcp.WithDescription("Create a new droplet"),
				mcp.WithString("Name", mcp.Required(), mcp.Description("Name of the droplet")),
				mcp.WithString("Size", mcp.Required(), mcp.Description("Slug of the droplet size (e.g., s-1vcpu-1gb)")),
				mcp.WithNumber("ImageID", mcp.Required(), mcp.Description("ID of the image to use")),
				mcp.WithString("Region", mcp.Required(), mcp.Description("Slug of the region (e.g., nyc3)")),
				mcp.WithBoolean("Backup", mcp.DefaultBool(false), mcp.Description("Whether to enable backups")),
				mcp.WithBoolean("Monitoring", mcp.DefaultBool(false), mcp.Description("Whether to enable monitoring")),
				mcp.WithBoolean("IPv6", mcp.DefaultBool(false), mcp.Description("Enable IPv6 networking")),
				mcp.WithString("VPCUUID", mcp.Description("VPC UUID to place the droplet into")),
				mcp.WithString("UserData", mcp.Description("Cloud-init user data to pass to droplet")),
				mcp.WithArray("Volumes", mcp.Description("Array of volume IDs to attach to the droplet")),
				mcp.WithBoolean("WithDropletAgent", mcp.DefaultBool(false), mcp.Description("Whether to enable the droplet agent")),
				mcp.WithString("BackupPolicy", mcp.Description("JSON encoded DropletBackupPolicyRequest (optional)")),
				mcp.WithArray("SSHKeys", mcp.Description("Array of SSH key IDs (numbers) or fingerprints (strings) to add to the droplet")),
				mcp.WithArray("Tags", mcp.Description("Array of tag names to apply to the droplet")),
			),
		},
		{
			Handler: d.deleteDroplet,
			Tool: mcp.NewTool("droplet-delete",
				mcp.WithDescription("Delete a droplet"),
				mcp.WithNumber("ID", mcp.Required(), mcp.Description("ID of the droplet to delete")),
			),
		},
		{
			Handler: d.enablePrivateNetworking,
			Tool: mcp.NewTool("droplet-enable-private-net",
				mcp.WithDescription("Enable private networking on a droplet"),
				mcp.WithNumber("ID", mcp.Required(), mcp.Description("ID of the droplet")),
			),
		},
		{
			Handler: d.getDropletNeighbors,
			Tool: mcp.NewTool("droplet-neighbors",
				mcp.WithDescription("List droplets that share the same physical host as the provided droplet"),
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
			Handler: d.getDropletByID,
			Tool: mcp.NewTool("droplet-get",
				mcp.WithDescription("Get a droplet by its ID"),
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
		{
			Handler: d.getDroplets,
			Tool: mcp.NewTool("droplet-list",
				mcp.WithDescription("List all droplets for the user. Supports pagination."),
				mcp.WithNumber("Page", mcp.DefaultNumber(1), mcp.Description("Page number")),
				mcp.WithNumber("PerPage", mcp.DefaultNumber(50), mcp.Description("Items per page")),
			),
		},
		{
			Handler: d.listDropletsWithGPUs,
			Tool: mcp.NewTool("droplet-list-gpus",
				mcp.WithDescription("List droplets that have GPUs attached. Supports pagination."),
				mcp.WithNumber("Page", mcp.DefaultNumber(1), mcp.Description("Page number")),
				mcp.WithNumber("PerPage", mcp.DefaultNumber(50), mcp.Description("Items per page")),
			),
		},
		{
			Handler: d.listDropletsByName,
			Tool: mcp.NewTool("droplet-list-by-name",
				mcp.WithDescription("List droplets filtered by an exact name. Supports pagination."),
				mcp.WithString("Name", mcp.Required(), mcp.Description("Exact droplet name (case-insensitive)")),
				mcp.WithNumber("Page", mcp.DefaultNumber(1), mcp.Description("Page number")),
				mcp.WithNumber("PerPage", mcp.DefaultNumber(50), mcp.Description("Items per page")),
			),
		},
		{
			Handler: d.listDropletsByTag,
			Tool: mcp.NewTool("droplet-list-by-tag",
				mcp.WithDescription("List droplets that share a given tag. Supports pagination."),
				mcp.WithString("Tag", mcp.Required(), mcp.Description("Tag of the droplets")),
				mcp.WithNumber("Page", mcp.DefaultNumber(1), mcp.Description("Page number")),
				mcp.WithNumber("PerPage", mcp.DefaultNumber(50), mcp.Description("Items per page")),
			),
		},
		{
			Handler: d.createMultipleDroplets,
			Tool: mcp.NewTool("droplet-create-multiple",
				mcp.WithDescription("Create multiple droplets with the same configuration."),
				mcp.WithArray("Names", mcp.Required(), mcp.Description("Array of droplet names")),
				mcp.WithString("Size", mcp.Required(), mcp.Description("Slug of the droplet size (e.g., s-1vcpu-1gb)")),
				mcp.WithNumber("ImageID", mcp.Required(), mcp.Description("ID of the image to use")),
				mcp.WithString("Region", mcp.Required(), mcp.Description("Slug of the region (e.g., nyc3)")),
				mcp.WithBoolean("Backup", mcp.DefaultBool(false), mcp.Description("Whether to enable backups")),
				mcp.WithBoolean("Monitoring", mcp.DefaultBool(false), mcp.Description("Whether to enable monitoring")),
				mcp.WithBoolean("IPv6", mcp.DefaultBool(false), mcp.Description("Enable IPv6 networking")),
				mcp.WithString("VPCUUID", mcp.Description("VPC UUID to place the droplets into")),
				mcp.WithString("UserData", mcp.Description("Cloud-init user data to pass to droplets")),
				mcp.WithArray("Volumes", mcp.Description("Array of volume IDs to attach to the droplets")),
				mcp.WithBoolean("WithDropletAgent", mcp.DefaultBool(false), mcp.Description("Whether to enable the droplet agent")),
				mcp.WithString("BackupPolicy", mcp.Description("JSON encoded DropletBackupPolicyRequest (optional)")),
				mcp.WithArray("SSHKeys", mcp.Description("Array of SSH key IDs or fingerprints to add to the droplets")),
				mcp.WithArray("Tags", mcp.Description("Array of tag names to apply to the droplets")),
			),
		},
		{
			Handler: d.deleteDropletsByTag,
			Tool: mcp.NewTool("droplet-delete-by-tag",
				mcp.WithDescription("Delete all droplets that share a specific tag."),
				mcp.WithString("Tag", mcp.Required(), mcp.Description("Tag of the droplets to delete")),
			),
		},
		{
			Handler: d.listDropletSnapshots,
			Tool: mcp.NewTool("droplet-snapshots",
				mcp.WithDescription("List snapshots for a specific droplet. Supports pagination."),
				mcp.WithNumber("ID", mcp.Required(), mcp.Description("Droplet ID")),
				mcp.WithNumber("Page", mcp.DefaultNumber(1), mcp.Description("Page number")),
				mcp.WithNumber("PerPage", mcp.DefaultNumber(50), mcp.Description("Items per page")),
			),
		},
		{
			Handler: d.listDropletBackups,
			Tool: mcp.NewTool("droplet-backups",
				mcp.WithDescription("List backups for a specific droplet. Supports pagination."),
				mcp.WithNumber("ID", mcp.Required(), mcp.Description("Droplet ID")),
				mcp.WithNumber("Page", mcp.DefaultNumber(1), mcp.Description("Page number")),
				mcp.WithNumber("PerPage", mcp.DefaultNumber(50), mcp.Description("Items per page")),
			),
		},
		{
			Handler: d.listAssociatedResourcesForDeletion,
			Tool: mcp.NewTool("droplet-associated-resources",
				mcp.WithDescription("List resources associated with a droplet that would be deleted if the droplet is removed."),
				mcp.WithNumber("ID", mcp.Required(), mcp.Description("Droplet ID")),
			),
		},
		{
			Handler: d.listBackupPolicies,
			Tool: mcp.NewTool("droplet-backup-policies-list",
				mcp.WithDescription("List all backup policies. Supports pagination."),
				mcp.WithNumber("Page", mcp.DefaultNumber(1), mcp.Description("Page number")),
				mcp.WithNumber("PerPage", mcp.DefaultNumber(50), mcp.Description("Items per page")),
			),
		},
		{
			Handler: d.listSupportedBackupPolicies,
			Tool: mcp.NewTool("droplet-backup-policies-supported",
				mcp.WithDescription("List supported backup policy options."),
			),
		},
		{
			Handler: d.listDropletActions,
			Tool: mcp.NewTool("droplet-actions-list",
				mcp.WithDescription("List all actions for a specific droplet. Supports pagination."),
				mcp.WithNumber("ID", mcp.Required(), mcp.Description("Droplet ID")),
				mcp.WithNumber("Page", mcp.DefaultNumber(1), mcp.Description("Page number")),
				mcp.WithNumber("PerPage", mcp.DefaultNumber(50), mcp.Description("Items per page")),
			),
		},
		{
			Handler: d.getDropletBackupPolicy,
			Tool: mcp.NewTool("droplet-backup-policy-get",
				mcp.WithDescription("Get the backup policy for a specific droplet."),
				mcp.WithNumber("ID", mcp.Required(), mcp.Description("Droplet ID")),
			),
		},
	}
	return tools
}
