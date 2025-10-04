# ZoteroFlow2 MCP 服务器使用指南

## 📋 概述

ZoteroFlow2 MCP 服务器为 AI 客户端（如 Claude Desktop）提供强大的学术文献管理能力。**重要特性**：自动使用项目现有的 AI 配置，无需重复设置。

## 🚀 快速开始

### 1. 准备工作

确保项目已正确配置：
```bash
# 检查 .env 文件是否存在
ls -la .env

# 查看 .env 文件内容（确认 AI 配置存在）
cat .env
```

### 2. Claude Desktop 配置

**最小配置（推荐）**：
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

### 3. 重启 Claude Desktop

配置完成后，重启 Claude Desktop 即可开始使用。

## 🔧 配置优先级

配置按以下优先级加载：

1. **MCP 客户端环境变量**（最高优先级）
2. **项目 .env 文件**
3. **代码默认值**（最低优先级）

### 配置示例对比

#### 方式1：仅设置必要参数（推荐）
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
- ✅ AI 配置自动从 `.env` 文件读取
- ✅ MinerU 配置自动从 `.env` 文件读取
- ✅ 使用默认超时和缓存设置

#### 方式2：覆盖特定配置
```json
{
  "mcpServers": {
    "zoteroflow2": {
      "command": "/path/to/zoteroflow2",
      "args": ["mcp"],
      "env": {
        "ZOTERO_DB_PATH": "/path/to/zotero.sqlite",
        "ZOTERO_DATA_DIR": "/path/to/zotero/storage",
        "AI_API_KEY": "special_api_key_for_claude",
        "AI_MODEL": "gpt-4o"
      }
    }
  }
}
```
- ✅ 覆盖了 AI 配置，使用 Claude 专用的 API 密钥
- ✅ 其他配置仍从 `.env` 文件读取

## 📚 可用工具

### Zotero 相关工具

| 工具名称 | 功能描述 | 参数 |
|---------|---------|------|
| `zotero_search` | 按标题搜索文献 | `query` (必需), `limit` (可选) |
| `zotero_list_items` | 列出有 PDF 的文献项 | `limit` (可选) |
| `zotero_find_by_doi` | 通过 DOI 查找文献 | `doi` (必需) |
| `zotero_get_stats` | 获取数据库统计信息 | 无 |

### PDF 解析工具

| 工具名称 | 功能描述 | 参数 |
|---------|---------|------|
| `mineru_parse` | 使用 MinerU 解析 PDF | `pdf_path` (必需) |

### AI 对话工具

| 工具名称 | 功能描述 | 参数 |
|---------|---------|------|
| `zotero_chat` | AI 学术文献对话 | `message` (必需), `doc_name` (可选) |

## 💡 使用示例

### 1. 搜索文献
```
AI：请搜索关于"机器学习"的相关文献
（AI 会调用 zotero_search 工具）
```

### 2. 获取文献统计
```
AI：显示我的文献库统计信息
（AI 会调用 zotero_get_stats 工具）
```

### 3. 解析 PDF
```
AI：请解析这个 PDF 文件：/path/to/paper.pdf
（AI 会调用 mineru_parse 工具）
```

### 4. 学术对话
```
AI：基于我的文献库，介绍一下深度学习的发展历程
（AI 会调用 zotero_chat 工具）
```

## ⚙️ 当前 AI 配置

项目已配置以下 AI 模型（自动使用）：

- **模型**：智谱 GLM-4.6
- **API 端点**：https://open.bigmodel.cn/api/coding/paas/v4
- **超时时间**：20 秒
- **上下文管理**：支持长文本处理

## 🛠️ 故障排除

### 常见问题

#### 1. 工具列表为空
**原因**：Zotero 数据库连接失败
**解决**：检查 `ZOTERO_DB_PATH` 和 `ZOTERO_DATA_DIR` 路径是否正确

#### 2. AI 对话无响应
**原因**：AI API 密钥未配置或无效
**解决**：检查 `.env` 文件中的 `AI_API_KEY` 配置

#### 3. PDF 解析失败
**原因**：MinerU API 配置问题
**解决**：检查 `.env` 文件中的 `MINERU_TOKEN` 配置

### 调试方法

#### 查看 MCP 服务器日志
```bash
# 手动启动 MCP 服务器查看详细日志
./bin/zoteroflow2 mcp
```

#### 测试配置
```bash
# 测试基础功能
./bin/zoteroflow2 list
./bin/zoteroflow2 search "测试"
```

## 📖 高级配置

### 自定义配置文件
可以在项目根目录创建不同的 `.env` 文件：

```bash
# 生产环境配置
cp .env .env.production

# 开发环境配置
cp .env .env.development
```

然后在启动时指定：
```bash
export ENV_FILE=.env.production
./bin/zoteroflow2 mcp
```

### 性能优化

```bash
# 在 .env 文件中添加优化配置
CACHE_DIR=/tmp/zoteroflow_cache    # 使用更快的缓存目录
AI_TIMEOUT=30                       # 增加超时时间
MINERU_TIMEOUT=120                  # 增加 PDF 解析超时
```

## 🤝 集成其他 MCP 服务器

ZoteroFlow2 可以与其他 MCP 服务器同时使用：

```json
{
  "mcpServers": {
    "zoteroflow2": {
      "command": "/path/to/zoteroflow2",
      "args": ["mcp"],
      "env": {"ZOTERO_DB_PATH": "/path/to/zotero.sqlite"}
    },
    "article_mcp": {
      "command": "uv",
      "args": ["tool", "run", "article-mcp", "server"]
    },
    "websearch": {
      "command": "python3",
      "args": ["/path/to/websearch-mcp/server.py"]
    }
  }
}
```

## 📞 技术支持

如遇问题，请检查：
1. Zotero 数据库文件是否存在
2. `.env` 文件配置是否正确
3. 网络连接是否正常（用于 AI API 调用）

更多帮助请查看项目文档或提交 Issue。