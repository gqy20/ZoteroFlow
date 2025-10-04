# ZoteroFlow2 MCP 集成方案

## 📋 概述

本文档详细描述了 ZoteroFlow2 的 MCP (Model Context Protocol) 集成方案，旨在将项目打造成一个功能强大的 MCP 网关，既提供本地文献管理能力，又能无缝集成外部 MCP 服务器。

## 🎯 目标

- **作为 MCP 服务器**: 对 AI 客户端（如 Claude Desktop）提供统一的工具接口
- **作为 MCP 客户端**: 连接外部 MCP 服务器扩展功能
- **双向集成**: 实现本地工具和外部工具的无缝集成
- **保持兼容**: 完全保持现有 CLI 功能不变

## 🏗️ 架构设计

### 整体架构

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   AI Client     │    │  ZoteroFlow2    │    │  External MCP   │
│   (Claude等)    │◄──►│   MCP Hub       │◄──►│   Servers       │
└─────────────────┘    └─────────────────┘    └─────────────────┘
                               ▲
                               │
                    ┌─────────────────┐
                    │  Local Tools    │
                    │  - ZoteroDB     │
                    │  - MinerU       │
                    │  - AI Chat      │
                    └─────────────────┘
```

### 核心组件

#### 1. MCP 服务器 (`server/mcp/server.go`)
- **功能**: 对 AI 客户端提供 MCP 服务器接口
- **协议**: JSON-RPC 2.0
- **端口**: 支持 stdio 和 HTTP 两种模式
- **代码量**: 约 200 行

```go
type MCPServer struct {
    config     *config.Config
    zoteroDB   *core.ZoteroDB
    mineru     *core.MinerUClient
    parser     *core.PDFParser
    aiClient   *core.GLMClient

    // MCP协议相关
    clients    map[string]MCPClient
    tools      []MCPTool
    resources  []MCPResource

    // 外部MCP管理
    mcpManager *MCPClientManager
}
```

#### 2. MCP 客户端管理器 (`server/mcp/client_manager.go`)
- **功能**: 管理多个外部 MCP 服务器连接
- **基于**: test_complete.go 中的 MCPClient 实现
- **特性**: 连接池、重试机制、错误处理
- **代码量**: 约 150 行

```go
type MCPClientManager struct {
    clients map[string]*MCPClient  // 外部MCP客户端
    config  *config.Config
    mu      sync.RWMutex
}

// 连接到外部MCP服务器
func (m *MCPClientManager) ConnectToServer(name, command string, args []string) error

// 调用外部MCP工具
func (m *MCPClientManager) CallTool(serverName, toolName string, args map[string]interface{}) (*CallToolResult, error)
```

#### 3. 统一工具接口 (`server/mcp/unified_tools.go`)
- **功能**: 统一管理本地工具和外部工具代理
- **特性**: 工具发现、路由、统一错误处理
- **代码量**: 约 100 行

#### 4. 协议处理 (`server/mcp/protocol.go`)
- **功能**: JSON-RPC 2.0 消息处理
- **特性**: 初始化协商、工具调用路由、错误处理
- **代码量**: 约 100 行

## 🛠️ 工具实现

### 本地工具

| 工具名称 | 功能描述 | 基于现有代码 |
|---------|---------|-------------|
| `zotero_search` | 搜索本地文献 | SearchByTitle |
| `zotero_get_item` | 获取文献详情 | GetItemsWithPDF |
| `zotero_list_items` | 列出文献 | GetItemsWithPDF |
| `mineru_parse` | 解析PDF | ParsePDF |
| `mineru_batch_parse` | 批量解析 | BatchParseDocuments |
| `zotero_get_stats` | 获取统计信息 | GetStats |
| `zotero_chat` | AI文献对话 | AIConversationManager |
| `zotero_find_by_doi` | DOI查找 | SearchByTitle |

### 外部工具代理

支持通过外部 MCP 服务器扩展的工具：

#### Article MCP 工具
- `article_search_pmc` - 搜索 Europe PMC 文献数据库
- `article_search_arxiv` - 搜索 arXiv 预印本
- `article_get_details` - 获取文章详细信息
- `article_get_citations` - 获取引用文献
- `article_get_references` - 获取参考文献

#### 其他外部 MCP 服务器
- 网络搜索 MCP
- 生物医学 MCP (BioMCP)
- 代码分析 MCP
- 文件操作 MCP

## ⚙️ 配置设计

### 环境变量配置

```bash
# MCP 服务器配置
export MCP_ENABLED=true
export MCP_PORT=8080
export MCP_STDIO=true
export MCP_CORS=*

# 外部 MCP 服务器配置
export ARTICLE_MCP_ENABLED=true
export ARTICLE_MCP_COMMAND="uv tool run article-mcp server"
export ARTICLE_MCP_TIMEOUT=30

export WEB_SEARCH_MCP_ENABLED=true
export WEB_SEARCH_MCP_COMMAND="python3 /path/to/websearch-mcp/server.py"
export WEB_SEARCH_MCP_TIMEOUT=30

# 原有配置保持不变
export ZOTERO_DB_PATH=/path/to/zotero.sqlite
export ZOTERO_DATA_DIR=/path/to/zotero/storage
export MINERU_API_URL=https://mineru-api.example.com
export MINERU_TOKEN=your_token_here
```

### 配置文件扩展

```go
// config/config.go 新增
type MCPConfig struct {
    Enabled      bool   `env:"MCP_ENABLED" envDefault:"true"`
    ServerPort   int    `env:"MCP_PORT" envDefault:"8080"`
    StdioMode    bool   `env:"MCP_STDIO" envDefault:"true"`
    AllowOrigins string `env:"MCP_CORS" envDefault:"*"`
}

type ExternalMCPConfig struct {
    Enabled bool   `env:"enabled" envDefault:"true"`
    Command string `env:"command"`
    Timeout int    `env:"timeout" envDefault:"30"`
}

type Config struct {
    // 原有配置...

    // MCP 服务器配置
    MCP MCPConfig `env-prefix:"MCP_"`

    // 外部 MCP 服务器配置
    ExternalMCP struct {
        ArticleMCP   ExternalMCPConfig `env-prefix:"ARTICLE_MCP_"`
        WebSearch   ExternalMCPConfig `env-prefix:"WEB_SEARCH_MCP_"`
        BioMCP      ExternalMCPConfig `env-prefix:"BIO_MCP_"`
    }
}
```

## 🚀 使用方式

### 启动模式

```bash
# MCP 模式（stdio，推荐用于 Claude Desktop）
./zoteroflow2 mcp

# MCP 模式（HTTP，用于 Web 客户端）
./zoteroflow2 mcp --port=8080

# 原有 CLI 模式保持不变
./zoteroflow2 list
./zoteroflow2 search "关键词"
./zoteroflow2 chat "问题"
```

### AI 客户端配置

#### Claude Desktop 配置
```json
{
  "mcpServers": {
    "zoteroflow2": {
      "command": "/path/to/zoteroflow2",
      "args": ["mcp"],
      "env": {
        "ZOTERO_DB_PATH": "/path/to/zotero.sqlite",
        "ZOTERO_DATA_DIR": "/path/to/zotero/storage"
      }
    }
  }
}
```

**配置说明**：
- **自动使用现有配置**：AI、MinerU等配置会自动从`.env`文件读取
- **仅需必要配置**：Zotero数据库路径和数据目录是必须的
- **推荐做法**：在项目目录中创建`.env`文件，包含完整配置

#### 完整 .env 配置示例
```bash
# Zotero 数据库配置（必须）
ZOTERO_DB_PATH=/path/to/zotero.sqlite
ZOTERO_DATA_DIR=/path/to/zotero/storage

# AI 模型配置（自动使用，无需在MCP客户端重复配置）
AI_API_KEY=your_api_key_here
AI_BASE_URL=https://open.bigmodel.cn/api/coding/paas/v4
AI_MODEL=glm-4.6

# MinerU PDF 解析配置（自动使用）
MINERU_API_URL=https://mineru.net/api/v4
MINERU_TOKEN=your_token_here

# 其他配置（可选）
AI_TIMEOUT=20
MINERU_TIMEOUT=60
RESULTS_DIR=data/results
CACHE_DIR=~/.zoteroflow/cache
```

#### 简化配置方式

**方式1：最小配置（推荐）**
- 仅在MCP客户端中设置Zotero路径
- AI和其他配置自动从`.env`文件读取

**方式2：环境变量优先**
- MCP客户端环境变量 > .env文件 > 默认值
- 支持覆盖特定配置（如使用不同的API密钥）
```

#### 工具调用示例

AI 客户端可以调用统一接口：

```json
{
  "tools": [
    // 本地工具
    "zotero_search",          // 搜索本地文献
    "mineru_parse",           // 解析PDF
    "zotero_chat",            // AI对话

    // 外部工具（通过Article MCP）
    "article_search_pmc",     // 搜索Europe PMC
    "article_search_arxiv",   // 搜索arXiv
    "article_get_details",    // 获取文章详情

    // 其他外部工具...
  ]
}
```

## 实现状态

### ✅ 已完成 (v0.8)

#### 核心MCP服务器 (server/mcp/)
- [x] MCP协议基础实现 (server.go: 272行)
- [x] 现有工具适配 (tools.go: 234行)
- [x] 配置系统集成 (handlers.go: 189行)
- [x] CLI集成 (main.go: 更新)

#### 外部工具支持
- [x] JSON配置加载 (external-mcp-servers.json)
- [x] 外部服务器管理框架
- [x] 示例配置 (article-mcp)

#### 测试和文档
- [x] MCP协议测试 (tests/test_mcp_integration.go)
- [x] 配置验证工具 (tests/test-mcp-ai-config.go)
- [x] 工具列表文档 (docs/api/mcp-tools-list.md)
- [x] 外部配置文档 (docs/external-mcp-configuration.md)

### 🔧 待实现

#### MCP客户端集成 (可选)
- [ ] 外部MCP服务器代理功能
- [ ] 动态工具发现
- [ ] 服务器健康检查

#### 高级功能
- [ ] 工具权限管理
- [ ] 性能监控
- [ ] 插件系统

## 实际代码统计

- **核心MCP服务器**: 695行 (已实现)
- **配置和工具**: 374行 (已实现)
- **测试代码**: 412行 (已实现)
- **文档**: 3个完整文档
- **总计**: ~1481行 (超预算但功能完整)

## 核心特性

1. **完整的MCP v2024-11-05协议支持**
2. **本地工具适配**: ZoteroDB、MinerU、AI对话
3. **外部MCP服务器JSON配置**
4. **Claude Desktop集成支持**
5. **Article MCP集成示例**
6. **完整的测试和文档**

## 💡 核心优势

### 1. 代码复用
- 充分利用现有的 2000+ 行核心功能
- 基于验证的 test_complete.go MCP 客户端实现
- 最小化新代码开发

### 2. 模块化设计
- MCP 层独立，不影响现有 CLI 功能
- 清晰的组件分离和职责划分
- 易于测试和维护

### 3. 标准化
- 完全兼容 MCP v2024-11-05 规范
- 支持主流 AI 客户端
- 标准化工具接口

### 4. 扩展性
- 支持任意数量的外部 MCP 服务器
- 插件化工具注册机制
- 灵活的配置管理

### 5. 容错性
- 外部 MCP 服务器失败不影响本地功能
- 连接池和重试机制
- 优雅的错误处理

## 🔧 技术细节

### 消息格式

```json
// MCP 请求
{
  "jsonrpc": "2.0",
  "id": 1,
  "method": "tools/call",
  "params": {
    "name": "zotero_search",
    "arguments": {
      "query": "machine learning"
    }
  }
}

// MCP 响应
{
  "jsonrpc": "2.0",
  "id": 1,
  "result": {
    "content": [
      {
        "type": "text",
        "text": "Found 5 articles matching 'machine learning'"
      }
    ]
  }
}
```

### 工具调用流程

1. AI 客户端发送工具调用请求
2. ZoteroFlow2 MCP 服务器接收请求
3. 判断工具类型（本地/外部）
4. 本地工具：直接调用核心功能
5. 外部工具：通过 MCP 客户端管理器转发
6. 整合结果并返回给 AI 客户端

### 错误处理

- 本地工具错误：直接返回错误信息
- 外部工具错误：包含服务器标识的错误信息
- 连接错误：自动重试或降级处理
- 超时处理：可配置的超时时间

## 📊 预期效果

### 对用户
- **统一接口**: 通过一个 MCP 服务器访问所有功能
- **扩展能力**: 集成外部 MCP 服务器的强大功能
- **无缝体验**: 无需关心工具来源
- **灵活配置**: 可选择性启用/禁用功能

### 对开发者
- **模块化**: 清晰的代码结构
- **可维护**: 良好的错误处理和日志
- **可扩展**: 易于添加新工具和功能
- **标准化**: 遵循 MCP 协议规范

### 对系统
- **高性能**: 连接池和缓存机制
- **高可用**: 容错和重试机制
- **低耦合**: 组件间松耦合设计
- **易部署**: 简单的配置和启动方式

## 📝 后续步骤

1. **确认方案**: 审核并确认此集成方案
2. **开始实施**: 按照实施计划逐步实现
3. **测试验证**: 确保功能正常工作
4. **文档完善**: 更新用户文档和配置说明
5. **发布部署**: 集成到主分支并发布

## 📚 参考资料

- [MCP Protocol Specification](https://spec.modelcontextprotocol.io/)
- [test_complete.go](../tests/test_complete.go) - MCP 客户端实现参考
- [Article MCP Server](https://github.com/example/article-mcp) - 外部 MCP 服务器示例

---

**文档版本**: v1.0
**创建日期**: 2025-01-01
**最后更新**: 2025-01-01
**维护者**: ZoteroFlow2 开发团队