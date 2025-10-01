package core

import (
	"database/sql"
	"fmt"
	"log"
	"sort"
	"strings"
)

// SearchResult 搜索结果
type SearchResult struct {
	ItemID  int     `json:"item_id"`
	Title   string  `json:"title"`
	Authors string  `json:"authors"`
	DOI     string  `json:"doi"`
	Journal string  `json:"journal"`
	Year    string  `json:"year"`
	PDFPath string  `json:"pdf_path"`
	Score   float64 `json:"score"`
}

// SearchEngine 统一搜索引擎 - 极简版本
type SearchEngine struct {
	db *sql.DB
}

// NewSearchEngine 创建搜索引擎
func NewSearchEngine(db *sql.DB) *SearchEngine {
	return &SearchEngine{db: db}
}

// Search 主搜索 - 只搜索标题，速度最快
func (se *SearchEngine) Search(query string, limit int) ([]SearchResult, error) {
	log.Printf("搜索: %s", query)

	query = strings.TrimSpace(strings.ToLower(query))
	if query == "" {
		return nil, fmt.Errorf("搜索查询不能为空")
	}
	if limit <= 0 {
		limit = 20
	}

	// 极简SQL查询 - 只搜索标题
	sqlQuery := `
		SELECT
			i.itemID,
			COALESCE(idv.value, '') as title,
			ia.path as attachment_path
		FROM items i
		LEFT JOIN itemData id ON i.itemID = id.itemID
		LEFT JOIN fieldsCombined fc ON id.fieldID = fc.fieldID AND fc.fieldName = 'title'
		LEFT JOIN itemDataValues idv ON id.valueID = idv.valueID
		LEFT JOIN itemAttachments ia ON i.itemID = ia.parentItemID
		WHERE ia.contentType = 'application/pdf'
		AND i.itemTypeID NOT IN (SELECT itemTypeID FROM itemTypes WHERE typeName IN ('attachment', 'note', 'annotation'))
		ORDER BY i.dateAdded DESC
		LIMIT ?
	`

	rows, err := se.db.Query(sqlQuery, limit*3) // 获取更多数据用于筛选
	if err != nil {
		return nil, fmt.Errorf("数据库查询失败: %w", err)
	}
	defer rows.Close()

	var results []SearchResult
	count := 0

	for rows.Next() {
		var itemID int
		var title, attachmentPath string

		if err := rows.Scan(&itemID, &title, &attachmentPath); err != nil {
			continue
		}

		// 简单匹配
		if strings.Contains(strings.ToLower(title), query) {
			// 构建PDF路径
			pdfPath := se.buildPDFPath(attachmentPath)

			result := SearchResult{
				ItemID:  itemID,
				Title:   title,
				PDFPath: pdfPath,
				Score:   float64(len(query)) / float64(len(title)) * 100, // 简单评分
			}

			results = append(results, result)
			count++

			if count >= limit {
				break
			}
		}
	}

	// 按评分排序
	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})

	log.Printf("搜索完成，找到 %d 个结果", len(results))
	return results, nil
}

// buildPDFPath 构建PDF路径
func (se *SearchEngine) buildPDFPath(attachmentPath string) string {
	if attachmentPath == "" {
		return ""
	}

	if strings.HasPrefix(attachmentPath, "storage:") {
		key := strings.TrimPrefix(attachmentPath, "storage:")
		zdb := &ZoteroDB{db: se.db}
		return zdb.findPDFInStorage(key + ".pdf")
	}

	return attachmentPath
}

// SearchByTitle ZoteroDB的标题搜索方法
func (z *ZoteroDB) SearchByTitle(query string, limit int) ([]SearchResult, error) {
	engine := NewSearchEngine(z.db)
	return engine.Search(query, limit)
}
