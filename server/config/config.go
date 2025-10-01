package config

import (
	"fmt"
	"github.com/joho/godotenv"
	"log"
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

	// 数据目录配置
	ResultsDir string `json:"results_dir"`
	RecordsDir string `json:"records_dir"`

	// 超时配置 (秒)
	AITimeout     int `json:"ai_timeout"`
	MineruTimeout int `json:"mineru_timeout"`

	// 文本长度限制
	AbstractLength int `json:"abstract_length"`
}

// Load 加载配置 (约50行)
func Load() (*Config, error) {
	// 1. 加载.env文件 - 尝试多个路径
	envPaths := []string{
		".env",       // 当前目录
		"../.env",    // 上级目录
		"../../.env", // 上上级目录
	}

	envLoaded := false
	for _, envPath := range envPaths {
		if err := godotenv.Load(envPath); err == nil {
			log.Printf("成功加载环境文件: %s", envPath)
			envLoaded = true
			break
		}
	}

	if !envLoaded {
		log.Printf("未找到.env文件，使用环境变量或默认值")
	}

	// 调试环境变量
	zoteroDataDir := getEnv("ZOTERO_DATA_DIR", expandPath("~/Zotero/storage"))
	log.Printf("环境变量 ZOTERO_DATA_DIR: %s", zoteroDataDir)

	config := &Config{
		ZoteroDBPath:   getEnv("ZOTERO_DB_PATH", expandPath("~/Zotero/zotero.sqlite")),
		ZoteroDataDir:  zoteroDataDir,
		MineruAPIURL:   getEnv("MINERU_API_URL", "https://mineru.net/api/v4"),
		MineruToken:    getEnv("MINERU_TOKEN", ""),
		AIAPIKey:       getEnv("AI_API_KEY", ""),
		AIBaseURL:      getEnv("AI_BASE_URL", "https://open.bigmodel.cn/api/coding/paas/v4"),
		AIModel:        getEnv("AI_MODEL", "glm-4.6"),
		CacheDir:       getEnv("CACHE_DIR", expandPath("~/.zoteroflow/cache")),
		ResultsDir:     getEnv("RESULTS_DIR", "data/results"),
		RecordsDir:     getEnv("RECORDS_DIR", "data/records"),
		AITimeout:      getIntEnv("AI_TIMEOUT", 20),
		MineruTimeout:  getIntEnv("MINERU_TIMEOUT", 60),
		AbstractLength: getIntEnv("ABSTRACT_LENGTH", 200),
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

func getIntEnv(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		var intValue int
		if n, err := fmt.Sscanf(value, "%d", &intValue); err == nil && n == 1 && intValue > 0 {
			return intValue
		}
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
