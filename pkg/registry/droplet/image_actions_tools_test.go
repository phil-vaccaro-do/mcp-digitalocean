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

func setupImageActionsToolWithMocks(actions *MockImageActionsService) *ImageActionsTool {
	client := func(ctx context.Context) (*godo.Client, error) {
		return &godo.Client{ImageActions: actions}, nil
	}

	return NewImageActionsTool(client)
}

func TestImageActionsTool_transferImage(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	testAction := &godo.Action{ID: 1, Status: "in-progress", Type: "transfer"}

	tests := []struct {
		name        string
		args        map[string]any
		mockSetup   func(*MockImageActionsService)
		expectError bool
	}{
		{
			name: "Successful transfer",
			args: map[string]any{"ID": float64(123), "Region": "nyc3"},
			mockSetup: func(m *MockImageActionsService) {
				expectedReq := &godo.ActionRequest{
					"type":   "transfer",
					"region": "nyc3",
				}
				m.EXPECT().
					Transfer(gomock.Any(), 123, expectedReq).
					Return(testAction, &godo.Response{}, nil).
					Times(1)
			},
		},
		{
			name:        "Missing ID",
			args:        map[string]any{"Region": "nyc3"},
			expectError: true,
		},
		{
			name:        "Missing Region",
			args:        map[string]any{"ID": float64(123)},
			expectError: true,
		},
		{
			name: "API Error",
			args: map[string]any{"ID": float64(456), "Region": "ams3"},
			mockSetup: func(m *MockImageActionsService) {
				m.EXPECT().
					Transfer(gomock.Any(), 456, gomock.Any()).
					Return(nil, nil, errors.New("api error")).
					Times(1)
			},
			expectError: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockActions := NewMockImageActionsService(ctrl)
			if tc.mockSetup != nil {
				tc.mockSetup(mockActions)
			}
			tool := setupImageActionsToolWithMocks(mockActions)

			req := mcp.CallToolRequest{Params: mcp.CallToolParams{Arguments: tc.args}}
			resp, err := tool.transferImage(context.Background(), req)

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

func TestImageActionsTool_convertImageToSnapshot(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	testAction := &godo.Action{ID: 2, Status: "completed", Type: "convert"}

	tests := []struct {
		name        string
		args        map[string]any
		mockSetup   func(*MockImageActionsService)
		expectError bool
	}{
		{
			name: "Successful convert",
			args: map[string]any{"ID": float64(123)},
			mockSetup: func(m *MockImageActionsService) {
				m.EXPECT().
					Convert(gomock.Any(), 123).
					Return(testAction, &godo.Response{}, nil).
					Times(1)
			},
		},
		{
			name:        "Missing ID",
			args:        map[string]any{},
			expectError: true,
		},
		{
			name: "API Error",
			args: map[string]any{"ID": float64(456)},
			mockSetup: func(m *MockImageActionsService) {
				m.EXPECT().
					Convert(gomock.Any(), 456).
					Return(nil, nil, errors.New("api error")).
					Times(1)
			},
			expectError: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockActions := NewMockImageActionsService(ctrl)
			if tc.mockSetup != nil {
				tc.mockSetup(mockActions)
			}
			tool := setupImageActionsToolWithMocks(mockActions)

			req := mcp.CallToolRequest{Params: mcp.CallToolParams{Arguments: tc.args}}
			resp, err := tool.convertImageToSnapshot(context.Background(), req)

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

func TestImageActionsTool_getImageAction(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	testAction := &godo.Action{ID: 999, Status: "completed"}

	tests := []struct {
		name        string
		args        map[string]any
		mockSetup   func(*MockImageActionsService)
		expectError bool
	}{
		{
			name: "Successful get",
			args: map[string]any{"ImageID": float64(123), "ActionID": float64(999)},
			mockSetup: func(m *MockImageActionsService) {
				m.EXPECT().
					Get(gomock.Any(), 123, 999).
					Return(testAction, &godo.Response{}, nil).
					Times(1)
			},
		},
		{
			name:        "Missing ImageID",
			args:        map[string]any{"ActionID": float64(999)},
			expectError: true,
		},
		{
			name:        "Missing ActionID",
			args:        map[string]any{"ImageID": float64(123)},
			expectError: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockActions := NewMockImageActionsService(ctrl)
			if tc.mockSetup != nil {
				tc.mockSetup(mockActions)
			}
			tool := setupImageActionsToolWithMocks(mockActions)

			req := mcp.CallToolRequest{Params: mcp.CallToolParams{Arguments: tc.args}}
			resp, err := tool.getImageAction(context.Background(), req)

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
