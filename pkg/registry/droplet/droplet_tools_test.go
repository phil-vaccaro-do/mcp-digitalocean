package droplet

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/digitalocean/godo"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func setupDropletToolWithMocks(droplets *MockDropletsService, actions *MockDropletActionsService) *DropletTool {
	client := func(ctx context.Context) (*godo.Client, error) {
		return &godo.Client{
			Droplets:       droplets,
			DropletActions: actions,
		}, nil
	}
	return NewDropletTool(client)
}

func TestDropletTool_createDroplet(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	testDroplet := &godo.Droplet{
		ID:   123,
		Name: "test-droplet",
	}
	tests := []struct {
		name        string
		args        map[string]any
		mockSetup   func(*MockDropletsService)
		expectError bool
	}{
		{
			name: "Successful create",
			args: map[string]any{
				"Name":       "test-droplet",
				"Size":       "s-1vcpu-1gb",
				"ImageID":    float64(456),
				"Region":     "nyc1",
				"Backup":     true,
				"Monitoring": false,
			},
			mockSetup: func(m *MockDropletsService) {
				m.EXPECT().
					Create(gomock.Any(), &godo.DropletCreateRequest{
						Name:       "test-droplet",
						Region:     "nyc1",
						Size:       "s-1vcpu-1gb",
						Image:      godo.DropletCreateImage{ID: 456},
						Backups:    true,
						Monitoring: false,
					}).
					Return(testDroplet, nil, nil).
					Times(1)
			},
		},
		{
			name: "API error",
			args: map[string]any{
				"Name":       "fail-droplet",
				"Size":       "s-1vcpu-1gb",
				"ImageID":    float64(789),
				"Region":     "nyc3",
				"Backup":     false,
				"Monitoring": true,
			},
			mockSetup: func(m *MockDropletsService) {
				m.EXPECT().
					Create(gomock.Any(), &godo.DropletCreateRequest{
						Name:       "fail-droplet",
						Region:     "nyc3",
						Size:       "s-1vcpu-1gb",
						Image:      godo.DropletCreateImage{ID: 789},
						Backups:    false,
						Monitoring: true,
					}).
					Return(nil, nil, errors.New("api error")).
					Times(1)
			},
			expectError: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockDroplets := NewMockDropletsService(ctrl)
			mockActions := NewMockDropletActionsService(ctrl)
			if tc.mockSetup != nil {
				tc.mockSetup(mockDroplets)
			}
			tool := setupDropletToolWithMocks(mockDroplets, mockActions)
			req := mcp.CallToolRequest{Params: mcp.CallToolParams{Arguments: tc.args}}
			resp, err := tool.createDroplet(context.Background(), req)
			if tc.expectError {
				require.NotNil(t, resp)
				require.True(t, resp.IsError)
				return
			}
			require.NoError(t, err)
			require.NotNil(t, resp)
			require.False(t, resp.IsError)
		})
	}
}

func TestDropletTool_getDropletByID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	testDroplet := &godo.Droplet{
		ID:   123,
		Name: "test-droplet",
	}
	tests := []struct {
		name        string
		args        map[string]any
		mockSetup   func(*MockDropletsService)
		expectError bool
	}{
		{
			name: "Successful get",
			args: map[string]any{"ID": float64(123)},
			mockSetup: func(m *MockDropletsService) {
				m.EXPECT().
					Get(gomock.Any(), 123).
					Return(testDroplet, nil, nil).
					Times(1)
			},
		},
		{
			name: "API error",
			args: map[string]any{"ID": float64(456)},
			mockSetup: func(m *MockDropletsService) {
				m.EXPECT().
					Get(gomock.Any(), 456).
					Return(nil, nil, errors.New("api error")).
					Times(1)
			},
			expectError: true,
		},
		{
			name:        "Missing ID argument",
			args:        map[string]any{},
			mockSetup:   nil,
			expectError: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockDroplets := NewMockDropletsService(ctrl)
			mockActions := NewMockDropletActionsService(ctrl)
			if tc.mockSetup != nil {
				tc.mockSetup(mockDroplets)
			}
			tool := setupDropletToolWithMocks(mockDroplets, mockActions)
			req := mcp.CallToolRequest{Params: mcp.CallToolParams{Arguments: tc.args}}
			resp, err := tool.getDropletByID(context.Background(), req)
			if tc.expectError {
				require.NotNil(t, resp)
				require.True(t, resp.IsError)
				return
			}
			require.NoError(t, err)
			require.NotNil(t, resp)
			require.False(t, resp.IsError)
			var outDroplet godo.Droplet
			require.NoError(t, json.Unmarshal([]byte(resp.Content[0].(mcp.TextContent).Text), &outDroplet))
			require.Equal(t, testDroplet.ID, outDroplet.ID)
		})
	}
}

func TestDropletTool_getDropletActionByID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	testAction := &godo.Action{
		ID:     789,
		Status: "completed",
	}
	tests := []struct {
		name        string
		args        map[string]any
		mockSetup   func(*MockDropletActionsService)
		expectError bool
	}{
		{
			name: "Successful get action",
			args: map[string]any{"DropletID": float64(123), "ActionID": float64(789)},
			mockSetup: func(m *MockDropletActionsService) {
				m.EXPECT().
					Get(gomock.Any(), 123, 789).
					Return(testAction, nil, nil).
					Times(1)
			},
		},
		{
			name: "API error",
			args: map[string]any{"DropletID": float64(456), "ActionID": float64(999)},
			mockSetup: func(m *MockDropletActionsService) {
				m.EXPECT().
					Get(gomock.Any(), 456, 999).
					Return(nil, nil, errors.New("api error")).
					Times(1)
			},
			expectError: true,
		},
		{
			name:        "Missing DropletID argument",
			args:        map[string]any{"ActionID": float64(789)},
			mockSetup:   nil,
			expectError: true,
		},
		{
			name:        "Missing ActionID argument",
			args:        map[string]any{"DropletID": float64(123)},
			mockSetup:   nil,
			expectError: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockDroplets := NewMockDropletsService(ctrl)
			mockActions := NewMockDropletActionsService(ctrl)
			if tc.mockSetup != nil {
				tc.mockSetup(mockActions)
			}
			tool := setupDropletToolWithMocks(mockDroplets, mockActions)
			req := mcp.CallToolRequest{Params: mcp.CallToolParams{Arguments: tc.args}}
			resp, err := tool.getDropletActionByID(context.Background(), req)
			if tc.expectError {
				require.NotNil(t, resp)
				require.True(t, resp.IsError)
				return
			}
			require.NoError(t, err)
			require.NotNil(t, resp)
			require.False(t, resp.IsError)
			var outAction godo.Action
			require.NoError(t, json.Unmarshal([]byte(resp.Content[0].(mcp.TextContent).Text), &outAction))
			require.Equal(t, testAction.ID, outAction.ID)
		})
	}
}

func TestDropletTool_deleteDroplet(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	tests := []struct {
		name        string
		args        map[string]any
		mockSetup   func(*MockDropletsService)
		expectError bool
		expectText  string
	}{
		{
			name: "Successful delete",
			args: map[string]any{"ID": float64(123)},
			mockSetup: func(m *MockDropletsService) {
				m.EXPECT().
					Delete(gomock.Any(), 123).
					Return(&godo.Response{}, nil).
					Times(1)
			},
			expectText: "Droplet deleted successfully",
		},
		{
			name: "API error",
			args: map[string]any{"ID": float64(456)},
			mockSetup: func(m *MockDropletsService) {
				m.EXPECT().
					Delete(gomock.Any(), 456).
					Return(nil, errors.New("api error")).
					Times(1)
			},
			expectError: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockDroplets := NewMockDropletsService(ctrl)
			mockActions := NewMockDropletActionsService(ctrl)
			if tc.mockSetup != nil {
				tc.mockSetup(mockDroplets)
			}
			tool := setupDropletToolWithMocks(mockDroplets, mockActions)
			req := mcp.CallToolRequest{Params: mcp.CallToolParams{Arguments: tc.args}}
			resp, err := tool.deleteDroplet(context.Background(), req)
			if tc.expectError {
				require.NotNil(t, resp)
				require.True(t, resp.IsError)
				return
			}
			require.NoError(t, err)
			require.NotNil(t, resp)
			require.False(t, resp.IsError)
			require.Contains(t, resp.Content[0].(mcp.TextContent).Text, tc.expectText)
		})
	}
}

func TestDropletTool_getDropletNeighbors(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	neighbor := godo.Droplet{ID: 999, Name: "neighbor-1"}

	tests := []struct {
		name        string
		args        map[string]any
		mockSetup   func(*MockDropletsService)
		expectError bool
	}{
		{
			name: "successful neighbors fetch",
			args: map[string]any{"ID": float64(123)},
			mockSetup: func(m *MockDropletsService) {
				m.EXPECT().
					Neighbors(gomock.Any(), 123).
					Return([]godo.Droplet{neighbor}, nil, nil).
					Times(1)
			},
		},
		{
			name: "api error",
			args: map[string]any{"ID": float64(123)},
			mockSetup: func(m *MockDropletsService) {
				m.EXPECT().
					Neighbors(gomock.Any(), 123).
					Return(nil, nil, errors.New("api error")).
					Times(1)
			},
			expectError: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockDroplets := NewMockDropletsService(ctrl)
			mockActions := NewMockDropletActionsService(ctrl)
			if tc.mockSetup != nil {
				tc.mockSetup(mockDroplets)
			}

			tool := setupDropletToolWithMocks(mockDroplets, mockActions)
			req := mcp.CallToolRequest{Params: mcp.CallToolParams{Arguments: tc.args}}
			resp, err := tool.getDropletNeighbors(context.Background(), req)

			if tc.expectError {
				require.NotNil(t, resp)
				require.True(t, resp.IsError)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, resp)
			require.False(t, resp.IsError)

			var neighbors []godo.Droplet
			require.NoError(t, json.Unmarshal([]byte(resp.Content[0].(mcp.TextContent).Text), &neighbors))
			require.Len(t, neighbors, 1)
			require.Equal(t, neighbor.ID, neighbors[0].ID)
		})
	}
}

func TestDropletTool_listDropletsWithGPUs(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	testDroplet := godo.Droplet{ID: 123}

	tests := []struct {
		name        string
		args        map[string]any
		mockSetup   func(*MockDropletsService)
		expectError bool
	}{
		{
			name: "successful list",
			args: map[string]any{"Page": float64(1), "PerPage": float64(20)},
			mockSetup: func(m *MockDropletsService) {
				m.EXPECT().
					ListWithGPUs(gomock.Any(), &godo.ListOptions{Page: 1, PerPage: 20}).
					Return([]godo.Droplet{testDroplet}, nil, nil).
					Times(1)
			},
		},
		{
			name: "api error",
			args: map[string]any{},
			mockSetup: func(m *MockDropletsService) {
				m.EXPECT().
					ListWithGPUs(gomock.Any(), &godo.ListOptions{Page: 1, PerPage: 50}).
					Return(nil, nil, errors.New("api error")).
					Times(1)
			},
			expectError: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockDroplets := NewMockDropletsService(ctrl)
			mockActions := NewMockDropletActionsService(ctrl)
			if tc.mockSetup != nil {
				tc.mockSetup(mockDroplets)
			}

			tool := setupDropletToolWithMocks(mockDroplets, mockActions)
			req := mcp.CallToolRequest{Params: mcp.CallToolParams{Arguments: tc.args}}
			resp, err := tool.listDropletsWithGPUs(context.Background(), req)

			if tc.expectError {
				require.NotNil(t, resp)
				require.True(t, resp.IsError)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, resp)
			require.False(t, resp.IsError)
		})
	}
}

func TestDropletTool_listDropletsByName(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	testDroplet := godo.Droplet{ID: 456}

	tests := []struct {
		name        string
		args        map[string]any
		mockSetup   func(*MockDropletsService)
		expectError bool
	}{
		{
			name: "successful list",
			args: map[string]any{"Name": "web-1", "Page": float64(2)},
			mockSetup: func(m *MockDropletsService) {
				m.EXPECT().
					ListByName(gomock.Any(), "web-1", &godo.ListOptions{Page: 2, PerPage: 50}).
					Return([]godo.Droplet{testDroplet}, nil, nil).
					Times(1)
			},
		},
		{
			name:        "missing name",
			args:        map[string]any{},
			mockSetup:   nil,
			expectError: true,
		},
		{
			name: "api error",
			args: map[string]any{"Name": "web-1"},
			mockSetup: func(m *MockDropletsService) {
				m.EXPECT().
					ListByName(gomock.Any(), "web-1", &godo.ListOptions{Page: 1, PerPage: 50}).
					Return(nil, nil, errors.New("api error")).
					Times(1)
			},
			expectError: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockDroplets := NewMockDropletsService(ctrl)
			mockActions := NewMockDropletActionsService(ctrl)
			if tc.mockSetup != nil {
				tc.mockSetup(mockDroplets)
			}

			tool := setupDropletToolWithMocks(mockDroplets, mockActions)
			req := mcp.CallToolRequest{Params: mcp.CallToolParams{Arguments: tc.args}}
			resp, err := tool.listDropletsByName(context.Background(), req)

			if tc.expectError {
				require.NotNil(t, resp)
				require.True(t, resp.IsError)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, resp)
			require.False(t, resp.IsError)
		})
	}
}

func TestDropletTool_listDropletsByTag(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	testDroplet := godo.Droplet{ID: 789}

	tests := []struct {
		name        string
		args        map[string]any
		mockSetup   func(*MockDropletsService)
		expectError bool
	}{
		{
			name: "successful list",
			args: map[string]any{"Tag": "prod", "PerPage": float64(10)},
			mockSetup: func(m *MockDropletsService) {
				m.EXPECT().
					ListByTag(gomock.Any(), "prod", &godo.ListOptions{Page: 1, PerPage: 10}).
					Return([]godo.Droplet{testDroplet}, nil, nil).
					Times(1)
			},
		},
		{
			name:        "missing tag",
			args:        map[string]any{},
			mockSetup:   nil,
			expectError: true,
		},
		{
			name: "api error",
			args: map[string]any{"Tag": "prod"},
			mockSetup: func(m *MockDropletsService) {
				m.EXPECT().
					ListByTag(gomock.Any(), "prod", &godo.ListOptions{Page: 1, PerPage: 50}).
					Return(nil, nil, errors.New("api error")).
					Times(1)
			},
			expectError: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockDroplets := NewMockDropletsService(ctrl)
			mockActions := NewMockDropletActionsService(ctrl)
			if tc.mockSetup != nil {
				tc.mockSetup(mockDroplets)
			}

			tool := setupDropletToolWithMocks(mockDroplets, mockActions)
			req := mcp.CallToolRequest{Params: mcp.CallToolParams{Arguments: tc.args}}
			resp, err := tool.listDropletsByTag(context.Background(), req)

			if tc.expectError {
				require.NotNil(t, resp)
				require.True(t, resp.IsError)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, resp)
			require.False(t, resp.IsError)
		})
	}
}

func TestDropletTool_createMultipleDroplets(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	testDroplet := godo.Droplet{ID: 101}
	tests := []struct {
		name        string
		args        map[string]any
		mockSetup   func(*MockDropletsService)
		expectError bool
	}{
		{
			name: "successful create multiple",
			args: map[string]any{
				"Names":      []any{"web-1", "web-2"},
				"Size":       "s-1vcpu-1gb",
				"ImageID":    float64(123),
				"Region":     "nyc1",
				"Backup":     true,
				"Monitoring": true,
				"SSHKeys":    []any{float64(1), "fingerprint"},
				"Tags":       []any{"prod", "web"},
			},
			mockSetup: func(m *MockDropletsService) {
				expectedReq := &godo.DropletMultiCreateRequest{
					Names:      []string{"web-1", "web-2"},
					Region:     "nyc1",
					Size:       "s-1vcpu-1gb",
					Image:      godo.DropletCreateImage{ID: 123},
					Backups:    true,
					Monitoring: true,
					SSHKeys: []godo.DropletCreateSSHKey{
						{ID: 1},
						{Fingerprint: "fingerprint"},
					},
					Tags: []string{"prod", "web"},
				}
				m.EXPECT().
					CreateMultiple(gomock.Any(), expectedReq).
					Return([]godo.Droplet{testDroplet}, nil, nil).
					Times(1)
			},
		},
		{
			name: "invalid names array",
			args: map[string]any{
				"Names":   []any{},
				"Size":    "s-1vcpu-1gb",
				"ImageID": float64(123),
				"Region":  "nyc1",
			},
			mockSetup:   nil,
			expectError: true,
		},
		{
			name: "api error",
			args: map[string]any{
				"Names":   []any{"web-1"},
				"Size":    "s-1vcpu-1gb",
				"ImageID": float64(123),
				"Region":  "nyc1",
			},
			mockSetup: func(m *MockDropletsService) {
				m.EXPECT().
					CreateMultiple(gomock.Any(), gomock.Any()).
					Return(nil, nil, errors.New("api error")).
					Times(1)
			},
			expectError: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockDroplets := NewMockDropletsService(ctrl)
			mockActions := NewMockDropletActionsService(ctrl)
			if tc.mockSetup != nil {
				tc.mockSetup(mockDroplets)
			}

			tool := setupDropletToolWithMocks(mockDroplets, mockActions)
			req := mcp.CallToolRequest{Params: mcp.CallToolParams{Arguments: tc.args}}
			resp, err := tool.createMultipleDroplets(context.Background(), req)

			if tc.expectError {
				require.NotNil(t, resp)
				require.True(t, resp.IsError)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, resp)
			require.False(t, resp.IsError)
		})
	}
}

func TestDropletTool_deleteDropletsByTag(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	tests := []struct {
		name        string
		args        map[string]any
		mockSetup   func(*MockDropletsService)
		expectError bool
	}{
		{
			name: "successful delete",
			args: map[string]any{"Tag": "cleanup"},
			mockSetup: func(m *MockDropletsService) {
				m.EXPECT().
					DeleteByTag(gomock.Any(), "cleanup").
					Return(&godo.Response{}, nil).
					Times(1)
			},
		},
		{
			name:        "missing tag",
			args:        map[string]any{},
			mockSetup:   nil,
			expectError: true,
		},
		{
			name: "api error",
			args: map[string]any{"Tag": "cleanup"},
			mockSetup: func(m *MockDropletsService) {
				m.EXPECT().
					DeleteByTag(gomock.Any(), "cleanup").
					Return(nil, errors.New("api error")).
					Times(1)
			},
			expectError: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockDroplets := NewMockDropletsService(ctrl)
			mockActions := NewMockDropletActionsService(ctrl)
			if tc.mockSetup != nil {
				tc.mockSetup(mockDroplets)
			}

			tool := setupDropletToolWithMocks(mockDroplets, mockActions)
			req := mcp.CallToolRequest{Params: mcp.CallToolParams{Arguments: tc.args}}
			resp, err := tool.deleteDropletsByTag(context.Background(), req)

			if tc.expectError {
				require.NotNil(t, resp)
				require.True(t, resp.IsError)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, resp)
			require.False(t, resp.IsError)
			require.Contains(t, resp.Content[0].(mcp.TextContent).Text, "cleanup")
		})
	}
}

func TestDropletTool_getDroplets(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	testDroplet := godo.Droplet{
		ID:               123,
		Name:             "test-droplet",
		Memory:           2048,
		Vcpus:            2,
		Disk:             50,
		Region:           &godo.Region{Slug: "nyc1", Name: "New York 1"},
		Image:            &godo.Image{ID: 456, Name: "ubuntu-20-04-x64", Distribution: "Ubuntu"},
		Size:             &godo.Size{Slug: "s-1vcpu-2gb", Memory: 2048, Vcpus: 2, Disk: 50},
		SizeSlug:         "s-1vcpu-2gb",
		BackupIDs:        []int{1, 2},
		NextBackupWindow: &godo.BackupWindow{},
		SnapshotIDs:      []int{3, 4},
		Features:         []string{"ipv6", "private_networking"},
		Locked:           false,
		Status:           "active",
		Networks:         &godo.Networks{},
		Created:          "2023-01-01T00:00:00Z",
		Kernel:           &godo.Kernel{ID: 789, Name: "kernel-1", Version: "1.0.0"},
		Tags:             []string{"web", "prod"},
		VolumeIDs:        []string{"vol-1", "vol-2"},
		VPCUUID:          "vpc-uuid-123",
	}

	tests := []struct {
		name        string
		args        map[string]any
		mockSetup   func(*MockDropletsService)
		expectError bool
	}{
		{
			name: "Successful list",
			args: map[string]any{"Page": float64(1), "PerPage": float64(1)},
			mockSetup: func(m *MockDropletsService) {
				m.EXPECT().List(gomock.Any(), &godo.ListOptions{Page: 1, PerPage: 1}).Return([]godo.Droplet{testDroplet}, nil, nil).Times(1)
			},
		},
		{
			name: "API error",
			args: map[string]any{"Page": float64(1), "PerPage": float64(1)},
			mockSetup: func(m *MockDropletsService) {
				m.EXPECT().List(gomock.Any(), &godo.ListOptions{Page: 1, PerPage: 1}).Return(nil, nil, errors.New("api error")).Times(1)
			},
			expectError: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockDroplets := NewMockDropletsService(ctrl)
			mockActions := NewMockDropletActionsService(ctrl)
			if tc.mockSetup != nil {
				tc.mockSetup(mockDroplets)
			}
			tool := setupDropletToolWithMocks(mockDroplets, mockActions)
			req := mcp.CallToolRequest{Params: mcp.CallToolParams{Arguments: tc.args}}
			resp, err := tool.getDroplets(context.Background(), req)
			if tc.expectError {
				require.NotNil(t, resp)
				require.True(t, resp.IsError)
				return
			}
			require.NoError(t, err)
			require.NotNil(t, resp)
			require.False(t, resp.IsError)
			var outDroplets []map[string]any
			require.NoError(t, json.Unmarshal([]byte(resp.Content[0].(mcp.TextContent).Text), &outDroplets))
			require.Len(t, outDroplets, 1)
			out := outDroplets[0]
			// Check that all expected fields are present
			for _, field := range []string{
				"id", "name", "memory", "vcpus", "disk", "region", "image", "size", "size_slug", "backup_ids", "next_backup_window", "snapshot_ids", "features", "locked", "status", "networks", "created_at", "kernel", "tags", "volume_ids", "vpc_uuid",
			} {
				require.Contains(t, out, field)
			}
			// Spot check a few values
			require.Equal(t, float64(testDroplet.ID), out["id"])
			require.Equal(t, testDroplet.Name, out["name"])
			require.Equal(t, testDroplet.SizeSlug, out["size_slug"])
		})
	}
}

func TestDropletTool_listDropletSnapshots(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	testSnapshot := godo.Image{
		ID:            789,
		Name:          "test-snapshot",
		Type:          "snapshot",
		Distribution:  "Ubuntu",
		Slug:          "",
		Public:        false,
		Regions:       []string{"nyc1"},
		MinDiskSize:   20,
		SizeGigaBytes: 2.5,
	}

	tests := []struct {
		name        string
		args        map[string]any
		mockSetup   func(*MockDropletsService)
		expectError bool
	}{
		{
			name: "successful list",
			args: map[string]any{"ID": float64(123), "Page": float64(1), "PerPage": float64(20)},
			mockSetup: func(m *MockDropletsService) {
				m.EXPECT().
					Snapshots(gomock.Any(), 123, &godo.ListOptions{Page: 1, PerPage: 20}).
					Return([]godo.Image{testSnapshot}, nil, nil).
					Times(1)
			},
		},
		{
			name:        "missing ID",
			args:        map[string]any{"Page": float64(1)},
			mockSetup:   nil,
			expectError: true,
		},
		{
			name: "api error",
			args: map[string]any{"ID": float64(123)},
			mockSetup: func(m *MockDropletsService) {
				m.EXPECT().
					Snapshots(gomock.Any(), 123, &godo.ListOptions{Page: 1, PerPage: 50}).
					Return(nil, nil, errors.New("api error")).
					Times(1)
			},
			expectError: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockDroplets := NewMockDropletsService(ctrl)
			mockActions := NewMockDropletActionsService(ctrl)
			if tc.mockSetup != nil {
				tc.mockSetup(mockDroplets)
			}

			tool := setupDropletToolWithMocks(mockDroplets, mockActions)
			req := mcp.CallToolRequest{Params: mcp.CallToolParams{Arguments: tc.args}}
			resp, err := tool.listDropletSnapshots(context.Background(), req)

			if tc.expectError {
				require.NotNil(t, resp)
				require.True(t, resp.IsError)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, resp)
			require.False(t, resp.IsError)
			var outSnapshots []godo.Image
			require.NoError(t, json.Unmarshal([]byte(resp.Content[0].(mcp.TextContent).Text), &outSnapshots))
			require.Len(t, outSnapshots, 1)
			require.Equal(t, testSnapshot.ID, outSnapshots[0].ID)
		})
	}
}

func TestDropletTool_listDropletBackups(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	testBackup := godo.Image{
		ID:            456,
		Name:          "test-backup",
		Type:          "backup",
		Distribution:  "Ubuntu",
		Slug:          "",
		Public:        false,
		Regions:       []string{"nyc1"},
		MinDiskSize:   20,
		SizeGigaBytes: 2.5,
	}

	tests := []struct {
		name        string
		args        map[string]any
		mockSetup   func(*MockDropletsService)
		expectError bool
	}{
		{
			name: "successful list",
			args: map[string]any{"ID": float64(123), "Page": float64(1), "PerPage": float64(20)},
			mockSetup: func(m *MockDropletsService) {
				m.EXPECT().
					Backups(gomock.Any(), 123, &godo.ListOptions{Page: 1, PerPage: 20}).
					Return([]godo.Image{testBackup}, nil, nil).
					Times(1)
			},
		},
		{
			name:        "missing ID",
			args:        map[string]any{"Page": float64(1)},
			mockSetup:   nil,
			expectError: true,
		},
		{
			name: "api error",
			args: map[string]any{"ID": float64(123)},
			mockSetup: func(m *MockDropletsService) {
				m.EXPECT().
					Backups(gomock.Any(), 123, &godo.ListOptions{Page: 1, PerPage: 50}).
					Return(nil, nil, errors.New("api error")).
					Times(1)
			},
			expectError: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockDroplets := NewMockDropletsService(ctrl)
			mockActions := NewMockDropletActionsService(ctrl)
			if tc.mockSetup != nil {
				tc.mockSetup(mockDroplets)
			}

			tool := setupDropletToolWithMocks(mockDroplets, mockActions)
			req := mcp.CallToolRequest{Params: mcp.CallToolParams{Arguments: tc.args}}
			resp, err := tool.listDropletBackups(context.Background(), req)

			if tc.expectError {
				require.NotNil(t, resp)
				require.True(t, resp.IsError)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, resp)
			require.False(t, resp.IsError)
			var outBackups []godo.Image
			require.NoError(t, json.Unmarshal([]byte(resp.Content[0].(mcp.TextContent).Text), &outBackups))
			require.Len(t, outBackups, 1)
			require.Equal(t, testBackup.ID, outBackups[0].ID)
		})
	}
}

func TestDropletTool_listDropletActions(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	testAction := godo.Action{
		ID:           123,
		Status:       "completed",
		Type:         "snapshot",
		ResourceID:   456,
		ResourceType: "droplet",
	}

	tests := []struct {
		name        string
		args        map[string]any
		mockSetup   func(*MockDropletsService)
		expectError bool
	}{
		{
			name: "successful list",
			args: map[string]any{"ID": float64(123), "Page": float64(1), "PerPage": float64(20)},
			mockSetup: func(m *MockDropletsService) {
				m.EXPECT().
					Actions(gomock.Any(), 123, &godo.ListOptions{Page: 1, PerPage: 20}).
					Return([]godo.Action{testAction}, nil, nil).
					Times(1)
			},
		},
		{
			name:        "missing ID",
			args:        map[string]any{"Page": float64(1)},
			mockSetup:   nil,
			expectError: true,
		},
		{
			name: "api error",
			args: map[string]any{"ID": float64(123)},
			mockSetup: func(m *MockDropletsService) {
				m.EXPECT().
					Actions(gomock.Any(), 123, &godo.ListOptions{Page: 1, PerPage: 50}).
					Return(nil, nil, errors.New("api error")).
					Times(1)
			},
			expectError: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockDroplets := NewMockDropletsService(ctrl)
			mockActions := NewMockDropletActionsService(ctrl)
			if tc.mockSetup != nil {
				tc.mockSetup(mockDroplets)
			}

			tool := setupDropletToolWithMocks(mockDroplets, mockActions)
			req := mcp.CallToolRequest{Params: mcp.CallToolParams{Arguments: tc.args}}
			resp, err := tool.listDropletActions(context.Background(), req)

			if tc.expectError {
				require.NotNil(t, resp)
				require.True(t, resp.IsError)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, resp)
			require.False(t, resp.IsError)
			var outActions []godo.Action
			require.NoError(t, json.Unmarshal([]byte(resp.Content[0].(mcp.TextContent).Text), &outActions))
			require.Len(t, outActions, 1)
			require.Equal(t, testAction.ID, outActions[0].ID)
		})
	}
}

func TestDropletTool_getDropletBackupPolicy(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	testPolicy := &godo.DropletBackupPolicy{
		DropletID:     123,
		BackupEnabled: true,
		BackupPolicy: &godo.DropletBackupPolicyConfig{
			Plan:    "BASIC",
			Weekday: "monday",
			Hour:    2,
		},
		NextBackupWindow: &godo.BackupWindow{},
	}

	tests := []struct {
		name        string
		args        map[string]any
		mockSetup   func(*MockDropletsService)
		expectError bool
	}{
		{
			name: "Successful get",
			args: map[string]any{"ID": float64(123)},
			mockSetup: func(m *MockDropletsService) {
				m.EXPECT().GetBackupPolicy(gomock.Any(), 123).Return(testPolicy, nil, nil).Times(1)
			},
		},
		{
			name: "API error",
			args: map[string]any{"ID": float64(456)},
			mockSetup: func(m *MockDropletsService) {
				m.EXPECT().GetBackupPolicy(gomock.Any(), 456).Return(nil, nil, errors.New("api error")).Times(1)
			},
			expectError: true,
		},
		{
			name:        "Missing ID argument",
			args:        map[string]any{},
			mockSetup:   nil,
			expectError: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockDroplets := NewMockDropletsService(ctrl)
			mockActions := NewMockDropletActionsService(ctrl)
			if tc.mockSetup != nil {
				tc.mockSetup(mockDroplets)
			}

			tool := setupDropletToolWithMocks(mockDroplets, mockActions)
			req := mcp.CallToolRequest{Params: mcp.CallToolParams{Arguments: tc.args}}
			resp, err := tool.getDropletBackupPolicy(context.Background(), req)

			if tc.expectError {
				require.NotNil(t, resp)
				require.True(t, resp.IsError)
				return
			}
			require.NoError(t, err)
			require.NotNil(t, resp)
			require.False(t, resp.IsError)
			var outPolicy godo.DropletBackupPolicy
			require.NoError(t, json.Unmarshal([]byte(resp.Content[0].(mcp.TextContent).Text), &outPolicy))
			require.Equal(t, testPolicy.DropletID, outPolicy.DropletID)
			require.Equal(t, testPolicy.BackupEnabled, outPolicy.BackupEnabled)
			require.NotNil(t, outPolicy.BackupPolicy)
			require.Equal(t, testPolicy.BackupPolicy.Plan, outPolicy.BackupPolicy.Plan)
		})
	}
}

func TestDropletTool_listBackupPolicies(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	testPolicy := &godo.DropletBackupPolicy{
		DropletID:     123,
		BackupEnabled: true,
		BackupPolicy: &godo.DropletBackupPolicyConfig{
			Plan:    "BASIC",
			Weekday: "monday",
			Hour:    2,
		},
		NextBackupWindow: &godo.BackupWindow{},
	}

	tests := []struct {
		name        string
		args        map[string]any
		mockSetup   func(*MockDropletsService)
		expectError bool
	}{
		{
			name: "successful list",
			args: map[string]any{"Page": float64(1), "PerPage": float64(20)},
			mockSetup: func(m *MockDropletsService) {
				pmap := map[int]*godo.DropletBackupPolicy{
					123: testPolicy,
				}
				m.EXPECT().ListBackupPolicies(gomock.Any(), &godo.ListOptions{Page: 1, PerPage: 20}).Return(pmap, nil, nil).Times(1)
			},
		},
		{
			name: "api error",
			args: map[string]any{"Page": float64(1)},
			mockSetup: func(m *MockDropletsService) {
				m.EXPECT().ListBackupPolicies(gomock.Any(), &godo.ListOptions{Page: 1, PerPage: 50}).Return(nil, nil, errors.New("api error")).Times(1)
			},
			expectError: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockDroplets := NewMockDropletsService(ctrl)
			mockActions := NewMockDropletActionsService(ctrl)
			if tc.mockSetup != nil {
				tc.mockSetup(mockDroplets)
			}

			tool := setupDropletToolWithMocks(mockDroplets, mockActions)
			req := mcp.CallToolRequest{Params: mcp.CallToolParams{Arguments: tc.args}}
			resp, err := tool.listBackupPolicies(context.Background(), req)

			if tc.expectError {
				require.NotNil(t, resp)
				require.True(t, resp.IsError)
				return
			}
			require.NoError(t, err)
			require.NotNil(t, resp)
			require.False(t, resp.IsError)

			var out map[string]any
			require.NoError(t, json.Unmarshal([]byte(resp.Content[0].(mcp.TextContent).Text), &out))
			// Ensure map contains the expected key as string
			require.Contains(t, out, "123")
		})
	}
}

func TestDropletTool_listSupportedBackupPolicies(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	testSupported := &godo.SupportedBackupPolicy{}

	tests := []struct {
		name        string
		args        map[string]any
		mockSetup   func(*MockDropletsService)
		expectError bool
	}{
		{
			name: "successful list supported",
			args: map[string]any{},
			mockSetup: func(m *MockDropletsService) {
				m.EXPECT().ListSupportedBackupPolicies(gomock.Any()).Return([]*godo.SupportedBackupPolicy{testSupported}, nil, nil).Times(1)
			},
		},
		{
			name: "api error",
			args: map[string]any{},
			mockSetup: func(m *MockDropletsService) {
				m.EXPECT().ListSupportedBackupPolicies(gomock.Any()).Return(nil, nil, errors.New("api error")).Times(1)
			},
			expectError: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockDroplets := NewMockDropletsService(ctrl)
			mockActions := NewMockDropletActionsService(ctrl)
			if tc.mockSetup != nil {
				tc.mockSetup(mockDroplets)
			}

			tool := setupDropletToolWithMocks(mockDroplets, mockActions)
			req := mcp.CallToolRequest{Params: mcp.CallToolParams{Arguments: tc.args}}
			resp, err := tool.listSupportedBackupPolicies(context.Background(), req)

			if tc.expectError {
				require.NotNil(t, resp)
				require.True(t, resp.IsError)
				return
			}
			require.NoError(t, err)
			require.NotNil(t, resp)
			require.False(t, resp.IsError)

			var out []godo.SupportedBackupPolicy
			require.NoError(t, json.Unmarshal([]byte(resp.Content[0].(mcp.TextContent).Text), &out))
			require.Len(t, out, 1)
			require.NotNil(t, out[0])
		})
	}
}

// New tests: ensure droplet create and create-multiple support enhanced fields (Step 5)

func TestDropletTool_createDroplet_with_extra_fields(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	testDroplet := &godo.Droplet{
		ID:   555,
		Name: "extra-droplet",
	}
	pHour := 2
	policyReq := &godo.DropletBackupPolicyRequest{
		Plan:    "BASIC",
		Weekday: "wednesday",
		Hour:    &pHour,
	}
	policyJSON, _ := json.Marshal(policyReq)

	args := map[string]any{
		"Name":             "extra-droplet",
		"Size":             "s-1vcpu-1gb",
		"ImageID":          float64(456),
		"Region":           "nyc1",
		"Backup":           true,
		"Monitoring":       true,
		"IPv6":             true,
		"VPCUUID":          "vpc-123",
		"UserData":         "#cloud-config\n",
		"Volumes":          []any{"vol-1", "vol-2"},
		"WithDropletAgent": true,
		"BackupPolicy":     string(policyJSON),
	}

	mockSetup := func(m *MockDropletsService) {
		m.EXPECT().Create(gomock.Any(), gomock.AssignableToTypeOf(&godo.DropletCreateRequest{})).Return(testDroplet, nil, nil).Times(1)
	}

	mockDroplets := NewMockDropletsService(ctrl)
	mockActions := NewMockDropletActionsService(ctrl)
	mockSetup(mockDroplets)

	tool := setupDropletToolWithMocks(mockDroplets, mockActions)
	req := mcp.CallToolRequest{Params: mcp.CallToolParams{Arguments: args}}
	resp, err := tool.createDroplet(context.Background(), req)

	require.NoError(t, err)
	require.NotNil(t, resp)
	require.False(t, resp.IsError)

	var out godo.Droplet
	require.NoError(t, json.Unmarshal([]byte(resp.Content[0].(mcp.TextContent).Text), &out))
	require.Equal(t, testDroplet.ID, out.ID)
}

func TestDropletTool_createMultipleDroplets_with_extra_fields(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	testDroplet := godo.Droplet{ID: 777}
	pHour := 3
	policyReq := &godo.DropletBackupPolicyRequest{
		Plan:    "BASIC",
		Weekday: "thursday",
		Hour:    &pHour,
	}
	policyJSON, _ := json.Marshal(policyReq)

	args := map[string]any{
		"Names":            []any{"multi-1", "multi-2"},
		"Size":             "s-1vcpu-1gb",
		"ImageID":          float64(123),
		"Region":           "nyc1",
		"Backup":           true,
		"Monitoring":       true,
		"IPv6":             true,
		"VPCUUID":          "vpc-multi-123",
		"UserData":         "#cloud-config\n",
		"Volumes":          []any{"vol-a", "vol-b"},
		"WithDropletAgent": true,
		"BackupPolicy":     string(policyJSON),
		"SSHKeys":          []any{float64(1), "fingerprint"},
		"Tags":             []any{"prod", "web"},
	}

	mockSetup := func(m *MockDropletsService) {
		m.EXPECT().CreateMultiple(gomock.Any(), gomock.AssignableToTypeOf(&godo.DropletMultiCreateRequest{})).Return([]godo.Droplet{testDroplet}, nil, nil).Times(1)
	}

	mockDroplets := NewMockDropletsService(ctrl)
	mockActions := NewMockDropletActionsService(ctrl)
	mockSetup(mockDroplets)

	tool := setupDropletToolWithMocks(mockDroplets, mockActions)
	req := mcp.CallToolRequest{Params: mcp.CallToolParams{Arguments: args}}
	resp, err := tool.createMultipleDroplets(context.Background(), req)

	require.NoError(t, err)
	require.NotNil(t, resp)
	require.False(t, resp.IsError)

	// response contains formatted summaries
	var out []map[string]any
	require.NoError(t, json.Unmarshal([]byte(resp.Content[0].(mcp.TextContent).Text), &out))
	require.Len(t, out, 1)
	require.Equal(t, float64(testDroplet.ID), out[0]["id"])
}
