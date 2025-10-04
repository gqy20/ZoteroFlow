# 文件组织接口 API

## 概述

文件组织接口提供了 MinerU 解析结果的后处理功能，包括 ZIP 文件解压、文件结构整理、元数据生成、索引链接创建等完整的文件组织流程。

## 数据结构

### ParsedFileInfo

```go
type ParsedFileInfo struct {
    Title    string `json:"title"`    // 文件标题
    Authors  string `json:"authors"`  // 作者信息
    Date     string `json:"date"`     // 处理日期
    Size     int64  `json:"size"`     // 文件大小
    Duration int64  `json:"duration"` // 处理耗时
    Path     string `json:"path"`     // 文件路径
}
```

## 核心接口

### OrganizeResult

解压并组织文件的核心函数。

```go
func OrganizeResult(zipPath, pdfPath string) error
```

**参数**:
- `zipPath` (string): MinerU 生成的 ZIP 文件路径
- `pdfPath` (string): 原始 PDF 文件路径

**返回值**:
- `error`: 错误信息

**处理流程**:
1. 创建目标目录
2. 解压 ZIP 文件
3. 复制原始 PDF 文件
4. 移动 ZIP 文件
5. 生成元数据文件
6. 整理文件结构
7. 创建索引链接

**示例**:
```go
err := core.OrganizeResult(
    "data/results/temp.zip",
    "/path/to/document.pdf",
)
if err != nil {
    log.Printf("文件组织失败: %v", err)
    return
}

fmt.Println("文件组织完成")
```

**目录结构**:
```
data/results/
├── 文献标题_20241201/
│   ├── full.md           # 完整Markdown内容
│   ├── meta.json         # 元数据文件
│   ├── source.pdf        # 原始PDF文件
│   ├── raw.zip          # 原始ZIP文件
│   ├── images/           # 图片文件目录
│   │   ├── img1.png
│   │   ├── img2.jpg
│   │   └── ...
│   └── tables/           # 表格文件目录
│       ├── table1.csv
│       └── ...
└── latest/               # 最新结果软链接
    └── -> 文献标题_20241201/
```

## 辅助功能

### sanitizeFilename

清理文件名，移除特殊字符。

```go
func sanitizeFilename(name string) string
```

**参数**:
- `name` (string): 原始文件名

**返回值**:
- `string`: 清理后的文件名

**处理规则**:
- 保留中文、英文、数字、连字符和下划线
- 移除其他特殊字符
- 合并多个下划线
- 限制长度为30个字符

**示例**:
```go
cleanName := core.sanitizeFilename("机器学习基础教程(2024).pdf")
// 返回: "机器学习基础教程2024_pdf"

cleanName := core.sanitizeFilename("Deep Learning@#$%^&*")
// 返回: "Deep_Learning"
```

### extractTitle

从 PDF 路径提取标题。

```go
func extractTitle(pdfPath string) string
```

**参数**:
- `pdfPath` (string): PDF 文件路径

**返回值**:
- `string`: 提取的标题

**处理逻辑**:
- 提取文件名（去除扩展名）
- 移除常见前缀（2025_, 2024_, doi_, jcr_）
- 限制长度为20个字符

**示例**:
```go
title := core.extractTitle("/path/to/2024_机器学习基础.pdf")
// 返回: "机器学习基础"

title := core.extractTitle("/path/to/doi_10.1234_ml.2024.pdf")
// 返回: "ml.2024"
```

### unzipFile

解压 ZIP 文件到目标目录。

```go
func unzipFile(zipPath, targetDir string) error
```

**参数**:
- `zipPath` (string): ZIP 文件路径
- `targetDir` (string): 目标目录

**返回值**:
- `error`: 错误信息

**特性**:
- 自动创建 images 目录
- 路径安全检查
- 图片文件自动分类
- 错误处理和日志记录

**示例**:
```go
err := core.unzipFile("result.zip", "output/")
if err != nil {
    log.Printf("解压失败: %v", err)
    return
}

fmt.Println("解压完成")
```

### extractFile

提取单个文件。

```go
func extractFile(file *zip.File, targetPath, imagesDir string) error
```

**参数**:
- `file` (*zip.File): ZIP 文件对象
- `targetPath` (string): 目标路径
- `imagesDir` (string): 图片目录

**特性**:
- 图片文件自动移动到 images 目录
- 保持原始文件权限
- 支持大文件处理

### isImageFile

检查是否为图片文件。

```go
func isImageFile(filename string) bool
```

**参数**:
- `filename` (string): 文件名

**返回值**:
- `bool`: 是否为图片文件

**支持的格式**:
- `.jpg`, `.jpeg`
- `.png`
- `.gif`

### generateMeta

生成元数据文件。

```go
func generateMeta(targetDir, originalPath, pdfPath string) error
```

**参数**:
- `targetDir` (string): 目标目录
- `originalPath` (string): 原始路径
- `pdfPath` (string): PDF 文件路径

**元数据内容**:
```json
{
    "title": "机器学习基础教程",
    "authors": "张三; 李四",
    "date": "2024-12-01",
    "size": 2048576,
    "duration": 75000,
    "path": "/path/to/result"
}
```

### extractBasicInfo

从内容中提取基本信息。

```go
func extractBasicInfo(content string) string
```

**参数**:
- `content` (string): 文档内容

**返回值**:
- `string`: 基本信息（前10行）

### extractAuthors

提取作者信息。

```go
func extractAuthors(content string) string
```

**参数**:
- `content` (string): 文档内容

**返回值**:
- `string`: 作者信息

**搜索逻辑**:
- 查找包含 "作者" 或 "Author" 的行
- 返回该行的内容

### organizeFiles

整理文件结构。

```go
func organizeFiles(targetDir string) error
```

**参数**:
- `targetDir` (string): 目标目录

**处理逻辑**:
- 移动 markdown 文件到根目录
- 重命名为 full.md
- 保持其他文件结构不变

### createSymlink

创建索引链接。

```go
func createSymlink(targetDir, folderName string) error
```

**参数**:
- `targetDir` (string): 目标目录
- `folderName` (string): 文件夹名称

**功能**:
- 创建 latest 软链接
- 指向最新的解析结果
- 自动删除旧链接

### copyFile

复制文件。

```go
func copyFile(src, dst string) error
```

**参数**:
- `src` (string): 源文件路径
- `dst` (string): 目标文件路径

**特性**:
- 支持大文件复制
- 保持文件权限
- 错误处理

## 使用示例

### 基础使用

```go
package main

import (
    "fmt"
    "log"
    "zoteroflow2-server/core"
)

func main() {
    // 组织解析结果
    zipPath := "data/results/temp.zip"
    pdfPath := "/path/to/document.pdf"
    
    err := core.OrganizeResult(zipPath, pdfPath)
    if err != nil {
        log.Fatalf("文件组织失败: %v", err)
    }
    
    fmt.Println("文件组织完成")
}
```

### 自定义文件名处理

```go
func customFilenameExample() {
    // 测试文件名清理
    testNames := []string{
        "机器学习基础教程(2024).pdf",
        "Deep Learning@#$%^&*.pdf",
        "2024_人工智能_研究_进展.pdf",
        "jcr_期刊论文_编号123.pdf",
    }
    
    for _, name := range testNames {
        clean := core.sanitizeFilename(name)
        fmt.Printf("原始: %s -> 清理: %s\n", name, clean)
    }
}
```

### 批量处理

```go
func batchOrganizeExample() {
    // 获取所有需要组织的ZIP文件
    zipFiles, err := filepath.Glob("data/results/temp/*.zip")
    if err != nil {
        log.Printf("获取ZIP文件失败: %v", err)
        return
    }
    
    for i, zipFile := range zipFiles {
        // 对应的PDF文件
        pdfFile := strings.Replace(zipFile, ".zip", ".pdf", 1)
        
        fmt.Printf("处理第 %d 个文件: %s\n", i+1, zipFile)
        
        err := core.OrganizeResult(zipFile, pdfFile)
        if err != nil {
            log.Printf("处理失败: %s, 错误: %v", zipFile, err)
            continue
        }
        
        fmt.Printf("✅ 处理完成: %s\n", zipFile)
    }
    
    fmt.Printf("批量处理完成，共处理 %d 个文件\n", len(zipFiles))
}
```

### 元数据分析

```go
func metadataAnalysisExample() {
    // 读取元数据文件
    metaFile := "data/results/机器学习基础_20241201/meta.json"
    
    data, err := os.ReadFile(metaFile)
    if err != nil {
        log.Printf("读取元数据失败: %v", err)
        return
    }
    
    var info core.ParsedFileInfo
    if err := json.Unmarshal(data, &info); err != nil {
        log.Printf("解析元数据失败: %v", err)
        return
    }
    
    fmt.Printf("文件信息:\n")
    fmt.Printf("  标题: %s\n", info.Title)
    fmt.Printf("  作者: %s\n", info.Authors)
    fmt.Printf("  日期: %s\n", info.Date)
    fmt.Printf("  大小: %.2f MB\n", float64(info.Size)/1024/1024)
    fmt.Printf("  耗时: %.2f 秒\n", float64(info.Duration)/1000)
    fmt.Printf("  路径: %s\n", info.Path)
}
```

## 错误处理

### 常见错误类型

1. **文件系统错误**
   ```go
   if os.IsNotExist(err) {
       // 文件不存在
   }
   ```

2. **权限错误**
   ```go
   if os.IsPermission(err) {
       // 权限不足
   }
   ```

3. **ZIP文件错误**
   ```go
   if strings.Contains(err.Error(), "zip") {
       // ZIP文件损坏或格式错误
   }
   ```

4. **磁盘空间不足**
   ```go
   if strings.Contains(err.Error(), "no space") {
       // 磁盘空间不足
   }
   ```

### 错误处理示例

```go
func safeOrganizeResult(zipPath, pdfPath string) error {
    // 检查文件存在性
    if _, err := os.Stat(zipPath); os.IsNotExist(err) {
        return fmt.Errorf("ZIP文件不存在: %s", zipPath)
    }
    
    if _, err := os.Stat(pdfPath); os.IsNotExist(err) {
        return fmt.Errorf("PDF文件不存在: %s", pdfPath)
    }
    
    // 检查磁盘空间
    var stat syscall.Statfs_t
    if err := syscall.Statfs(".", &stat); err == nil {
        freeSpace := stat.Bavail * uint64(stat.Bsize)
        requiredSpace := uint64(100 * 1024 * 1024) // 100MB
        
        if freeSpace < requiredSpace {
            return fmt.Errorf("磁盘空间不足，需要 %d MB，可用 %d MB", 
                requiredSpace/1024/1024, freeSpace/1024/1024)
        }
    }
    
    // 执行组织操作
    err := OrganizeResult(zipPath, pdfPath)
    if err != nil {
        return fmt.Errorf("文件组织失败: %w", err)
    }
    
    return nil
}
```

## 性能优化

### 并发处理

```go
func concurrentOrganize(zipFiles, pdfFiles []string) error {
    var wg sync.WaitGroup
    var mu sync.Mutex
    var errors []error
    
    semaphore := make(chan struct{}, 5) // 限制并发数
    
    for i := 0; i < len(zipFiles); i++ {
        wg.Add(1)
        go func(index int) {
            defer wg.Done()
            
            semaphore <- struct{}{}
            defer func() { <-semaphore }()
            
            err := OrganizeResult(zipFiles[index], pdfFiles[index])
            if err != nil {
                mu.Lock()
                errors = append(errors, fmt.Errorf("文件 %s: %w", zipFiles[index], err))
                mu.Unlock()
            }
        }(i)
    }
    
    wg.Wait()
    
    if len(errors) > 0 {
        log.Printf("并发处理完成，失败 %d 个文件", len(errors))
        for _, err := range errors {
            log.Printf("错误: %v", err)
        }
    }
    
    return nil
}
```

### 内存优化

```go
func optimizedUnzip(zipPath, targetDir string) error {
    reader, err := zip.OpenReader(zipPath)
    if err != nil {
        return err
    }
    defer reader.Close()
    
    // 创建images目录
    imagesDir := filepath.Join(targetDir, "images")
    if err := os.MkdirAll(imagesDir, 0755); err != nil {
        return err
    }
    
    // 缓冲区大小优化
    buffer := make([]byte, 32*1024) // 32KB缓冲区
    
    for _, file := range reader.File {
        path := filepath.Join(targetDir, file.Name)
        
        // 安全检查
        if !strings.HasPrefix(path, targetDir) {
            continue
        }
        
        if file.FileInfo().IsDir() {
            os.MkdirAll(path, file.FileInfo().Mode())
            continue
        }
        
        // 处理文件
        if err := extractFileOptimized(file, path, imagesDir, buffer); err != nil {
            log.Printf("提取文件失败 %s: %v", file.Name, err)
            continue
        }
    }
    
    return nil
}

func extractFileOptimized(file *zip.File, targetPath, imagesDir string, buffer []byte) error {
    src, err := file.Open()
    if err != nil {
        return err
    }
    defer src.Close()
    
    // 图片文件处理
    if isImageFile(file.Name) {
        filename := filepath.Base(file.Name)
        targetPath = filepath.Join(imagesDir, filename)
    }
    
    dst, err := os.OpenFile(targetPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, file.FileInfo().Mode())
    if err != nil {
        return err
    }
    defer dst.Close()
    
    // 使用缓冲区复制
    _, err = io.CopyBuffer(dst, src, buffer)
    return err
}
```

## 配置建议

### 环境变量

```bash
# 文件组织配置
ORGANIZER_RESULTS_DIR=data/results      # 结果目录
ORGANIZER_IMAGES_DIR=images            # 图片目录名
ORGANIZER_MAX_FILENAME_LEN=30         # 最大文件名长度
ORGANIZER_CREATE_SYMLINK=true           # 创建软链接

# 性能配置
ORGANIZER_MAX_CONCURRENT=5             # 最大并发数
ORGANIZER_BUFFER_SIZE=32768             # 缓冲区大小
ORGANIZER_TIMEOUT=300s                   # 处理超时时间
```

### 目录结构配置

```go
type OrganizerConfig struct {
    ResultsDir    string `json:"results_dir"`    // 结果目录
    ImagesDir     string `json:"images_dir"`     // 图片目录名
    MaxFilenameLen int    `json:"max_filename_len"` // 最大文件名长度
    CreateSymlink  bool   `json:"create_symlink"`  // 是否创建软链接
}

func DefaultOrganizerConfig() *OrganizerConfig {
    return &OrganizerConfig{
        ResultsDir:    "data/results",
        ImagesDir:     "images",
        MaxFilenameLen: 30,
        CreateSymlink:  true,
    }
}
```

## 监控和日志

### 处理进度监控

```go
func monitorOrganizingProgress(zipPath string) {
    ticker := time.NewTicker(5 * time.Second)
    defer ticker.Stop()
    
    initialSize, _ := os.Stat(zipPath)
    targetDir := filepath.Dir(zipPath)
    
    for range ticker.C {
        currentSize, err := os.Stat(targetDir)
        if err != nil {
            continue
        }
        
        progress := float64(currentSize.Size()) / float64(initialSize.Size()) * 100
        log.Printf("文件组织进度: %.1f%%", progress)
        
        if progress >= 100 {
            break
        }
    }
    
    log.Printf("文件组织完成")
}
```

### 文件统计

```go
type FileStats struct {
    TotalFiles    int           `json:"total_files"`
    MarkdownFiles int           `json:"markdown_files"`
    ImageFiles    int           `json:"image_files"`
    OtherFiles    int           `json:"other_files"`
    TotalSize     int64         `json:"total_size"`
    ProcessTime   time.Duration `json:"process_time"`
}

func collectFileStats(targetDir string) (*FileStats, error) {
    stats := &FileStats{}
    startTime := time.Now()
    
    err := filepath.Walk(targetDir, func(path string, info os.FileInfo, err error) error {
        if err != nil {
            return nil
        }
        
        if info.IsDir() {
            return nil
        }
        
        stats.TotalFiles++
        stats.TotalSize += info.Size()
        
        switch filepath.Ext(path) {
        case ".md":
            stats.MarkdownFiles++
        case ".jpg", ".jpeg", ".png", ".gif":
            stats.ImageFiles++
        default:
            stats.OtherFiles++
        }
        
        return nil
    })
    
    stats.ProcessTime = time.Since(startTime)
    return stats, err
}

func (s *FileStats) String() string {
    return fmt.Sprintf(
        "文件统计: 总数=%d, Markdown=%d, 图片=%d, 其他=%d, 总大小=%.2f MB, 耗时=%v",
        s.TotalFiles, s.MarkdownFiles, s.ImageFiles, s.OtherFiles,
        float64(s.TotalSize)/1024/1024, s.ProcessTime,
    )
}
```

## 最佳实践

### 1. 文件安全
- 验证文件路径安全性
- 检查文件权限
- 避免路径遍历攻击

### 2. 错误处理
- 区分不同类型的错误
- 提供详细的错误信息
- 实现重试机制

### 3. 性能优化
- 使用并发处理提高效率
- 优化缓冲区大小
- 合理设置并发限制

### 4. 资源管理
- 及时释放文件句柄
- 监控磁盘空间使用
- 定期清理临时文件

## 版本兼容性

- **ZIP格式**: 支持 ZIP 2.0+
- **图片格式**: 支持 JPEG, PNG, GIF
- **文件系统**: 支持 Unix/Linux, Windows, macOS
- **Go 1.21+**: 支持

## 注意事项

1. **磁盘空间**: 确保有足够的磁盘空间存储解压后的文件
2. **文件权限**: 确保对目标目录有读写权限
3. **路径安全**: 验证文件路径，避免路径遍历攻击
4. **并发控制**: 避免过多的并发操作导致系统资源耗尽
5. **文件编码**: 注意文件名的编码问题，特别是中文文件名