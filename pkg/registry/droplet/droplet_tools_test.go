package droplet

import (
	"context"
	"errors"
	"testing"

	"github.com/digitalocean/godo"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

// setupDropletToolWithMocks is a helper available to the entire package
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

	testDroplet := &godo.Droplet{ID: 123, Name: "test-droplet"}
	tests := []struct {
		name        string
		args        map[string]any
		mockSetup   func(*MockDropletsService)
		expectError bool
	}{
		{
			name: "Successful create",
			args: map[string]any{
				"Name": "test-droplet", "Size": "s-1vcpu-1gb", "ImageID": float64(456), "Region": "nyc1",
			},
			mockSetup: func(m *MockDropletsService) {
				m.EXPECT().Create(gomock.Any(), gomock.Any()).Return(testDroplet, nil, nil).Times(1)
			},
		},
		{
			name: "API error",
			args: map[string]any{
				"Name": "fail-droplet", "Size": "s-1vcpu-1gb", "ImageID": float64(789), "Region": "nyc3",
			},
			mockSetup: func(m *MockDropletsService) {
				m.EXPECT().Create(gomock.Any(), gomock.Any()).Return(nil, nil, errors.New("api error")).Times(1)
			},
			expectError: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockDroplets := NewMockDropletsService(ctrl)
			if tc.mockSetup != nil {
				tc.mockSetup(mockDroplets)
			}
			tool := setupDropletToolWithMocks(mockDroplets, nil)
			req := mcp.CallToolRequest{Params: mcp.CallToolParams{Arguments: tc.args}}
			resp, err := tool.createDroplet(context.Background(), req)
			if tc.expectError {
				require.True(t, resp.IsError)
				return
			}
			require.NoError(t, err)
			require.False(t, resp.IsError)
		})
	}
}

func TestDropletTool_createMultipleDroplets(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	tests := []struct {
		name        string
		args        map[string]any
		mockSetup   func(*MockDropletsService)
		expectError bool
	}{
		{
			name: "Successful create multiple",
			args: map[string]any{
				"Names": []interface{}{"d1", "d2"}, "Size": "s-1vcpu-1gb", "ImageID": float64(456), "Region": "nyc1",
			},
			mockSetup: func(m *MockDropletsService) {
				m.EXPECT().CreateMultiple(gomock.Any(), gomock.Any()).Return([]godo.Droplet{{ID: 1}}, nil, nil).Times(1)
			},
		},
		{
			name: "API Error",
			args: map[string]any{
				"Names": []interface{}{"d1"}, "Size": "s-1vcpu-1gb", "ImageID": float64(456), "Region": "nyc1",
			},
			mockSetup: func(m *MockDropletsService) {
				m.EXPECT().CreateMultiple(gomock.Any(), gomock.Any()).Return(nil, nil, errors.New("api error")).Times(1)
			},
			expectError: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockDroplets := NewMockDropletsService(ctrl)
			if tc.mockSetup != nil {
				tc.mockSetup(mockDroplets)
			}
			tool := setupDropletToolWithMocks(mockDroplets, nil)
			req := mcp.CallToolRequest{Params: mcp.CallToolParams{Arguments: tc.args}}
			resp, err := tool.createMultipleDroplets(context.Background(), req)
			if tc.expectError {
				require.True(t, resp.IsError)
			} else {
				require.NoError(t, err)
				require.False(t, resp.IsError)
			}
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
	}{
		{
			name: "Successful delete",
			args: map[string]any{"ID": float64(123)},
			mockSetup: func(m *MockDropletsService) {
				m.EXPECT().Delete(gomock.Any(), 123).Return(&godo.Response{}, nil).Times(1)
			},
		},
		{
			name: "API error",
			args: map[string]any{"ID": float64(456)},
			mockSetup: func(m *MockDropletsService) {
				m.EXPECT().Delete(gomock.Any(), 456).Return(nil, errors.New("api error")).Times(1)
			},
			expectError: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockDroplets := NewMockDropletsService(ctrl)
			if tc.mockSetup != nil {
				tc.mockSetup(mockDroplets)
			}
			tool := setupDropletToolWithMocks(mockDroplets, nil)
			req := mcp.CallToolRequest{Params: mcp.CallToolParams{Arguments: tc.args}}
			resp, err := tool.deleteDroplet(context.Background(), req)
			if tc.expectError {
				require.True(t, resp.IsError)
			} else {
				require.NoError(t, err)
				require.False(t, resp.IsError)
			}
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
			name: "Successful delete by tag",
			args: map[string]any{"Tag": "env:prod"},
			mockSetup: func(m *MockDropletsService) {
				m.EXPECT().DeleteByTag(gomock.Any(), "env:prod").Return(&godo.Response{}, nil).Times(1)
			},
		},
		{
			name: "API error",
			args: map[string]any{"Tag": "env:prod"},
			mockSetup: func(m *MockDropletsService) {
				m.EXPECT().DeleteByTag(gomock.Any(), "env:prod").Return(nil, errors.New("api error")).Times(1)
			},
			expectError: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockDroplets := NewMockDropletsService(ctrl)
			if tc.mockSetup != nil {
				tc.mockSetup(mockDroplets)
			}
			tool := setupDropletToolWithMocks(mockDroplets, nil)
			req := mcp.CallToolRequest{Params: mcp.CallToolParams{Arguments: tc.args}}
			resp, err := tool.deleteDropletsByTag(context.Background(), req)
			if tc.expectError {
				require.True(t, resp.IsError)
			} else {
				require.NoError(t, err)
				require.False(t, resp.IsError)
			}
		})
	}
}

func TestDropletTool_getDropletByID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

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
				m.EXPECT().Get(gomock.Any(), 123).Return(&godo.Droplet{ID: 123}, nil, nil).Times(1)
			},
		},
		{
			name:        "Missing ID",
			args:        map[string]any{},
			expectError: true,
		},
		{
			name: "API Error",
			args: map[string]any{"ID": float64(123)},
			mockSetup: func(m *MockDropletsService) {
				m.EXPECT().Get(gomock.Any(), 123).Return(nil, nil, errors.New("api error")).Times(1)
			},
			expectError: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockDroplets := NewMockDropletsService(ctrl)
			if tc.mockSetup != nil {
				tc.mockSetup(mockDroplets)
			}
			tool := setupDropletToolWithMocks(mockDroplets, nil)
			req := mcp.CallToolRequest{Params: mcp.CallToolParams{Arguments: tc.args}}
			resp, err := tool.getDropletByID(context.Background(), req)
			if tc.expectError {
				require.True(t, resp.IsError)
			} else {
				require.NoError(t, err)
				require.False(t, resp.IsError)
			}
		})
	}
}

func TestDropletTool_getDroplets(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	tests := []struct {
		name        string
		args        map[string]any
		mockSetup   func(*MockDropletsService)
		expectError bool
	}{
		{
			name: "Successful list",
			args: map[string]any{},
			mockSetup: func(m *MockDropletsService) {
				m.EXPECT().List(gomock.Any(), gomock.Any()).Return([]godo.Droplet{{ID: 1}}, nil, nil).Times(1)
			},
		},
		{
			name: "API Error",
			args: map[string]any{},
			mockSetup: func(m *MockDropletsService) {
				m.EXPECT().List(gomock.Any(), gomock.Any()).Return(nil, nil, errors.New("api error")).Times(1)
			},
			expectError: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockDroplets := NewMockDropletsService(ctrl)
			if tc.mockSetup != nil {
				tc.mockSetup(mockDroplets)
			}
			tool := setupDropletToolWithMocks(mockDroplets, nil)
			req := mcp.CallToolRequest{Params: mcp.CallToolParams{Arguments: tc.args}}
			resp, err := tool.getDroplets(context.Background(), req)
			if tc.expectError {
				require.True(t, resp.IsError)
			} else {
				require.NoError(t, err)
				require.False(t, resp.IsError)
			}
		})
	}
}

func TestDropletTool_listDropletsWithGPUs(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	tests := []struct {
		name        string
		mockSetup   func(*MockDropletsService)
		expectError bool
	}{
		{
			name: "Successful list with GPUs",
			mockSetup: func(m *MockDropletsService) {
				m.EXPECT().ListWithGPUs(gomock.Any(), gomock.Any()).Return([]godo.Droplet{{ID: 1}}, nil, nil).Times(1)
			},
		},
		{
			name: "API Error",
			mockSetup: func(m *MockDropletsService) {
				m.EXPECT().ListWithGPUs(gomock.Any(), gomock.Any()).Return(nil, nil, errors.New("api error")).Times(1)
			},
			expectError: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockDroplets := NewMockDropletsService(ctrl)
			tc.mockSetup(mockDroplets)
			tool := setupDropletToolWithMocks(mockDroplets, nil)
			resp, err := tool.listDropletsWithGPUs(context.Background(), mcp.CallToolRequest{})
			if tc.expectError {
				require.True(t, resp.IsError)
			} else {
				require.NoError(t, err)
				require.False(t, resp.IsError)
			}
		})
	}
}

func TestDropletTool_listDropletsByName(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	tests := []struct {
		name        string
		args        map[string]any
		mockSetup   func(*MockDropletsService)
		expectError bool
	}{
		{
			name: "Successful list by name",
			args: map[string]any{"Name": "test"},
			mockSetup: func(m *MockDropletsService) {
				m.EXPECT().ListByName(gomock.Any(), "test", gomock.Any()).Return([]godo.Droplet{{ID: 1}}, nil, nil).Times(1)
			},
		},
		{
			name:        "Missing Name",
			args:        map[string]any{},
			mockSetup:   nil,
			expectError: true,
		},
		{
			name: "API Error",
			args: map[string]any{"Name": "test"},
			mockSetup: func(m *MockDropletsService) {
				m.EXPECT().ListByName(gomock.Any(), "test", gomock.Any()).Return(nil, nil, errors.New("api error")).Times(1)
			},
			expectError: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockDroplets := NewMockDropletsService(ctrl)
			if tc.mockSetup != nil {
				tc.mockSetup(mockDroplets)
			}
			tool := setupDropletToolWithMocks(mockDroplets, nil)
			req := mcp.CallToolRequest{Params: mcp.CallToolParams{Arguments: tc.args}}
			resp, err := tool.listDropletsByName(context.Background(), req)
			if tc.expectError {
				require.True(t, resp.IsError)
			} else {
				require.NoError(t, err)
				require.False(t, resp.IsError)
			}
		})
	}
}

func TestDropletTool_listDropletsByTag(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	tests := []struct {
		name        string
		args        map[string]any
		mockSetup   func(*MockDropletsService)
		expectError bool
	}{
		{
			name: "Successful list by tag",
			args: map[string]any{"Tag": "prod"},
			mockSetup: func(m *MockDropletsService) {
				m.EXPECT().ListByTag(gomock.Any(), "prod", gomock.Any()).Return([]godo.Droplet{{ID: 1}}, nil, nil).Times(1)
			},
		},
		{
			name:        "Missing Tag",
			args:        map[string]any{},
			mockSetup:   nil,
			expectError: true,
		},
		{
			name: "API Error",
			args: map[string]any{"Tag": "prod"},
			mockSetup: func(m *MockDropletsService) {
				m.EXPECT().ListByTag(gomock.Any(), "prod", gomock.Any()).Return(nil, nil, errors.New("api error")).Times(1)
			},
			expectError: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockDroplets := NewMockDropletsService(ctrl)
			if tc.mockSetup != nil {
				tc.mockSetup(mockDroplets)
			}
			tool := setupDropletToolWithMocks(mockDroplets, nil)
			req := mcp.CallToolRequest{Params: mcp.CallToolParams{Arguments: tc.args}}
			resp, err := tool.listDropletsByTag(context.Background(), req)
			if tc.expectError {
				require.True(t, resp.IsError)
			} else {
				require.NoError(t, err)
				require.False(t, resp.IsError)
			}
		})
	}
}
