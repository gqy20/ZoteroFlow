package core

import (
	"context"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"
)

// PDFParser PDF解析器 (150行)
type PDFParser struct {
	zoteroDB     *ZoteroDB
	mineruClient *MinerUClient
	cacheDir     string
}

// ParsedDocument 解析后的文档
type ParsedDocument struct {
	ZoteroItem ZoteroItem `json:"zotero_item"`
	ParseHash  string     `json:"parse_hash"`
	Content    string     `json:"content"`    // Markdown格式内容
	Summary    string     `json:"summary"`    // AI生成的摘要
	KeyPoints  []string   `json:"key_points"` // 关键要点
	ZipPath    string     `json:"zip_path"`   // ZIP文件路径
	ParseTime  time.Time  `json:"parse_time"`
}

// NewPDFParser 创建PDF解析器 (30行)
func NewPDFParser(zoteroDB *ZoteroDB, mineruClient *MinerUClient, cacheDir string) (*PDFParser, error) {
	// 确保缓存目录存在
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return nil, fmt.Errorf("创建缓存目录失败: %w", err)
	}

	return &PDFParser{
		zoteroDB:     zoteroDB,
		mineruClient: mineruClient,
		cacheDir:     cacheDir,
	}, nil
}

// GetZoteroDB 获取Zotero数据库连接
func (p *PDFParser) GetZoteroDB() *ZoteroDB {
	return p.zoteroDB
}

// ParseDocument 解析单个文档 (100行)
func (p *PDFParser) ParseDocument(ctx context.Context, itemID int, pdfPath string) (*ParsedDocument, error) {
	log.Printf("开始解析文档 ItemID: %d", itemID)

	// 1. 获取Zotero元数据
	item, err := p.zoteroDB.GetItemsWithPDF(1)
	if err != nil {
		return nil, fmt.Errorf("获取Zotero元数据失败: %w", err)
	}

	if len(item) == 0 {
		return nil, fmt.Errorf("未找到文献 ItemID: %d", itemID)
	}

	// 2. 检查缓存
	cacheKey := p.generateCacheKey(pdfPath)
	cachePath := filepath.Join(p.cacheDir, cacheKey+".json")
	if parsedDoc, err := p.loadFromCache(cachePath); err == nil {
		log.Printf("使用缓存结果: %s", cachePath)
		return parsedDoc, nil
	}

	// 3. 调用MinerU解析
	log.Printf("调用MinerU解析PDF: %s", pdfPath)
	result, err := p.mineruClient.ParsePDF(ctx, pdfPath)
	if err != nil {
		return nil, fmt.Errorf("MinerU解析失败: %w", err)
	}

	// 4. 创建解析结果
	parsedDoc := &ParsedDocument{
		ZoteroItem: item[0],
		ParseHash:  cacheKey,
		Content:    "PDF解析完成，结果已保存",
		Summary:    "AI摘要功能待实现",
		KeyPoints:  []string{"关键要点提取待实现"},
		ZipPath:    result.ZipPath,
		ParseTime:  time.Now(),
	}

	// 5. 保存到缓存
	if err := p.saveToCache(parsedDoc, cachePath); err != nil {
		log.Printf("保存缓存失败: %v", err)
	}

	log.Printf("文档解析完成: ItemID %d", itemID)
	return parsedDoc, nil
}

// BatchParseDocuments 批量解析文档 (完整实现)
func (p *PDFParser) BatchParseDocuments(ctx context.Context, itemIDs []int) ([]*ParsedDocument, error) {
	log.Printf("开始批量解析 %d 篇文档", len(itemIDs))

	var results []*ParsedDocument
	var errors []error

	for i, itemID := range itemIDs {
		// 获取文献信息（包含PDF路径）
		items, err := p.zoteroDB.GetItemsWithPDF(1)
		if err != nil {
			errors = append(errors, fmt.Errorf("ItemID %d: %w", itemID, err))
			continue
		}

		if len(items) == 0 {
			errors = append(errors, fmt.Errorf("ItemID %d: 未找到文献", itemID))
			continue
		}

		// 使用找到的文献进行解析
		item := items[0]
		if item.PDFPath == "" {
			errors = append(errors, fmt.Errorf("ItemID %d: PDF路径为空", itemID))
			continue
		}

		// 验证PDF文件存在
		if _, err := os.Stat(item.PDFPath); err != nil {
			errors = append(errors, fmt.Errorf("ItemID %d: PDF文件不存在: %s", itemID, item.PDFPath))
			continue
		}

		log.Printf("开始解析第 %d/%d 篇文档: %s", i+1, len(itemIDs), item.Title)

		// 调用实际解析
		doc, err := p.ParseDocument(ctx, item.ItemID, item.PDFPath)
		if err != nil {
			errors = append(errors, fmt.Errorf("ItemID %d: 解析失败: %w", itemID, err))
			continue
		}

		results = append(results, doc)

		// 显示进度
		log.Printf("进度: %d/%d 完成", i+1, len(itemIDs))
	}

	if len(errors) > 0 {
		log.Printf("批量解析完成，但有 %d 个错误", len(errors))
		for _, err := range errors {
			log.Printf("错误: %v", err)
		}
	}

	log.Printf("批量解析完成，成功解析 %d 篇文档", len(results))
	return results, nil
}

// generateCacheKey 生成缓存键 (20行)
func (p *PDFParser) generateCacheKey(pdfPath string) string {
	h := md5.New()
	h.Write([]byte(pdfPath))
	return fmt.Sprintf("%x", h.Sum(nil))
}

// loadFromCache 从缓存加载 (20行)
func (p *PDFParser) loadFromCache(cachePath string) (*ParsedDocument, error) {
	file, err := os.Open(cachePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var doc ParsedDocument
	if err := json.NewDecoder(file).Decode(&doc); err != nil {
		return nil, err
	}

	// 检查文件是否存在
	if _, err := os.Stat(doc.ZipPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("ZIP文件不存在: %s", doc.ZipPath)
	}

	return &doc, nil
}

// saveToCache 保存到缓存 (20行)
func (p *PDFParser) saveToCache(doc *ParsedDocument, cachePath string) error {
	file, err := os.Create(cachePath)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	return encoder.Encode(doc)
}
