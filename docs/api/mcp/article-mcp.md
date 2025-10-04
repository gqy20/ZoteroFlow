# Article MCP 集成 API

## 概述

Article MCP 是一个专业的学术文献搜索和分析服务，通过 MCP 协议为 ZoteroFlow2 提供了强大的文献检索能力。它集成了 Europe PMC 和 arXiv 两大权威学术数据库，支持文献搜索、质量评估、引用分析等功能。

## 支持的工具

### 1. 搜索工具

#### search_europe_pmc

搜索 Europe PMC 数据库中的学术文献。

```json
{
    "name": "search_europe_pmc",
    "description": "搜索 Europe PMC 数据库中的文献",
    "inputSchema": {
        "type": "object",
        "properties": {
            "keyword": {
                "type": "string",
                "description": "搜索关键词，支持布尔运算符"
            },
            "email": {
                "type": "string",
                "description": "联系邮箱，用于获得更高的API访问限制"
            },
            "start_date": {
                "type": "string",
                "description": "开始日期，格式：YYYY-MM-DD"
            },
            "end_date": {
                "type": "string",
                "description": "结束日期，格式：YYYY-MM-DD"
            },
            "max_results": {
                "type": "integer",
                "description": "最大返回结果数量",
                "default": 10,
                "maximum": 100
            }
        },
        "required": ["keyword"]
    }
}
```

**参数说明**:
- `keyword` (必需): 搜索关键词，支持布尔运算符（AND、OR、NOT）
- `email` (可选): 联系邮箱，提供后可获得更高的API访问限制
- `start_date` (可选): 限制搜索的起始日期
- `end_date` (可选): 限制搜索的结束日期
- `max_results` (可选): 最大返回结果数量，默认10，最大100

**响应示例**:
```json
{
    "content": [
        {
            "type": "text",
            "text": "找到 10 篇相关文献，包括机器学习、深度学习、神经网络等主题。"
        }
    ],
    "isError": false
}
```

#### search_arxiv_papers

搜索 arXiv 预印本数据库中的学术论文。

```json
{
    "name": "search_arxiv_papers",
    "description": "搜索 arXiv 预印本数据库中的学术论文",
    "inputSchema": {
        "type": "object",
        "properties": {
            "keyword": {
                "type": "string",
                "description": "搜索关键词，支持复杂查询语法"
            },
            "email": {
                "type": "string",
                "description": "联系邮箱，用于获得更好的API服务"
            },
            "start_date": {
                "type": "string",
                "description": "开始日期，格式：YYYY-MM-DD"
            },
            "end_date": {
                "type": "string",
                "description": "结束日期，格式：YYYY-MM-DD"
            },
            "max_results": {
                "type": "integer",
                "description": "最大返回结果数量",
                "default": 10,
                "maximum": 1000
            }
        },
        "required": ["keyword"]
    }
}
```

**参数说明**:
- `keyword` (必需): 搜索关键词，支持复杂查询语法
- `email` (可选): 联系邮箱
- `start_date` (可选): 开始日期
- `end_date` (可选): 结束日期
- `max_results` (可选): 最大结果数量，默认10，最大1000

### 2. 文献详情工具

#### get_article_details

获取特定文献的详细信息。

```json
{
    "name": "get_article_details",
    "description": "获取特定文献的详细信息",
    "inputSchema": {
        "type": "object",
        "properties": {
            "identifier": {
                "type": "string",
                "description": "文献标识符（PMID、DOI或PMCID）"
            },
            "id_type": {
                "type": "string",
                "description": "标识符类型：pmid、doi 或 pmcid",
                "default": "pmid",
                "enum": ["pmid", "doi", "pmcid"]
            },
            "include_fulltext": {
                "type": "boolean",
                "description": "是否包含全文内容",
                "default": false
            }
        },
        "required": ["identifier"]
    }
}
```

**参数说明**:
- `identifier` (必需): 文献标识符
- `id_type` (可选): 标识符类型，默认为 "pmid"
- `include_fulltext` (可选): 是否包含全文内容，默认为 false

**响应示例**:
```json
{
    "content": [
        {
            "type": "text",
            "text": "文献详细信息：标题、作者、摘要、关键词、发表信息等。"
        }
    ],
    "isError": false
}
```

### 3. 关联分析工具

#### get_references_by_doi

通过 DOI 获取参考文献列表。

```json
{
    "name": "get_references_by_doi",
    "description": "通过 DOI 获取参考文献列表",
    "inputSchema": {
        "type": "object",
        "properties": {
            "doi": {
                "type": "string",
                "description": "数字对象标识符"
            }
        },
        "required": ["doi"]
    }
}
```

**参数说明**:
- `doi` (必需): 数字对象标识符

#### get_similar_articles

根据文献标识符获取相似文章。

```json
{
    "name": "get_similar_articles",
    "description": "根据文献标识符获取相似文章",
    "inputSchema": {
        "type": "object",
        "properties": {
            "identifier": {
                "type": "string",
                "description": "文献标识符"
            },
            "id_type": {
                "type": "string",
                "description": "标识符类型：doi、pmid 或 pmcid",
                "default": "doi",
                "enum": ["doi", "pmid", "pmcid"]
            },
            "email": {
                "type": "string",
                "description": "联系邮箱，用于获得更好的API服务"
            },
            "max_results": {
                "type": "integer",
                "description": "返回的最大相似文章数量",
                "default": 20
            }
        },
        "required": ["identifier"]
    }
}
```

#### get_citing_articles

获取引用该文献的文献信息。

```json
{
    "name": "get_citing_articles",
    "description": "获取引用该文献的文献信息",
    "inputSchema": {
        "type": "object",
        "properties": {
            "identifier": {
                "type": "string",
                "description": "文献标识符"
            },
            "id_type": {
                "type": "string",
                "description": "标识符类型：doi、pmid 或 pmcid",
                "default": "pmid",
                "enum": ["doi", "pmid", "pmcid"]
            },
            "max_results": {
                "type": "integer",
                "description": "返回的最大引用文献数量",
                "default": 20
            },
            "email": {
                "type": "string",
                "description": "联系邮箱，用于获得更好的API服务"
            }
        },
        "required": ["identifier"]
    }
}
```

### 4. 质量评估工具

#### evaluate_articles_quality

批量评估文献的期刊质量。

```json
{
    "name": "evaluate_articles_quality",
    "description": "批量评估文献的期刊质量",
    "inputSchema": {
        "type": "object",
        "properties": {
            "articles": {
                "type": "array",
                "description": "文献列表，来自搜索结果",
                "items": {
                    "type": "object"
                }
            },
            "secret_key": {
                "type": "string",
                "description": "EasyScholar API密钥"
            }
        },
        "required": ["articles"]
    }
}
```

### 5. 关联分析工具

#### get_literature_relations

获取文献的所有关联信息。

```json
{
    "name": "get_literature_relations",
    "description": "获取文献的所有关联信息（参考文献、相似文献、引用文献）",
    "inputSchema": {
        "type": "object",
        "properties": {
            "identifier": {
                "type": "string",
                "description": "文献标识符"
            },
            "id_type": {
                "type": "string",
                "description": "标识符类型：doi、pmid 或 pmcid",
                "default": "doi",
                "enum": ["doi", "pmid", "pmcid"]
            },
            "max_results": {
                "type": "integer",
                "description": "每种关联文献的最大返回数量",
                "default": 20
            }
        },
        "required": ["identifier"]
    }
}
```

## 使用示例

### 基础搜索

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
    _, err := client.Initialize("zoteroflow-test", "1.0.0")
    if err != nil {
        log.Fatalf("初始化失败: %v", err)
    }
    
    // 搜索 Europe PMC
    fmt.Println("=== 搜索 Europe PMC ===")
    result, err := client.CallTool("search_europe_pmc", map[string]interface{}{
        "keyword":     "machine learning",
        "max_results":  5,
    })
    if err != nil {
        log.Printf("搜索失败: %v", err)
        return
    }
    
    displayResult(result)
    
    // 搜索 arXiv
    fmt.Println("\n=== 搜索 arXiv ===")
    result, err = client.CallTool("search_arxiv_papers", map[string]interface{}{
        "keyword":     "deep learning",
        "max_results":  3,
    })
    if err != nil {
        log.Printf("搜索失败: %v", err)
        return
    }
    
    displayResult(result)
}

func displayResult(result *mcp.CallToolResult) {
    for _, content := range result.Content {
        if content.Type == "text" {
            fmt.Printf("结果: %s\n", content.Text)
        }
    }
}
```

### 文献详情获取

```go
func getArticleDetails(client *mcp.MCPClient, pmid string) {
    fmt.Printf("=== 获取文献详情 (PMID: %s) ===\n", pmid)
    
    result, err := client.CallTool("get_article_details", map[string]interface{}{
        "identifier": pmid,
        "id_type":    "pmid",
    })
    if err != nil {
        log.Printf("获取详情失败: %v", err)
        return
    }
    
    displayResult(result)
}
```

### 智能文献分析

```go
func intelligentLiteratureAnalysis(client *mcp.MCPClient, query string) {
    fmt.Printf("=== 智能文献分析: %s ===\n", query)
    
    // 1. 搜索 Europe PMC
    fmt.Println("步骤1: 搜索 Europe PMC 数据库")
    europeResult, err := client.CallTool("search_europe_pmc", map[string]interface{}{
        "keyword":     query,
        "max_results":  10,
    })
    if err != nil {
        log.Printf("Europe PMC 搜索失败: %v", err)
    } else {
        fmt.Printf("✅ Europe PMC 搜索完成\n")
        displayResult(europeResult)
    }
    
    // 2. 搜索 arXiv
    fmt.Println("\n步骤2: 搜索 arXiv 数据库")
    arxivResult, err := client.CallTool("search_arxiv_papers", map[string]interface{}{
        "keyword":     query,
        "max_results":  5,
    })
    if err != nil {
        log.Printf("arXiv 搜索失败: %v", err)
    } else {
        fmt.Printf("✅ arXiv 搜索完成\n")
        displayResult(arxivResult)
    }
    
    // 3. 获取相似文献（假设有第一篇文献的DOI）
    if europeResult != nil && len(europeResult.Content) > 0 {
        // 这里需要从搜索结果中提取DOI
        // 简化示例，使用固定DOI
        doi := "10.1038/nature12345"
        
        fmt.Println("\n步骤3: 获取相似文献")
        similarResult, err := client.CallTool("get_similar_articles", map[string]interface{}{
            "identifier": doi,
            "id_type":    "doi",
            "max_results": 5,
        })
        if err != nil {
            log.Printf("获取相似文献失败: %v", err)
        } else {
            fmt.Printf("✅ 相似文献获取完成\n")
            displayResult(similarResult)
        }
    }
}
```

### 文献质量评估

```go
func evaluateArticleQuality(client *mcp.MCPClient, articles []interface{}) {
    fmt.Println("=== 文献质量评估 ===")
    
    result, err := client.CallTool("evaluate_articles_quality", map[string]interface{}{
        "articles": articles,
    })
    if err != nil {
        log.Printf("质量评估失败: %v", err)
        return
    }
    
    displayResult(result)
}
```

### 关联分析

```go
func analyzeLiteratureRelations(client *mcp.MCPClient, identifier string) {
    fmt.Printf("=== 文献关联分析: %s ===\n", identifier)
    
    result, err := client.CallTool("get_literature_relations", map[string]interface{}{
        "identifier": identifier,
        "id_type":    "doi",
        "max_results": 10,
    })
    if err != nil {
        log.Printf("关联分析失败: %v", err)
        return
    }
    
    displayResult(result)
}
```

## 高级用法

### 复合搜索

```go
func complexSearch(client *mcp.MCPClient) {
    // 使用布尔运算符搜索
    complexQuery := "machine learning AND (neural network OR deep learning) NOT reinforcement learning"
    
    result, err := client.CallTool("search_europe_pmc", map[string]interface{}{
        "keyword":     complexQuery,
        "max_results":  20,
        "start_date":  "2020-01-01",
        "end_date":    "2024-12-31",
    })
    if err != nil {
        log.Printf("复合搜索失败: %v", err)
        return
    }
    
    fmt.Printf("复合搜索结果: %s\n", result.Content[0].Text)
}
```

### 批量处理

```go
func batchProcess(client *mcp.MCPClient, queries []string) {
    fmt.Printf("=== 批量处理 %d 个查询 ===\n", len(queries))
    
    for i, query := range queries {
        fmt.Printf("\n[%d] 处理查询: %s\n", i+1, query)
        
        result, err := client.CallTool("search_europe_pmc", map[string]interface{}{
            "keyword":     query,
            "max_results":  5,
        })
        if err != nil {
            log.Printf("查询失败: %s, 错误: %v", query, err)
            continue
        }
        
        fmt.Printf("结果: %s\n", result.Content[0].Text)
    }
    
    fmt.Printf("\n批量处理完成\n")
}
```

### 缓存优化

```go
type CachedArticleMCP struct {
    client    *mcp.MCPClient
    cache     map[string]*CacheEntry
    cacheTTL  time.Duration
    mutex     sync.RWMutex
}

type CacheEntry struct {
    Result    *mcp.CallToolResult
    Timestamp time.Time
    TTL       time.Duration
}

func (c *CachedArticleMCP) CallTool(toolName string, args map[string]interface{}) (*mcp.CallToolResult, error) {
    // 生成缓存键
    cacheKey := c.generateCacheKey(toolName, args)
    
    // 检查缓存
    c.mutex.RLock()
    if entry, exists := c.cache[cacheKey]; exists {
        if time.Since(entry.Timestamp) < entry.TTL {
            c.mutex.RUnlock()
            return entry.Result, nil
        }
    }
    c.mutex.RUnlock()
    
    // 调用原始方法
    result, err := c.client.CallTool(toolName, args)
    if err != nil {
        return nil, err
    }
    
    // 缓存结果
    c.mutex.Lock()
    c.cache[cacheKey] = &CacheEntry{
        Result:    result,
        Timestamp: time.Now(),
        TTL:       c.cacheTTL,
    }
    c.mutex.Unlock()
    
    return result, nil
}

func (c *CachedArticleMCP) generateCacheKey(toolName string, args map[string]interface{}) string {
    h := md5.New()
    h.Write([]byte(toolName))
    
    // 对参数进行排序以确保一致性
    keys := make([]string, 0, len(args))
    for k := range args {
        keys = append(keys, k)
    }
    sort.Strings(keys)
    
    for _, k := range keys {
        h.Write([]byte(fmt.Sprintf("%s=%v", k, args[k])))
    }
    
    return fmt.Sprintf("%x", h.Sum(nil))
}
```

## 错误处理

### 常见错误类型

1. **连接错误**
   ```go
   if strings.Contains(err.Error(), "connection") {
       // MCP 服务器连接失败
   }
   ```

2. **工具不存在错误**
   ```go
   if strings.Contains(err.Error(), "tool not found") {
       // 请求的工具不存在
   }
   ```

3. **参数错误**
   ```go
   if strings.Contains(err.Error(), "invalid arguments") {
       // 参数格式或内容错误
   }
   ```

4. **API限制错误**
   ```go
   if strings.Contains(err.Error(), "rate limit") {
       // API 调用频率超限
   }
   ```

### 错误处理示例

```go
func safeArticleMCPCall(client *mcp.MCPClient, toolName string, args map[string]interface{}) (*mcp.CallToolResult, error) {
    result, err := client.CallTool(toolName, args)
    if err != nil {
        switch {
        case strings.Contains(err.Error(), "connection"):
            return nil, fmt.Errorf("Article MCP 连接失败，请检查服务器状态")
        case strings.Contains(err.Error(), "tool not found"):
            return nil, fmt.Errorf("工具不存在: %s", toolName)
        case strings.Contains(err.Error(), "invalid arguments"):
            return nil, fmt.Errorf("参数无效: %v", args)
        case strings.Contains(err.Error(), "rate limit"):
            return nil, fmt.Errorf("API 调用频率超限，请稍后重试")
        default:
            return nil, fmt.Errorf("Article MCP 调用失败: %w", err)
        }
    }
    
    // 检查响应是否为错误
    if len(result.Content) > 0 && result.Content[0].Type == "text" {
        if strings.Contains(result.Content[0].Text, "error") {
            return nil, fmt.Errorf("API 返回错误: %s", result.Content[0].Text)
        }
    }
    
    return result, nil
}
```

## 性能优化

### 并发请求

```go
func concurrentSearch(client *mcp.MCPClient, queries []string) map[string]*mcp.CallToolResult {
    var wg sync.WaitGroup
    var mu sync.Mutex
    results := make(map[string]*mcp.CallToolResult)
    
    semaphore := make(chan struct{}, 3) // 限制并发数
    
    for _, query := range queries {
        wg.Add(1)
        go func(q string) {
            defer wg.Done()
            
            semaphore <- struct{}{}
            defer func() { <-semaphore }()
            
            result, err := client.CallTool("search_europe_pmc", map[string]interface{}{
                "keyword":     q,
                "max_results": 5,
            })
            
            mu.Lock()
            if err != nil {
                results[q] = &mcp.CallToolResult{
                    Content: []mcp.Content{
                        {Type: "text", Text: fmt.Sprintf("搜索失败: %v", err)}},
                    IsError: true,
                }
            } else {
                results[q] = result
            }
            mu.Unlock()
        }(query)
    }
    
    wg.Wait()
    return results
}
```

### 智能重试

```go
func retryCall(client *mcp.MCPClient, toolName string, args map[string]interface{}, maxRetries int) (*mcp.CallToolResult, error) {
    var lastErr error
    
    for i := 0; i < maxRetries; i++ {
        result, err := client.CallTool(toolName, args)
        if err == nil {
            return result, nil
        }
        
        lastErr = err
        
        // 检查是否为可重试的错误
        if !isRetryableError(err) {
            break
        }
        
        // 指数退避
        backoff := time.Duration(1<<uint(i)) * time.Second
        time.Sleep(backoff)
        
        log.Printf("重试 %d/%d: %v\n", i+1, maxRetries, err)
    }
    
    return nil, lastErr
}

func isRetryableError(err error) bool {
    errStr := err.Error()
    
    // 可重试的错误类型
    retryableErrors := []string{
        "connection",
        "timeout",
        "rate limit",
        "temporary",
        "service unavailable",
    }
    
    for _, retryable := range retryableErrors {
        if strings.Contains(errStr, retryable) {
            return true
        }
    }
    
    return false
}
```

## 配置建议

### 环境变量

```bash
# Article MCP 配置
ARTICLE_MCP_COMMAND="uv tool run article-mcp server"
ARTICLE_MCP_TIMEOUT=30s

# API 配置
ARTICLE_MCP_EMAIL=your-email@example.com
ARTICLE_MCP_MAX_RESULTS=20

# 缓存配置
ARTICLE_MCP_CACHE_TTL=3600s
ARTICLE_MCP_CACHE_SIZE=100

# 性能配置
ARTICLE_MCP_MAX_CONCURRENT=3
ARTICLE_MCP_RETRY_COUNT=3
ARTICLE_MCP_RETRY_DELAY=1s
```

### 配置文件

```toml
# ~/.zoteroflow/article-mcp.toml

[article_mcp]
enabled = true
command = "uv"
args = ["tool", "run", "article-mcp", "server"]
timeout = "30s"

[article_mcp.cache]
enabled = true
ttl = "1h"
max_size = 100

[article_mcp.performance]
max_concurrent = 3
retry_count = 3
retry_delay = "1s"
```

## 监控和日志

### 使用统计

```go
type ArticleMCPStats struct {
    TotalCalls       int64         `json:"total_calls"`
    SuccessfulCalls  int64         `json:"successful_calls"`
    FailedCalls      int64         `json:"failed_calls"`
    CacheHits        int64         `json:"cache_hits"`
    CacheMisses      int64         `json:"cache_misses"`
    AverageLatency   time.Duration `json:"average_latency"`
    ToolUsage       map[string]int64 `json:"tool_usage"`
}

func (s *ArticleMCPStats) RecordCall(toolName string, success bool, latency time.Duration, cacheHit bool) {
    s.TotalCalls++
    s.ToolUsage[toolName]++
    
    if success {
        s.SuccessfulCalls++
    } else {
        s.FailedCalls++
    }
    
    if cacheHit {
        s.CacheHits++
    } else {
        s.CacheMisses++
    }
    
    // 更新平均延迟（简化计算）
    s.AverageLatency = (s.AverageLatency + latency) / 2
}

func (s *ArticleMCPStats) GetCacheHitRate() float64 {
    total := s.CacheHits + s.CacheMisses
    if total == 0 {
        return 0
    }
    return float64(s.CacheHits) / float64(total) * 100
}

func (s *ArticleMCPStats) GetSuccessRate() float64 {
    if s.TotalCalls == 0 {
        return 0
    }
    return float64(s.SuccessfulCalls) / float64(s.TotalCalls) * 100
}
```

## 最佳实践

### 1. 查询优化
- 使用具体的搜索关键词
- 合理设置结果数量限制
- 利用日期范围过滤

### 2. 缓存策略
- 缓存常用查询结果
- 设置合理的TTL时间
- 定期清理过期缓存

### 3. 错误处理
- 实现智能重试机制
- 区分不同类型的错误
- 提供用户友好的错误信息

### 4. 性能考虑
- 使用并发请求提高效率
- 限制并发数量避免过载
- 监控API调用频率

## 版本兼容性

- **Article MCP**: 完全支持
- **Europe PMC API**: 完全支持
- **arXiv API**: 完全支持
- **MCP v2024-11-05**: 完全支持

## 注意事项

1. **API限制**: 注意 Europe PMC 和 arXiv 的API调用限制
2. **邮箱要求**: 提供邮箱可获得更好的API服务
3. **查询语法**: 不同数据库支持的查询语法可能不同
4. **结果格式**: 注意不同工具返回的结果格式差异
5. **网络依赖**: 确保网络连接稳定，避免长时间阻塞

## 故障排除

### 常见问题及解决方案

1. **MCP 服务器启动失败**
   ```bash
   # 检查 uv 工具是否安装
   uv --version
   
   # 检查 article-mcp 是否可用
   uv tool run article-mcp --help
   ```

2. **工具调用失败**
   ```bash
   # 检查网络连接
   curl -I https://www.ebi.ac.uk/europepmc/api/
   
   # 检查 arXiv API
   curl -I http://export.arxiv.org/api/query
   ```

3. **搜索无结果**
   - 检查关键词拼写
   - 尝试更通用的关键词
   - 移除特殊字符和布尔运算符

4. **性能问题**
   - 减少 max_results 数量
   - 使用缓存避免重复查询
   - 限制并发请求数量