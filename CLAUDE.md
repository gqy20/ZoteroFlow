# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Task Master AI Instructions
**Import Task Master's development workflow commands and guidelines, treat as if import is in the main CLAUDE.md file.**
@./.taskmaster/CLAUDE.md

## 项目概述

ZoteroFlow2 是一个基于 Go 语言开发的学术文献管理工具，集成了 Zotero 数据库访问、MinerU PDF 解析、AI 智能分析和 MCP (Model Context Protocol) 服务器功能。

## 常用开发命令

### 构建和运行
```bash
# 进入服务器目录
cd server

# 标准构建
make build

# 生产构建（含UPX压缩）
make build-prod

# 开发构建（保留调试信息）
make build-dev

# 运行主程序
make run

# 直接运行开发版本
make dev
```

### 测试命令
```bash
# 基础单元测试
make test

# 带覆盖率的测试
make test-coverage

# MCP基础功能测试
go run tests/mcp/test_mcp_basic.go

# 集成测试
go run tests/integration/test_main_article_mcp.go

# Python集成测试
cd tests && python3 test_article_mcp.py
```

### 代码质量检查
```bash
# 快速检查（格式化 + vet）
make quick

# 完整检查（格式化 + lint + 测试）
make check

# 单独检查
make fmt    # 格式化代码
make lint   # golangci-lint检查
make vet    # go vet检查
```

### 项目状态检查
```bash
# 全面项目状态检查
go run tools/MCP_STATUS_CHECK.go

# MCP服务器功能测试
./server/bin/zoteroflow2 mcp

# Article MCP独立运行
uvx article-mcp server
```

## 核心架构

### 主要组件

1. **ZoteroDB** (`server/core/zotero.go`) - Zotero SQLite 数据库访问层
   - 只读模式访问 Zotero 数据库
   - 解析 PDF 元数据和文件路径
   - 处理 Zotero 存储系统格式 (storage:XXXXXX.pdf)

2. **MinerUClient** (`server/core/mineru.go`) - MinerU PDF 解析 API 客户端
   - 支持单文件和批量处理
   - 处理文件上传和结果轮询
   - 使用命名类型而非匿名结构体

3. **AIClient** (`server/core/ai.go`) - AI 模型集成
   - 支持智谱 GLM-4.6 模型
   - 集成学术文献分析功能
   - 异步处理和错误恢复

4. **MCPManager** (`server/mcp_manager.go`) - MCP 服务器管理器
   - 配置驱动的 MCP 服务器集成
   - 支持本地和外部 MCP 服务器
   - JSON-RPC 2.0 协议实现

5. **RelatedLiterature** (`server/related_literature.go`) - 相关文献分析
   - 基于 AI 的文献关联分析
   - 支持多种学术数据库搜索

### 配置管理

配置优先级（从高到低）：
1. MCP 客户端环境变量
2. 项目 `.env` 文件
3. 代码默认值

关键配置项：
```bash
# Zotero 配置
ZOTERO_DB_PATH=/path/to/zotero.sqlite
ZOTERO_DATA_DIR=/path/to/zotero/storage

# AI 配置
AI_API_KEY=your_api_key
AI_BASE_URL=https://open.bigmodel.cn/api/coding/paas/v4
AI_MODEL=glm-4.6

# MinerU 配置
MINERU_API_URL=https://mineru.net/api/v4
MINERU_TOKEN=your_token

# 目录配置
RESULTS_DIR=data/results
RECORDS_DIR=data/records
CACHE_DIR=~/.zoteroflow/cache
```

### 数据流程

1. **配置加载**：从 `.env` 和环境变量加载配置
2. **数据库连接**：以只读模式连接 Zotero 数据库
3. **PDF 发现**：从 Zotero 存储系统发现 PDF 文件
4. **PDF 解析**：上传至 MinerU 进行解析
5. **结果缓存**：将解析结果缓存到 `CACHE_DIR`
6. **AI 分析**：使用 AI 模型进行智能分析
7. **MCP 服务**：通过 MCP 协议暴露功能

## MCP 集成

### 本地 MCP 工具

项目内置 6 个本地 MCP 工具：
- `list_literature` - 列出文献
- `search_literature` - 搜索文献
- `get_literature_details` - 获取文献详情
- `parse_pdf` - 解析 PDF
- `analyze_literature` - AI 分析文献
- `find_related_literature` - 查找相关文献

### 外部 MCP 集成

支持 Article MCP 服务器集成，提供 10 个学术工具：
- Europe PMC 文献搜索
- arXiv 预印本搜索
- 文献详细信息获取
- 参考文献管理
- 期刊质量评估
- 文献关联分析

### Claude Desktop 配置

最小配置示例：
```json
{
  "mcpServers": {
    "zoteroflow2": {
      "command": "/home/qy113/workspace/note/zo/ZoteroFlow2/server/bin/zoteroflow2",
      "args": ["mcp"],
      "env": {
        "ZOTERO_DB_PATH": "/home/qy113/workspace/note/zo/zotero_file/zotero.sqlite",
        "ZOTERO_DATA_DIR": "/home/qy113/workspace/note/zo/articles"
      }
    }
  }
}
```

## 测试框架

### 测试目录结构

```
tests/
├── core/                    # 核心功能测试
├── integration/             # 集成测试
├── mcp/                    # MCP 协议测试
├── test_article_mcp.py     # Python 集成测试
└── test_related_literature.go  # 相关文献测试
```

### 测试覆盖范围

- ✅ MCP 协议通信（JSON-RPC 2.0）
- ✅ Article MCP 服务器连接
- ✅ Go MCP 客户端实现
- ✅ AI 智能分析集成
- ✅ 多数据库文献搜索
- ✅ 文献质量评估
- ✅ 配置文件管理

## 开发工具

### 状态检查工具

`tools/MCP_STATUS_CHECK.go` - 全面的项目状态检查工具：
- 项目结构完整性检查
- 编译状态验证
- MCP 服务器功能测试
- 外部 MCP 集成验证

### 依赖管理

```bash
# 下载依赖
make deps

# 升级依赖
make mod-upgrade

# 整理依赖
go mod tidy
```

## 关键文件说明

### 服务器核心文件
- `server/main.go` - 主程序入口，CLI 命令处理
- `server/mcp_manager.go` - MCP 服务器管理器
- `server/related_literature.go` - 相关文献分析功能
- `server/config/` - 配置管理模块

### 配置文件
- `.env` - 环境变量配置
- `server/mcp_config.json` - MCP 服务器配置
- `server/external-mcp-servers.json` - 外部 MCP 服务器配置

### 文档
- `docs/PROJECT_STRUCTURE.md` - 详细项目结构说明
- `docs/mcp-usage-guide.md` - MCP 使用指南
- `tests/README.md` - 测试说明

## 开发最佳实践

### 开发流程

1. **开发前**：运行 `go run tools/MCP_STATUS_CHECK.go` 检查状态
2. **编码后**：运行 `make quick` 进行快速检查
3. **提交前**：运行 `make check` 进行完整检查
4. **发布前**：运行完整测试套件

### 代码规范

- 使用英文日志输出
- 优先使用命名类型而非匿名结构体
- 遵循约定式提交格式：`<type>(<scope>): <description>`
- 代码覆盖率要求：80%+

### 调试技巧

- 使用 `make build-dev` 构建开发版本进行调试
- MCP 连接问题：检查 `server/mcp_config.json` 配置
- AI 集成问题：验证 `.env` 文件中的 AI 配置
- PDF 解析问题：确认 MinerU API 令牌有效性

## 常见问题解决

### 编译问题
- 确保使用 Go 1.21+
- 运行 `make deps` 更新依赖
- 检查 `go.mod` 和 `go.sum` 文件

### MCP 连接问题
- 验证二进制文件路径：`ls -la server/bin/`
- 检查环境变量配置
- 运行 `go run tools/MCP_STATUS_CHECK.go` 诊断

### AI 集成问题
- 检查 `.env` 文件中的 AI 配置
- 验证 API 密钥和模型名称
- 确认网络连接和 API 端点可访问性

### Zotero 数据库问题
- 确认 Zotero 数据库路径正确
- 检查文件权限（只读访问）
- 验证 Zotero 存储目录路径

## 部署注意事项

### 生产环境构建
```bash
cd server
make build-prod
```

### 环境变量配置
确保生产环境中正确设置所有必需的环境变量，特别是：
- `ZOTERO_DB_PATH`
- `ZOTERO_DATA_DIR`
- `AI_API_KEY`
- `MINERU_TOKEN`

### 权限要求
- Zotero 数据库：只读访问权限
- 缓存目录：读写权限
- 日志目录：写入权限

---

**文档版本**: v2.0
**创建日期**: 2025-10-04
**维护者**: ZoteroFlow2 开发团队
**更新频率**: 根据项目变化更新