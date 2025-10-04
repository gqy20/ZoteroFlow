# MinerU PDF 解析接口 API

## 概述

MinerU 接口提供了与 MinerU PDF 解析服务的集成功能，支持 PDF 文件上传、解析状态监控、结果下载等完整的 PDF 处理流程。

## 数据结构

### MinerUClient

```go
type MinerUClient struct {
    BaseURL    string        // MinerU API 基础URL
    Token      string        // 认证令牌
    HTTPClient *http.Client  // HTTP客户端
    MaxRetry   int           // 最大重试次数
    Timeout    time.Duration // 超时时间
}
```

### FileInfo

```go
type FileInfo struct {
    Name  string `json:"name"`   // 文件名
    IsOCR bool   `json:"is_ocr"` // 是否启用OCR
}
```

### BatchRequest

```go
type BatchRequest struct {
    Language string     `json:"language"` // 处理语言
    Files    []FileInfo `json:"files"`    // 文件列表
}
```

### BatchResponse

```go
type BatchResponse struct {
    Data BatchData `json:"data"`
}

type BatchData struct {
    BatchID  string   `json:"batch_id"`  // 批次ID
    FileURLs []string `json:"file_urls"` // 上传URL列表
}
```

### ExtractResult

```go
type ExtractResult struct {
    FileName   string `json:"file_name"`    // 文件名
    State      string `json:"state"`       // 处理状态
    FullZipURL string `json:"full_zip_url,omitempty"` // 完整ZIP下载URL
}
```

### StatusResponse

```go
type StatusResponse struct {
    Data StatusData `json:"data"`
}

type StatusData struct {
    ExtractResult []ExtractResult `json:"extract_result"`
}
```

### ParseResult

```go
type ParseResult struct {
    TaskID    string    `json:"task_id"`     // 任务ID
    Status    string    `json:"status"`      // 解析状态
    Content   string    `json:"content"`     // 解析内容
    Message   string    `json:"message"`     // 状态消息
    ErrorCode string    `json:"error_code"`  // 错误代码
    ZipPath   string    `json:"zip_path"`    // ZIP文件路径
    ParseTime time.Time `json:"parse_time"`  // 解析时间
    PDFPath   string    `json:"pdf_path"`    // 原始PDF路径
    FileName  string    `json:"file_name"`   // 文件名
    FileSize  int64     `json:"file_size"`   // 文件大小
    Duration  int64     `json:"duration_ms"` // 解析耗时（毫秒）
}
```

### ParseRecord

```go
type ParseRecord struct {
    ID           string    `csv:"id"`            // 唯一标识
    TaskID       string    `csv:"task_id"`       // MinerU任务ID
    FileName     string    `csv:"file_name"`     // 文件名
    PDFPath      string    `csv:"pdf_path"`      // PDF路径
    FileSize     int64     `csv:"file_size"`     // 文件大小（字节）
    Status       string    `csv:"status"`        // 解析状态
    ZipPath      string    `csv:"zip_path"`      // 结果ZIP路径
    ParseTime    time.Time `csv:"parse_time"`    // 解析时间
    Duration     int64     `csv:"duration_ms"`   // 解析耗时（毫秒）
    ErrorMessage string    `csv:"error_message"` // 错误信息
}
```

## 核心接口

### NewMinerUClient

创建新的 MinerU 客户端实例。

```go
func NewMinerUClient(apiURL, token string) *MinerUClient
```

**参数**:
- `apiURL` (string): MinerU API 基础URL
- `token` (string): 认证令牌

**返回值**:
- `*MinerUClient`: MinerU 客户端实例

**示例**:
```go
client := core.NewMinerUClient(
    "https://mineru.net/api/v4",
    "your_mineru_token_here",
)
```

**默认配置**:
- HTTP超时: 120秒
- 最大重试次数: 3次
- 连接超时: 3分钟

### ParsePDF

解析 PDF 文件的主要接口。

```go
func (c *MinerUClient) ParsePDF(ctx context.Context, pdfPath string) (*ParseResult, error)
```

**参数**:
- `ctx` (context.Context): 上下文对象，用于超时控制
- `pdfPath` (string): PDF 文件路径

**返回值**:
- `*ParseResult`: 解析结果
- `error`: 错误信息

**示例**:
```go
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
defer cancel()

result, err := client.ParsePDF(ctx, "/path/to/document.pdf")
if err != nil {
    log.Printf("PDF解析失败: %v", err)
    return
}

fmt.Printf("解析完成，任务ID: %s\n", result.TaskID)
fmt.Printf("解析耗时: %d 毫秒\n", result.Duration)
fmt.Printf("结果保存到: %s\n", result.ZipPath)
```

**处理流程**:
1. 验证 PDF 文件存在性
2. 提交批量解析任务
3. 上传 PDF 文件
4. 轮询处理状态
5. 下载解析结果
6. 保存解析记录
7. 异步组织文件结构

## 内部方法

### submitBatchTask

提交批量解析任务。

```go
func (c *MinerUClient) submitBatchTask(ctx context.Context, fileName string) (*BatchResponse, error)
```

**请求示例**:
```json
{
    "language": "ch",
    "files": [
        {
            "name": "document.pdf",
            "is_ocr": true
        }
    ]
}
```

**响应示例**:
```json
{
    "data": {
        "batch_id": "batch_123456789",
        "file_urls": [
            "https://mineru.net/upload/upload_abc123"
        ]
    }
}
```

### uploadFile

上传 PDF 文件到 MinerU 服务器。

```go
func (c *MinerUClient) uploadFile(ctx context.Context, uploadURL, filePath string) error
```

**参数**:
- `uploadURL` (string): 上传URL
- `filePath` (string): 本地文件路径

**HTTP方法**: PUT
**内容类型**: application/octet-stream

### pollStatus

轮询解析任务状态。

```go
func (c *MinerUClient) pollStatus(ctx context.Context, batchID string) (string, error)
```

**参数**:
- `batchID` (string): 批次任务ID

**返回值**:
- `string`: 完整ZIP文件的下载URL
- `error`: 错误信息

**轮询策略**:
- 间隔: 10秒
- 最大轮询次数: 18次（3分钟超时）
- 状态检查: `done`, `failed`, `processing`

**状态响应示例**:
```json
{
    "data": {
        "extract_result": [
            {
                "file_name": "document.pdf",
                "state": "done",
                "full_zip_url": "https://mineru.net/results/result_123456.zip"
            }
        ]
    }
}
```

### checkStatus

检查单次任务状态。

```go
func (c *MinerUClient) checkStatus(ctx context.Context, batchID string) (*StatusResponse, error)
```

**HTTP方法**: GET
**URL**: `/extract-results/batch/{batchID}`

### downloadResult

下载解析结果文件。

```go
func (c *MinerUClient) downloadResult(ctx context.Context, resultURL, outputPath string) error
```

**参数**:
- `resultURL` (string): 结果文件下载URL
- `outputPath` (string): 本地保存路径

**特性**:
- 自动创建输出目录
- 支持大文件下载
- 进度监控（可扩展）

## 记录管理

### saveParseRecord

保存解析记录到 CSV 文件。

```go
func (c *MinerUClient) saveParseRecord(record ParseRecord) error
```

**文件位置**: `data/records/mineru_parse_records_{date}.csv`

**CSV格式**:
```csv
id,task_id,file_name,pdf_path,file_size,status,zip_path,parse_time,duration_ms,error_message
123456789,batch_123,document.pdf,/path/to/doc.pdf,2048576,completed,/path/to/result.zip,2024-12-01 10:30:00,75000,
```

### GetParseRecords

获取指定日期的解析记录。

```go
func GetParseRecords(date string) ([]ParseRecord, error)
```

**参数**:
- `date` (string): 日期字符串（格式：2006-01-02），为空时使用今天

**返回值**:
- `[]ParseRecord`: 解析记录列表
- `error`: 错误信息

**示例**:
```go
// 获取今天的记录
records, err := core.GetParseRecords("")
if err != nil {
    log.Printf("获取记录失败: %v", err)
    return
}

// 获取指定日期的记录
records, err := core.GetParseRecords("2024-12-01")
if err != nil {
    log.Printf("获取记录失败: %v", err)
    return
}

for _, record := range records {
    fmt.Printf("文件: %s, 状态: %s, 耗时: %dms\n", 
        record.FileName, record.Status, record.Duration)
}
```

## 错误处理

### 错误类型

1. **文件相关错误**
   - 文件不存在
   - 文件读取失败
   - 文件大小超限

2. **网络相关错误**
   - 连接超时
   - API调用失败
   - 认证失败

3. **解析相关错误**
   - 任务提交失败
   - 文件上传失败
   - 解析处理失败
   - 结果下载失败

### 错误处理示例

```go
func parseWithErrorHandling(client *core.MinerUClient, pdfPath string) {
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
    defer cancel()

    result, err := client.ParsePDF(ctx, pdfPath)
    if err != nil {
        switch {
        case strings.Contains(err.Error(), "file not found"):
            log.Printf("文件不存在: %s", pdfPath)
        case strings.Contains(err.Error(), "timeout"):
            log.Printf("解析超时，请稍后重试")
        case strings.Contains(err.Error(), "authentication"):
            log.Printf("认证失败，请检查API密钥")
        case strings.Contains(err.Error(), "processing failed"):
            log.Printf("解析失败，可能是文件格式不支持")
        default:
            log.Printf("未知错误: %v", err)
        }
        return
    }

    log.Printf("解析成功: %s", result.TaskID)
}
```

## 性能优化

### 并发控制

```go
type ConcurrentMinerUClient struct {
    *MinerUClient
    semaphore chan struct{} // 限制并发数
}

func NewConcurrentMinerUClient(apiURL, token string, maxConcurrent int) *ConcurrentMinerUClient {
    return &ConcurrentMinerUClient{
        MinerUClient: NewMinerUClient(apiURL, token),
        semaphore:    make(chan struct{}, maxConcurrent),
    }
}

func (c *ConcurrentMinerUClient) ParsePDF(ctx context.Context, pdfPath string) (*ParseResult, error) {
    c.semaphore <- struct{}{} // 获取信号量
    defer func() { <-c.semaphore }() // 释放信号量
    
    return c.MinerUClient.ParsePDF(ctx, pdfPath)
}
```

### 批量处理

```go
func batchParse(client *core.MinerUClient, pdfPaths []string) ([]*core.ParseResult, error) {
    var results []*core.ParseResult
    var errors []error
    
    for _, pdfPath := range pdfPaths {
        ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
        result, err := client.ParsePDF(ctx, pdfPath)
        cancel()
        
        if err != nil {
            errors = append(errors, fmt.Errorf("%s: %w", pdfPath, err))
            continue
        }
        
        results = append(results, result)
    }
    
    if len(errors) > 0 {
        log.Printf("批量解析完成，成功: %d, 失败: %d", len(results), len(errors))
        for _, err := range errors {
            log.Printf("错误: %v", err)
        }
    }
    
    return results, nil
}
```

### 缓存策略

```go
type CachedMinerUClient struct {
    *MinerUClient
    cache map[string]*core.ParseResult
    mutex sync.RWMutex
}

func (c *CachedMinerUClient) ParsePDF(ctx context.Context, pdfPath string) (*core.ParseResult, error) {
    // 生成缓存键
    cacheKey := c.generateCacheKey(pdfPath)
    
    // 检查缓存
    c.mutex.RLock()
    if cached, exists := c.cache[cacheKey]; exists {
        c.mutex.RUnlock()
        return cached, nil
    }
    c.mutex.RUnlock()
    
    // 调用原始方法
    result, err := c.MinerUClient.ParsePDF(ctx, pdfPath)
    if err != nil {
        return nil, err
    }
    
    // 缓存结果
    c.mutex.Lock()
    c.cache[cacheKey] = result
    c.mutex.Unlock()
    
    return result, nil
}

func (c *CachedMinerUClient) generateCacheKey(pdfPath string) string {
    h := md5.New()
    h.Write([]byte(pdfPath))
    return fmt.Sprintf("%x", h.Sum(nil))
}
```

## 使用示例

### 基础使用

```go
package main

import (
    "context"
    "log"
    "time"
    "zoteroflow2-server/core"
)

func main() {
    // 创建客户端
    client := core.NewMinerUClient(
        "https://mineru.net/api/v4",
        "your_token_here",
    )
    
    // 解析PDF
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
    defer cancel()
    
    result, err := client.ParsePDF(ctx, "/path/to/document.pdf")
    if err != nil {
        log.Fatalf("解析失败: %v", err)
    }
    
    log.Printf("解析成功:")
    log.Printf("任务ID: %s", result.TaskID)
    log.Printf("文件大小: %d 字节", result.FileSize)
    log.Printf("解析耗时: %d 毫秒", result.Duration)
    log.Printf("结果路径: %s", result.ZipPath)
}
```

### 批量处理

```go
func batchProcessExample() {
    client := core.NewMinerUClient("https://mineru.net/api/v4", "token")
    
    pdfFiles := []string{
        "/path/to/doc1.pdf",
        "/path/to/doc2.pdf",
        "/path/to/doc3.pdf",
    }
    
    results, err := batchParse(client, pdfFiles)
    if err != nil {
        log.Printf("批量处理失败: %v", err)
        return
    }
    
    for i, result := range results {
        log.Printf("[%d] %s: %s", i+1, result.FileName, result.Status)
    }
}
```

### 记录查询

```go
func recordQueryExample() {
    // 获取今天的解析记录
    records, err := core.GetParseRecords("")
    if err != nil {
        log.Printf("获取记录失败: %v", err)
        return
    }
    
    log.Printf("今天共解析了 %d 个文件", len(records))
    
    // 统计成功率
    successCount := 0
    totalDuration := int64(0)
    
    for _, record := range records {
        if record.Status == "completed" {
            successCount++
        }
        totalDuration += record.Duration
    }
    
    if len(records) > 0 {
        successRate := float64(successCount) / float64(len(records)) * 100
        avgDuration := totalDuration / int64(len(records))
        
        log.Printf("成功率: %.1f%%", successRate)
        log.Printf("平均耗时: %d 毫秒", avgDuration)
    }
}
```

## 配置建议

### 环境变量

```bash
# MinerU API 配置
MINERU_API_URL=https://mineru.net/api/v4
MINERU_TOKEN=your_token_here

# 性能配置
MINERU_TIMEOUT=300          # 超时时间（秒）
MINERU_MAX_RETRY=3          # 最大重试次数
MINERU_MAX_CONCURRENT=5     # 最大并发数
```

### 超时配置

```go
// 根据文件大小调整超时时间
func getTimeout(fileSize int64) time.Duration {
    if fileSize < 10*1024*1024 { // < 10MB
        return 2 * time.Minute
    } else if fileSize < 50*1024*1024 { // < 50MB
        return 5 * time.Minute
    } else {
        return 10 * time.Minute
    }
}
```

## 监控和日志

### 解析状态监控

```go
func monitorParsing(client *core.MinerUClient, batchID string) {
    ticker := time.NewTicker(30 * time.Second)
    defer ticker.Stop()
    
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
    defer cancel()
    
    for {
        select {
        case <-ctx.Done():
            log.Printf("监控超时")
            return
        case <-ticker.C:
            resp, err := client.checkStatus(ctx, batchID)
            if err != nil {
                log.Printf("状态检查失败: %v", err)
                continue
            }
            
            if len(resp.Data.ExtractResult) > 0 {
                state := resp.Data.ExtractResult[0].State
                log.Printf("任务状态: %s", state)
                
                if state == "done" || state == "failed" {
                    return
                }
            }
        }
    }
}
```

## 最佳实践

### 1. 资源管理
- 使用 context 控制超时
- 及时释放 HTTP 连接
- 合理设置并发限制

### 2. 错误处理
- 区分不同类型的错误
- 提供用户友好的错误信息
- 实现重试机制

### 3. 性能优化
- 使用并发处理提高效率
- 实现缓存避免重复解析
- 根据文件大小调整超时时间

### 4. 监控和日志
- 记录详细的解析过程
- 监控解析状态和性能
- 定期清理过期的缓存文件

## 版本兼容性

- **MinerU API v4**: 完全支持
- **HTTP/1.1**: 支持
- **HTTP/2**: 支持
- **TLS 1.2+**: 支持

## 注意事项

1. **API限制**: 注意 MinerU API 的调用频率限制
2. **文件大小**: 单个文件建议不超过 100MB
3. **并发控制**: 避免同时提交过多任务
4. **网络稳定性**: 长时间解析需要考虑网络中断情况
5. **存储空间**: 确保有足够的磁盘空间存储解析结果