package config

import (
	"os"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	// 保存原始环境变量
	originalDB := os.Getenv("ZOTERO_DB_PATH")
	originalDataDir := os.Getenv("ZOTERO_DATA_DIR")
	originalAPIKey := os.Getenv("AI_API_KEY")
	originalToken := os.Getenv("MINERU_TOKEN")

	// 测试完成后恢复原始环境变量
	defer func() {
		os.Setenv("ZOTERO_DB_PATH", originalDB)
		os.Setenv("ZOTERO_DATA_DIR", originalDataDir)
		os.Setenv("AI_API_KEY", originalAPIKey)
		os.Setenv("MINERU_TOKEN", originalToken)
	}()

	// 创建临时测试文件
	tempDB := t.TempDir() + "/test_zotero.sqlite"
	file, err := os.Create(tempDB)
	if err != nil {
		t.Fatal("无法创建临时数据库文件:", err)
	}
	file.Close()

	tests := []struct {
		name    string
		envVars map[string]string
		want    string // 检查某个特定字段
	}{
		{
			name: "从环境变量加载配置",
			envVars: map[string]string{
				"ZOTERO_DB_PATH":  tempDB,
				"ZOTERO_DATA_DIR": "/test/storage",
				"AI_API_KEY":      "test_api_key",
				"MINERU_TOKEN":    "test_token",
				"RESULTS_DIR":     "/test/results",
			},
			want: tempDB,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 设置测试环境变量
			for k, v := range tt.envVars {
				os.Setenv(k, v)
			}

			cfg, err := Load()
			if err != nil {
				t.Errorf("Load() 返回错误: %v", err)
				return
			}

			if cfg.ZoteroDBPath != tt.want {
				t.Errorf("Load().ZoteroDBPath = %v, want %v", cfg.ZoteroDBPath, tt.want)
			}

			// 验证其他字段
			if cfg.AIAPIKey != tt.envVars["AI_API_KEY"] {
				t.Errorf("Load().AIAPIKey = %v, want %v", cfg.AIAPIKey, tt.envVars["AI_API_KEY"])
			}
		})
	}
}

func TestLoadConfigError(t *testing.T) {
	// 保存原始环境变量
	originalDB := os.Getenv("ZOTERO_DB_PATH")
	defer func() {
		os.Setenv("ZOTERO_DB_PATH", originalDB)
	}()

	// 测试无效数据库路径
	os.Setenv("ZOTERO_DB_PATH", "/nonexistent/zotero.sqlite")

	_, err := Load()
	if err == nil {
		t.Error("Load() 应该返回错误，但没有返回")
	}
}
