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

	// 4. 删除原始ZIP文件（节省空间）
	if err := os.Remove(zipPath); err != nil {
		log.Printf("删除ZIP文件失败: %v", err)
	} else {
		log.Printf("已删除原始ZIP文件: %s", zipPath)
	}

	// 5. 生成简单元数据
	if err := generateMeta(targetDir, pdfPath, targetPDF); err != nil {
		log.Printf("生成元数据失败: %v", err)
	}

	// 6. 整理文件结构
	if err := organizeFiles(targetDir); err != nil {
		log.Printf("整理文件失败: %v", err)
	}

	// 7. 创建索引链接（简化版：跳过不必要的软链接）
	// 软链接增加复杂性且不符合实际需求，已移除
	log.Printf("跳过创建软链接（latest），保持简单架构")

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
	log.Printf("开始生成元数据文件: %s", targetDir)

	// 获取文件信息
	stat, err := os.Stat(pdfPath)
	if err != nil {
		log.Printf("获取PDF文件信息失败: %v", err)
		return err
	}

	// 读取内容提取信息
	contentFile := filepath.Join(targetDir, "full.md")
	content := ""

	// 检查full.md是否存在且有内容
	if contentInfo, err := os.Stat(contentFile); err == nil && contentInfo.Size() > 0 {
		if data, err := os.ReadFile(contentFile); err == nil {
			content = string(data)
			log.Printf("成功读取内容文件，大小: %d bytes", len(content))
		} else {
			log.Printf("读取内容文件失败: %v", err)
		}
	} else {
		log.Printf("内容文件不存在或为空，尝试从其他文件提取信息")
		// 尝试从其他markdown文件中提取内容
		if extractedContent := extractFromOtherFiles(targetDir); extractedContent != "" {
			content = extractedContent
			log.Printf("从其他文件提取到内容，大小: %d bytes", len(content))
		}
	}

	// 提取标题和作者
	title := extractTitle(originalPath)
	authors := extractAuthors(content)

	// 改进作者信息提取
	if authors == "未知" || authors == "" {
		// 尝试从文件名中提取作者信息
		if extractedAuthors := extractAuthorsFromFilename(originalPath); extractedAuthors != "" {
			authors = extractedAuthors
			log.Printf("从文件名提取作者: %s", authors)
		} else {
			authors = "解析中..."
		}
	}

	// 尝试改进标题提取
	if title == "" || len(title) < 5 {
		// 尝试从内容中提取更好的标题
		if contentTitle := extractTitleFromContent(content); contentTitle != "" {
			title = contentTitle
			log.Printf("从内容提取标题: %s", title)
		}
	}

	info := ParsedFileInfo{
		Title:    title,
		Authors:  authors,
		Date:     time.Now().Format("2006-01-02"),
		Size:     stat.Size(),
		Duration: 0, // 这里可以后续从解析记录获取
		Path:     targetDir,
	}

	metaFile := filepath.Join(targetDir, "meta.json")
	data, err := json.MarshalIndent(info, "", "  ")
	if err != nil {
		log.Printf("序列化元数据失败: %v", err)
		return err
	}

	if err := os.WriteFile(metaFile, data, 0644); err != nil {
		log.Printf("写入元数据文件失败: %v", err)
		return err
	}

	log.Printf("成功生成元数据文件: %s (标题: %s, 作者: %s)", metaFile, title, authors)
	return nil
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
	if content == "" {
		return "未知"
	}

	lines := strings.Split(content, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.Contains(line, "作者") || strings.Contains(line, "Author") {
			// 尝试提取更准确的作者信息
			if strings.Contains(line, "：") {
				parts := strings.Split(line, "：")
				if len(parts) > 1 {
					return strings.TrimSpace(parts[1])
				}
			}
			if strings.Contains(line, ":") {
				parts := strings.Split(line, ":")
				if len(parts) > 1 {
					return strings.TrimSpace(parts[1])
				}
			}
			return strings.TrimSpace(line)
		}
	}
	return "未知"
}

// extractFromOtherFiles 从其他markdown文件中提取内容
func extractFromOtherFiles(targetDir string) string {
	entries, err := os.ReadDir(targetDir)
	if err != nil {
		return ""
	}

	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".md") && entry.Name() != "full.md" {
			filePath := filepath.Join(targetDir, entry.Name())
			if data, err := os.ReadFile(filePath); err == nil {
				content := string(data)
				if len(content) > 100 { // 确保内容不为空
					return content
				}
			}
		}
	}
	return ""
}

// extractAuthorsFromFilename 从文件名中提取作者信息
func extractAuthorsFromFilename(filename string) string {
	// 移除常见的文件前缀和后缀
	basename := filepath.Base(filename)
	basename = strings.TrimSuffix(basename, filepath.Ext(basename))

	// 移除常见的日期前缀
	prefixes := []string{"2025_", "2024_", "2023_", "2022_", "2021_", "2020_"}
	for _, prefix := range prefixes {
		if strings.HasPrefix(basename, prefix) {
			basename = basename[len(prefix):]
			break
		}
	}

	// 移除DOI前缀
	if strings.HasPrefix(basename, "10.") {
		if idx := strings.Index(basename, "_"); idx > 0 {
			basename = basename[:idx]
		}
		return "学术期刊"
	}

	// 如果包含中文姓名，尝试提取
	if strings.Contains(basename, "--") {
		parts := strings.Split(basename, "--")
		if len(parts) > 1 && len(parts[0]) > 1 && len(parts[0]) < 10 {
			return parts[0]
		}
	}

	return ""
}

// extractTitleFromContent 从内容中提取标题
func extractTitleFromContent(content string) string {
	if content == "" {
		return ""
	}

	lines := strings.Split(content, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		// 跳过空行和太短的行
		if len(line) < 5 {
			continue
		}

		// 查找可能的标题行
		if strings.HasPrefix(line, "#") ||
			strings.Contains(line, "标题") ||
			strings.Contains(line, "Title") ||
			(len(line) < 100 && !strings.Contains(line, "作者") &&
				!strings.Contains(line, "Abstract") && !strings.Contains(line, "摘要")) {

			// 清理标题
			title := strings.TrimPrefix(line, "#")
			title = strings.TrimPrefix(title, "##")
			title = strings.TrimPrefix(title, "###")
			title = strings.TrimSpace(title)

			// 移除标题标记
			if strings.Contains(title, "：") {
				parts := strings.Split(title, "：")
				if len(parts) > 1 {
					title = strings.TrimSpace(parts[1])
				}
			}
			if strings.Contains(title, ":") {
				parts := strings.Split(title, ":")
				if len(parts) > 1 {
					title = strings.TrimSpace(parts[1])
				}
			}

			if len(title) > 5 && len(title) < 200 {
				return title
			}
		}
	}
	return ""
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

// createSymlink 创建索引链接（已移除，符合简单原则）
// 注释：根据Linus原则，不必要的复杂性应该移除
// 软链接增加了维护成本，实际使用价值不高
func createSymlink(targetDir, folderName string) error {
	return nil
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

// RegenerateMissingMeta 重新生成缺失的meta.json文件
func RegenerateMissingMeta() error {
	resultsDir := "data/results"

	entries, err := os.ReadDir(resultsDir)
	if err != nil {
		return fmt.Errorf("读取结果目录失败: %w", err)
	}

	regeneratedCount := 0
	for _, entry := range entries {
		if !entry.IsDir() || entry.Name() == "latest" {
			continue
		}

		targetDir := filepath.Join(resultsDir, entry.Name())
		metaFile := filepath.Join(targetDir, "meta.json")
		pdfFile := filepath.Join(targetDir, "source.pdf")

		// 检查是否需要重新生成meta.json
		needsRegeneration := false
		reason := ""

		if _, err := os.Stat(metaFile); os.IsNotExist(err) {
			needsRegeneration = true
			reason = "meta.json文件不存在"
		} else {
			// 检查meta.json是否包含无效信息
			data, err := os.ReadFile(metaFile)
			if err == nil {
				var info ParsedFileInfo
				if json.Unmarshal(data, &info) == nil {
					if info.Authors == "解析中..." || info.Authors == "# Authors'contribution" {
						needsRegeneration = true
						reason = "包含无效的作者信息"
					} else if info.Title == "" || len(info.Title) < 3 {
						needsRegeneration = true
						reason = "标题无效或过短"
					}
				}
			}
		}

		if needsRegeneration {
			log.Printf("重新生成meta.json: %s (原因: %s)", entry.Name(), reason)

			// 查找PDF文件路径
			if _, err := os.Stat(pdfFile); os.IsNotExist(err) {
				// 如果source.pdf不存在，尝试从CSV记录中查找对应的信息
				if pdfPath := findPDFPathFromRecord(entry.Name()); pdfPath != "" {
					if err := generateMeta(targetDir, pdfPath, pdfPath); err != nil {
						log.Printf("重新生成meta.json失败 %s: %v", entry.Name(), err)
					} else {
						regeneratedCount++
						log.Printf("✅ 成功重新生成meta.json: %s", entry.Name())
					}
				} else {
					log.Printf("无法找到PDF路径: %s", entry.Name())
				}
			} else {
				if err := generateMeta(targetDir, pdfFile, pdfFile); err != nil {
					log.Printf("重新生成meta.json失败 %s: %v", entry.Name(), err)
				} else {
					regeneratedCount++
					log.Printf("✅ 成功重新生成meta.json: %s", entry.Name())
				}
			}
		}
	}

	if regeneratedCount > 0 {
		log.Printf("✅ 总共重新生成了 %d 个meta.json文件", regeneratedCount)
	} else {
		log.Printf("✅ 所有meta.json文件都正常，无需重新生成")
	}

	return nil
}

// findPDFPathFromRecord 从CSV记录中查找PDF路径（简化版，暂时不依赖GetParseRecords）
func findPDFPathFromRecord(folderName string) string {
	// 尝试根据文件夹名称推断PDF路径
	// 这里可以根据实际需要扩展逻辑

	// 如果文件夹名包含DOI信息，返回默认路径
	if strings.Contains(folderName, "10.") {
		return ""
	}

	// 对于其他情况，暂时返回空字符串
	// 在实际使用中，可以从配置或数据库中查找
	return ""
}

// CleanupRedundantZIPs 清理多余的ZIP文件
func CleanupRedundantZIPs() error {
	resultsDir := "data/results"

	entries, err := os.ReadDir(resultsDir)
	if err != nil {
		return fmt.Errorf("读取结果目录失败: %w", err)
	}

	cleanedCount := 0
	totalSize := int64(0)

	for _, entry := range entries {
		if !entry.IsDir() || entry.Name() == "latest" {
			continue
		}

		targetDir := filepath.Join(resultsDir, entry.Name())

		// 查找目录中的ZIP文件
		dirEntries, err := os.ReadDir(targetDir)
		if err != nil {
			continue
		}

		for _, dirEntry := range dirEntries {
			if !dirEntry.IsDir() && strings.HasSuffix(dirEntry.Name(), ".zip") {
				zipPath := filepath.Join(targetDir, dirEntry.Name())
				info, err := dirEntry.Info()
				if err == nil {
					totalSize += info.Size()
				}

				// 删除ZIP文件
				if err := os.Remove(zipPath); err != nil {
					log.Printf("删除ZIP文件失败: %s, 错误: %v", zipPath, err)
				} else {
					cleanedCount++
					log.Printf("已删除冗余ZIP文件: %s", zipPath)
				}
			}
		}
	}

	// 查找results根目录中的ZIP文件
	rootEntries, err := os.ReadDir(resultsDir)
	if err == nil {
		for _, entry := range rootEntries {
			if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".zip") {
				zipPath := filepath.Join(resultsDir, entry.Name())
				info, err := entry.Info()
				if err == nil {
					totalSize += info.Size()
				}

				if err := os.Remove(zipPath); err != nil {
					log.Printf("删除根目录ZIP文件失败: %s, 错误: %v", zipPath, err)
				} else {
					cleanedCount++
					log.Printf("已删除根目录冗余ZIP文件: %s", entry.Name())
				}
			}
		}
	}

	if cleanedCount > 0 {
		log.Printf("✅ 清理完成，删除了 %d 个冗余ZIP文件，释放空间: %.1f MB", cleanedCount, float64(totalSize)/1024/1024)
	} else {
		log.Printf("✅ 没有发现冗余的ZIP文件")
	}

	return nil
}
