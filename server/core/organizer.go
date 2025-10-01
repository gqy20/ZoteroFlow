package core

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

// ParsedFileInfo 解析后的文件信息
type ParsedFileInfo struct {
	Title    string `json:"title"`
	Authors  string `json:"authors"`
	Date     string `json:"date"`
	Size     int64  `json:"size"`
	Duration int64  `json:"duration"`
	Path     string `json:"path"`
}

// OrganizeResult 解压并组织文件 - 核心函数
func OrganizeResult(zipPath, pdfPath string) error {
	log.Printf("开始组织文件: %s", zipPath)

	// 1. 创建目标目录
	baseDir := "data/results"
	title := extractTitle(pdfPath)
	folderName := sanitizeFilename(fmt.Sprintf("%s_%s", title, time.Now().Format("20060102")))
	targetDir := filepath.Join(baseDir, folderName)

	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return fmt.Errorf("创建目录失败: %w", err)
	}

	// 2. 解压ZIP文件
	if err := unzipFile(zipPath, targetDir); err != nil {
		return fmt.Errorf("解压失败: %w", err)
	}

	// 3. 移动原始PDF到目录
	targetPDF := filepath.Join(targetDir, "source.pdf")
	if err := copyFile(pdfPath, targetPDF); err != nil {
		log.Printf("复制PDF失败: %v", err)
	}

	// 4. 移动ZIP文件（处理跨设备移动）
	targetZIP := filepath.Join(targetDir, "raw.zip")
	if err := os.Rename(zipPath, targetZIP); err != nil {
		log.Printf("移动ZIP失败，尝试复制: %v", err)
		// 如果是跨设备问题，使用复制然后删除的方式
		if err := copyFile(zipPath, targetZIP); err != nil {
			log.Printf("复制ZIP失败: %v", err)
		} else {
			os.Remove(zipPath)
		}
	}

	// 5. 生成简单元数据
	if err := generateMeta(targetDir, pdfPath, targetPDF); err != nil {
		log.Printf("生成元数据失败: %v", err)
	}

	// 6. 整理文件结构
	if err := organizeFiles(targetDir); err != nil {
		log.Printf("整理文件失败: %v", err)
	}

	// 7. 创建索引链接
	if err := createSymlink(targetDir, folderName); err != nil {
		log.Printf("创建链接失败: %v", err)
	}

	log.Printf("文件组织完成: %s", targetDir)
	return nil
}

// sanitizeFilename 清理文件名
func sanitizeFilename(name string) string {
	// 移除特殊字符，保留中文、英文、数字、连字符和下划线
	re := regexp.MustCompile(`[^\w\x{4e00}-\x{9fff}\-_.]`)
	clean := re.ReplaceAllString(name, "_")

	// 移除多余的下划线
	clean = regexp.MustCompile(`_+`).ReplaceAllString(clean, "_")
	clean = strings.Trim(clean, "_")

	// 限制长度
	if len(clean) > 30 {
		clean = clean[:30]
	}

	return clean
}

// extractTitle 从PDF路径提取标题
func extractTitle(pdfPath string) string {
	filename := filepath.Base(pdfPath)
	title := strings.TrimSuffix(filename, filepath.Ext(filename))

	// 移除常见的文件前缀
	prefixes := []string{"2025_", "2024_", "doi_", "jcr_"}
	for _, prefix := range prefixes {
		if strings.HasPrefix(strings.ToLower(title), prefix) {
			title = title[len(prefix):]
			break
		}
	}

	// 如果标题太长，截取
	if len(title) > 20 {
		title = title[:20]
	}

	return title
}

// unzipFile 解压ZIP文件
func unzipFile(zipPath, targetDir string) error {
	reader, err := zip.OpenReader(zipPath)
	if err != nil {
		return err
	}
	defer reader.Close()

	// 创建目标目录和images目录
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return fmt.Errorf("创建目标目录失败: %w", err)
	}

	imagesDir := filepath.Join(targetDir, "images")
	os.MkdirAll(imagesDir, 0755)

	for _, file := range reader.File {
		path := filepath.Join(targetDir, file.Name)

		// 确保路径安全
		if !strings.HasPrefix(path, targetDir) {
			continue
		}

		if file.FileInfo().IsDir() {
			os.MkdirAll(path, file.FileInfo().Mode())
			continue
		}

		// 处理文件
		if err := extractFile(file, path, imagesDir); err != nil {
			log.Printf("提取文件失败 %s: %v", file.Name, err)
			continue
		}
	}

	return nil
}

// extractFile 提取单个文件
func extractFile(file *zip.File, targetPath, imagesDir string) error {
	log.Printf("正在提取文件: %s -> %s", file.Name, targetPath)

	// 确保目标目录存在
	targetDir := filepath.Dir(targetPath)
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		log.Printf("创建目标目录失败 %s: %v", targetDir, err)
		return fmt.Errorf("创建目标目录失败: %w", err)
	}

	src, err := file.Open()
	if err != nil {
		log.Printf("打开ZIP内文件失败 %s: %v", file.Name, err)
		return fmt.Errorf("打开ZIP文件失败: %w", err)
	}
	defer src.Close()

	// 如果是图片，放到images目录
	if isImageFile(file.Name) {
		filename := filepath.Base(file.Name)
		targetPath = filepath.Join(imagesDir, filename)
		log.Printf("图片文件，重新定位到: %s", targetPath)

		// 确保images目录存在
		if err := os.MkdirAll(imagesDir, 0755); err != nil {
			log.Printf("创建images目录失败 %s: %v", imagesDir, err)
			return fmt.Errorf("创建images目录失败: %w", err)
		}
	}

	// 检查目标路径的目录
	targetDir = filepath.Dir(targetPath)
	if targetDir != "." {
		if err := os.MkdirAll(targetDir, 0755); err != nil {
			log.Printf("创建最终目标目录失败 %s: %v", targetDir, err)
			return fmt.Errorf("创建最终目标目录失败: %w", err)
		}
	}

	log.Printf("准备创建文件: %s (权限: %v)", targetPath, file.FileInfo().Mode())

	dst, err := os.OpenFile(targetPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, file.FileInfo().Mode())
	if err != nil {
		log.Printf("创建目标文件失败 %s: %v", targetPath, err)
		return fmt.Errorf("创建目标文件失败: %w", err)
	}
	defer dst.Close()

	// 获取文件大小用于进度显示
	fileSize := file.FileInfo().Size()
	log.Printf("开始复制文件: %s (大小: %d bytes)", file.Name, fileSize)

	written, err := io.Copy(dst, src)
	if err != nil {
		log.Printf("复制文件失败 %s: %v", file.Name, err)
		return fmt.Errorf("文件复制失败: %w", err)
	}

	// 验证写入的字节数
	if written != fileSize {
		log.Printf("警告: 文件大小不匹配 %s: 期望 %d bytes, 实际 %d bytes", file.Name, fileSize, written)
	} else {
		log.Printf("成功提取文件: %s (%d bytes)", file.Name, written)
	}

	// 验证文件是否真的被创建了
	if stat, err := os.Stat(targetPath); err == nil {
		log.Printf("文件验证成功: %s (大小: %d bytes)", targetPath, stat.Size())
	} else {
		log.Printf("文件验证失败: %s: %v", targetPath, err)
	}

	return nil
}

// isImageFile 检查是否为图片文件
func isImageFile(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	return ext == ".jpg" || ext == ".jpeg" || ext == ".png" || ext == ".gif"
}

// generateMeta 生成元数据文件
func generateMeta(targetDir, originalPath, pdfPath string) error {
	// 获取文件信息
	stat, err := os.Stat(pdfPath)
	if err != nil {
		return err
	}

	// 读取内容提取信息
	contentFile := filepath.Join(targetDir, "full.md")
	content := ""
	if data, err := os.ReadFile(contentFile); err == nil {
		content = extractBasicInfo(string(data))
	}

	info := ParsedFileInfo{
		Title:    extractTitle(originalPath),
		Authors:  extractAuthors(content),
		Date:     time.Now().Format("2006-01-02"),
		Size:     stat.Size(),
		Duration: 0, // 这里可以后续从解析记录获取
		Path:     targetDir,
	}

	metaFile := filepath.Join(targetDir, "meta.json")
	data, err := json.MarshalIndent(info, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(metaFile, data, 0644)
}

// extractBasicInfo 从内容中提取基本信息
func extractBasicInfo(content string) string {
	lines := strings.Split(content, "\n")
	if len(lines) > 10 {
		return strings.Join(lines[:10], "\n")
	}
	return content
}

// extractAuthors 提取作者信息
func extractAuthors(content string) string {
	// 简单的作者提取逻辑
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		if strings.Contains(line, "作者") || strings.Contains(line, "Author") {
			return strings.TrimSpace(line)
		}
	}
	return "未知"
}

// organizeFiles 整理文件结构
func organizeFiles(targetDir string) error {
	// 移动markdown文件到根目录
	entries, err := os.ReadDir(targetDir)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".md") {
			if entry.Name() != "full.md" {
				oldPath := filepath.Join(targetDir, entry.Name())
				newPath := filepath.Join(targetDir, "full.md")
				os.Rename(oldPath, newPath)
				break
			}
		}
	}

	return nil
}

// createSymlink 创建索引链接
func createSymlink(targetDir, folderName string) error {
	latestDir := filepath.Join("data/results", "latest")

	// 删除旧的链接
	os.Remove(latestDir)

	// 创建新的链接
	return os.Symlink(folderName, latestDir)
}

// copyFile 复制文件
func copyFile(src, dst string) error {
	source, err := os.Open(src)
	if err != nil {
		return err
	}
	defer source.Close()

	destination, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destination.Close()

	_, err = io.Copy(destination, source)
	return err
}
