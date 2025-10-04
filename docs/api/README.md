# ZoteroFlow2 API 文档

## 概述

ZoteroFlow2 是一个智能文献分析系统，提供多种 API 接口支持文献管理、PDF 解析、AI 分析等功能。

## 文档结构

```
docs/api/
├── README.md                 # 本文档
├── cli/                      # CLI API 文档
│   ├── commands.md          # 命令行接口
│   └── examples.md          # 使用示例
├── core/                     # 核心模块 API
│   ├── zotero.md            # Zotero 数据库接口
│   ├── mineru.md            # MinerU PDF 解析接口
│   ├── parser.md            # 文档解析接口
│   ├── organizer.md         # 文件组织接口
│   └── ai.md                # AI 分析接口
├── mcp/                      # MCP 集成文档
│   ├── overview.md          # MCP 协议概述
│   ├── article-mcp.md       # Article MCP 集成
│   └── custom-mcp.md        # 自定义 MCP 服务
├── examples/                 # 使用示例
│   ├── basic-usage.md       # 基础使用
│   ├── advanced-usage.md    # 高级用法
│   └── integration.md       # 集成示例
├── deployment/               # 部署文档
│   ├── installation.md      # 安装指南
│   ├── configuration.md     # 配置说明
│   └── troubleshooting.md   # 故障排除
├── openapi/                  # OpenAPI 规范
│   ├── specification.yaml    # API 规范文件
│   └── schemas.md           # 数据模型
└── developer/                # 开发者指南
    ├── architecture.md      # 架构说明
    ├── contributing.md       # 贡献指南
    └── testing.md          # 测试指南
```

## 快速开始

### 1. CLI 接口

```bash
# 列出文献
./zoteroflow2 list

# 搜索文献
./zoteroflow2 search "机器学习"

# AI 对话
./zoteroflow2 chat "什么是深度学习？"
```

### 2. 核心模块

```go
// 连接 Zotero 数据库
zoteroDB, err := core.NewZoteroDB(dbPath, dataDir)

// 创建 MinerU 客户端
mineruClient := core.NewMinerUClient(apiURL, token)

// 解析 PDF
result, err := mineruClient.ParsePDF(ctx, pdfPath)
```

### 3. AI 对话

```go
// 创建 AI 客户端
client := core.NewGLMClient(apiKey, baseURL, model)

// 创建对话管理器
chatManager := core.NewAIConversationManager(client, zoteroDB)

// 开始对话
conv, err := chatManager.StartConversation(ctx, message, nil)
```

## API 版本

- 当前版本: v1.0.0
- 协议版本: MCP v2024-11-05
- Go 版本: 1.21+

## 支持

- 📧 邮箱: support@zoteroflow2.com
- 📖 文档: https://docs.zoteroflow2.com
- 🐛 问题反馈: https://github.com/zoteroflow2/issues