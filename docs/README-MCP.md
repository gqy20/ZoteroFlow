# ZoteroFlow2 MCP 服务器

## 🎯 一句话说明

ZoteroFlow2 MCP 服务器让 AI 客户端（如 Claude Desktop）能够直接访问和管理你的 Zotero 文献库。

## 🚀 5分钟快速上手

### 1. 确保项目配置正确
```bash
# 检查配置文件
cat .env | grep -E "(AI_|ZOTERO_|MINERU_)"
```

### 2. 编译项目
```bash
cd server
make build
```

### 3. Claude Desktop 配置
复制以下配置到 Claude Desktop：

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

### 4. 重启 Claude Desktop
配置完成后重启 Claude Desktop 即可使用。

## ✨ 功能特性

- **📚 文献搜索**：AI 可以搜索你的 Zotero 文献库
- **📄 PDF 解析**：使用 MinerU 解析 PDF 文件内容
- **🤖 AI 对话**：基于文献的智能问答
- **📊 统计分析**：获取文献库统计信息
- **🔗 DOI 查找**：通过 DOI 快速定位文献

## 💡 核心优势

- **零配置 AI**：自动使用项目现有的 GLM-4.6 配置
- **标准协议**：完全兼容 MCP v2024-11-05 规范
- **即开即用**：无需复杂的额外设置

## 🛠️ 支持的工具

| 工具名称 | 功能 | 示例 |
|---------|------|------|
| `zotero_search` | 搜索文献 | "搜索机器学习相关论文" |
| `zotero_chat` | AI 文献对话 | "分析这篇论文的主要贡献" |
| `mineru_parse` | PDF 解析 | "解析这个 PDF 的内容" |
| `zotero_get_stats` | 统计信息 | "我的文献库有多少篇文章？" |

## 🔧 故障排除

**工具列表为空** → 检查 Zotero 数据库路径
**AI 无响应** → 检查网络连接和 API 配置
**PDF 解析失败** → 检查 MinerU 配置

## 📖 更多文档

- [完整使用指南](mcp-usage-guide.md)
- [集成方案详情](mcp-integration-plan.md)
- [Claude Desktop 配置示例](claude-desktop-config.json)

---
**当前 AI 配置**：智谱 GLM-4.6（自动使用，无需额外配置）