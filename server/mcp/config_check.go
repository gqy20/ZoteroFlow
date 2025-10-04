package mcp

import (
	"os"
	"path/filepath"
)

// IsMCPConfigured 检查MCP是否已配置
func IsMCPConfigured() bool {
	// 检查MCP配置文件是否存在
	configFile := filepath.Join(".", "mcp_config.json")
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		return false
	}

	// 检查外部MCP服务器配置文件是否存在
	externalConfigFile := filepath.Join(".", "external-mcp-servers.json")
	if _, err := os.Stat(externalConfigFile); os.IsNotExist(err) {
		return false
	}

	// 检查基本的环境变量
	if os.Getenv("MCP_SERVERS") == "" {
		return false
	}

	return true
}

// GetMCPStatus 获取MCP服务状态
func GetMCPStatus() map[string]interface{} {
	status := make(map[string]interface{})

	if IsMCPConfigured() {
		status["configured"] = true
		status["local_servers"] = true
		status["external_servers"] = true
	} else {
		status["configured"] = false
		status["local_servers"] = false
		status["external_servers"] = false
	}

	// 检查具体的服务器配置
	configFile := filepath.Join(".", "mcp_config.json")
	if _, err := os.Stat(configFile); err == nil {
		status["config_file"] = configFile
	}

	externalConfigFile := filepath.Join(".", "external-mcp-servers.json")
	if _, err := os.Stat(externalConfigFile); err == nil {
		status["external_config_file"] = externalConfigFile
	}

	return status
}
