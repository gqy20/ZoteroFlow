# ZoteroFlow2 MCP 使用指南

## 🚀 快速开始

### 1. 系统要求

- Go 1.19+ (用于编译)
- Python 3.8+ (用于Article MCP)
- Zotero 数据库文件
- AI API Key (可选，用于对话功能)

### 2. 安装和配置

#### 2.1 编译项目
```bash
cd server/
make build
```

#### 2.2 配置环境变量
编辑 `server/.env` 文件：
```bash
# Zotero 数据库配置（必须）
ZOTERO_DB_PATH=/path/to/zotero.sqlite
ZOTERO_DATA_DIR=/path/to/zotero/storage

# AI 对话功能（可选）
AI_API_KEY=your_api_key_here
AI_BASE_URL=https://open.bigmodel.cn/api/coding/paas/v4
AI_MODEL=glm-4.6

# MinerU PDF 解析（可选）
MINERU_API_URL=https://mineru.net/api/v4
MINERU_TOKEN=your_token_here
```

#### 2.3 验证配置
```bash
go run -tags test ../tests/test-mcp-ai-config.go
```

### 3. 启动 MCP 服务器

```bash
./server/bin/zoteroflow2 mcp
```

服务器将在 stdio 模式下启动，等待 MCP 客户端连接。

## 🛠️ 可用工具

### 本地工具 (6个)

| 工具名称 | 功能描述 | 参数 |
|---------|---------|------|
| `zotero_search` | 搜索本地文献库 | `query`, `limit` |
| `zotero_list_items` | 列出文献项目 | `limit`, `offset` |
| `zotero_find_by_doi` | DOI精确查找 | `doi` |
| `zotero_get_stats` | 数据库统计信息 | 无 |
| `mineru_parse` | PDF解析 | `file_path`, `output_format` |
| `zotero_chat` | AI文献对话 | `message`, `document_id` |

### 外部工具 (Article MCP)

| 工具名称 | 功能描述 |
|---------|---------|
| `search_europe_pmc` | Europe PMC文献搜索 |
| `search_arxiv_papers` | arXiv预印本搜索 |
| `get_article_details` | 获取文献详细信息 |
| `get_references_by_doi` | 获取参考文献列表 |
| `batch_enrich_references_by_dois` | 批量DOI信息补全 |
| `get_similar_articles` | 获取相似文章 |
| `get_citing_articles` | 获取引用文献 |
| `get_literature_relations` | 获取文献关联信息 |
| `get_journal_quality` | 期刊质量评估 |
| `evaluate_articles_quality` | 批量文献质量评估 |

## 🔗 Claude Desktop 集成

### 1. 获取配置

运行配置验证工具会自动生成 Claude Desktop 配置：
```bash
go run -tags test ../tests/test-mcp-ai-config.go
```

配置文件保存在：`docs/claude-desktop-config.json`

### 2. 配置 Claude Desktop

将生成的配置添加到 Claude Desktop 的配置文件中：

```json
{
  "mcpServers": {
    "zoteroflow2": {
      "command": "/path/to/zoteroflow2/server/bin/zoteroflow2",
      "args": ["mcp"],
      "env": {
        "ZOTERO_DB_PATH": "/path/to/zotero.sqlite",
        "ZOTERO_DATA_DIR": "/path/to/zotero/storage"
      }
    }
  }
}
```

### 3. 重启 Claude Desktop

重启 Claude Desktop 后，就可以开始使用 ZoteroFlow2 的 MCP 工具了。

## 📝 使用示例

### 示例 1: 搜索本地文献

在 Claude Desktop 中询问：
```
请搜索我本地文献库中关于"机器学习"的文献，限制返回5个结果。
```

Claude 会调用 `zotero_search` 工具：
```json
{
  "name": "zotero_search",
  "arguments": {
    "query": "机器学习",
    "limit": 5
  }
}
```

### 示例 2: AI 文献对话

```
这篇论文的主要贡献是什么？请基于文献ID 12345 进行回答。
```

Claude 会调用 `zotero_chat` 工具：
```json
{
  "name": "zotero_chat",
  "arguments": {
    "message": "这篇论文的主要贡献是什么？",
    "document_id": "12345"
  }
}
```

### 示例 3: PDF 解析

```
请帮我解析这个PDF文件：/path/to/paper.pdf
```

Claude 会调用 `mineru_parse` 工具：
```json
{
  "name": "mineru_parse",
  "arguments": {
    "file_path": "/path/to/paper.pdf",
    "output_format": "json"
  }
}
```

## 🔧 外部 MCP 服务器

### 配置外部服务器

编辑 `server/external-mcp-servers.json`：

```json
{
  "external_mcp_servers": {
    "article_mcp": {
      "enabled": true,
      "command": "uvx",
      "args": ["article-mcp", "server"],
      "timeout": 30,
      "auto_start": true,
      "env": {
        "PYTHONUNBUFFERED": "1"
      }
    }
  }
}
```

### 使用 Article MCP

**当前状态**: Article MCP 需要独立启动

```bash
# 终端1: 启动 ZoteroFlow2 MCP
./server/bin/zoteroflow2 mcp

# 终端2: 启动 Article MCP (可选)
uvx article-mcp server
```

**未来版本**: 将支持外部 MCP 服务器的自动代理。

## 🧪 测试和验证

### 1. 基础功能测试
```bash
cd tests/
go run test_mcp_basic.go
```

### 2. Article MCP 集成测试
```bash
cd tests/
python3 test_article_mcp.py
```

### 3. 完整项目状态检查
```bash
go run MCP_STATUS_CHECK.go
```

## 🔍 故障排除

### 常见问题

#### 1. MCP 服务器启动失败
- 检查 Zotero 数据库路径是否正确
- 确认环境变量配置正确
- 查看错误日志信息

#### 2. 工具调用失败
- 确认工具参数格式正确
- 检查相关依赖服务（如 MinerU API）
- 验证权限设置

#### 3. Claude Desktop 连接问题
- 确认配置文件路径正确
- 检查二进制文件权限
- 重启 Claude Desktop

### 调试模式

启用详细日志：
```bash
export LOG_LEVEL=debug
./server/bin/zoteroflow2 mcp
```

## 📊 性能优化

### 1. 本地工具性能
- Zotero 数据库查询：通常 < 100ms
- PDF 解析：取决于文件大小和网络
- AI 对话：取决于模型和上下文长度

### 2. Article MCP 性能
- Europe PMC 搜索：2-5秒
- arXiv 搜索：1-3秒
- 批量处理：比逐个处理快 6-10 倍

## 🔮 未来功能

### 计划中的功能
- [ ] 外部 MCP 服务器自动代理
- [ ] 工具权限管理
- [ ] 性能监控面板
- [ ] Web 管理界面
- [ ] 多用户支持

### 扩展性
- 支持更多外部 MCP 服务器
- 插件系统
- 自定义工具开发
- 云端同步功能

## 📚 更多资源

- [MCP 协议规范](https://spec.modelcontextprotocol.io/)
- [Article MCP 项目](https://github.com/gqy20/article-mcp)
- [项目架构文档](docs/mcp-integration-plan.md)
- [外部配置指南](docs/external-mcp-configuration.md)
- [工具详细说明](docs/api/mcp-tools-list.md)

## 🤝 贡献和支持

如有问题或建议，请：
1. 检查本文档的故障排除部分
2. 查看项目的 GitHub Issues
3. 运行诊断工具：`go run MCP_STATUS_CHECK.go`
4. 提交新的 Issue 或 Pull Request

---

**最后更新**: 2025-10-04
**版本**: v0.8
**状态**: ✅ 生产就绪