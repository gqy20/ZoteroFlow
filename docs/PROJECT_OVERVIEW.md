# ZoteroFlow2 - AI-Powered Literature Deep Reading Assistant

## 项目概述

一个连接 Zotero 文献库、多 MCP 服务和 AI 的智能文献搜索与分析系统。通过 MCP (Model Context Protocol) 为 AI 助手提供结构化的文献搜索、质量评估和深度分析能力。

---

## 核心能力

### 1. 本地文献解析层 (🔧 核心功能)
- **MinerU PDF 解析**: 结构化提取本地PDF内容
- **Zotero 数据库访问**: 只读访问文献元数据和附件路径
- **文档结构化**: 提取章节、图表、参考文献等结构信息
- **本地缓存管理**: 解析结果存储，避免重复处理

### 2. MCP 兼容服务层 (🚀 已验证基础)
- **多 MCP 支持**: 标准化接口，支持接入多个MCP服务
- **Article MCP 集成**: ✅ 已验证 - Europe PMC + arXiv 文献搜索
- **可扩展架构**: 支持未来接入更多专业MCP服务
- **统一协议**: 完全兼容 MCP v2024-11-05 标准

### 3. AI 智能分析引擎 (🎯 核心价值)
- **本地文档分析**: 基于MinerU解析结果进行深度内容分析
- **外部文献关联**: 通过MCP服务搜索相关研究，扩展分析视野
- **智能推荐**: 结合本地库和外部搜索，推荐相关文献
- **研究洞察**: 生成研究趋势、质量评估等深度分析

---

## 技术架构 (🔄 正确架构)

```
┌─────────────────────────────────────────┐
│              AI Client                  │
│    (Claude Desktop / Continue.dev)      │
│         + GLM-4.6 模型集成              │
└─────────────────┬───────────────────────┘
                  │ MCP Protocol (stdio)
                  ▼
┌─────────────────────────────────────────┐
│        ZoteroFlow Go MCP Server        │
│                                         │
│  ┌───────────────────────────────────┐  │
│  │        MCP 兼容服务层              │  │
│  │  ┌──────────┐  ┌──────────────┐  │  │
│  │  │Article   │  │  其他MCP服务  │  │  │
│  │  │MCP ✅   │  │   (未来扩展)  │  │  │
│  │  └────┬─────┘  └──────────────┘  │  │
│  └───────┼─────────────────────────┘  │
│          │                              │
│  ┌───────┴──────────────────────────┐  │
│  │      本地文献解析核心层            │  │
│  │  ┌─────────────┐  ┌────────────┐ │  │
│  │  │   MinerU    │  │  Zotero DB │ │  │
│  │  │ PDF 解析    │  │  元数据访问 │ │  │
│  │  └─────────────┘  └────────────┘ │  │
│  └───────────────────────────────────┘  │
│                                         │
│  ┌───────────────────────────────────┐  │
│  │      AI 智能分析引擎               │  │
│  │  ┌─────────────┐  ┌────────────┐ │  │
│  │  │  本地文档    │  │  外部文献   │ │  │
│  │  │  深度分析    │  │  关联分析   │ │  │
│  │  └─────────────┘  └────────────┘ │  │
│  └───────────────────────────────────┘  │
└─────────────────────────────────────────┘
                    │
                    ▼
        ┌──────────────────────────┐
        │        外部数据源         │
        │  ┌─────────┐ ┌─────────┐ │
        │  │EuropePMC│ │ arXiv   │ │
        │  └─────────┘ └─────────┘ │
        │  ┌─────────┐ ┌─────────┐ │
        │  │更多数据源│ │专业MCP  │ │
        │  │ (未来)   │ │ 服务    │ │
        │  └─────────┘ └─────────┘ │
        └──────────────────────────┘
```

### 🎯 数据流设计

1. **本地解析优先**:
   ```
   用户请求 → MinerU解析本地PDF → 结构化内容 → AI深度分析
   ```

2. **外部搜索增强**:
   ```
   本地分析结果 → Article MCP搜索相关文献 → 关联分析 → 完整洞察
   ```

3. **智能推荐**:
   ```
   本地库特征 + 外部搜索结果 → ML推荐算法 → 个性化文献推荐
   ```

---

## 数据流

1. **启动阶段**
   - MCP Server 连接 Zotero 数据库
   - 初始化 MinerU 解析服务
   - 注册多个 MCP 服务（Article MCP + 其他）
   - 加载本地缓存和索引

2. **用户查询流程**
   ```
   用户: "分析我的机器学习相关论文，并推荐最新研究"

   AI → search_zotero_papers(query="机器学习")
   AI → parse_local_pdfs(item_ids=[1,2,3])
   AI → search_external_papers(query="machine learning", via="article_mcp")
   AI → analyze_correlations(local_docs + external_docs)
   AI → generate_recommendations()
   ```

3. **缓存策略**
   - 本地PDF解析结果存储到 `~/.zoteroflow/parsed/`
   - 外部搜索结果缓存到 `~/.zoteroflow/cache/`
   - 智能增量更新，避免重复处理

---

## 核心数据结构

### Zotero Item
```go
type ZoteroItem struct {
    ItemID    int      `json:"item_id"`
    Title     string   `json:"title"`
    Authors   []string `json:"authors"`
    Year      int      `json:"year"`
    ItemType  string   `json:"item_type"`
    Tags      []string `json:"tags"`
    PDFPaths  []string `json:"pdf_paths"`
}
```

### Parsed Document
```go
type ParsedDocument struct {
    ZoteroItem    ZoteroItem `json:"zotero_item"`
    ParseHash     string     `json:"parse_hash"`
    Sections      []Section  `json:"sections"`
    Figures       []Figure   `json:"figures"`
    Tables        []Table    `json:"tables"`
    References    []Ref      `json:"references"`
    Summary       string     `json:"ai_summary"`
    KeyPoints     []string   `json:"key_points"`
    ParseTime     time.Time  `json:"parse_time"`
}
```

### MCP Tool Interface
```go
type MCPTool interface {
    Name() string
    Description() string
    Execute(ctx context.Context, args map[string]interface{}) (interface{}, error)
}

// Article MCP 工具示例
type ArticleMCPTool struct {
    client *MCPClient
}

// 未来可扩展其他 MCP 工具
type ChemicalMCPTool struct {
    client *MCPClient
}

type PatentMCPTool struct {
    client *MCPClient
}
```

---

## 技术选型

| 组件 | 技术方案 | 理由 |
|------|---------|------|
| 数据库访问 | `sqlite3` (Go) | Zotero 用 SQLite，直接读 |
| PDF 解析 | MinerU API | 结构化提取质量高，支持图表 |
| MCP 实现 | Go + stdio | 官方 SDK，高性能异步处理 |
| 缓存 | 文件系统 + SQLite索引 | 简单可靠，支持复杂查询 |
| AI 分析 | GLM-4.6 API | 中文友好，分析能力强 |
| 配置 | TOML + .env | 人类可读 + 安全分离 |

---

## 🎯 当前 MVP 状态 (✅ 已实现 + 🔧 待完善)

### ✅ 已完成核心组件
1. **Article MCP 集成** (🚀 生产就绪)
   - 10个文献搜索和分析工具已验证
   - Europe PMC + arXiv 数据源接入
   - 完整的参考文献和引用分析

2. **Go MCP 客户端框架** (🚀 生产就绪)
   - 完整的 MCP v2024-11-05 协议实现
   - 异步消息处理和错误恢复
   - 标准化工具调用接口

3. **AI 智能分析** (🚀 生产就绪)
   - GLM-4.6 模型集成
   - 智能文献分析和总结
   - 基于配置文件的模型管理

### 🔧 待完善核心功能
1. **MinerU 集成层** (🔧 高优先级)
   - MinerU API 调用实现
   - 本地PDF批量解析
   - 解析结果结构化存储

2. **Zotero 数据库访问** (🔧 高优先级)
   - SQLite 数据库连接
   - 文献元数据提取
   - PDF 附件路径定位

3. **多 MCP 兼容层** (📋 中优先级)
   - MCP 服务抽象接口
   - 动态 MCP 服务加载
   - 统一的工具调用格式

---

## 🚀 实施计划

### 阶段 1: 完善核心解析层 (3-5 天)
- [ ] 实现 MinerU API 集成
- [ ] 实现 Zotero 数据库访问
- [ ] 建立本地文档解析工作流
- [ ] 单元测试覆盖

### 阶段 2: MCP 兼容架构 (2-3 天)
- [ ] 设计多 MCP 服务抽象接口
- [ ] 实现 Article MCP 以外的扩展接口
- [ ] 统一工具调用协议
- [ ] 错误处理和服务降级

### 阶段 3: AI 分析引擎 (3-4 天)
- [ ] 实现本地文档 AI 分析
- [ ] 结合外部搜索的关联分析
- [ ] 智能推荐算法实现
- [ ] 分析结果缓存优化

### 阶段 4: 集成测试优化 (2 天)
- [ ] 端到端工作流测试
- [ ] 性能基准测试
- [ ] 边界情况处理
- [ ] 用户体验优化

---

## 关键设计决策

### ✅ 核心原则
1. **本地优先**: 先处理用户已有文献，再扩展外部搜索
2. **MCP 兼容**: 标准化接口，支持多种 MCP 服务接入
3. **渐进增强**: 基础功能先行，高级功能逐步添加
4. **缓存智能**: 避免重复处理，提升响应速度

### 🎯 架构优势
- **数据主权**: 本地文献不离开用户设备
- **可扩展性**: 基于 MCP 协议，易于接入新服务
- **分析深度**: 结合本地理解和外部视野
- **性能优化**: 多级缓存，异步处理

### ⚠️ 技术挑战
1. **MinerU API 限制**: 并发调用和配额管理
   - 解决：智能队列 + 批量处理 + 本地缓存
2. **多 MCP 服务管理**: 不同服务的协议差异
   - 解决：适配器模式 + 统一接口封装
3. **AI 分析成本**: Token 消耗和响应时间
   - 解决：分段分析 + 结果缓存 + 优先级调度

---

## 配置文件示例

```toml
# ~/.zoteroflow/config.toml

[zotero]
database_path = "~/Zotero/zotero.sqlite"
storage_dir = "~/Zotero/storage"

[mineru]
# MinerU API 配置
api_url = "https://mineru.net/api/v4"
token = "${MINERU_TOKEN}"  # 从环境变量读取
max_concurrent = 3
timeout_seconds = 120

[mcp]
# 支持的 MCP 服务
[mcp.article_mcp]
enabled = true
command = "uv"
args = ["run", "article-mcp"]
timeout = 30

[mcp.chemical_mcp]  # 未来扩展
enabled = false
command = "chemical-mcp-server"
args = []

[mcp.patent_mcp]    # 未来扩展
enabled = false
command = "patent-mcp-server"
args = []

[ai]
# AI 模型配置
provider = "zhipuai"
model = "glm-4.6"
api_key = "${AI_API_KEY}"
base_url = "${AI_BASE_URL}"

[cache]
# 缓存配置
parsed_docs_dir = "~/.zoteroflow/parsed"
search_cache_dir = "~/.zoteroflow/cache"
max_cache_size_gb = 20
cleanup_days = 30
```

---

## 🎉 项目现状总结

### ✅ 已完成的里程碑
1. **MCP 协议突破**: 解决了 Article MCP 的初始化问题，实现了完整的通信流程
2. **Go 架构验证**: 证明了 Go 语言实现 MCP 客户端的可行性和高性能
3. **AI 集成成功**: GLM-4.6 模型与文献搜索的无缝集成
4. **多 MCP 服务设计**: 建立了可扩展的 MCP 服务兼容架构

### 🚀 核心竞争力
- **技术先进性**: 基于最新的 MCP 协议和 AI 模型
- **架构优雅性**: Go 语言的高性能异步处理
- **功能完整性**: 本地解析 + 外部搜索的完整研究工具
- **可扩展性**: 标准化的 MCP 接口，易于接入新的数据源

### 🎯 即时价值 + 未来价值
**当前可用**:
1. 搜索全球学术文献（Europe PMC + arXiv）
2. AI驱动的文献分析和总结
3. 深度文献关联和引用分析

**完成后可用**:
1. 本地PDF文献的深度解析和理解
2. 基于个人文献库的智能推荐
3. 完整的研究工作流自动化

---

## 💡 Linus 式评审（更新版）

### 🟢 优秀的部分
- **MCP 协议实现正确**: 没有偷工减料，完整实现标准
- **Go 架构简洁高效**: 没有过度抽象，代码直截了当
- **AI 集成务实**: 使用成熟的 GLM-4.6 API，不重复造轮子
- **多 MCP 设计前瞻**: 预留了扩展接口，架构合理

### 🔴 保持警惕
- **避免功能膨胀**: 核心是本地解析，外部搜索是增强
- **保持简单**: 能用就不改，稳定运行比新功能重要
- **性能第一**: Go 的高性能优势不能被复杂的业务逻辑拖累
- **用户体验**: 本地PDF解析必须稳定可靠

### 最终建议
**"先做好本地PDF解析这个核心功能，再用外部搜索锦上添花。不要本末倒置。"**