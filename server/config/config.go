package config

import (
	"fmt"
	"github.com/joho/godotenv"
	"os"
	"path/filepath"
)

type Config struct {
	// Zotero配置
	ZoteroDBPath  string `json:"zotero_db_path"`
	ZoteroDataDir string `json:"zotero_data_dir"`

	// MinerU配置
	MineruAPIURL string `json:"mineru_api_url"`
	MineruToken  string `json:"mineru_token"`

	// AI配置
	AIAPIKey  string `json:"ai_api_key"`
	AIBaseURL string `json:"ai_base_url"`
	AIModel   string `json:"ai_model"`

	// 缓存配置
	CacheDir string `json:"cache_dir"`
}

// Load 加载配置 (约50行)
func Load() (*Config, error) {
	// 1. 加载.env文件
	if err := godotenv.Load(); err != nil {
		// .env文件不存在时继续，使用环境变量
	}

	config := &Config{
		ZoteroDBPath:  getEnv("ZOTERO_DB_PATH", expandPath("~/Zotero/zotero.sqlite")),
		ZoteroDataDir: getEnv("ZOTERO_DATA_DIR", expandPath("~/Zotero/storage")),
		MineruAPIURL:  getEnv("MINERU_API_URL", "https://mineru.net/api/v4"),
		MineruToken:   getEnv("MINERU_TOKEN", ""),
		AIAPIKey:      getEnv("AI_API_KEY", ""),
		AIBaseURL:     getEnv("AI_BASE_URL", "https://open.bigmodel.cn/api/coding/paas/v4"),
		AIModel:       getEnv("AI_MODEL", "glm-4.6"),
		CacheDir:      getEnv("CACHE_DIR", expandPath("~/.zoteroflow/cache")),
	}

	// 2. 验证必要配置
	if !fileExists(config.ZoteroDBPath) {
		return nil, fmt.Errorf("Zotero数据库文件不存在: %s", config.ZoteroDBPath)
	}

	return config, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// expandPath 展开用户目录路径
func expandPath(path string) string {
	if len(path) > 0 && path[0] == '~' {
		home, _ := os.UserHomeDir()
		if home != "" {
			return filepath.Join(home, path[1:])
		}
	}
	return path
}

// fileExists 检查文件是否存在
func fileExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}
