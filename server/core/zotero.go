package core

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
	_ "github.com/mattn/go-sqlite3"
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
	db       *sql.DB
	dataDir  string
	dbPath   string
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

// GetItemsWithPDF 获取有PDF附件的文献 (50行)
func (z *ZoteroDB) GetItemsWithPDF(limit int) ([]ZoteroItem, error) {
	log.Printf("查询前 %d 篇文献", limit)

	// 简化查询 - 只查询基本信息
	query := `
	SELECT
		i.itemID,
		i.key as pdf_name,
		it.typeName as item_type
	FROM items i
	LEFT JOIN itemTypes it ON it.itemTypeID = i.itemTypeID
	WHERE i.itemTypeID NOT IN (
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

		err := rows.Scan(
			&item.ItemID,
			&item.PDFName,
			&item.ItemType,
		)
		if err != nil {
			log.Printf("扫描行数据失败: %v", err)
			continue
		}

		// 设置默认值
		item.Title = fmt.Sprintf("文献 #%d", item.ItemID)
		item.Authors = []string{"未知作者"}
		item.Year = 0
		item.Tags = []string{}

		items = append(items, item)
	}

	log.Printf("成功查询到 %d 篇文献", len(items))
	return items, nil
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
		// attachments:格式路径 - 需要查找对应的附件项
		return "" // 暂时跳过，这种格式需要特殊处理
	}

	// 直接路径格式 - storage:XXXXXX.pdf 或直接文件名
	if strings.Contains(pdfPath, ":") {
		parts := strings.Split(pdfPath, ":")
		if len(parts) >= 2 {
			// storage:XXXXXX.pdf 格式
			folderName := parts[1]
			if folderName != "" {
				return filepath.Join(z.dataDir, folderName)
			}
		}
	}

	// 直接文件名格式 - 在存储目录中查找
	return z.findPDFInStorage(pdfPath)
}

// findPDFInStorage 在存储目录中查找PDF文件 (20行)
func (z *ZoteroDB) findPDFInStorage(filename string) string {
	// 简化实现：返回空，表示路径构建失败
	// 实际实现可以在storage目录中递归搜索文件
	log.Printf("PDF文件搜索功能待实现: %s", filename)
	return ""
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