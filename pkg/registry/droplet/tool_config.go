package droplet

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/digitalocean/godo"
	"github.com/mark3labs/mcp-go/mcp"
)

// ArgumentType represents the type of an argument
type ArgumentType string

const (
	ArgumentTypeString  ArgumentType = "string"
	ArgumentTypeNumber  ArgumentType = "number"
	ArgumentTypeBoolean ArgumentType = "boolean"
	ArgumentTypeArray   ArgumentType = "array"
	ArgumentTypeObject  ArgumentType = "object"
)

// ArgumentConfig defines the configuration for a tool argument
type ArgumentConfig struct {
	Name         string
	Type         ArgumentType
	Description  string
	Required     bool
	DefaultValue interface{}
}

// ToolConfig defines the configuration for a tool
type ToolConfig struct {
	Name        string
	Description string
	Arguments   []ArgumentConfig
	Handler     HandlerFunc
}

// HandlerFunc is the function signature for tool handlers
type HandlerFunc func(ctx context.Context, client *godo.Client, args map[string]interface{}) (interface{}, error)

// BuildMCPTool converts a ToolConfig into an MCP Tool definition
func (tc *ToolConfig) BuildMCPTool() mcp.Tool {
	properties := make(map[string]interface{})
	required := []string{}

	for _, arg := range tc.Arguments {
		prop := map[string]interface{}{
			"type":        string(arg.Type),
			"description": arg.Description,
		}

		if arg.DefaultValue != nil {
			prop["default"] = arg.DefaultValue
		}

		properties[arg.Name] = prop

		if arg.Required {
			required = append(required, arg.Name)
		}
	}

	inputSchema := mcp.ToolInputSchema{
		Type:       "object",
		Properties: properties,
	}

	if len(required) > 0 {
		inputSchema.Required = required
	}

	return mcp.Tool{
		Name:        tc.Name,
		Description: tc.Description,
		InputSchema: inputSchema,
	}
}

// GetArgumentString safely retrieves a string argument
func GetArgumentString(args map[string]interface{}, name string) string {
	if val, ok := args[name]; ok {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return ""
}

// GetArgumentNumber safely retrieves a number argument as int
func GetArgumentNumber(args map[string]interface{}, name string) int {
	if val, ok := args[name]; ok {
		switch v := val.(type) {
		case float64:
			return int(v)
		case int:
			return v
		case json.Number:
			if i, err := v.Int64(); err == nil {
				return int(i)
			}
		}
	}
	return 0
}

// GetArgumentBoolean safely retrieves a boolean argument
func GetArgumentBoolean(args map[string]interface{}, name string) bool {
	if val, ok := args[name]; ok {
		if b, ok := val.(bool); ok {
			return b
		}
	}
	return false
}

// GetArgumentArray safely retrieves an array argument
func GetArgumentArray(args map[string]interface{}, name string) []interface{} {
	if val, ok := args[name]; ok {
		if arr, ok := val.([]interface{}); ok {
			return arr
		}
	}
	return nil
}

// GetArgumentObject safely retrieves an object argument
func GetArgumentObject(args map[string]interface{}, name string) map[string]interface{} {
	if val, ok := args[name]; ok {
		if obj, ok := val.(map[string]interface{}); ok {
			return obj
		}
	}
	return nil
}

// ValidateArguments validates that all required arguments are present
func (tc *ToolConfig) ValidateArguments(args map[string]interface{}) error {
	for _, arg := range tc.Arguments {
		if arg.Required {
			if _, ok := args[arg.Name]; !ok {
				return fmt.Errorf("missing required argument: %s", arg.Name)
			}
		}
	}
	return nil
}
