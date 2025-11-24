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

func setupImageToolWithMocks(images *MockImagesService) *ImageTool {
	client := func(ctx context.Context) (*godo.Client, error) {
		return &godo.Client{Images: images}, nil
	}

	return NewImageTool(client)
}

func TestImageTool_listImages(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	testImages := []godo.Image{
		{ID: 1, Name: "Ubuntu 20.04", Type: "distribution", Slug: "ubuntu-20-04-x64"},
		{ID: 2, Name: "My Backup", Type: "snapshot"},
	}

	tests := []struct {
		name        string
		args        map[string]any
		mockSetup   func(*MockImagesService)
		expectError bool
	}{
		{
			name: "List all (default)",
			args: map[string]any{"Page": float64(1), "PerPage": float64(10)},
			mockSetup: func(m *MockImagesService) {
				m.EXPECT().
					List(gomock.Any(), &godo.ListOptions{Page: 1, PerPage: 10}).
					Return(testImages, &godo.Response{}, nil).
					Times(1)
			},
		},
		{
			name: "List distributions",
			args: map[string]any{"Type": "distribution"},
			mockSetup: func(m *MockImagesService) {
				m.EXPECT().
					ListDistribution(gomock.Any(), &godo.ListOptions{Page: 1, PerPage: 50}).
					Return([]godo.Image{testImages[0]}, &godo.Response{}, nil).
					Times(1)
			},
		},
		{
			name: "List applications",
			args: map[string]any{"Type": "application"},
			mockSetup: func(m *MockImagesService) {
				m.EXPECT().
					ListApplication(gomock.Any(), &godo.ListOptions{Page: 1, PerPage: 50}).
					Return([]godo.Image{}, &godo.Response{}, nil).
					Times(1)
			},
		},
		{
			name: "List user images",
			args: map[string]any{"Type": "user"},
			mockSetup: func(m *MockImagesService) {
				m.EXPECT().
					ListUser(gomock.Any(), &godo.ListOptions{Page: 1, PerPage: 50}).
					Return([]godo.Image{testImages[1]}, &godo.Response{}, nil).
					Times(1)
			},
		},
		{
			name: "API Error",
			args: map[string]any{},
			mockSetup: func(m *MockImagesService) {
				m.EXPECT().
					List(gomock.Any(), gomock.Any()).
					Return(nil, nil, errors.New("api error")).
					Times(1)
			},
			expectError: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockImages := NewMockImagesService(ctrl)
			if tc.mockSetup != nil {
				tc.mockSetup(mockImages)
			}
			tool := setupImageToolWithMocks(mockImages)

			req := mcp.CallToolRequest{Params: mcp.CallToolParams{Arguments: tc.args}}
			resp, err := tool.listImages(context.Background(), req)

			if tc.expectError {
				require.NotNil(t, resp)
				require.True(t, resp.IsError)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, resp)
			require.False(t, resp.IsError)
			require.NotEmpty(t, resp.Content)
		})
	}
}

func TestImageTool_getImageByID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	testImage := &godo.Image{ID: 123, Name: "test-image"}

	tests := []struct {
		name        string
		args        map[string]any
		mockSetup   func(*MockImagesService)
		expectError bool
	}{
		{
			name: "Successful get",
			args: map[string]any{"ID": float64(123)},
			mockSetup: func(m *MockImagesService) {
				m.EXPECT().
					GetByID(gomock.Any(), 123).
					Return(testImage, &godo.Response{}, nil).
					Times(1)
			},
		},
		{
			name: "API Error",
			args: map[string]any{"ID": float64(456)},
			mockSetup: func(m *MockImagesService) {
				m.EXPECT().
					GetByID(gomock.Any(), 456).
					Return(nil, nil, errors.New("not found")).
					Times(1)
			},
			expectError: true,
		},
		{
			name:        "Missing ID",
			args:        map[string]any{},
			expectError: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockImages := NewMockImagesService(ctrl)
			if tc.mockSetup != nil {
				tc.mockSetup(mockImages)
			}
			tool := setupImageToolWithMocks(mockImages)

			req := mcp.CallToolRequest{Params: mcp.CallToolParams{Arguments: tc.args}}
			resp, err := tool.getImageByID(context.Background(), req)

			if tc.expectError {
				require.NotNil(t, resp)
				require.True(t, resp.IsError)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, resp)
			require.False(t, resp.IsError)

			var outImage map[string]any
			require.NoError(t, json.Unmarshal([]byte(resp.Content[0].(mcp.TextContent).Text), &outImage))
			require.Equal(t, testImage.Name, outImage["name"])
		})
	}
}

func TestImageTool_createImage(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	testImage := &godo.Image{ID: 123, Name: "custom-image", Distribution: "Ubuntu"}

	tests := []struct {
		name        string
		args        map[string]any
		mockSetup   func(*MockImagesService)
		expectError bool
	}{
		{
			name: "Successful create",
			args: map[string]any{
				"Name":         "custom-image",
				"Url":          "http://example.com/image.iso",
				"Region":       "nyc3",
				"Distribution": "Ubuntu",
				"Description":  "A custom image",
				"Tags":         []any{"custom"},
			},
			mockSetup: func(m *MockImagesService) {
				expectedReq := &godo.CustomImageCreateRequest{
					Name:         "custom-image",
					Url:          "http://example.com/image.iso",
					Region:       "nyc3",
					Distribution: "Ubuntu",
					Description:  "A custom image",
					Tags:         []string{"custom"},
				}
				m.EXPECT().
					Create(gomock.Any(), expectedReq).
					Return(testImage, &godo.Response{}, nil).
					Times(1)
			},
		},
		{
			name: "Missing Name",
			args: map[string]any{
				"Url":    "http://example.com/image.iso",
				"Region": "nyc3",
			},
			expectError: true,
		},
		{
			name: "Missing Url",
			args: map[string]any{
				"Name":   "custom-image",
				"Region": "nyc3",
			},
			expectError: true,
		},
		{
			name: "Missing Region",
			args: map[string]any{
				"Name": "custom-image",
				"Url":  "http://example.com/image.iso",
			},
			expectError: true,
		},
		{
			name: "API Error",
			args: map[string]any{
				"Name":   "custom-image",
				"Url":    "http://example.com/image.iso",
				"Region": "nyc3",
			},
			mockSetup: func(m *MockImagesService) {
				m.EXPECT().
					Create(gomock.Any(), gomock.Any()).
					Return(nil, nil, errors.New("api error")).
					Times(1)
			},
			expectError: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockImages := NewMockImagesService(ctrl)
			if tc.mockSetup != nil {
				tc.mockSetup(mockImages)
			}
			tool := setupImageToolWithMocks(mockImages)

			req := mcp.CallToolRequest{Params: mcp.CallToolParams{Arguments: tc.args}}
			resp, err := tool.createImage(context.Background(), req)

			if tc.expectError {
				require.NotNil(t, resp)
				require.True(t, resp.IsError)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, resp)
			require.False(t, resp.IsError)

			var outImage map[string]any
			require.NoError(t, json.Unmarshal([]byte(resp.Content[0].(mcp.TextContent).Text), &outImage))
			require.Equal(t, testImage.Name, outImage["name"])
		})
	}
}

func TestImageTool_updateImage(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	testImage := &godo.Image{ID: 123, Name: "new-name"}

	tests := []struct {
		name        string
		args        map[string]any
		mockSetup   func(*MockImagesService)
		expectError bool
	}{
		{
			name: "Successful update",
			args: map[string]any{"ID": float64(123), "Name": "new-name"},
			mockSetup: func(m *MockImagesService) {
				m.EXPECT().
					Update(gomock.Any(), 123, &godo.ImageUpdateRequest{Name: "new-name"}).
					Return(testImage, &godo.Response{}, nil).
					Times(1)
			},
		},
		{
			name:        "Missing Name",
			args:        map[string]any{"ID": float64(123)},
			expectError: true,
		},
		{
			name:        "Missing ID",
			args:        map[string]any{"Name": "new-name"},
			expectError: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockImages := NewMockImagesService(ctrl)
			if tc.mockSetup != nil {
				tc.mockSetup(mockImages)
			}
			tool := setupImageToolWithMocks(mockImages)

			req := mcp.CallToolRequest{Params: mcp.CallToolParams{Arguments: tc.args}}
			resp, err := tool.updateImage(context.Background(), req)

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

func TestImageTool_deleteImage(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	tests := []struct {
		name        string
		args        map[string]any
		mockSetup   func(*MockImagesService)
		expectError bool
	}{
		{
			name: "Successful delete",
			args: map[string]any{"ID": float64(123)},
			mockSetup: func(m *MockImagesService) {
				m.EXPECT().
					Delete(gomock.Any(), 123).
					Return(&godo.Response{}, nil).
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
			mockSetup: func(m *MockImagesService) {
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
			mockImages := NewMockImagesService(ctrl)
			if tc.mockSetup != nil {
				tc.mockSetup(mockImages)
			}
			tool := setupImageToolWithMocks(mockImages)

			req := mcp.CallToolRequest{Params: mcp.CallToolParams{Arguments: tc.args}}
			resp, err := tool.deleteImage(context.Background(), req)

			if tc.expectError {
				require.NotNil(t, resp)
				require.True(t, resp.IsError)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, resp)
			require.False(t, resp.IsError)
			require.Contains(t, resp.Content[0].(mcp.TextContent).Text, "deleted successfully")
		})
	}
}
