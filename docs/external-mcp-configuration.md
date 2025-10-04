# 外部MCP服务器配置

ZoteroFlow2 MCP 服务器支持通过JSON配置文件导入外部MCP服务器，扩展其功能。

## 配置文件格式

配置文件位于 `server/external-mcp-servers.json`：

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
    },
    "filesystem": {
      "enabled": false,
      "command": "npx",
      "args": ["-y", "@modelcontextprotocol/server-filesystem", "/tmp"],
      "timeout": 10,
      "auto_start": false
    }
  }
}
```

## 配置字段说明

### 服务器配置字段

- `enabled`: 是否启用该服务器
- `command`: 启动命令
- `args`: 命令参数数组
- `timeout`: 启动超时时间（秒）
- `auto_start`: 是否自动启动
- `env`: 环境变量（可选）

## 示例配置

### 1. Article MCP（学术文献搜索）

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

### 2. 文件系统访问

```json
{
  "external_mcp_servers": {
    "filesystem": {
      "enabled": true,
      "command": "npx",
      "args": ["-y", "@modelcontextprotocol/server-filesystem", "/tmp"],
      "timeout": 10,
      "auto_start": true
    }
  }
}
```

### 3. GitHub 集成

```json
{
  "external_mcp_servers": {
    "github": {
      "enabled": true,
      "command": "npx",
      "args": ["-y", "@modelcontextprotocol/server-github"],
      "timeout": 15,
      "auto_start": true,
      "env": {
        "GITHUB_TOKEN": "your_github_token_here"
      }
    }
  }
}
```

## 使用方法

1. 编辑 `server/external-mcp-servers.json` 文件
2. 添加需要的外部MCP服务器配置
3. 启动 ZoteroFlow2 MCP 服务器：

```bash
./server/bin/zoteroflow2 mcp
```

4. 服务器会自动加载启用的外部MCP服务器

## 支持的外部MCP服务器

### 官方服务器

- **@modelcontextprotocol/server-filesystem**: 文件系统访问
- **@modelcontextprotocol/server-github**: GitHub 集成
- **@modelcontextprotocol/server-sqlite**: SQLite 数据库访问
- **@modelcontextprotocol/server-postgres**: PostgreSQL 数据库访问

### 社区服务器

- **article-mcp**: 学术文献搜索 (Europe PMC + arXiv)
- **更多服务器持续添加中...**

## 注意事项

1. **安全性**: 外部服务器可能访问系统资源，请谨慎配置
2. **性能**: 外部服务器启动会增加响应时间
3. **兼容性**: 确保外部服务器与当前MCP协议版本兼容
4. **环境变量**: 某些服务器需要特定的环境变量或API密钥

## 故障排除

### 常见问题

1. **服务器启动失败**
   - 检查命令路径是否正确
   - 确认参数格式无误
   - 查看错误日志

2. **超时问题**
   - 增加 `timeout` 值
   - 检查网络连接
   - 确认服务器响应正常

3. **权限问题**
   - 检查文件权限
   - 确认环境变量设置
   - 验证API密钥有效性

### 调试方法

启用详细日志：

```bash
export LOG_LEVEL=debug
./server/bin/zoteroflow2 mcp
```

## 配置验证

使用配置验证工具：

```bash
go run -tags test server/tests/test-mcp-ai-config.go
```

这将验证配置文件格式并生成Claude Desktop配置。