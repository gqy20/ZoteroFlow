# 改进的MCP集成方案

## 当前问题分析

1. **手动MCP协议实现复杂**：当前代码手动实现JSON-RPC协议，容易出错
2. **Article MCP参数格式问题**：工具调用参数格式不匹配
3. **缺乏统一的MCP客户端**：没有使用成熟的MCP客户端库

## 推荐的解决方案

### 方案1: 使用现有MCP客户端库

```go
// 使用成熟的MCP客户端库，如 github.com/modelcontextprotocol/go-sdk
import "github.com/modelcontextprotocol/go-sdk/client"

func createMCPClient() (*client.Client, error) {
    // 创建标准MCP客户端
    return client.New(client.Options{
        ClientInfo: client.ClientInfo{
            Name:    "zoteroflow2",
            Version: "1.0.0",
        },
    })
}
```

### 方案2: 简化的外部调用方案

```go
// 直接调用uvx article-mcp，包装成简单命令
func searchGlobalLiterature(keyword string) ([]DocumentSummary, error) {
    cmd := exec.Command("uvx", "article-mcp", "search", "--keyword", keyword, "--limit", "5")
    output, err := cmd.Output()
    if err != nil {
        return nil, err
    }

    // 解析JSON输出
    var results []DocumentSummary
    json.Unmarshal(output, &results)
    return results, nil
}
```

### 方案3: 配置文件驱动的MCP集成

```json
{
  "mcpServers": {
    "article-mcp": {
      "command": "uvx",
      "args": ["article-mcp", "server"],
      "tools": ["search_europe_pmc", "search_arxiv", "get_article_details"]
    },
    "context7": {
      "command": "npx",
      "args": ["-y", "@upstash/context7-mcp"]
    }
  }
}
```

## 建议的实施步骤

1. **立即可用方案**：暂时简化为直接CLI调用
2. **中期优化**：集成成熟的MCP Go客户端库
3. **长期目标**：完全基于MCP协议的插件化架构

## 代码改进建议

### 替换当前的ArticleMCPClient

```go
// 新的简化版本
type SimpleArticleClient struct{}

func NewSimpleArticleClient() *SimpleArticleClient {
    return &SimpleArticleClient{}
}

func (c *SimpleArticleClient) Search(keyword string) ([]DocumentSummary, error) {
    // 调用uvx article-mcp search
    cmd := exec.Command("uvx", "article-mcp", "search", keyword, "--format", "json", "--limit", "5")

    var output bytes.Buffer
    cmd.Stdout = &output

    if err := cmd.Run(); err != nil {
        return nil, fmt.Errorf("搜索失败: %w", err)
    }

    // 解析结果
    var results struct {
        Articles []DocumentSummary `json:"articles"`
    }

    if err := json.Unmarshal(output.Bytes(), &results); err != nil {
        return nil, fmt.Errorf("解析结果失败: %w", err)
    }

    return results.Articles, nil
}
```

## 总结

当前的MCP集成确实不够完善，主要原因：
1. 手动实现MCP协议过于复杂
2. 缺乏成熟的客户端库支持
3. 测试和调试困难

建议采用更简单实用的方案，先保证功能可用，再逐步优化架构。