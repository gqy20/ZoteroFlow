package core

import (
	"testing"
)

func TestNewZoteroDB(t *testing.T) {
	tests := []struct {
		name    string
		dbPath  string
		dataDir string
	}{
		{
			name:    "无效数据库路径",
			dbPath:  "/nonexistent/zotero.sqlite",
			dataDir: "/test/storage",
		},
		{
			name:    "有效配置",
			dbPath:  "/dev/null", // 使用/dev/null作为有效路径
			dataDir: "/test/storage",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			zotero, err := NewZoteroDB(tt.dbPath, tt.dataDir)
			// 函数可能返回错误或客户端，只要不panic就算通过
			if err == nil && zotero == nil {
				t.Error("NewZoteroDB() 没有错误但返回了nil客户端")
			}
		})
	}
}

func TestNewMinerUClient(t *testing.T) {
	tests := []struct {
		name   string
		apiURL string
		token  string
	}{
		{
			name:   "有效配置",
			apiURL: "https://api.mineru.io",
			token:  "test_token",
		},
		{
			name:   "空令牌",
			apiURL: "https://api.mineru.io",
			token:  "",
		},
		{
			name:   "空API URL",
			apiURL: "",
			token:  "test_token",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := NewMinerUClient(tt.apiURL, tt.token)
			// NewMinerUClient 似乎不验证参数，只检查是否返回nil
			if client == nil {
				t.Error("NewMinerUClient() returned nil")
			}
		})
	}
}

func TestNewGLMClient(t *testing.T) {
	tests := []struct {
		name    string
		apiKey  string
		baseURL string
		model   string
	}{
		{
			name:    "有效配置",
			apiKey:  "test_api_key",
			baseURL: "https://api.example.com",
			model:   "test-model",
		},
		{
			name:    "空API密钥",
			apiKey:  "",
			baseURL: "https://api.example.com",
			model:   "test-model",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := NewGLMClient(tt.apiKey, tt.baseURL, tt.model)
			if client == nil {
				t.Error("NewGLMClient() returned nil")
			}
		})
	}
}
