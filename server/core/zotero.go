package core

import (
	"database/sql"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// ZoteroItem 简化的Zotero文献项结构 (30行)
type ZoteroItem struct {
	ItemID   int      `json:"item_id"`
	Title    string   `json:"title"`
	Authors  []string `json:"authors"`
	Year     int      `json:"year"`
	ItemType string   `json:"item_type"`
	Tags     []string `json:"tags"`
	PDFPath  string   `json:"pdf_path"`
	PDFName  string   `json:"pdf_name"`
}

// ZoteroDB Zotero数据库访问器
type ZoteroDB struct {
	db      *sql.DB
	dataDir string
	dbPath  string
}

// NewZoteroDB 连接Zotero数据库 (30行)
func NewZoteroDB(dbPath, dataDir string) (*ZoteroDB, error) {
	log.Printf("连接Zotero数据库: %s", dbPath)

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("连接数据库失败: %w", err)
	}

	// 设置只读模式和连接池
	db.SetMaxOpenConns(1) // 只读访问，一个连接足够
	db.SetMaxIdleConns(1)
	db.SetConnMaxLifetime(time.Hour)

	// 设置只读模式，避免锁定问题
	_, err = db.Exec("PRAGMA query_only = 1; PRAGMA journal_mode = WAL;")
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("设置只读模式失败: %w", err)
	}

	// 测试连接
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("数据库连接测试失败: %w", err)
	}

	log.Printf("成功连接到Zotero数据库")
	return &ZoteroDB{db: db, dataDir: dataDir, dbPath: dbPath}, nil
}

// Close 关闭数据库连接
func (z *ZoteroDB) Close() error {
	if z.db != nil {
		return z.db.Close()
	}
	return nil
}

// GetItemsWithPDF 获取有PDF附件的文献 (完整实现)
func (z *ZoteroDB) GetItemsWithPDF(limit int) ([]ZoteroItem, error) {
	log.Printf("查询前 %d 篇有PDF附件的文献", limit)

	// 完整查询 - 获取文献信息和PDF附件路径
	query := `
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
	`

	rows, err := z.db.Query(query, limit)
	if err != nil {
		return nil, fmt.Errorf("查询失败: %w", err)
	}
	defer rows.Close()

	var items []ZoteroItem
	for rows.Next() {
		var item ZoteroItem
		var attachmentPath, contentType string

		err := rows.Scan(
			&item.ItemID,
			&item.Title,
			&item.ItemType,
			&attachmentPath,
			&contentType,
		)
		if err != nil {
			log.Printf("扫描行数据失败: %v", err)
			continue
		}

		// 如果没有标题，使用默认值
		if item.Title == "" {
			item.Title = fmt.Sprintf("文献 #%d", item.ItemID)
		}

		// 设置默认值
		item.Authors = []string{"未知作者"}
		item.Year = 0
		item.Tags = []string{}

		// 构建PDF路径
		if attachmentPath != "" {
			item.PDFPath = z.buildPDFPath(attachmentPath)
			// 从路径中提取文件名
			item.PDFName = z.extractFilenameFromPath(attachmentPath)
		} else {
			log.Printf("ItemID %d 没有附件路径信息", item.ItemID)
			continue // 跳过没有PDF路径的记录
		}

		// 只有成功找到PDF路径才添加到结果中
		if item.PDFPath != "" {
			// 验证文件是否存在
			if _, err := os.Stat(item.PDFPath); err == nil {
				items = append(items, item)
				log.Printf("找到PDF文献: ID=%d, 标题=%s, 路径=%s",
					item.ItemID, item.Title, item.PDFPath)
			} else {
				log.Printf("PDF文件不存在: %s (ItemID: %d)", item.PDFPath, item.ItemID)
			}
		} else {
			log.Printf("无法构建PDF路径: ItemID=%d, attachment=%s",
				item.ItemID, attachmentPath)
		}
	}

	// 如果没有找到文献，尝试备用查询方法
	if len(items) == 0 {
		log.Printf("主查询未找到文献，尝试备用查询方法")
		return z.getItemsWithPDFFallback(limit)
	}

	log.Printf("成功查询到 %d 篇有PDF附件的文献", len(items))
	return items, nil
}

// getItemsWithPDFFallback 备用查询方法
func (z *ZoteroDB) getItemsWithPDFFallback(limit int) ([]ZoteroItem, error) {
	log.Printf("使用备用方法查询文献")

	// 简化的查询，尝试不同的数据库结构
	query := `
	SELECT
		i.itemID,
		i.key as item_key,
		'journalArticle' as item_type
	FROM items i
	WHERE i.itemID NOT IN (
		SELECT itemTypeID FROM itemTypes
		WHERE typeName IN ('attachment', 'note', 'annotation')
	)
	ORDER BY i.dateAdded DESC
	LIMIT ?
	`

	rows, err := z.db.Query(query, limit)
	if err != nil {
		return nil, fmt.Errorf("备用查询失败: %w", err)
	}
	defer rows.Close()

	var items []ZoteroItem
	for rows.Next() {
		var item ZoteroItem
		var itemKey string

		err := rows.Scan(&item.ItemID, &itemKey, &item.ItemType)
		if err != nil {
			log.Printf("扫描备用查询行数据失败: %v", err)
			continue
		}

		// 设置基本信息
		item.Title = fmt.Sprintf("文献 #%d", item.ItemID)
		item.Authors = []string{"未知作者"}
		item.Year = 0
		item.Tags = []string{}
		item.PDFName = itemKey + ".pdf"

		// 尝试直接在存储目录中搜索
		pdfPath := z.findPDFInStorage(item.PDFName)
		if pdfPath != "" {
			item.PDFPath = pdfPath
			items = append(items, item)
			log.Printf("备用方法找到PDF: ID=%d, 路径=%s", item.ItemID, pdfPath)
		}
	}

	log.Printf("备用方法找到 %d 篇文献", len(items))
	return items, nil
}

// extractFilenameFromPath 从路径中提取文件名
func (z *ZoteroDB) extractFilenameFromPath(path string) string {
	// 处理不同格式的路径
	if strings.Contains(path, ":") {
		parts := strings.Split(path, ":")
		if len(parts) >= 2 {
			filename := parts[len(parts)-1]
			if !strings.HasSuffix(strings.ToLower(filename), ".pdf") {
				filename += ".pdf"
			}
			return filename
		}
	}

	// 直接从路径中提取文件名
	filename := filepath.Base(path)
	if !strings.HasSuffix(strings.ToLower(filename), ".pdf") {
		filename += ".pdf"
	}

	return filename
}

// getItemTags 获取文献的标签 (40行)
func (z *ZoteroDB) getItemTags(itemID int) ([]string, error) {
	query := `
	SELECT t.name
	FROM itemTags it
	JOIN tags t ON it.tagID = t.tagID
	WHERE it.itemID = ?
	ORDER BY t.name
	`

	rows, err := z.db.Query(query, itemID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tags []string
	for rows.Next() {
		var tag string
		if err := rows.Scan(&tag); err != nil {
			continue
		}
		tags = append(tags, tag)
	}

	return tags, nil
}

// buildPDFPath 构建PDF文件路径 (30行)
func (z *ZoteroDB) buildPDFPath(pdfPath string) string {
	// 处理两种路径格式:
	// 1. attachments:分类:年份_标题.pdf
	// 2. 附件项的直接路径 (storage:XXXXXX.pdf)

	if strings.HasPrefix(pdfPath, "attachments:") {
		// attachments:格式路径 - 解析并查找实际文件
		return z.parseAttachmentsPath(pdfPath)
	}

	// 处理路径分隔符 - 统一为系统路径分隔符
	pdfPath = filepath.FromSlash(pdfPath)

	// 直接路径格式 - storage:XXXXXX.pdf 或直接文件名
	if strings.Contains(pdfPath, ":") {
		parts := strings.Split(pdfPath, ":")
		if len(parts) >= 2 {
			// storage:XXXXXX.pdf 格式
			folderName := parts[1]
			if folderName != "" {
				storagePath := filepath.Join(z.dataDir, folderName)

				// 首先尝试直接路径
				if _, err := os.Stat(storagePath); err == nil {
					return storagePath
				}

				// 如果是文件夹，尝试查找其中的PDF文件
				if stat, err := os.Stat(storagePath); err == nil && stat.IsDir() {
					return z.findPDFInDirectory(storagePath)
				}
			}
		}
	}

	// 直接文件名格式 - 在存储目录中查找
	return z.findPDFInStorage(pdfPath)
}

// findPDFInStorage 在存储目录中查找PDF文件 (完整实现)
func (z *ZoteroDB) findPDFInStorage(filename string) string {
	if filename == "" {
		return ""
	}

	// 1. 标准化文件名 - 移除特殊字符
	cleanFilename := z.normalizeFilename(filename)

	// 2. 检查storage目录是否存在
	storageDir := z.dataDir
	if _, err := os.Stat(storageDir); os.IsNotExist(err) {
		log.Printf("Storage目录不存在: %s", storageDir)
		return ""
	}

	// 3. 递归搜索PDF文件
	var foundPath string
	err := filepath.Walk(storageDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // 忽略访问错误，继续搜索
		}

		// 只检查PDF文件
		if !strings.EqualFold(filepath.Ext(path), ".pdf") {
			return nil
		}

		// 获取文件名进行比较
		baseName := filepath.Base(path)
		cleanBaseName := z.normalizeFilename(baseName)

		// 匹配策略：
		// 1. 完全匹配
		// 2. 去掉扩展名后匹配
		// 3. 模糊匹配（包含关系）
		if baseName == filename ||
			cleanBaseName == cleanFilename ||
			strings.Contains(cleanBaseName, cleanFilename) ||
			strings.Contains(cleanFilename, cleanBaseName) {
			foundPath = path
			return filepath.SkipAll // 找到后停止搜索
		}

		return nil
	})

	if err != nil {
		log.Printf("搜索PDF文件时出错: %v", err)
		return ""
	}

	if foundPath != "" {
		log.Printf("找到PDF文件: %s -> %s", filename, foundPath)
	} else {
		log.Printf("未找到PDF文件: %s", filename)
	}

	return foundPath
}

// findPDFInDirectory 在指定目录中查找PDF文件
func (z *ZoteroDB) findPDFInDirectory(directory string) string {
	// 在指定目录中查找PDF文件
	entries, err := os.ReadDir(directory)
	if err != nil {
		log.Printf("无法读取目录 %s: %v", directory, err)
		return ""
	}

	// 优先查找非 _.pdf 的文件（原始文件名）
	var originalFile string
	var fallbackFile string

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		filename := entry.Name()
		if strings.EqualFold(filepath.Ext(filename), ".pdf") {
			if filename == "_.pdf" {
				fallbackFile = filepath.Join(directory, filename)
			} else {
				originalFile = filepath.Join(directory, filename)
				break // 找到原始文件名，立即返回
			}
		}
	}

	// 优先返回原始文件名
	if originalFile != "" {
		log.Printf("找到原始PDF文件: %s", originalFile)
		return originalFile
	}

	if fallbackFile != "" {
		log.Printf("找到备用PDF文件: %s", fallbackFile)
		return fallbackFile
	}

	log.Printf("目录中未找到PDF文件: %s", directory)
	return ""
}

// normalizeFilename 标准化文件名，用于比较
func (z *ZoteroDB) normalizeFilename(filename string) string {
	// 移除扩展名
	name := strings.TrimSuffix(filename, filepath.Ext(filename))

	// 转换为小写
	name = strings.ToLower(name)

	// 移除常见特殊字符和空格
	replacements := []struct {
		old string
		new string
	}{
		{" ", "_"},
		{"-", "_"},
		{".", "_"},
		{",", "_"},
		{";", "_"},
		{":", "_"},
		{"'", ""},
		{"\"", ""},
		{"(", ""},
		{")", ""},
		{"[", ""},
		{"]", ""},
		{"{", ""},
		{"}", ""},
		{"__", "_"}, // 合并多个下划线
	}

	for _, rep := range replacements {
		name = strings.ReplaceAll(name, rep.old, rep.new)
	}

	// 移除开头和结尾的下划线
	name = strings.Trim(name, "_")

	return name
}

// parseAuthors 解析作者字符串 (20行)
func parseAuthors(authorsStr string) []string {
	if authorsStr == "" {
		return []string{"未知作者"}
	}

	authors := strings.Split(authorsStr, ";")
	var result []string
	for _, author := range authors {
		author = strings.TrimSpace(author)
		if author != "" {
			result = append(result, author)
		}
	}

	if len(result) == 0 {
		return []string{"未知作者"}
	}
	return result
}

// GetStats 获取数据库统计信息
func (z *ZoteroDB) GetStats() (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// 总文献数
	var totalItems int
	z.db.QueryRow("SELECT COUNT(*) FROM items").Scan(&totalItems)
	stats["total_items"] = totalItems

	// 有PDF附件的文献数
	var pdfItems int
	z.db.QueryRow(`
		SELECT COUNT(DISTINCT parentItemID)
		FROM itemAttachments
		WHERE contentType = 'application/pdf'
	`).Scan(&pdfItems)
	stats["pdf_items"] = pdfItems

	// 数据库文件大小
	if file, err := os.Stat(z.dbPath); err == nil {
		stats["db_size_mb"] = file.Size() / 1024 / 1024
	}

	return stats, nil
}

// parseAttachmentsPath 解析 attachments: 格式路径
func (z *ZoteroDB) parseAttachmentsPath(attachmentsPath string) string {
	// 移除前缀 "attachments:"
	pathPart := strings.TrimPrefix(attachmentsPath, "attachments:")
	if pathPart == "" {
		log.Printf("attachments 路径格式无效: %s", attachmentsPath)
		return ""
	}

	log.Printf("解析 attachments 路径: %s", pathPart)

	// 尝试通过数据库查询实际的附件路径
	// attachments:格式通常需要查询 itemAttachments 表
	query := `
		SELECT ia.path, ii.key as storage_key
		FROM itemAttachments ia
		LEFT JOIN items ii ON ia.itemID = ii.itemID
		WHERE ia.path LIKE ? AND ia.contentType = 'application/pdf'
		LIMIT 1
	`

	// 构建搜索模式 - 支持模糊匹配
	searchPattern := "%" + pathPart + "%"

	var attachmentPath, storageKey string
	err := z.db.QueryRow(query, searchPattern).Scan(&attachmentPath, &storageKey)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Printf("未找到匹配的附件记录: %s", pathPart)
		} else {
			log.Printf("查询附件记录时出错: %v", err)
		}

		// 如果数据库查询失败，尝试从路径中提取文件名进行搜索
		filename := z.extractFilenameFromAttachmentsPath(pathPart)
		if filename != "" {
			return z.findPDFInStorage(filename)
		}

		return ""
	}

	// 处理查询到的路径
	if attachmentPath != "" {
		// 如果路径以 storage: 开头，解析存储路径
		if strings.HasPrefix(attachmentPath, "storage:") {
			parts := strings.Split(attachmentPath, ":")
			if len(parts) >= 2 {
				storageFolder := parts[1]
				storagePath := filepath.Join(z.dataDir, storageFolder)

				// 检查存储路径是否存在
				if _, err := os.Stat(storagePath); err == nil {
					log.Printf("通过storage路径找到附件: %s", storagePath)
					return storagePath
				}
			}
		}

		// 直接使用查询到的路径
		if filepath.IsAbs(attachmentPath) {
			if _, err := os.Stat(attachmentPath); err == nil {
				log.Printf("通过绝对路径找到附件: %s", attachmentPath)
				return attachmentPath
			}
		}
	}

	// 如果通过 storageKey 也能定位
	if storageKey != "" {
		storagePath := filepath.Join(z.dataDir, storageKey)
		if _, err := os.Stat(storagePath); err == nil {
			log.Printf("通过storage key找到附件: %s", storagePath)
			return storagePath
		}
	}

	// 最后尝试文件名搜索
	filename := z.extractFilenameFromAttachmentsPath(pathPart)
	if filename != "" {
		return z.findPDFInStorage(filename)
	}

	log.Printf("无法解析 attachments 路径: %s", attachmentsPath)
	return ""
}

// extractFilenameFromAttachmentsPath 从 attachments 路径中提取文件名
func (z *ZoteroDB) extractFilenameFromAttachmentsPath(pathPart string) string {
	// pathPart 可能的格式:
	// 1. 分类:年份_标题.pdf
	// 2. 年份_标题.pdf
	// 3. 标题.pdf
	// 4. 直接文件名

	parts := strings.Split(pathPart, ":")
	var filename string

	if len(parts) >= 2 {
		// 取最后一部分作为文件名
		filename = parts[len(parts)-1]
	} else {
		filename = pathPart
	}

	// 确保有 .pdf 扩展名
	if !strings.HasSuffix(strings.ToLower(filename), ".pdf") {
		filename += ".pdf"
	}

	log.Printf("从 attachments 路径提取文件名: %s -> %s", pathPart, filename)
	return filename
}
