# Zotero 数据库接口 API

## 概述

Zotero 数据库接口提供了对 Zotero 文献数据库的只读访问功能，支持文献查询、PDF 文件定位、元数据提取等操作。

## 数据结构

### ZoteroItem

```go
type ZoteroItem struct {
    ItemID   int      `json:"item_id"`    // Zotero 项目ID
    Title    string   `json:"title"`      // 文献标题
    Authors  []string `json:"authors"`     // 作者列表
    Year     int      `json:"year"`       // 发表年份
    ItemType string   `json:"item_type"`  // 文献类型
    Tags     []string `json:"tags"`       // 标签列表
    PDFPath  string   `json:"pdf_path"`   // PDF文件路径
    PDFName  string   `json:"pdf_name"`   // PDF文件名
}
```

### SearchResult

```go
type SearchResult struct {
    ZoteroItem
    Score   float64 `json:"score"`    // 搜索评分
    DOI     string  `json:"doi"`      // DOI标识符
    Journal string  `json:"journal"`  // 期刊名称
}
```

### ZoteroDB

```go
type ZoteroDB struct {
    db      *sql.DB  // 数据库连接
    dataDir string    // 数据存储目录
    dbPath  string    // 数据库文件路径
}
```

## 核心接口

### NewZoteroDB

创建新的 Zotero 数据库连接实例。

```go
func NewZoteroDB(dbPath, dataDir string) (*ZoteroDB, error)
```

**参数**:
- `dbPath` (string): Zotero SQLite 数据库文件路径
- `dataDir` (string): Zotero 存储目录路径

**返回值**:
- `*ZoteroDB`: Zotero 数据库实例
- `error`: 错误信息

**示例**:
```go
zoteroDB, err := core.NewZoteroDB(
    "~/Zotero/zotero.sqlite",
    "~/Zotero/storage",
)
if err != nil {
    log.Fatalf("连接数据库失败: %v", err)
}
defer zoteroDB.Close()
```

**特性**:
- 自动设置只读模式，避免数据损坏
- 配置连接池参数（最大连接数：1）
- 执行连接测试验证数据库可访问性

### GetItemsWithPDF

获取包含 PDF 附件的文献列表。

```go
func (z *ZoteroDB) GetItemsWithPDF(limit int) ([]ZoteroItem, error)
```

**参数**:
- `limit` (int): 返回结果数量限制

**返回值**:
- `[]ZoteroItem`: 文献项目列表
- `error`: 错误信息

**示例**:
```go
items, err := zoteroDB.GetItemsWithPDF(10)
if err != nil {
    log.Printf("查询失败: %v", err)
    return
}

for _, item := range items {
    fmt.Printf("标题: %s\n", item.Title)
    fmt.Printf("作者: %s\n", strings.Join(item.Authors, "; "))
    fmt.Printf("PDF路径: %s\n", item.PDFPath)
}
```

**查询逻辑**:
```sql
SELECT DISTINCT
    i.itemID,
    COALESCE(idv.value, '') as title,
    it.typeName as item_type,
    ia.path as attachment_path,
    ia.contentType as content_type
FROM items i
LEFT JOIN itemData id ON i.itemID = id.itemID
LEFT JOIN fieldsCombined fc ON id.fieldID = fc.fieldID AND fc.fieldName = 'title'
LEFT JOIN itemDataValues idv ON id.valueID = idv.valueID
LEFT JOIN itemAttachments ia ON i.itemID = ia.parentItemID
LEFT JOIN itemTypes it ON it.itemTypeID = i.itemTypeID
WHERE ia.contentType = 'application/pdf'
AND i.itemTypeID NOT IN (
    SELECT itemTypeID FROM itemTypes
    WHERE typeName IN ('attachment', 'note', 'annotation')
)
ORDER BY i.dateAdded DESC
LIMIT ?
```

### SearchByTitle

按标题关键词搜索文献。

```go
func (z *ZoteroDB) SearchByTitle(query string, limit int) ([]SearchResult, error)
```

**参数**:
- `query` (string): 搜索关键词
- `limit` (int): 返回结果数量限制

**返回值**:
- `[]SearchResult`: 搜索结果列表
- `error`: 错误信息

**示例**:
```go
results, err := zoteroDB.SearchByTitle("机器学习", 5)
if err != nil {
    log.Printf("搜索失败: %v", err)
    return
}

for _, result := range results {
    fmt.Printf("标题: %s (评分: %.1f)\n", result.Title, result.Score)
    fmt.Printf("作者: %s\n", strings.Join(result.Authors, "; "))
    if result.DOI != "" {
        fmt.Printf("DOI: %s\n", result.DOI)
    }
}
```

**评分算法**:
- 完全匹配: 100.0 分
- 前缀/后缀匹配: 80.0 分
- 部分匹配: 50.0 + 出现次数 × 10.0 分
- 包含匹配: 30.0 分

### GetStats

获取数据库统计信息。

```go
func (z *ZoteroDB) GetStats() (map[string]interface{}, error)
```

**返回值**:
- `map[string]interface{}`: 统计信息
- `error`: 错误信息

**示例**:
```go
stats, err := zoteroDB.GetStats()
if err != nil {
    log.Printf("获取统计信息失败: %v", err)
    return
}

fmt.Printf("总文献数: %d\n", stats["total_items"])
fmt.Printf("有PDF附件的文献数: %d\n", stats["pdf_items"])
fmt.Printf("数据库大小: %d MB\n", stats["db_size_mb"])
```

**统计指标**:
- `total_items`: 总文献数量
- `pdf_items`: 包含PDF附件的文献数量
- `db_size_mb`: 数据库文件大小（MB）

## 辅助功能

### 文件路径处理

#### buildPDFPath

构建 PDF 文件的完整路径。

```go
func (z *ZoteroDB) buildPDFPath(pdfPath string) string
```

**支持的路径格式**:
- `attachments:分类:年份_标题.pdf`
- `storage:XXXXXX.pdf`
- 直接文件名

**示例**:
```go
// 处理 attachments: 格式
path := z.buildPDFPath("attachments:机器学习:2024_基础教程.pdf")

// 处理 storage: 格式
path := z.buildPDFPath("storage:ABC123DEF456")

// 处理直接文件名
path := z.buildPDFPath("机器学习基础.pdf")
```

#### findPDFInStorage

在存储目录中查找 PDF 文件。

```go
func (z *ZoteroDB) findPDFInStorage(filename string) string
```

**搜索策略**:
1. 完全匹配文件名
2. 去掉扩展名后匹配
3. 模糊匹配（包含关系）

**示例**:
```go
pdfPath := z.findPDFInStorage("机器学习基础.pdf")
if pdfPath != "" {
    fmt.Printf("找到PDF文件: %s\n", pdfPath)
} else {
    fmt.Println("未找到PDF文件")
}
```

### 元数据提取

#### getItemTags

获取文献的标签信息。

```go
func (z *ZoteroDB) getItemTags(itemID int) ([]string, error)
```

**示例**:
```go
tags, err := zoteroDB.getItemTags(12345)
if err != nil {
    log.Printf("获取标签失败: %v", err)
    return
}

fmt.Printf("标签: %s\n", strings.Join(tags, ", "))
```

#### parseAuthors

解析作者字符串。

```go
func parseAuthors(authorsStr string) []string
```

**示例**:
```go
authors := parseAuthors("张三; 李四; 王五")
// 返回: ["张三", "李四", "王五"]
```

## 错误处理

### 常见错误类型

1. **数据库连接错误**
   ```go
   if err != nil && strings.Contains(err.Error(), "database is locked") {
       // 数据库被锁定，可能是 Zotero 正在使用
   }
   ```

2. **文件不存在错误**
   ```go
   if err != nil && os.IsNotExist(err) {
       // PDF 文件不存在
   }
   ```

3. **权限错误**
   ```go
   if err != nil && strings.Contains(err.Error(), "permission denied") {
       // 权限不足
   }
   ```

### 错误处理示例

```go
func safeGetItems(zoteroDB *core.ZoteroDB) {
    items, err := zoteroDB.GetItemsWithPDF(10)
    if err != nil {
        switch {
        case strings.Contains(err.Error(), "database is locked"):
            log.Println("数据库被锁定，请关闭 Zotero 后重试")
        case strings.Contains(err.Error(), "no such file"):
            log.Println("数据库文件不存在，请检查路径配置")
        default:
            log.Printf("未知错误: %v", err)
        }
        return
    }
    
    log.Printf("成功获取 %d 篇文献", len(items))
}
```

## 性能优化

### 连接池配置

```go
// 设置只读模式和连接池
db.SetMaxOpenConns(1)        // 只读访问，一个连接足够
db.SetMaxIdleConns(1)        // 最大空闲连接数
db.SetConnMaxLifetime(time.Hour) // 连接最大生存时间
```

### 查询优化

1. **使用索引**: 确保常用查询字段有索引
2. **限制结果集**: 使用 LIMIT 限制返回数量
3. **只读模式**: 避免写锁竞争

### 缓存策略

```go
type CachedZoteroDB struct {
    *ZoteroDB
    cache map[string]interface{}
    mutex sync.RWMutex
}

func (c *CachedZoteroDB) GetItemsWithPDF(limit int) ([]ZoteroItem, error) {
    cacheKey := fmt.Sprintf("items_pdf_%d", limit)
    
    c.mutex.RLock()
    if cached, exists := c.cache[cacheKey]; exists {
        c.mutex.RUnlock()
        return cached.([]ZoteroItem), nil
    }
    c.mutex.RUnlock()
    
    items, err := c.ZoteroDB.GetItemsWithPDF(limit)
    if err != nil {
        return nil, err
    }
    
    c.mutex.Lock()
    c.cache[cacheKey] = items
    c.mutex.Unlock()
    
    return items, nil
}
```

## 使用示例

### 完整的文献查询流程

```go
package main

import (
    "fmt"
    "log"
    "zoteroflow2-server/core"
)

func main() {
    // 1. 连接数据库
    zoteroDB, err := core.NewZoteroDB(
        "/home/user/Zotero/zotero.sqlite",
        "/home/user/Zotero/storage",
    )
    if err != nil {
        log.Fatalf("连接数据库失败: %v", err)
    }
    defer zoteroDB.Close()

    // 2. 获取统计信息
    stats, err := zoteroDB.GetStats()
    if err != nil {
        log.Printf("获取统计信息失败: %v", err)
    } else {
        fmt.Printf("数据库统计: %+v\n", stats)
    }

    // 3. 搜索文献
    results, err := zoteroDB.SearchByTitle("机器学习", 5)
    if err != nil {
        log.Printf("搜索失败: %v", err)
        return
    }

    // 4. 处理搜索结果
    for i, result := range results {
        fmt.Printf("\n[%d] %s\n", i+1, result.Title)
        fmt.Printf("作者: %s\n", strings.Join(result.Authors, "; "))
        fmt.Printf("评分: %.1f\n", result.Score)
        
        if result.PDFPath != "" {
            fmt.Printf("PDF路径: %s\n", result.PDFPath)
            
            // 检查文件是否存在
            if _, err := os.Stat(result.PDFPath); err == nil {
                fmt.Printf("✅ PDF文件存在\n")
            } else {
                fmt.Printf("❌ PDF文件不存在\n")
            }
        }
        
        // 获取标签
        if tags, err := zoteroDB.getItemTags(result.ItemID); err == nil {
            fmt.Printf("标签: %s\n", strings.Join(tags, ", "))
        }
    }
}
```

## 最佳实践

### 1. 资源管理
- 总是调用 `Close()` 关闭数据库连接
- 使用 `defer` 确保资源释放

### 2. 错误处理
- 检查所有返回的错误
- 根据错误类型提供用户友好的错误信息

### 3. 性能考虑
- 合理设置查询限制
- 避免频繁的小查询
- 考虑使用缓存

### 4. 安全性
- 只使用只读模式访问数据库
- 验证文件路径的安全性
- 不要暴露敏感的数据库信息

## 版本兼容性

- **Zotero 6.x**: 完全支持
- **Zotero 7.x**: 完全支持
- **SQLite 3.x**: 支持 3.35+ 版本

## 注意事项

1. **数据库锁定**: 确保 Zotero 客户端未在运行时进行批量操作
2. **路径配置**: 使用绝对路径避免相对路径问题
3. **文件权限**: 确保对数据库文件和存储目录有读取权限
4. **并发限制**: 由于使用只读模式，建议限制并发连接数