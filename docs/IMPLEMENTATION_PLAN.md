# ZoteroFlow2 技术实现方案

## 📋 执行摘要

**项目定位**: Go 实现的智能文献分析 MCP Server
**核心价值**: 通过 MinerU 解析本地 PDF，结合 Article MCP 扩展搜索，为 AI 提供结构化文献访问能力
**技术栈**: Go + MinerU API + SQLite + MCP Protocol v2024-11-05
**代码约束**: 总实现不超过 1000 行（核心功能优先）

## 🎯 MVP 状态更新

### ✅ 已完成核心功能 (当前版本 v0.8)
1. **Article MCP 集成** - 10个文献搜索分析工具已验证可用 (~300行)
2. **Go MCP 客户端** - 完整实现 MCP v2024-11-05 协议 (~200行)
3. **AI 智能分析** - GLM-4.6 模型集成 (~100行)

### 🔧 核心功能待实现 (剩余 ~400行预算)
1. **MinerU PDF 解析** - 核心功能 (~200行)
2. **Zotero 数据库访问** - 基础数据读取 (~100行)
3. **整合工作流** - 连接所有组件 (~100行)

---

## 🚀 简化的技术架构

```
AI Client (Claude/Continue.dev)
    │ MCP Protocol
    ▼
┌─────────────────────────────────┐
│     ZoteroFlow Go Server        │  ← 1000行代码约束
│  ┌─────────────────────────────┐│
│  │ 核心层 (400行)              ││
│  │ ┌──────────┐ ┌────────────┐ ││
│  │ │ MinerU   │ │ Zotero DB  │ ││
│  │ │ PDF解析  │ │ 元数据读取  │ ││
│  │ └──────────┘ └────────────┘ ││
│  └─────────────────────────────┘│
│  ┌─────────────────────────────┐│
│  │ 扩展层 (600行，已完成)       ││
│  │ ┌──────────┐ ┌────────────┐ ││
│  │ │ArticleMCP│ │ AI分析     │ ││
│  │ │外部搜索  │ │ GLM-4.6    │ ││
│  │ └──────────┘ └────────────┘ ││
│  └─────────────────────────────┘│
└─────────────────────────────────┘
```

---

## 🛠️ 核心实现方案 (总控制 1000 行)

### 1. 项目结构 (简化版)

```
zoteroflow2/
├── main.go                 (50 行) - 主入口
├── config/
│   └── config.go          (80 行) - 配置管理
├── core/
│   ├── zotero.go          (120行) - 数据库访问
│   ├── mineru.go          (180行) - PDF解析
│   └── parser.go          (100行) - 内容解析
├── mcp/
│   ├── client.go          (200行) - MCP客户端 (已完成)
│   ├── article.go         (150行) - Article MCP (已完成)
│   └── server.go          (120行) - MCP服务器
└── ai/
    └── analyzer.go        (100行) - AI分析 (已完成)
总计: ~1100行 (包含注释和错误处理)
```

### 2. 核心数据结构 (简化版)

```go
// 核心文档结构 (50行)
type Document struct {
    ID          string    `json:"id"`
    ZoteroID    int       `json:"zotero_id"`
    Title       string    `json:"title"`
    Authors     []string  `json:"authors"`
    Content     string    `json:"content"`      // MinerU解析的Markdown内容
    Summary     string    `json:"summary"`      // AI生成的摘要
    KeyPoints   []string  `json:"key_points"`  // 关键要点
    References  []Ref     `json:"references"`   // 参考文献
    ParseTime   time.Time `json:"parse_time"`
}

// 简化的Zotero项目 (30行)
type ZoteroItem struct {
    ItemID   int      `json:"item_id"`
    Title    string   `json:"title"`
    Authors  []string `json:"authors"`
    Year     int      `json:"year"`
    PDFPath  string   `json:"pdf_path"`
}
```

### 3. MinerU API 集成 (200行实现)

```go
// mineru.go - 核心PDF解析
package core

import (
    "bytes"
    "context"
    "encoding/json"
    "fmt"
    "io"
    "net/http"
    "os"
    "time"
)

type MinerUClient struct {
    APIURL     string
    Token      string
    HTTPClient *http.Client
    MaxRetry   int
}

// NewMinerUClient 初始化MinerU客户端 (20行)
func NewMinerUClient(apiURL, token string) *MinerUClient {
    return &MinerUClient{
        APIURL:   apiURL,
        Token:    token,
        HTTPClient: &http.Client{Timeout: 120 * time.Second},
        MaxRetry: 3,
    }
}

// ParsePDF 解析PDF文件返回Markdown (80行)
func (c *MinerUClient) ParsePDF(ctx context.Context, pdfPath string) (*ParseResult, error) {
    // 1. 上传文件
    uploadURL := fmt.Sprintf("%s/api/v4/file-urls", c.APIURL)

    fileContent, err := os.ReadFile(pdfPath)
    if err != nil {
        return nil, fmt.Errorf("读取PDF失败: %w", err)
    }

    // 2. 调用解析API
    taskResp, err := c.submitParseTask(ctx, fileContent)
    if err != nil {
        return nil, err
    }

    // 3. 轮询结果
    result, err := c.pollForResult(ctx, taskResp.TaskID)
    if err != nil {
        return nil, err
    }

    return result, nil
}

// submitParseTask 提交解析任务 (40行)
func (c *MinerUClient) submitParseTask(ctx context.Context, content []byte) (*TaskResponse, error) {
    // 实现文件上传和任务提交逻辑
    // 简化版实现，重点在核心功能
}

// pollForResult 轮询解析结果 (40行)
func (c *MinerUClient) pollForResult(ctx context.Context, taskID string) (*ParseResult, error) {
    // 实现结果轮询逻辑
    // 支持超时和重试机制
}

type ParseResult struct {
    TaskID    string `json:"task_id"`
    Status    string `json:"status"`
    Content   string `json:"content"`    // Markdown格式内容
    Message   string `json:"message"`
    ErrorCode string `json:"error_code"`
}
```

### 4. Zotero 数据库访问 (120行实现)

```go
// zotero.go - 简化的数据库访问
package core

import (
    "database/sql"
    "fmt"
    "path/filepath"
    _ "github.com/mattn/go-sqlite3"
)

type ZoteroDB struct {
    db       *sql.DB
    dataDir  string
}

// NewZoteroDB 连接Zotero数据库 (30行)
func NewZoteroDB(dbPath string) (*ZoteroDB, error) {
    db, err := sql.Open("sqlite3", dbPath)
    if err != nil {
        return nil, fmt.Errorf("连接数据库失败: %w", err)
    }

    // 设置只读模式，避免锁定问题
    _, err = db.Exec("PRAGMA query_only = 1")
    if err != nil {
        return nil, fmt.Errorf("设置只读模式失败: %w", err)
    }

    return &ZoteroDB{db: db}, nil
}

// GetItemsWithPDF 获取有PDF附件的文献 (50行)
func (z *ZoteroDB) GetItemsWithPDF(limit int) ([]ZoteroItem, error) {
    query := `
    SELECT
        i.itemID,
        i.fieldValue as title,
        GROUP_CONCAT(creator.lastName || ', ' || creator.firstName, '; ') as authors,
        SUBSTR(i.fieldValue, 1, 4) as year,
        ia.path as pdf_path
    FROM items i
    LEFT JOIN itemCreators ic ON i.itemID = ic.itemID
    LEFT JOIN creators ON ic.creatorID = creators.creatorID
    LEFT JOIN itemAttachments ia ON i.itemID = ia.parentItemID
    WHERE ia.contentType = 'application/pdf'
    GROUP BY i.itemID
    LIMIT ?
    `

    rows, err := z.db.Query(query, limit)
    if err != nil {
        return nil, fmt.Errorf("查询失败: %w", err)
    }
    defer rows.Close()

    var items []ZoteroItem
    for rows.Next() {
        var item ZoteroItem
        var authorsStr string
        err := rows.Scan(&item.ItemID, &item.Title, &authorsStr, &item.Year, &item.PDFPath)
        if err != nil {
            continue
        }

        // 解析作者字符串
        item.Authors = parseAuthors(authorsStr)
        items = append(items, item)
    }

    return items, nil
}

// parseAuthors 解析作者字符串 (20行)
func parseAuthors(authorsStr string) []string {
    // 简化的作者解析逻辑
    if authorsStr == "" {
        return []string{}
    }

    // 实际实现需要更复杂的解析
    return []string{authorsStr}
}
```

### 5. 配置管理 (80行实现)

```go
// config.go - 简化的配置管理
package config

import (
    "os"
    "github.com/joho/godotenv"
)

type Config struct {
    // Zotero配置
    ZoteroDBPath string `json:"zotero_db_path"`
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

// Load 加载配置 (50行)
func Load() (*Config, error) {
    // 1. 加载.env文件
    if err := godotenv.Load(); err != nil {
        // .env文件不存在时继续，使用环境变量
    }

    config := &Config{
        ZoteroDBPath:  getEnv("ZOTERO_DB_PATH", "~/Zotero/zotero.sqlite"),
        ZoteroDataDir: getEnv("ZOTERO_DATA_DIR", "~/Zotero/storage"),
        MineruAPIURL:  getEnv("MINERU_API_URL", "https://mineru.net/api/v4"),
        MineruToken:   getEnv("MINERU_TOKEN", ""),
        AIAPIKey:      getEnv("AI_API_KEY", ""),
        AIBaseURL:     getEnv("AI_BASE_URL", "https://open.bigmodel.cn/api/coding/paas/v4"),
        AIModel:       getEnv("AI_MODEL", "glm-4.6"),
        CacheDir:      getEnv("CACHE_DIR", "~/.zoteroflow/cache"),
    }

    // 2. 验证必要配置
    if config.MineruToken == "" {
        return nil, fmt.Errorf("MINERU_TOKEN 必须设置")
    }

    return config, nil
}

func getEnv(key, defaultValue string) string {
    if value := os.Getenv(key); value != "" {
        return value
    }
    return defaultValue
}
```

---

## 🎯 MCP 工具设计 (简化版，不超过10个工具)

### 核心工具列表 (遵循Linus原则: 5个工具足够)

```go
// mcp/tools.go - MCP工具定义
package mcp

var Tools = []MCPTool{
    &ListDocumentsTool{},      // 列出本地文献
    &ParsePDFTool{},           // 解析PDF
    &SearchDocumentsTool{},    // 搜索本地内容
    &AnalyzeDocumentTool{},    // AI分析文档
    &SearchExternalTool{},     // 外部搜索增强
}
```

### 工具实现示例 (每个工具不超过50行)

```go
// ListDocumentsTool 列出文献 (40行)
type ListDocumentsTool struct {
    zoteroDB *core.ZoteroDB
}

func (t *ListDocumentsTool) Name() string {
    return "list_documents"
}

func (t *ListDocumentsTool) Description() string {
    return "列出Zotero库中的文献及其PDF文件"
}

func (t *ListDocumentsTool) Execute(ctx context.Context, args map[string]interface{}) (interface{}, error) {
    limit := 20
    if l, ok := args["limit"].(float64); ok {
        limit = int(l)
    }

    items, err := t.zoteroDB.GetItemsWithPDF(limit)
    if err != nil {
        return nil, err
    }

    return map[string]interface{}{
        "documents": items,
        "total":     len(items),
    }, nil
}
```

---

## 📊 代码分配预算 (1000行约束)

| 模块 | 行数预算 | 状态 | 说明 |
|------|---------|------|------|
| main + config | 130行 | ✅ | 主入口和配置管理 |
| core/zotero | 120行 | 🔧 | 数据库访问 |
| core/mineru | 200行 | 🔧 | PDF解析API |
| core/parser | 100行 | 🔧 | 内容结构化 |
| mcp/client | 200行 | ✅ | MCP客户端 (已完成) |
| mcp/article | 150行 | ✅ | Article MCP (已完成) |
| mcp/server | 120行 | ✅ | MCP服务器 |
| ai/analyzer | 100行 | ✅ | AI分析 (已完成) |
| **总计** | **1120行** | | 包含注释和错误处理 |

**优化策略**: 如果超出1000行，优先简化错误处理和日志记录，专注核心功能。

---

## 🚀 实施计划 (2周完成)

### 第1周: 核心功能实现
- **Day 1-2**: MinerU API集成 (200行)
- **Day 3-4**: Zotero数据库访问 (120行)
- **Day 5**: 内容解析器 (100行)

### 第2周: 集成和测试
- **Day 6-7**: MCP工具集成 (300行)
- **Day 8-9**: AI分析优化 (100行)
- **Day 10**: 测试和文档 (50行)

---

## 💡 Linus式设计原则

### ✅ 遵循的原则
1. **简单优于复杂**: 优先实现核心的PDF解析功能
2. **数据驱动设计**: 基于实际的Zotero数据结构设计接口
3. **只读优先原则**: 不修改Zotero数据库，避免数据损坏
4. **代码量控制**: 严格控制在1000行以内，避免过度工程
5. **标准API遵从**: 严格按照MCP v2024-11-05协议实现

### 🔴 避免的陷阱
1. **不要过度抽象**: 不设计复杂的插件系统
2. **不要贪多求全**: 5个核心工具足够，不追求功能完整
3. **不要过早优化**: 先让基本功能工作，再考虑性能
4. **不要复杂配置**: 简单的环境变量配置即可

### 最终建议
**"用1000行代码解决80%的问题，剩下的20%留给下一个版本。"**

---

## 🎯 成功指标

### MVP成功标准
1. **功能完整**: 能解析本地PDF并生成AI分析
2. **性能可用**: 单篇PDF解析不超过2分钟
3. **稳定可靠**: 连续处理10篇文献不崩溃
4. **接口简洁**: 5个MCP工具覆盖主要使用场景

### 质量保证
1. **代码审查**: 每个模块不超过200行
2. **单元测试**: 核心功能测试覆盖率80%
3. **集成测试**: 端到端工作流验证
4. **性能测试**: 100篇文献库处理基准

这个实现方案确保在1000行代码约束下，提供最核心的文献分析能力，为后续扩展奠定坚实基础。