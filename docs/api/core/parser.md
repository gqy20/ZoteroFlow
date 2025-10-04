# 文档解析接口 API

## 概述

文档解析接口提供了 PDF 文档解析的统一管理功能，集成了 Zotero 数据库访问和 MinerU PDF 解析服务，支持缓存管理、批量处理等高级功能。

## 数据结构

### PDFParser

```go
type PDFParser struct {
    zoteroDB     *ZoteroDB      // Zotero数据库实例
    mineruClient *MinerUClient   // MinerU客户端实例
    cacheDir     string          // 缓存目录
}
```

### ParsedDocument

```go
type ParsedDocument struct {
    ZoteroItem ZoteroItem `json:"zotero_item"` // Zotero文献信息
    ParseHash  string     `json:"parse_hash"`  // 解析哈希值
    Content    string     `json:"content"`     // Markdown格式内容
    Summary    string     `json:"summary"`     // AI生成的摘要
    KeyPoints  []string   `json:"key_points"`  // 关键要点
    ZipPath    string     `json:"zip_path"`    // ZIP文件路径
    ParseTime  time.Time  `json:"parse_time"`  // 解析时间
}
```

## 核心接口

### NewPDFParser

创建新的 PDF 解析器实例。

```go
func NewPDFParser(zoteroDB *ZoteroDB, mineruClient *MinerUClient, cacheDir string) (*PDFParser, error)
```

**参数**:
- `zoteroDB` (*ZoteroDB): Zotero 数据库实例
- `mineruClient` (*MinerUClient): MinerU 客户端实例
- `cacheDir` (string): 缓存目录路径

**返回值**:
- `*PDFParser`: PDF 解析器实例
- `error`: 错误信息

**示例**:
```go
// 创建Zotero数据库连接
zoteroDB, err := core.NewZoteroDB(dbPath, dataDir)
if err != nil {
    log.Fatalf("连接数据库失败: %v", err)
}

// 创建MinerU客户端
mineruClient := core.NewMinerUClient(apiURL, token)

// 创建PDF解析器
parser, err := core.NewPDFParser(zoteroDB, mineruClient, "~/.zoteroflow/cache")
if err != nil {
    log.Fatalf("创建解析器失败: %v", err)
}
```

**特性**:
- 自动创建缓存目录
- 验证组件连接状态
- 初始化缓存系统

### ParseDocument

解析单个文档。

```go
func (p *PDFParser) ParseDocument(ctx context.Context, itemID int, pdfPath string) (*ParsedDocument, error)
```

**参数**:
- `ctx` (context.Context): 上下文对象
- `itemID` (int): Zotero 项目ID
- `pdfPath` (string): PDF 文件路径

**返回值**:
- `*ParsedDocument`: 解析后的文档对象
- `error`: 错误信息

**处理流程**:
1. 获取 Zotero 元数据
2. 检查缓存
3. 调用 MinerU 解析
4. 创建解析结果
5. 保存到缓存

**示例**:
```go
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
defer cancel()

doc, err := parser.ParseDocument(ctx, 12345, "/path/to/document.pdf")
if err != nil {
    log.Printf("解析失败: %v", err)
    return
}

fmt.Printf("解析完成:")
fmt.Printf("  标题: %s\n", doc.ZoteroItem.Title)
fmt.Printf("  作者: %s\n", strings.Join(doc.ZoteroItem.Authors, "; "))
fmt.Printf("  解析时间: %s\n", doc.ParseTime.Format("2006-01-02 15:04:05"))
fmt.Printf("  ZIP路径: %s\n", doc.ZipPath)
```

### BatchParseDocuments

批量解析文档。

```go
func (p *PDFParser) BatchParseDocuments(ctx context.Context, itemIDs []int) ([]*ParsedDocument, error)
```

**参数**:
- `ctx` (context.Context): 上下文对象
- `itemIDs` ([]int): Zotero 项目ID列表

**返回值**:
- `[]*ParsedDocument`: 解析结果列表
- `error`: 错误信息

**示例**:
```go
itemIDs := []int{12345, 12346, 12347}

ctx, cancel := context.WithTimeout(context.Background(), 15*time.Minute)
defer cancel()

docs, err := parser.BatchParseDocuments(ctx, itemIDs)
if err != nil {
    log.Printf("批量解析失败: %v", err)
    return
}

fmt.Printf("批量解析完成，成功解析 %d 篇文档\n", len(docs))
for i, doc := range docs {
    fmt.Printf("[%d] %s\n", i+1, doc.ZoteroItem.Title)
}
```

**处理特性**:
- 并行处理多个文档
- 错误隔离，单个失败不影响整体
- 进度监控和日志记录

## 缓存管理

### generateCacheKey

生成缓存键。

```go
func (p *PDFParser) generateCacheKey(pdfPath string) string
```

**参数**:
- `pdfPath` (string): PDF 文件路径

**返回值**:
- `string`: MD5 哈希值

**示例**:
```go
cacheKey := parser.generateCacheKey("/path/to/document.pdf")
// 返回: "a1b2c3d4e5f6..."
```

### loadFromCache

从缓存加载解析结果。

```go
func (p *PDFParser) loadFromCache(cachePath string) (*ParsedDocument, error)
```

**参数**:
- `cachePath` (string): 缓存文件路径

**返回值**:
- `*ParsedDocument`: 缓存的解析结果
- `error`: 错误信息

**验证逻辑**:
- 检查缓存文件是否存在
- 验证 ZIP 文件是否仍然存在
- 反序列化 JSON 数据

### saveToCache

保存解析结果到缓存。

```go
func (p *PDFParser) saveToCache(doc *ParsedDocument, cachePath string) error
```

**参数**:
- `doc` (*ParsedDocument): 解析结果
- `cachePath` (string): 缓存文件路径

**特性**:
- JSON 序列化
- 原子写入操作
- 错误处理和日志记录

## 辅助功能

### GetZoteroDB

获取 Zotero 数据库连接。

```go
func (p *PDFParser) GetZoteroDB() *ZoteroDB
```

**返回值**:
- `*ZoteroDB`: Zotero 数据库实例

**用途**:
- 直接访问数据库进行查询
- 获取文献元数据
- 执行搜索操作

## 使用示例

### 基础解析

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
    zoteroDB, err := core.NewZoteroDB("~/Zotero/zotero.sqlite", "~/Zotero/storage")
    if err != nil {
        log.Fatalf("连接数据库失败: %v", err)
    }
    defer zoteroDB.Close()
    
    mineruClient := core.NewMinerUClient("https://mineru.net/api/v4", "token")
    
    // 创建解析器
    parser, err := core.NewPDFParser(zoteroDB, mineruClient, "~/.zoteroflow/cache")
    if err != nil {
        log.Fatalf("创建解析器失败: %v", err)
    }
    
    // 获取文献列表
    items, err := zoteroDB.GetItemsWithPDF(1)
    if err != nil {
        log.Fatalf("获取文献失败: %v", err)
    }
    
    if len(items) == 0 {
        fmt.Println("没有找到PDF文献")
        return
    }
    
    // 解析第一个文献
    item := items[0]
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
    defer cancel()
    
    doc, err := parser.ParseDocument(ctx, item.ItemID, item.PDFPath)
    if err != nil {
        log.Printf("解析失败: %v", err)
        return
    }
    
    // 输出结果
    fmt.Printf("解析成功:\n")
    fmt.Printf("  标题: %s\n", doc.ZoteroItem.Title)
    fmt.Printf("  作者: %s\n", strings.Join(doc.ZoteroItem.Authors, "; "))
    fmt.Printf("  解析时间: %s\n", doc.ParseTime.Format("2006-01-02 15:04:05"))
    fmt.Printf("  缓存键: %s\n", doc.ParseHash)
}
```

### 批量处理

```go
func batchProcessingExample() {
    // 初始化组件
    zoteroDB, _ := core.NewZoteroDB(dbPath, dataDir)
    defer zoteroDB.Close()
    
    mineruClient := core.NewMinerUClient(apiURL, token)
    parser, _ := core.NewPDFParser(zoteroDB, mineruClient, cacheDir)
    
    // 获取所有文献
    items, err := zoteroDB.GetItemsWithPDF(10)
    if err != nil {
        log.Printf("获取文献失败: %v", err)
        return
    }
    
    // 提取项目ID
    var itemIDs []int
    for _, item := range items {
        itemIDs = append(itemIDs, item.ItemID)
    }
    
    // 批量解析
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
    defer cancel()
    
    docs, err := parser.BatchParseDocuments(ctx, itemIDs)
    if err != nil {
        log.Printf("批量解析失败: %v", err)
        return
    }
    
    // 统计结果
    successCount := 0
    for _, doc := range docs {
        if doc != nil {
            successCount++
            fmt.Printf("✅ %s\n", doc.ZoteroItem.Title)
        }
    }
    
    fmt.Printf("批量解析完成: 成功 %d/%d\n", successCount, len(itemIDs))
}
```

### 缓存管理

```go
func cacheManagementExample() {
    // 初始化组件
    zoteroDB, _ := core.NewZoteroDB(dbPath, dataDir)
    defer zoteroDB.Close()
    
    mineruClient := core.NewMinerUClient(apiURL, token)
    parser, _ := core.NewPDFParser(zoteroDB, mineruClient, cacheDir)
    
    // 获取文献
    items, _ := zoteroDB.GetItemsWithPDF(1)
    if len(items) == 0 {
        return
    }
    
    item := items[0]
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
    defer cancel()
    
    // 第一次解析（会缓存）
    fmt.Println("第一次解析...")
    doc1, err := parser.ParseDocument(ctx, item.ItemID, item.PDFPath)
    if err != nil {
        log.Printf("解析失败: %v", err)
        return
    }
    
    // 第二次解析（使用缓存）
    fmt.Println("第二次解析（使用缓存）...")
    doc2, err := parser.ParseDocument(ctx, item.ItemID, item.PDFPath)
    if err != nil {
        log.Printf("解析失败: %v", err)
        return
    }
    
    // 验证缓存效果
    if doc1.ParseHash == doc2.ParseHash {
        fmt.Println("✅ 缓存生效，两次解析结果一致")
    } else {
        fmt.Println("❌ 缓存失效，重新解析")
    }
    
    fmt.Printf("解析时间对比:\n")
    fmt.Printf("  第一次: %s\n", doc1.ParseTime.Format("15:04:05.000"))
    fmt.Printf("  第二次: %s\n", doc2.ParseTime.Format("15:04:05.000"))
}
```

## 错误处理

### 常见错误类型

1. **数据库错误**
   ```go
   if strings.Contains(err.Error(), "database") {
       // 数据库连接或查询失败
   }
   ```

2. **文件系统错误**
   ```go
   if os.IsNotExist(err) {
       // PDF文件不存在
   }
   ```

3. **MinerU API错误**
   ```go
   if strings.Contains(err.Error(), "mineru") {
       // MinerU API调用失败
   }
   ```

4. **缓存错误**
   ```go
   if strings.Contains(err.Error(), "cache") {
       // 缓存读写失败
   }
   ```

### 错误处理示例

```go
func safeParseDocument(parser *core.PDFParser, itemID int, pdfPath string) (*core.ParsedDocument, error) {
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
    defer cancel()
    
    doc, err := parser.ParseDocument(ctx, itemID, pdfPath)
    if err != nil {
        switch {
        case strings.Contains(err.Error(), "file not found"):
            return nil, fmt.Errorf("PDF文件不存在: %s", pdfPath)
        case strings.Contains(err.Error(), "database"):
            return nil, fmt.Errorf("数据库访问失败")
        case strings.Contains(err.Error(), "mineru"):
            return nil, fmt.Errorf("MinerU解析失败")
        case strings.Contains(err.Error(), "timeout"):
            return nil, fmt.Errorf("解析超时")
        default:
            return nil, fmt.Errorf("解析失败: %w", err)
        }
    }
    
    return doc, nil
}
```

## 性能优化

### 并发处理

```go
type ConcurrentPDFParser struct {
    *PDFParser
    semaphore chan struct{}
    workers  int
}

func NewConcurrentPDFParser(zoteroDB *core.ZoteroDB, mineruClient *core.MinerUClient, cacheDir string, workers int) *ConcurrentPDFParser {
    return &ConcurrentPDFParser{
        PDFParser: NewPDFParser(zoteroDB, mineruClient, cacheDir),
        semaphore: make(chan struct{}, workers),
        workers:   workers,
    }
}

func (c *ConcurrentPDFParser) BatchParseDocuments(ctx context.Context, itemIDs []int) ([]*core.ParsedDocument, error) {
    var wg sync.WaitGroup
    var mu sync.Mutex
    var results []*core.ParsedDocument
    var errors []error
    
    for _, itemID := range itemIDs {
        wg.Add(1)
        go func(id int) {
            defer wg.Done()
            
            c.semaphore <- struct{}{}
            defer func() { <-c.semaphore }()
            
            // 获取PDF路径
            items, err := c.zoteroDB.GetItemsWithPDF(1)
            if err != nil {
                mu.Lock()
                errors = append(errors, fmt.Errorf("itemID %d: %w", id, err))
                mu.Unlock()
                return
            }
            
            if len(items) == 0 {
                mu.Lock()
                errors = append(errors, fmt.Errorf("itemID %d: 未找到文献", id))
                mu.Unlock()
                return
            }
            
            item := items[0]
            doc, err := c.ParseDocument(ctx, item.ItemID, item.PDFPath)
            if err != nil {
                mu.Lock()
                errors = append(errors, fmt.Errorf("itemID %d: %w", id, err))
                mu.Unlock()
                return
            }
            
            mu.Lock()
            results = append(results, doc)
            mu.Unlock()
        }(itemID)
    }
    
    wg.Wait()
    
    if len(errors) > 0 {
        log.Printf("并发解析完成，成功: %d, 失败: %d", len(results), len(errors))
        for _, err := range errors {
            log.Printf("错误: %v", err)
        }
    }
    
    return results, nil
}
```

### 智能缓存

```go
type SmartCache struct {
    cache    map[string]*CacheEntry
    mutex    sync.RWMutex
    maxSize  int
    ttl      time.Duration
}

type CacheEntry struct {
    Document *core.ParsedDocument
    Created  time.Time
    Accessed time.Time
    Size     int64
}

func (c *SmartCache) Get(key string) (*core.ParsedDocument, bool) {
    c.mutex.RLock()
    defer c.mutex.RUnlock()
    
    entry, exists := c.cache[key]
    if !exists {
        return nil, false
    }
    
    // 检查TTL
    if time.Since(entry.Created) > c.ttl {
        return nil, false
    }
    
    // 更新访问时间
    entry.Accessed = time.Now()
    return entry.Document, true
}

func (c *SmartCache) Put(key string, doc *core.ParsedDocument) {
    c.mutex.Lock()
    defer c.mutex.Unlock()
    
    // 检查缓存大小
    if len(c.cache) >= c.maxSize {
        c.evictLRU()
    }
    
    entry := &CacheEntry{
        Document: doc,
        Created:  time.Now(),
        Accessed: time.Now(),
        Size:     int64(len(doc.Content)),
    }
    
    c.cache[key] = entry
}

func (c *SmartCache) evictLRU() {
    var oldestKey string
    var oldestTime time.Time
    
    for key, entry := range c.cache {
        if oldestTime.IsZero() || entry.Accessed.Before(oldestTime) {
            oldestKey = key
            oldestTime = entry.Accessed
        }
    }
    
    if oldestKey != "" {
        delete(c.cache, oldestKey)
    }
}
```

## 配置建议

### 环境变量

```bash
# 缓存配置
PARSER_CACHE_DIR=~/.zoteroflow/cache
PARSER_CACHE_SIZE=1000        # 最大缓存条目数
PARSER_CACHE_TTL=24h           # 缓存TTL时间

# 并发配置
PARSER_MAX_WORKERS=5            # 最大并发工作数
PARSER_BATCH_SIZE=10            # 批量处理大小

# 超时配置
PARSER_TIMEOUT=300s             # 解析超时时间
PARSER_BATCH_TIMEOUT=1800s       # 批量处理超时时间
```

### 缓存策略

```go
type CacheConfig struct {
    Dir      string        `json:"dir"`       // 缓存目录
    MaxSize  int           `json:"max_size"`  // 最大缓存条目数
    TTL      time.Duration `json:"ttl"`       // 缓存TTL
    Compress bool          `json:"compress"`  // 是否压缩缓存
}

func DefaultCacheConfig() *CacheConfig {
    return &CacheConfig{
        Dir:      "~/.zoteroflow/cache",
        MaxSize:  1000,
        TTL:      24 * time.Hour,
        Compress: true,
    }
}
```

## 监控和日志

### 解析进度监控

```go
func monitorParsingProgress(parser *core.PDFParser, total int) {
    ticker := time.NewTicker(10 * time.Second)
    defer ticker.Stop()
    
    processed := 0
    
    for range ticker.C {
        // 检查缓存目录中的文件数量
        files, _ := filepath.Glob(parser.cacheDir + "/*.json")
        current := len(files)
        
        if current > processed {
            processed = current
            progress := float64(processed) / float64(total) * 100
            log.Printf("解析进度: %d/%d (%.1f%%)", processed, total, progress)
        }
        
        if processed >= total {
            break
        }
    }
    
    log.Printf("解析完成: %d/%d", processed, total)
}
```

### 性能统计

```go
type ParsingStats struct {
    TotalDocuments    int           `json:"total_documents"`
    SuccessfulParse   int           `json:"successful_parse"`
    FailedParse       int           `json:"failed_parse"`
    CacheHits         int           `json:"cache_hits"`
    CacheMisses       int           `json:"cache_misses"`
    AverageParseTime  time.Duration `json:"average_parse_time"`
    TotalParseTime    time.Duration `json:"total_parse_time"`
}

func (s *ParsingStats) RecordParse(success bool, duration time.Duration, cacheHit bool) {
    s.TotalDocuments++
    
    if success {
        s.SuccessfulParse++
    } else {
        s.FailedParse++
    }
    
    if cacheHit {
        s.CacheHits++
    } else {
        s.CacheMisses++
    }
    
    s.TotalParseTime += duration
    s.AverageParseTime = s.TotalParseTime / time.Duration(s.TotalDocuments)
}

func (s *ParsingStats) String() string {
    cacheHitRate := float64(s.CacheHits) / float64(s.CacheHits+s.CacheMisses) * 100
    successRate := float64(s.SuccessfulParse) / float64(s.TotalDocuments) * 100
    
    return fmt.Sprintf(
        "解析统计: 总数=%d, 成功=%d (%.1f%%), 失败=%d, 缓存命中率=%.1f%%, 平均耗时=%v",
        s.TotalDocuments, s.SuccessfulParse, successRate, s.FailedParse,
        cacheHitRate, s.AverageParseTime,
    )
}
```

## 最佳实践

### 1. 资源管理
- 及时释放数据库连接
- 合理设置并发数量
- 定期清理过期缓存

### 2. 错误处理
- 区分不同类型的错误
- 实现重试机制
- 提供详细的错误信息

### 3. 性能优化
- 使用缓存避免重复解析
- 并行处理提高效率
- 根据文件大小调整超时时间

### 4. 监控和日志
- 记录解析过程和结果
- 监控性能指标
- 定期分析统计数据

## 版本兼容性

- **Zotero 6.x**: 完全支持
- **Zotero 7.x**: 完全支持
- **MinerU API v4**: 完全支持
- **Go 1.21+**: 支持

## 注意事项

1. **缓存一致性**: 确保缓存与实际文件状态一致
2. **并发安全**: 在多线程环境下注意数据竞争
3. **存储空间**: 监控缓存目录大小，避免磁盘空间不足
4. **网络依赖**: MinerU API 调用需要稳定的网络连接
5. **文件权限**: 确保对PDF文件和缓存目录有适当的读写权限