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

func TestDropletTool_getDropletNeighbors(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	tests := []struct {
		name        string
		mockSetup   func(*MockDropletsService)
		expectError bool
	}{
		{
			name: "Success",
			mockSetup: func(m *MockDropletsService) {
				m.EXPECT().Neighbors(gomock.Any(), 123).Return([]godo.Droplet{{ID: 456}}, nil, nil).Times(1)
			},
		},
		{
			name: "Error",
			mockSetup: func(m *MockDropletsService) {
				m.EXPECT().Neighbors(gomock.Any(), 123).Return(nil, nil, errors.New("api error")).Times(1)
			},
			expectError: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockDroplets := NewMockDropletsService(ctrl)
			tc.mockSetup(mockDroplets)
			tool := setupDropletToolWithMocks(mockDroplets, nil)
			req := mcp.CallToolRequest{Params: mcp.CallToolParams{Arguments: map[string]any{"ID": float64(123)}}}
			resp, err := tool.getDropletNeighbors(context.Background(), req)
			if tc.expectError {
				require.True(t, resp.IsError)
			} else {
				require.NoError(t, err)
				require.False(t, resp.IsError)
			}
		})
	}
}

func TestDropletTool_enablePrivateNetworking(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	tests := []struct {
		name        string
		mockSetup   func(*MockDropletActionsService)
		expectError bool
	}{
		{
			name: "Success",
			mockSetup: func(m *MockDropletActionsService) {
				m.EXPECT().EnablePrivateNetworking(gomock.Any(), 123).Return(&godo.Action{ID: 1}, nil, nil).Times(1)
			},
		},
		{
			name: "Error",
			mockSetup: func(m *MockDropletActionsService) {
				m.EXPECT().EnablePrivateNetworking(gomock.Any(), 123).Return(nil, nil, errors.New("api error")).Times(1)
			},
			expectError: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockActions := NewMockDropletActionsService(ctrl)
			tc.mockSetup(mockActions)
			tool := setupDropletToolWithMocks(nil, mockActions)
			req := mcp.CallToolRequest{Params: mcp.CallToolParams{Arguments: map[string]any{"ID": float64(123)}}}
			resp, err := tool.enablePrivateNetworking(context.Background(), req)
			if tc.expectError {
				require.True(t, resp.IsError)
			} else {
				require.NoError(t, err)
				require.False(t, resp.IsError)
			}
		})
	}
}

func TestDropletTool_getDropletKernels(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	tests := []struct {
		name        string
		mockSetup   func(*MockDropletsService)
		expectError bool
	}{
		{
			name: "Success",
			mockSetup: func(m *MockDropletsService) {
				m.EXPECT().Kernels(gomock.Any(), 123, gomock.Any()).Return([]godo.Kernel{{ID: 1}}, nil, nil).Times(1)
			},
		},
		{
			name: "Error",
			mockSetup: func(m *MockDropletsService) {
				m.EXPECT().Kernels(gomock.Any(), 123, gomock.Any()).Return(nil, nil, errors.New("api error")).Times(1)
			},
			expectError: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockDroplets := NewMockDropletsService(ctrl)
			tc.mockSetup(mockDroplets)
			tool := setupDropletToolWithMocks(mockDroplets, nil)
			req := mcp.CallToolRequest{Params: mcp.CallToolParams{Arguments: map[string]any{"ID": float64(123)}}}
			resp, err := tool.getDropletKernels(context.Background(), req)
			if tc.expectError {
				require.True(t, resp.IsError)
			} else {
				require.NoError(t, err)
				require.False(t, resp.IsError)
			}
		})
	}
}

func TestDropletTool_listDropletSnapshots(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	tests := []struct {
		name        string
		mockSetup   func(*MockDropletsService)
		expectError bool
	}{
		{
			name: "Success",
			mockSetup: func(m *MockDropletsService) {
				m.EXPECT().Snapshots(gomock.Any(), 123, gomock.Any()).Return([]godo.Image{{ID: 1}}, nil, nil).Times(1)
			},
		},
		{
			name: "Error",
			mockSetup: func(m *MockDropletsService) {
				m.EXPECT().Snapshots(gomock.Any(), 123, gomock.Any()).Return(nil, nil, errors.New("api error")).Times(1)
			},
			expectError: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockDroplets := NewMockDropletsService(ctrl)
			tc.mockSetup(mockDroplets)
			tool := setupDropletToolWithMocks(mockDroplets, nil)
			req := mcp.CallToolRequest{Params: mcp.CallToolParams{Arguments: map[string]any{"ID": float64(123)}}}
			resp, err := tool.listDropletSnapshots(context.Background(), req)
			if tc.expectError {
				require.True(t, resp.IsError)
			} else {
				require.NoError(t, err)
				require.False(t, resp.IsError)
			}
		})
	}
}

func TestDropletTool_listDropletBackups(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	tests := []struct {
		name        string
		mockSetup   func(*MockDropletsService)
		expectError bool
	}{
		{
			name: "Success",
			mockSetup: func(m *MockDropletsService) {
				m.EXPECT().Backups(gomock.Any(), 123, gomock.Any()).Return([]godo.Image{{ID: 1}}, nil, nil).Times(1)
			},
		},
		{
			name: "Error",
			mockSetup: func(m *MockDropletsService) {
				m.EXPECT().Backups(gomock.Any(), 123, gomock.Any()).Return(nil, nil, errors.New("api error")).Times(1)
			},
			expectError: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockDroplets := NewMockDropletsService(ctrl)
			tc.mockSetup(mockDroplets)
			tool := setupDropletToolWithMocks(mockDroplets, nil)
			req := mcp.CallToolRequest{Params: mcp.CallToolParams{Arguments: map[string]any{"ID": float64(123)}}}
			resp, err := tool.listDropletBackups(context.Background(), req)
			if tc.expectError {
				require.True(t, resp.IsError)
			} else {
				require.NoError(t, err)
				require.False(t, resp.IsError)
			}
		})
	}
}

func TestDropletTool_listDropletActions(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	tests := []struct {
		name        string
		mockSetup   func(*MockDropletsService)
		expectError bool
	}{
		{
			name: "Success",
			mockSetup: func(m *MockDropletsService) {
				m.EXPECT().Actions(gomock.Any(), 123, gomock.Any()).Return([]godo.Action{{ID: 1}}, nil, nil).Times(1)
			},
		},
		{
			name: "Error",
			mockSetup: func(m *MockDropletsService) {
				m.EXPECT().Actions(gomock.Any(), 123, gomock.Any()).Return(nil, nil, errors.New("api error")).Times(1)
			},
			expectError: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockDroplets := NewMockDropletsService(ctrl)
			tc.mockSetup(mockDroplets)
			tool := setupDropletToolWithMocks(mockDroplets, nil)
			req := mcp.CallToolRequest{Params: mcp.CallToolParams{Arguments: map[string]any{"ID": float64(123)}}}
			resp, err := tool.listDropletActions(context.Background(), req)
			if tc.expectError {
				require.True(t, resp.IsError)
			} else {
				require.NoError(t, err)
				require.False(t, resp.IsError)
			}
		})
	}
}

func TestDropletTool_getDropletBackupPolicy(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	tests := []struct {
		name        string
		mockSetup   func(*MockDropletsService)
		expectError bool
	}{
		{
			name: "Success",
			mockSetup: func(m *MockDropletsService) {
				m.EXPECT().GetBackupPolicy(gomock.Any(), 123).Return(&godo.DropletBackupPolicy{}, nil, nil).Times(1)
			},
		},
		{
			name: "Error",
			mockSetup: func(m *MockDropletsService) {
				m.EXPECT().GetBackupPolicy(gomock.Any(), 123).Return(nil, nil, errors.New("api error")).Times(1)
			},
			expectError: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockDroplets := NewMockDropletsService(ctrl)
			tc.mockSetup(mockDroplets)
			tool := setupDropletToolWithMocks(mockDroplets, nil)
			req := mcp.CallToolRequest{Params: mcp.CallToolParams{Arguments: map[string]any{"ID": float64(123)}}}
			resp, err := tool.getDropletBackupPolicy(context.Background(), req)
			if tc.expectError {
				require.True(t, resp.IsError)
			} else {
				require.NoError(t, err)
				require.False(t, resp.IsError)
			}
		})
	}
}

func TestDropletTool_listBackupPolicies(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	tests := []struct {
		name        string
		mockSetup   func(*MockDropletsService)
		expectError bool
	}{
		{
			name: "Success",
			mockSetup: func(m *MockDropletsService) {
				m.EXPECT().ListBackupPolicies(gomock.Any(), gomock.Any()).Return(map[int]*godo.DropletBackupPolicy{}, nil, nil).Times(1)
			},
		},
		{
			name: "Error",
			mockSetup: func(m *MockDropletsService) {
				m.EXPECT().ListBackupPolicies(gomock.Any(), gomock.Any()).Return(nil, nil, errors.New("api error")).Times(1)
			},
			expectError: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockDroplets := NewMockDropletsService(ctrl)
			tc.mockSetup(mockDroplets)
			tool := setupDropletToolWithMocks(mockDroplets, nil)
			resp, err := tool.listBackupPolicies(context.Background(), mcp.CallToolRequest{})
			if tc.expectError {
				require.True(t, resp.IsError)
			} else {
				require.NoError(t, err)
				require.False(t, resp.IsError)
			}
		})
	}
}

func TestDropletTool_listSupportedBackupPolicies(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	tests := []struct {
		name        string
		mockSetup   func(*MockDropletsService)
		expectError bool
	}{
		{
			name: "Success",
			mockSetup: func(m *MockDropletsService) {
				m.EXPECT().ListSupportedBackupPolicies(gomock.Any()).Return([]*godo.SupportedBackupPolicy{}, nil, nil).Times(1)
			},
		},
		{
			name: "Error",
			mockSetup: func(m *MockDropletsService) {
				m.EXPECT().ListSupportedBackupPolicies(gomock.Any()).Return(nil, nil, errors.New("api error")).Times(1)
			},
			expectError: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockDroplets := NewMockDropletsService(ctrl)
			tc.mockSetup(mockDroplets)
			tool := setupDropletToolWithMocks(mockDroplets, nil)
			resp, err := tool.listSupportedBackupPolicies(context.Background(), mcp.CallToolRequest{})
			if tc.expectError {
				require.True(t, resp.IsError)
			} else {
				require.NoError(t, err)
				require.False(t, resp.IsError)
			}
		})
	}
}

func TestDropletTool_listAssociatedResourcesForDeletion(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	tests := []struct {
		name        string
		mockSetup   func(*MockDropletsService)
		expectError bool
	}{
		{
			name: "Success",
			mockSetup: func(m *MockDropletsService) {
				m.EXPECT().ListAssociatedResourcesForDeletion(gomock.Any(), 123).Return(&godo.DropletAssociatedResources{}, nil, nil).Times(1)
			},
		},
		{
			name: "Error",
			mockSetup: func(m *MockDropletsService) {
				m.EXPECT().ListAssociatedResourcesForDeletion(gomock.Any(), 123).Return(nil, nil, errors.New("api error")).Times(1)
			},
			expectError: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockDroplets := NewMockDropletsService(ctrl)
			tc.mockSetup(mockDroplets)
			tool := setupDropletToolWithMocks(mockDroplets, nil)
			req := mcp.CallToolRequest{Params: mcp.CallToolParams{Arguments: map[string]any{"ID": float64(123)}}}
			resp, err := tool.listAssociatedResourcesForDeletion(context.Background(), req)
			if tc.expectError {
				require.True(t, resp.IsError)
			} else {
				require.NoError(t, err)
				require.False(t, resp.IsError)
			}
		})
	}
}

func TestDropletTool_getDropletActionByID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	tests := []struct {
		name        string
		mockSetup   func(*MockDropletActionsService)
		expectError bool
	}{
		{
			name: "Success",
			mockSetup: func(m *MockDropletActionsService) {
				m.EXPECT().Get(gomock.Any(), 123, 789).Return(&godo.Action{ID: 789}, nil, nil).Times(1)
			},
		},
		{
			name: "Error",
			mockSetup: func(m *MockDropletActionsService) {
				m.EXPECT().Get(gomock.Any(), 123, 789).Return(nil, nil, errors.New("api error")).Times(1)
			},
			expectError: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockDroplets := NewMockDropletsService(ctrl)
			mockActions := NewMockDropletActionsService(ctrl)
			tc.mockSetup(mockActions)
			tool := setupDropletToolWithMocks(mockDroplets, mockActions)
			req := mcp.CallToolRequest{Params: mcp.CallToolParams{Arguments: map[string]any{"DropletID": float64(123), "ActionID": float64(789)}}}
			resp, err := tool.getDropletActionByID(context.Background(), req)
			if tc.expectError {
				require.True(t, resp.IsError)
			} else {
				require.NoError(t, err)
				require.False(t, resp.IsError)
				var outAction godo.Action
				require.NoError(t, json.Unmarshal([]byte(resp.Content[0].(mcp.TextContent).Text), &outAction))
				require.Equal(t, 789, outAction.ID)
			}
		})
	}
}
