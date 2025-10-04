# ZoteroFlow2 基础使用示例

## 概述

本文档提供了 ZoteroFlow2 的基础使用示例，包括 CLI 命令、核心模块 API 调用、MCP 集成等常见使用场景的完整代码示例。

## 环境准备

### 1. 安装依赖

```bash
# 安装 Go 1.21+
go version

# 安装 uv (用于 Article MCP)
pip install uv

# 验证安装
uv --version
```

### 2. 配置环境变量

```bash
# 创建 .env 文件
cat > .env << EOF
# Zotero 配置
ZOTERO_DB_PATH=~/Zotero/zotero.sqlite
ZOTERO_DATA_DIR=~/Zotero/storage

# MinerU 配置
MINERU_API_URL=https://mineru.net/api/v4
MINERU_TOKEN=your_mineru_token_here

# AI 配置
AI_API_KEY=your_ai_api_key_here
AI_BASE_URL=https://open.bigmodel.cn/api/coding/paas/v4
AI_MODEL=glm-4.6

# 缓存配置
CACHE_DIR=~/.zoteroflow/cache
RESULTS_DIR=data/results
RECORDS_DIR=data/records
EOF
```

### 3. 构建项目

```bash
cd server/
make build
```

## CLI 使用示例

### 1. 基础文献管理

#### 列出所有文献

```bash
./bin/zoteroflow2 list
```

**预期输出**:
```
找到 3 个解析结果:

[1] 机器学习基础_20241201
     标题: 机器学习基础教程
     作者: 张三; 李四
     大小: 2.3 MB
     日期: 2024-12-01

[2] 深度学习研究_20241130
     标题: 深度学习在医疗诊断中的应用
     作者: 王五; 赵六
     大小: 1.8 MB
     日期: 2024-11-30
```

#### 搜索文献

```bash
# 按标题搜索
./bin/zoteroflow2 search "机器学习"

# 按 DOI 搜索
./bin/zoteroflow2 doi "10.1234/ml.2024.001"
```

#### 查看解析结果

```bash
# 打开文献文件夹
./bin/zoteroflow2 open "机器学习基础"

# 查看文件内容
ls -la data/results/机器学习基础_20241201/
```

### 2. AI 对话示例

#### 单次问答

```bash
./bin/zoteroflow2 chat "什么是机器学习？"
```

#### 交互式对话

```bash
./bin/zoteroflow2 chat
```

**对话示例**:
```
🤖 ZoteroFlow2 AI学术助手
输入 'help' 查看帮助，输入 'quit' 或 'exit' 退出
--------------------------------------------------
📚 您: 什么是深度学习？
🤖 助手: 深度学习是机器学习的一个分支...

📚 您: 深度学习和传统机器学习有什么区别？
🤖 助手: 主要区别包括：
1. 特征工程：传统ML需要手动设计特征，DL自动学习特征...

📚 您: quit
👋 再见!
```

#### 基于文献的对话

```bash
./bin/zoteroflow2 chat --doc="机器学习基础" "请总结这篇文章的主要内容"
```

## Go API 使用示例

### 1. Zotero 数据库访问

```go
package main

import (
    "fmt"
    "log"
    "strings"
    "zoteroflow2-server/core"
)

func main() {
    // 连接 Zotero 数据库
    zoteroDB, err := core.NewZoteroDB(
        "~/Zotero/zotero.sqlite",
        "~/Zotero/storage",
    )
    if err != nil {
        log.Fatalf("连接数据库失败: %v", err)
    }
    defer zoteroDB.Close()

    // 获取有 PDF 附件的文献
    items, err := zoteroDB.GetItemsWithPDF(5)
    if err != nil {
        log.Printf("查询失败: %v", err)
        return
    }

    fmt.Printf("找到 %d 篇有PDF附件的文献:\n", len(items))
    for i, item := range items {
        fmt.Printf("[%d] %s\n", i+1, item.Title)
        fmt.Printf("    作者: %s\n", strings.Join(item.Authors, "; "))
        fmt.Printf("    PDF路径: %s\n", item.PDFPath)
        fmt.Println()
    }
}
```

### 2. 文献搜索

```go
package main

import (
    "fmt"
    "log"
    "zoteroflow2-server/core"
)

func main() {
    // 连接数据库
    zoteroDB, err := core.NewZoteroDB(
        "~/Zotero/zotero.sqlite",
        "~/Zotero/storage",
    )
    if err != nil {
        log.Fatalf("连接数据库失败: %v", err)
    }
    defer zoteroDB.Close()

    // 搜索文献
    results, err := zoteroDB.SearchByTitle("机器学习", 5)
    if err != nil {
        log.Printf("搜索失败: %v", err)
        return
    }

    fmt.Printf("搜索 '%s' 找到 %d 篇文献:\n", "机器学习", len(results))
    for i, result := range results {
        fmt.Printf("[%d] %s (评分: %.1f)\n", i+1, result.Title, result.Score)
        fmt.Printf("    作者: %s\n", strings.Join(result.Authors, "; "))
        if result.DOI != "" {
            fmt.Printf("    DOI: %s\n", result.DOI)
        }
        fmt.Printf("    PDF路径: %s\n", result.PDFPath)
        fmt.Println()
    }
}
```

### 3. PDF 解析

```go
package main

import (
    "context"
    "fmt"
    "log"
    "time"
    "zoteroflow2-server/core"
)

func main() {
    // 初始化组件
    zoteroDB, err := core.NewZoteroDB(
        "~/Zotero/zotero.sqlite",
        "~/Zotero/storage",
    )
    if err != nil {
        log.Fatalf("连接数据库失败: %v", err)
    }
    defer zoteroDB.Close()

    mineruClient := core.NewMinerUClient(
        "https://mineru.net/api/v4",
        "your_token_here",
    )

    parser, err := core.NewPDFParser(zoteroDB, mineruClient, "~/.zoteroflow/cache")
    if err != nil {
        log.Fatalf("创建解析器失败: %v", err)
    }

    // 获取文献
    items, err := zoteroDB.GetItemsWithPDF(1)
    if err != nil {
        log.Fatalf("获取文献失败: %v", err)
    }

    if len(items) == 0 {
        fmt.Println("没有找到PDF文献")
        return
    }

    // 解析 PDF
    item := items[0]
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
    defer cancel()

    doc, err := parser.ParseDocument(ctx, item.ItemID, item.PDFPath)
    if err != nil {
        log.Printf("解析失败: %v", err)
        return
    }

    fmt.Printf("解析成功:\n")
    fmt.Printf("  标题: %s\n", doc.ZoteroItem.Title)
    fmt.Printf("  缓存键: %s\n", doc.ParseHash)
    fmt.Printf("  解析时间: %s\n", doc.ParseTime.Format("2006-01-02 15:04:05"))
    fmt.Printf("  ZIP路径: %s\n", doc.ZipPath)
}
```

### 4. AI 对话

```go
package main

import (
    "context"
    "fmt"
    "log"
    "zoteroflow2-server/core"
)

func main() {
    // 创建 AI 客户端
    aiClient := core.NewGLMClient(
        "your_api_key",
        "https://open.bigmodel.cn/api/coding/paas/v4",
        "glm-4.6",
    )

    // 创建对话请求
    messages := []core.ChatMessage{
        {
            Role:    "system",
            Content: "你是一个专业的学术文献助手，请用中文回答。",
        },
        {
            Role:    "user",
            Content: "什么是机器学习？请简要解释。",
        },
    }

    req := &core.AIRequest{
        Model:    "glm-4.6",
        Messages: messages,
        MaxTokens: 500,
    }

    // 发送请求
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()

    response, err := aiClient.Chat(ctx, req)
    if err != nil {
        log.Fatalf("AI调用失败: %v", err)
    }

    // 处理响应
    if len(response.Choices) > 0 {
        fmt.Printf("AI回答: %s\n", response.Choices[0].Message.Content)
        fmt.Printf("Token使用: %d\n", response.Usage.TotalTokens)
    }
}
```

### 5. 对话管理

```go
package main

import (
    "context"
    "fmt"
    "log"
    "zoteroflow2-server/core"
)

func main() {
    // 初始化组件
    aiClient := core.NewGLMClient(
        "your_api_key",
        "https://open.bigmodel.cn/api/coding/paas/v4",
        "glm-4.6",
    )

    zoteroDB, err := core.NewZoteroDB(
        "~/Zotero/zotero.sqlite",
        "~/Zotero/storage",
    )
    if err != nil {
        log.Fatalf("连接数据库失败: %v", err)
    }
    defer zoteroDB.Close()

    // 创建对话管理器
    chatManager := core.NewAIConversationManager(aiClient, zoteroDB)

    // 开始新对话
    conv, err := chatManager.StartConversation(
        context.Background(),
        "请分析一下机器学习在医疗领域的应用现状",
        nil, // 不指定特定文档
    )
    if err != nil {
        log.Fatalf("开始对话失败: %v", err)
    }

    fmt.Printf("对话ID: %s\n", conv.ID)
    if len(conv.Messages) >= 3 {
        aiResponse := conv.Messages[2]
        fmt.Printf("AI回答: %s\n", aiResponse.Content)
    }

    // 继续对话
    conv, err = chatManager.ContinueConversation(
        context.Background(),
        conv.ID,
        "能具体介绍一下深度学习在医学影像诊断中的应用吗？",
    )
    if err != nil {
        log.Printf("继续对话失败: %v", err)
        return
    }

    if len(conv.Messages) >= 2 {
        lastMsg := conv.Messages[len(conv.Messages)-1]
        if lastMsg.Role == "assistant" {
            fmt.Printf("AI回答: %s\n", lastMsg.Content)
        }
    }
}
```

## MCP 集成示例

### 1. Article MCP 基础使用

```go
package main

import (
    "fmt"
    "log"
    "zoteroflow2-server/mcp"
)

func main() {
    // 创建 MCP 客户端
    client := mcp.NewMCPClient([]string{
        "uv", "tool", "run", "article-mcp", "server",
    })

    // 启动服务器
    if err := client.Start(); err != nil {
        log.Fatalf("启动 MCP 服务器失败: %v", err)
    }
    defer client.Stop()

    // 初始化连接
    serverInfo, err := client.Initialize("zoteroflow-test", "1.0.0")
    if err != nil {
        log.Fatalf("初始化失败: %v", err)
    }

    fmt.Printf("MCP 连接成功: %s v%s\n",
        serverInfo.ServerInfo.Name,
        serverInfo.ServerInfo.Version)

    // 获取工具列表
    tools, err := client.ListTools()
    if err != nil {
        log.Printf("获取工具列表失败: %v", err)
        return
    }

    fmt.Printf("发现 %d 个工具:\n", len(tools.Tools))
    for i, tool := range tools.Tools {
        fmt.Printf("  %d. %s - %s\n", i+1, tool.Name, tool.Description)
    }
}
```

### 2. 文献搜索集成

```go
package main

import (
    "fmt"
    "log"
    "zoteroflow2-server/mcp"
)

func main() {
    // 创建 MCP 客户端
    client := mcp.NewMCPClient([]string{
        "uv", "tool", "run", "article-mcp", "server",
    })

    // 启动和初始化
    if err := client.Start(); err != nil {
        log.Fatalf("启动失败: %v", err)
    }
    defer client.Stop()

    _, err := client.Initialize("zoteroflow-test", "1.0.0")
    if err != nil {
        log.Fatalf("初始化失败: %v", err)
    }

    // 搜索 Europe PMC
    fmt.Println("=== 搜索 Europe PMC ===")
    europeResult, err := client.CallTool("search_europe_pmc", map[string]interface{}{
        "keyword":     "machine learning",
        "max_results":  5,
    })
    if err != nil {
        log.Printf("Europe PMC 搜索失败: %v", err)
    } else {
        fmt.Printf("✅ Europe PMC 搜索成功\n")
        displayResult(europeResult)
    }

    // 搜索 arXiv
    fmt.Println("\n=== 搜索 arXiv ===")
    arxivResult, err := client.CallTool("search_arxiv_papers", map[string]interface{}{
        "keyword":     "deep learning",
        "max_results":  3,
    })
    if err != nil {
        log.Printf("arXiv 搜索失败: %v", err)
    } else {
        fmt.Printf("✅ arXiv 搜索成功\n")
        displayResult(arxivResult)
    }
}

func displayResult(result *mcp.CallToolResult) {
    for _, content := range result.Content {
        if content.Type == "text" {
            text := content.Text
            if len(text) > 200 {
                text = text[:200] + "..."
            }
            fmt.Printf("📄 %s\n", text)
        }
    }
}
```

### 3. 智能文献分析

```go
package main

import (
    "fmt"
    "log"
    "zoteroflow2-server/mcp"
)

func main() {
    client := mcp.NewMCPClient([]string{
        "uv", "tool", "run", "article-mcp", "server",
    })

    if err := client.Start(); err != nil {
        log.Fatalf("启动失败: %v", err)
    }
    defer client.Stop()

    _, err := client.Initialize("zoteroflow-test", "1.0.0")
    if err != nil {
        log.Fatalf("初始化失败: %v", err)
    }

    // 智能文献分析
    query := "人工智能在医疗诊断中的应用"
    fmt.Printf("🔍 开始智能文献分析: %s\n", query)

    // 1. 搜索 Europe PMC
    fmt.Println("\n📚 步骤1: 搜索 Europe PMC 数据库")
    europeResult, err := client.CallTool("search_europe_pmc", map[string]interface{}{
        "keyword":     query,
        "max_results":  10,
    })
    if err != nil {
        fmt.Printf("❌ Europe PMC 搜索失败: %v\n", err)
    } else {
        fmt.Printf("✅ Europe PMC 搜索成功\n")
        displayResult(europeResult)
    }

    // 2. 搜索 arXiv
    fmt.Println("\n📖 步骤2: 搜索 arXiv 数据库")
    arxivResult, err := client.CallTool("search_arxiv_papers", map[string]interface{}{
        "keyword":     query,
        "max_results":  5,
    })
    if err != nil {
        fmt.Printf("❌ arXiv 搜索失败: %v\n", err)
    } else {
        fmt.Printf("✅ arXiv 搜索成功\n")
        displayResult(arxivResult)
    }

    // 3. 获取相似文献（假设有DOI）
    doi := "10.1038/s41586-021-03464-6"
    fmt.Println("\n🔗 步骤3: 获取相似文献")
    similarResult, err := client.CallTool("get_similar_articles", map[string]interface{}{
        "identifier": doi,
        "id_type":    "doi",
        "max_results": 5,
    })
    if err != nil {
        fmt.Printf("❌ 获取相似文献失败: %v\n", err)
    } else {
        fmt.Printf("✅ 相似文献获取成功\n")
        displayResult(similarResult)
    }

    fmt.Println("\n🎉 智能文献分析完成！")
}

func displayResult(result *mcp.CallToolResult) {
    for _, content := range result.Content {
        if content.Type == "text" {
            fmt.Printf("📄 %s\n", content.Text)
        }
    }
}
```

## 完整工作流示例

### 1. 文献搜索 → 解析 → AI 分析

```go
package main

import (
    "context"
    "fmt"
    "log"
    "strings"
    "time"
    "zoteroflow2-server/core"
    "zoteroflow2-server/mcp"
)

func main() {
    // 1. 初始化所有组件
    zoteroDB, err := core.NewZoteroDB(
        "~/Zotero/zotero.sqlite",
        "~/Zotero/storage",
    )
    if err != nil {
        log.Fatalf("连接数据库失败: %v", err)
    }
    defer zoteroDB.Close()

    mineruClient := core.NewMinerUClient(
        "https://mineru.net/api/v4",
        "your_token",
    )

    parser, err := core.NewPDFParser(zoteroDB, mineruClient, "~/.zoteroflow/cache")
    if err != nil {
        log.Fatalf("创建解析器失败: %v", err)
    }

    aiClient := core.NewGLMClient(
        "your_api_key",
        "https://open.bigmodel.cn/api/coding/paas/v4",
        "glm-4.6",
    )

    mcpClient := mcp.NewMCPClient([]string{
        "uv", "tool", "run", "article-mcp", "server",
    })

    if err := mcpClient.Start(); err != nil {
        log.Fatalf("启动MCP失败: %v", err)
    }
    defer mcpClient.Stop()

    _, err = mcpClient.Initialize("zoteroflow-workflow", "1.0.0")
    if err != nil {
        log.Fatalf("MCP初始化失败: %v", err)
    }

    // 2. 搜索本地文献
    fmt.Println("=== 步骤1: 搜索本地文献 ===")
    localItems, err := zoteroDB.SearchByTitle("机器学习", 3)
    if err != nil {
        log.Printf("本地搜索失败: %v", err)
        return
    }

    if len(localItems) == 0 {
        fmt.Println("未找到本地文献，继续外部搜索")
    } else {
        fmt.Printf("找到 %d 篇本地文献:\n", len(localItems))
        for i, item := range localItems {
            fmt.Printf("  [%d] %s\n", i+1, item.Title)
        }
    }

    // 3. 搜索外部文献
    fmt.Println("\n=== 步骤2: 搜索外部文献 ===")
    externalResult, err := mcpClient.CallTool("search_europe_pmc", map[string]interface{}{
        "keyword":     "machine learning",
        "max_results":  5,
    })
    if err != nil {
        log.Printf("外部搜索失败: %v", err)
    } else {
        fmt.Printf("✅ 外部搜索成功\n")
    }

    // 4. 解析本地PDF
    fmt.Println("\n=== 步骤3: 解析本地PDF ===")
    if len(localItems) > 0 {
        item := localItems[0]
        ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
        defer cancel()

        doc, err := parser.ParseDocument(ctx, item.ItemID, item.PDFPath)
        if err != nil {
            log.Printf("PDF解析失败: %v", err)
        } else {
            fmt.Printf("✅ PDF解析成功: %s\n", doc.ZoteroItem.Title)
        }
    }

    // 5. AI 分析
    fmt.Println("\n=== 步骤4: AI智能分析 ===")
    analysisPrompt := buildAnalysisPrompt(localItems, externalResult)
    
    messages := []core.ChatMessage{
        {
            Role:    "system",
            Content: "你是一个专业的学术文献分析师，请基于提供的文献信息进行深度分析。",
        },
        {
            Role:    "user",
            Content: analysisPrompt,
        },
    }

    req := &core.AIRequest{
        Model:    "glm-4.6",
        Messages: messages,
        MaxTokens: 1000,
    }

    ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
    defer cancel()

    response, err := aiClient.Chat(ctx, req)
    if err != nil {
        log.Printf("AI分析失败: %v", err)
    } else {
        fmt.Printf("✅ AI分析完成\n")
        if len(response.Choices) > 0 {
            fmt.Printf("🤖 分析结果:\n%s\n", response.Choices[0].Message.Content)
        }
    }

    fmt.Println("\n🎉 完整工作流执行完成！")
}

func buildAnalysisPrompt(localItems []core.SearchResult, externalResult *mcp.CallToolResult) string {
    var prompt strings.Builder
    
    prompt.WriteString("请基于以下文献信息进行分析：\n\n")
    
    if len(localItems) > 0 {
        prompt.WriteString("本地文献:\n")
        for i, item := range localItems {
            prompt.WriteString(fmt.Sprintf("%d. %s\n", i+1, item.Title))
            prompt.WriteString(fmt.Sprintf("   作者: %s\n", strings.Join(item.Authors, "; ")))
            if item.DOI != "" {
                prompt.WriteString(fmt.Sprintf("   DOI: %s\n", item.DOI))
            }
        }
        prompt.WriteString("\n")
    }
    
    if externalResult != nil && len(externalResult.Content) > 0 {
        prompt.WriteString("外部搜索结果:\n")
        for _, content := range externalResult.Content {
            if content.Type == "text" {
                prompt.WriteString(fmt.Sprintf("- %s\n", content.Text))
            }
        }
    }
    
    prompt.WriteString("\n请提供以下分析：\n")
    prompt.WriteString("1. 研究主题概述\n")
    prompt.WriteString("2. 主要研究方向和趋势\n")
    prompt.WriteString("3. 关键发现和贡献\n")
    prompt.WriteString("4. 研究建议和未来方向\n")
    
    return prompt.String()
}
```

### 2. 批量处理工作流

```go
package main

import (
    "context"
    "fmt"
    "log"
    "sync"
    "time"
    "zoteroflow2-server/core"
)

func main() {
    // 初始化组件
    zoteroDB, err := core.NewZoteroDB(
        "~/Zotero/zotero.sqlite",
        "~/Zotero/storage",
    )
    if err != nil {
        log.Fatalf("连接数据库失败: %v", err)
    }
    defer zoteroDB.Close()

    mineruClient := core.NewMinerUClient(
        "https://mineru.net/api/v4",
        "your_token",
    )

    parser, err := core.NewPDFParser(zoteroDB, mineruClient, "~/.zoteroflow/cache")
    if err != nil {
        log.Fatalf("创建解析器失败: %v", err)
    }

    // 获取所有文献
    fmt.Println("=== 获取文献列表 ===")
    allItems, err := zoteroDB.GetItemsWithPDF(10)
    if err != nil {
        log.Fatalf("获取文献失败: %v", err)
    }

    fmt.Printf("找到 %d 篇文献，开始批量处理...\n", len(allItems))

    // 批量解析
    fmt.Println("\n=== 批量解析PDF ===")
    var itemIDs []int
    for _, item := range allItems {
        itemIDs = append(itemIDs, item.ItemID)
    }

    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
    defer cancel()

    startTime := time.Now()
    docs, err := parser.BatchParseDocuments(ctx, itemIDs)
    if err != nil {
        log.Printf("批量解析失败: %v", err)
    }
    duration := time.Since(startTime)

    successCount := 0
    for _, doc := range docs {
        if doc != nil {
            successCount++
        }
    }

    fmt.Printf("批量解析完成:\n")
    fmt.Printf("  成功: %d/%d\n", successCount, len(allItems))
    fmt.Printf("  耗时: %v\n", duration)

    // 生成报告
    fmt.Println("\n=== 生成处理报告 ===")
    generateReport(allItems, docs, duration)
}

func generateReport(items []core.ZoteroItem, docs []*core.ParsedDocument, duration time.Duration) {
    var report strings.Builder
    
    report.WriteString("# ZoteroFlow2 批量处理报告\n\n")
    report.WriteString(fmt.Sprintf("处理时间: %s\n", time.Now().Format("2006-01-02 15:04:05")))
    report.WriteString(fmt.Sprintf("总耗时: %v\n\n", duration))
    
    report.WriteString("## 处理统计\n")
    report.WriteString(fmt.Sprintf("- 总文献数: %d\n", len(items)))
    report.WriteString(fmt.Sprintf("- 成功解析: %d\n", len(docs)))
    report.WriteString(fmt.Sprintf("- 失败数量: %d\n", len(items)-len(docs)))
    report.WriteString(fmt.Sprintf("- 成功率: %.1f%%\n\n", float64(len(docs))/float64(len(items))*100))
    
    report.WriteString("## 成功解析的文献\n")
    for i, doc := range docs {
        if doc != nil {
            report.WriteString(fmt.Sprintf("%d. %s\n", i+1, doc.ZoteroItem.Title))
            report.WriteString(fmt.Sprintf("   解析时间: %s\n", doc.ParseTime.Format("15:04:05")))
            report.WriteString(fmt.Sprintf("   缓存键: %s\n", doc.ParseHash))
        }
    }
    
    report.WriteString("\n## 失败的文献\n")
    failedCount := 0
    for i, item := range items {
        found := false
        for _, doc := range docs {
            if doc != nil && doc.ZoteroItem.ItemID == item.ItemID {
                found = true
                break
            }
        }
        if !found {
            report.WriteString(fmt.Sprintf("%d. %s\n", failedCount+1, item.Title))
            failedCount++
        }
    }
    
    // 保存报告
    reportFile := fmt.Sprintf("batch_report_%s.md", time.Now().Format("20060102_150405"))
    err := os.WriteFile(reportFile, []byte(report.String()), 0644)
    if err != nil {
        log.Printf("保存报告失败: %v", err)
    } else {
        fmt.Printf("报告已保存到: %s\n", reportFile)
    }
}
```

## 错误处理示例

### 1. 完整的错误处理

```go
package main

import (
    "fmt"
    "log"
    "os"
    "strings"
    "zoteroflow2-server/core"
)

func main() {
    // 连接数据库（带错误处理）
    zoteroDB, err := connectToZoteroDB()
    if err != nil {
        log.Fatalf("无法连接到Zotero数据库: %v", err)
    }
    defer zoteroDB.Close()

    // 搜索文献（带错误处理）
    items, err := searchLiterature(zoteroDB, "机器学习")
    if err != nil {
        log.Printf("文献搜索失败: %v", err)
        return
    }

    // 处理结果
    if len(items) == 0 {
        fmt.Println("未找到相关文献")
        return
    }

    fmt.Printf("找到 %d 篇文献\n", len(items))
    for _, item := range items {
        fmt.Printf("- %s\n", item.Title)
    }
}

func connectToZoteroDB() (*core.ZoteroDB, error) {
    // 尝试多个可能的路径
    paths := []string{
        "~/Zotero/zotero.sqlite",
        "/home/user/Zotero/zotero.sqlite",
        "./zotero.sqlite",
    }

    for _, path := range paths {
        expandedPath := expandPath(path)
        
        // 检查文件是否存在
        if _, err := os.Stat(expandedPath); err == nil {
            // 尝试连接
            db, err := core.NewZoteroDB(expandedPath, "~/Zotero/storage")
            if err == nil {
                fmt.Printf("成功连接到数据库: %s\n", expandedPath)
                return db, nil
            }
            log.Printf("连接失败: %s, 错误: %v\n", expandedPath, err)
        }
    }

    return nil, fmt.Errorf("无法找到有效的Zotero数据库文件")
}

func searchLiterature(zoteroDB *core.ZoteroDB, query string) ([]core.SearchResult, error) {
    // 参数验证
    if strings.TrimSpace(query) == "" {
        return nil, fmt.Errorf("搜索查询不能为空")
    }

    // 执行搜索
    results, err := zoteroDB.SearchByTitle(query, 10)
    if err != nil {
        return nil, fmt.Errorf("搜索执行失败: %w", err)
    }

    // 结果验证
    if len(results) == 0 {
        return nil, fmt.Errorf("未找到匹配的文献")
    }

    return results, nil
}

func expandPath(path string) string {
    if len(path) > 0 && path[0] == '~' {
        home, _ := os.UserHomeDir()
        if home != "" {
            return home + path[1:]
        }
    }
    return path
}
```

### 2. 重试机制

```go
package main

import (
    "fmt"
    "log"
    "time"
    "zoteroflow2-server/core"
)

func main() {
    zoteroDB, err := connectWithRetry()
    if err != nil {
        log.Fatalf("连接失败: %v", err)
    }
    defer zoteroDB.Close()

    // 使用连接
    items, err := zoteroDB.GetItemsWithPDF(5)
    if err != nil {
        log.Printf("查询失败: %v", err)
        return
    }

    fmt.Printf("找到 %d 篇文献\n", len(items))
}

func connectWithRetry() (*core.ZoteroDB, error) {
    var lastErr error
    
    for attempt := 0; attempt < 3; attempt++ {
        dbPath := "~/Zotero/zotero.sqlite"
        dataDir := "~/Zotero/storage"
        
        fmt.Printf("尝试连接 (第 %d 次)...\n", attempt+1)
        
        db, err := core.NewZoteroDB(dbPath, dataDir)
        if err == nil {
            // 测试连接
            if _, err := db.GetStats(); err == nil {
                fmt.Printf("连接成功！\n")
                return db, nil
            }
            lastErr = fmt.Errorf("连接测试失败: %w", err)
        } else {
            lastErr = err
        }
        
        if attempt < 2 {
            waitTime := time.Duration(attempt+1) * time.Second
            fmt.Printf("等待 %v 后重试...\n", waitTime)
            time.Sleep(waitTime)
        }
    }
    
    return nil, fmt.Errorf("连接失败，最后错误: %w", lastErr)
}
```

## 性能优化示例

### 1. 并发处理

```go
package main

import (
    "fmt"
    "log"
    "sync"
    "time"
    "zoteroflow2-server/core"
)

func main() {
    zoteroDB, err := core.NewZoteroDB(
        "~/Zotero/zotero.sqlite",
        "~/Zotero/storage",
    )
    if err != nil {
        log.Fatalf("连接失败: %v", err)
    }
    defer zoteroDB.Close()

    // 并发搜索多个查询
    queries := []string{
        "机器学习",
        "深度学习",
        "神经网络",
        "人工智能",
        "数据挖掘",
    }

    fmt.Printf("并发搜索 %d 个查询...\n", len(queries))
    
    startTime := time.Now()
    results := concurrentSearch(zoteroDB, queries)
    duration := time.Since(startTime)

    fmt.Printf("并发搜索完成，耗时: %v\n", duration)
    
    for query, items := range results {
        fmt.Printf("查询 '%s': 找到 %d 篇文献\n", query, len(items))
    }
}

func concurrentSearch(zoteroDB *core.ZoteroDB, queries []string) map[string][]core.SearchResult {
    var wg sync.WaitGroup
    var mu sync.Mutex
    results := make(map[string][]core.SearchResult)
    
    semaphore := make(chan struct{}, 3) // 限制并发数
    
    for _, query := range queries {
        wg.Add(1)
        go func(q string) {
            defer wg.Done()
            
            semaphore <- struct{}{}
            defer func() { <-semaphore }()
            
            items, err := zoteroDB.SearchByTitle(q, 5)
            if err != nil {
                log.Printf("查询 '%s' 失败: %v\n", q, err)
                return
            }
            
            mu.Lock()
            results[q] = items
            mu.Unlock()
        }(query)
    }
    
    wg.Wait()
    return results
}
```

### 2. 缓存优化

```go
package main

import (
    "crypto/md5"
    "fmt"
    "log"
    "sync"
    "time"
    "zoteroflow2-server/core"
)

type CachedSearch struct {
    zoteroDB *core.ZoteroDB
    cache    map[string]*CacheEntry
    mutex    sync.RWMutex
    ttl      time.Duration
}

type CacheEntry struct {
    Results []core.SearchResult
    Time    time.Time
}

func main() {
    zoteroDB, err := core.NewZoteroDB(
        "~/Zotero/zotero.sqlite",
        "~/Zotero/storage",
    )
    if err != nil {
        log.Fatalf("连接失败: %v", err)
    }
    defer zoteroDB.Close()

    cachedSearch := &CachedSearch{
        zoteroDB: zoteroDB,
        cache:    make(map[string]*CacheEntry),
        ttl:      30 * time.Minute,
    }

    // 第一次搜索（会缓存）
    fmt.Println("第一次搜索...")
    results1, err := cachedSearch.Search("机器学习")
    if err != nil {
        log.Printf("搜索失败: %v", err)
        return
    }
    fmt.Printf("找到 %d 篇文献\n", len(results1))

    // 第二次搜索（使用缓存）
    fmt.Println("\n第二次搜索（使用缓存）...")
    results2, err := cachedSearch.Search("机器学习")
    if err != nil {
        log.Printf("搜索失败: %v", err)
        return
    }
    fmt.Printf("找到 %d 篇文献\n", len(results2))

    // 验证缓存效果
    if len(results1) == len(results2) {
        fmt.Println("✅ 缓存生效，结果一致")
    } else {
        fmt.Println("❌ 缓存失效，结果不一致")
    }
}

func (c *CachedSearch) Search(query string) ([]core.SearchResult, error) {
    cacheKey := c.generateCacheKey(query)
    
    // 检查缓存
    c.mutex.RLock()
    if entry, exists := c.cache[cacheKey]; exists {
        if time.Since(entry.Time) < c.ttl {
            c.mutex.RUnlock()
            fmt.Printf("缓存命中: %s\n", query)
            return entry.Results, nil
        }
        c.mutex.RUnlock()
    }
    
    // 缓存未命中，执行搜索
    fmt.Printf("缓存未命中，执行搜索: %s\n", query)
    results, err := c.zoteroDB.SearchByTitle(query, 10)
    if err != nil {
        return nil, err
    }
    
    // 更新缓存
    c.mutex.Lock()
    c.cache[cacheKey] = &CacheEntry{
        Results: results,
        Time:    time.Now(),
    }
    c.mutex.Unlock()
    
    return results, nil
}

func (c *CachedSearch) generateCacheKey(query string) string {
    h := md5.New()
    h.Write([]byte(query))
    return fmt.Sprintf("%x", h.Sum(nil))
}
```

## 测试示例

### 1. 单元测试

```go
package main

import (
    "testing"
    "zoteroflow2-server/core"
)

func TestZoteroDBConnection(t *testing.T) {
    // 测试数据库连接
    db, err := core.NewZoteroDB(
        "~/Zotero/zotero.sqlite",
        "~/Zotero/storage",
    )
    
    if err != nil {
        t.Fatalf("连接失败: %v", err)
    }
    defer db.Close()
    
    // 测试基本查询
    stats, err := db.GetStats()
    if err != nil {
        t.Fatalf("获取统计信息失败: %v", err)
    }
    
    if stats["total_items"] == nil {
        t.Error("总文献数不应为空")
    }
    
    t.Logf("数据库连接测试通过，总文献数: %v", stats["total_items"])
}

func TestPDFParsing(t *testing.T) {
    // 模拟PDF解析测试
    zoteroDB, _ := core.NewZoteroDB(
        "~/Zotero/zotero.sqlite",
        "~/Zotero/storage",
    )
    defer zoteroDB.Close()
    
    mineruClient := core.NewMinerUClient(
        "https://mineru.net/api/v4",
        "test_token",
    )
    
    parser, err := core.NewPDFParser(zoteroDB, mineruClient, "/tmp/cache")
    if err != nil {
        t.Fatalf("创建解析器失败: %v", err)
    }
    
    // 测试解析器创建
    if parser == nil {
        t.Error("解析器不应为空")
    }
    
    t.Log("PDF解析器创建测试通过")
}
```

### 2. 集成测试

```go
package main

import (
    "fmt"
    "log"
    "testing"
    "time"
    "zoteroflow2-server/core"
    "zoteroflow2-server/mcp"
)

func TestCompleteWorkflow(t *testing.T) {
    // 初始化组件
    zoteroDB, err := core.NewZoteroDB(
        "~/Zotero/zotero.sqlite",
        "~/Zotero/storage",
    )
    if err != nil {
        t.Fatalf("连接数据库失败: %v", err)
    }
    defer zoteroDB.Close()

    mineruClient := core.NewMinerUClient(
        "https://mineru.net/api/v4",
        "test_token",
    )

    parser, err := core.NewPDFParser(zoteroDB, mineruClient, "/tmp/cache")
    if err != nil {
        t.Fatalf("创建解析器失败: %v", err)
    }

    // 测试完整工作流
    err = testWorkflow(zoteroDB, parser)
    if err != nil {
        t.Fatalf("工作流测试失败: %v", err)
    }

    t.Log("完整工作流测试通过")
}

func testWorkflow(zoteroDB *core.ZoteroDB, parser *core.PDFParser) error {
    // 1. 获取文献
    items, err := zoteroDB.GetItemsWithPDF(1)
    if err != nil {
        return fmt.Errorf("获取文献失败: %w", err)
    }

    if len(items) == 0 {
        return fmt.Errorf("没有找到PDF文献")
    }

    // 2. 解析文献
    item := items[0]
    ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
    defer cancel()

    doc, err := parser.ParseDocument(ctx, item.ItemID, item.PDFPath)
    if err != nil {
        return fmt.Errorf("解析失败: %w", err)
    }

    // 3. 验证结果
    if doc.ZoteroItem.ItemID != item.ItemID {
        return fmt.Errorf("解析结果不匹配")
    }

    fmt.Printf("工作流测试成功: %s\n", doc.ZoteroItem.Title)
    return nil
}
```

## 部署和运行

### 1. 构建和运行

```bash
# 构建项目
cd server/
make build

# 运行基础测试
./bin/zoteroflow2

# 运行CLI命令
./bin/zoteroflow2 list
./bin/zoteroflow2 search "机器学习"
```

### 2. 环境配置

```bash
# 设置环境变量
export ZOTERO_DB_PATH=~/Zotero/zotero.sqlite
export ZOTERO_DATA_DIR=~/Zotero/storage
export MINERU_TOKEN=your_token
export AI_API_KEY=your_ai_key

# 运行程序
./bin/zoteroflow2
```

### 3. 生产部署

```bash
# 生产环境构建
make build-prod

# 创建systemd服务
sudo cp scripts/zoteroflow2.service /etc/systemd/system/
sudo systemctl enable zoteroflow2
sudo systemctl start zoteroflow2

# 检查状态
sudo systemctl status zoteroflow2
```

这些示例涵盖了 ZoteroFlow2 的主要使用场景，从基础的 CLI 操作到复杂的 API 集成，从简单的单个功能调用到完整的工作流处理。通过这些示例，您可以快速上手并根据自己的需求进行定制开发。