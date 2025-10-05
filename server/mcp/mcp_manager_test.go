package mcp

import (
	"testing"
)

func TestNewMCPManager(t *testing.T) {
	tests := []struct {
		name       string
		configFile string
		expectErr  bool
	}{
		{
			name:       "创建MCP管理器",
			configFile: "mcp_config.json",
			expectErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manager, err := NewMCPManager(tt.configFile)
			if (err != nil) != tt.expectErr {
				t.Errorf("NewMCPManager() error = %v, expectErr %v", err, tt.expectErr)
				return
			}
			if manager == nil && !tt.expectErr {
				t.Error("NewMCPManager() returned nil")
			}
		})
	}
}
