package core

import (
	"bytes"
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// MinerUClient MinerU API客户端
type MinerUClient struct {
	BaseURL    string
	Token      string
	HTTPClient *http.Client
	MaxRetry   int
	Timeout    time.Duration
	ResultsDir string // 解析结果存储目录
}

// FileInfo 文件信息
type FileInfo struct {
	Name  string `json:"name"`
	IsOCR bool   `json:"is_ocr"`
}

// BatchRequest 批量提交请求
type BatchRequest struct {
	Language string     `json:"language"`
	Files    []FileInfo `json:"files"`
}

// BatchData 批量响应数据
type BatchData struct {
	BatchID  string   `json:"batch_id"`
	FileURLs []string `json:"file_urls"`
}

// BatchResponse 批量提交响应
type BatchResponse struct {
	Data BatchData `json:"data"`
}

// ExtractResult 提取结果
type ExtractResult struct {
	FileName   string `json:"file_name"`
	State      string `json:"state"`
	FullZipURL string `json:"full_zip_url,omitempty"`
}

// StatusData 状态查询数据
type StatusData struct {
	ExtractResult []ExtractResult `json:"extract_result"`
}

// StatusResponse 状态查询响应
type StatusResponse struct {
	Data StatusData `json:"data"`
}

// ParseResult 解析结果
type ParseResult struct {
	TaskID    string    `json:"task_id"`
	Status    string    `json:"status"`
	Content   string    `json:"content"`
	Message   string    `json:"message"`
	ErrorCode string    `json:"error_code"`
	ZipPath   string    `json:"zip_path"`
	ParseTime time.Time `json:"parse_time"`
	PDFPath   string    `json:"pdf_path"`
	FileName  string    `json:"file_name"`
	FileSize  int64     `json:"file_size"`
	Duration  int64     `json:"duration_ms"` // 解析耗时（毫秒）
}

// ParseRecord CSV记录结构
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

// NewMinerUClient 创建MinerU客户端
func NewMinerUClient(apiURL, token string) *MinerUClient {
	return NewMinerUClientWithResultsDir(apiURL, token, "data/results")
}

// NewMinerUClientWithResultsDir 创建MinerU客户端，指定结果目录
func NewMinerUClientWithResultsDir(apiURL, token, resultsDir string) *MinerUClient {
	return &MinerUClient{
		BaseURL:    apiURL,
		Token:      token,
		HTTPClient: &http.Client{Timeout: 120 * time.Second},
		MaxRetry:   3,
		Timeout:    3 * time.Minute,
		ResultsDir: resultsDir,
	}
}

// ParsePDF 解析PDF文件
func (c *MinerUClient) ParsePDF(ctx context.Context, pdfPath string) (*ParseResult, error) {
	startTime := time.Now()
	fileName := filepath.Base(pdfPath)

	// 获取文件信息
	fileInfo, err := os.Stat(pdfPath)
	if err != nil {
		return nil, fmt.Errorf("无法读取文件信息: %w", err)
	}
	fileSize := fileInfo.Size()

	log.Printf("Starting PDF parsing: %s (size: %d bytes)", fileName, fileSize)

	// 生成唯一ID
	recordID := fmt.Sprintf("%d_%s", time.Now().UnixNano(), fileName)

	// 1. 提交批量任务
	log.Printf("步骤1: 提交批量任务")
	batchResp, err := c.submitBatchTask(ctx, fileName)
	if err != nil {
		// 记录失败
		duration := time.Since(startTime).Milliseconds()
		recordErr := c.saveParseRecord(ParseRecord{
			ID:           recordID,
			TaskID:       "",
			FileName:     fileName,
			PDFPath:      pdfPath,
			FileSize:     fileSize,
			Status:       "failed",
			ZipPath:      "",
			ParseTime:    startTime,
			Duration:     duration,
			ErrorMessage: fmt.Sprintf("提交任务失败: %v", err),
		})
		if recordErr != nil {
			log.Printf("保存失败记录时出错: %v", recordErr)
		}
		return nil, fmt.Errorf("提交任务失败: %w", err)
	}

	batchID := batchResp.Data.BatchID
	uploadURL := batchResp.Data.FileURLs[0]
	log.Printf("任务ID: %s, 上传URL: %s...", batchID, uploadURL[:50])

	// 2. 上传文件
	log.Printf("步骤2: 上传PDF文件")
	if err := c.uploadFile(ctx, uploadURL, pdfPath); err != nil {
		// 记录失败
		duration := time.Since(startTime).Milliseconds()
		recordErr := c.saveParseRecord(ParseRecord{
			ID:           recordID,
			TaskID:       batchID,
			FileName:     fileName,
			PDFPath:      pdfPath,
			FileSize:     fileSize,
			Status:       "failed",
			ZipPath:      "",
			ParseTime:    startTime,
			Duration:     duration,
			ErrorMessage: fmt.Sprintf("上传文件失败: %v", err),
		})
		if recordErr != nil {
			log.Printf("保存失败记录时出错: %v", recordErr)
		}
		return nil, fmt.Errorf("上传文件失败: %w", err)
	}

	// 3. 轮询处理状态
	log.Printf("步骤3: 轮询处理状态")
	resultURL, err := c.pollStatus(ctx, batchID)
	if err != nil {
		// 记录失败
		duration := time.Since(startTime).Milliseconds()
		recordErr := c.saveParseRecord(ParseRecord{
			ID:           recordID,
			TaskID:       batchID,
			FileName:     fileName,
			PDFPath:      pdfPath,
			FileSize:     fileSize,
			Status:       "failed",
			ZipPath:      "",
			ParseTime:    startTime,
			Duration:     duration,
			ErrorMessage: fmt.Sprintf("处理失败: %v", err),
		})
		if recordErr != nil {
			log.Printf("保存失败记录时出错: %v", recordErr)
		}
		return nil, fmt.Errorf("处理失败: %w", err)
	}

	// 4. 下载结果到配置的结果目录
	log.Printf("步骤4: 下载解析结果")
	zipPath := filepath.Join(c.ResultsDir, fileName+".zip")
	if err := c.downloadResult(ctx, resultURL, zipPath); err != nil {
		// 记录失败
		duration := time.Since(startTime).Milliseconds()
		recordErr := c.saveParseRecord(ParseRecord{
			ID:           recordID,
			TaskID:       batchID,
			FileName:     fileName,
			PDFPath:      pdfPath,
			FileSize:     fileSize,
			Status:       "failed",
			ZipPath:      "",
			ParseTime:    startTime,
			Duration:     duration,
			ErrorMessage: fmt.Sprintf("下载结果失败: %v", err),
		})
		if recordErr != nil {
			log.Printf("保存失败记录时出错: %v", recordErr)
		}
		return nil, fmt.Errorf("下载结果失败: %w", err)
	}

	// 计算总耗时
	duration := time.Since(startTime).Milliseconds()

	// 创建成功结果
	result := &ParseResult{
		TaskID:    batchID,
		Status:    "completed",
		Content:   "解析完成，结果已保存到ZIP文件",
		ZipPath:   zipPath,
		ParseTime: startTime,
		PDFPath:   pdfPath,
		FileName:  fileName,
		FileSize:  fileSize,
		Duration:  duration,
	}

	// 保存成功记录
	if err := c.saveParseRecord(ParseRecord{
		ID:           recordID,
		TaskID:       batchID,
		FileName:     fileName,
		PDFPath:      pdfPath,
		FileSize:     fileSize,
		Status:       "completed",
		ZipPath:      zipPath,
		ParseTime:    startTime,
		Duration:     duration,
		ErrorMessage: "",
	}); err != nil {
		log.Printf("保存成功记录时出错: %v", err)
		// 不影响主流程
	}

	log.Printf("Parsing completed successfully! Duration: %dms, Result saved to: %s", duration, zipPath)

	// 同步组织文件，确保文件组织成功
	log.Printf("开始组织文件: %s", zipPath)
	if err := OrganizeResult(zipPath, pdfPath); err != nil {
		log.Printf("⚠️ 文件组织失败: %v", err)
		// 不影响主流程，但记录错误
	} else {
		log.Printf("✅ 文件组织完成")
	}

	return result, nil
}

// submitBatchTask 提交批量任务
func (c *MinerUClient) submitBatchTask(ctx context.Context, fileName string) (*BatchResponse, error) {
	payload := BatchRequest{
		Language: "ch",
		Files: []FileInfo{
			{Name: fileName, IsOCR: true},
		},
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("序列化请求失败: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.BaseURL+"/file-urls/batch", bytes.NewReader(jsonData))
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.Token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("发送请求失败: %w", err)
	}
	defer resp.Body.Close()

	var batchResp BatchResponse
	if err := json.NewDecoder(resp.Body).Decode(&batchResp); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	return &batchResp, nil
}

// uploadFile 上传文件
func (c *MinerUClient) uploadFile(ctx context.Context, uploadURL, filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("打开文件失败: %w", err)
	}
	defer file.Close()

	req, err := http.NewRequestWithContext(ctx, "PUT", uploadURL, file)
	if err != nil {
		return fmt.Errorf("创建上传请求失败: %w", err)
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("上传失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("上传失败，状态码: %d", resp.StatusCode)
	}

	return nil
}

// pollStatus 轮询处理状态
func (c *MinerUClient) pollStatus(ctx context.Context, batchID string) (string, error) {
	log.Printf("开始轮询处理状态，任务ID: %s", batchID)
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for i := 0; i < 18; i++ { // 3分钟超时
		log.Printf("轮询第 %d 次 (等待 %d 秒)", i+1, (i+1)*10)
		select {
		case <-ctx.Done():
			return "", ctx.Err()
		case <-ticker.C:
			statusResp, err := c.checkStatus(ctx, batchID)
			if err != nil {
				continue
			}

			if len(statusResp.Data.ExtractResult) > 0 {
				result := statusResp.Data.ExtractResult[0]
				log.Printf("[%ds] Status: %s", (i+1)*10, result.State)

				switch result.State {
				case "done":
					return result.FullZipURL, nil
				case "failed":
					return "", fmt.Errorf("processing failed")
				default:
					continue
				}
			}
		}
	}

	return "", fmt.Errorf("processing timeout (no completion within 3 minutes)")
}

// checkStatus 检查处理状态
func (c *MinerUClient) checkStatus(ctx context.Context, batchID string) (*StatusResponse, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", c.BaseURL+"/extract-results/batch/"+batchID, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+c.Token)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var statusResp StatusResponse
	if err := json.NewDecoder(resp.Body).Decode(&statusResp); err != nil {
		return nil, err
	}

	return &statusResp, nil
}

// downloadResult 下载结果
func (c *MinerUClient) downloadResult(ctx context.Context, resultURL, outputPath string) error {
	// 确保输出目录存在
	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		return fmt.Errorf("创建输出目录失败: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "GET", resultURL, nil)
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "Bearer "+c.Token)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	file, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = io.Copy(file, resp.Body)
	return err
}

// saveParseRecord 保存解析记录到对应的CSV文件
func (c *MinerUClient) saveParseRecord(record ParseRecord) error {
	recordsDir := "data/records"

	// 确保目录存在
	if err := os.MkdirAll(recordsDir, 0755); err != nil {
		return fmt.Errorf("创建记录目录失败: %w", err)
	}

	// 根据状态选择不同的CSV文件
	var csvPath string
	if record.Status == "failed" {
		csvPath = filepath.Join(recordsDir, "mineru_failed_records.csv")
	} else {
		csvPath = filepath.Join(recordsDir, "mineru_success_records.csv")
	}

	// 检查文件是否存在，如果不存在则创建并写入标题
	fileExists := true
	if _, err := os.Stat(csvPath); os.IsNotExist(err) {
		fileExists = false
	}

	// 打开CSV文件（追加模式）
	file, err := os.OpenFile(csvPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("打开CSV文件失败: %w", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// 如果文件不存在，写入标题行
	if !fileExists {
		headers := []string{
			"id", "task_id", "file_name", "pdf_path", "file_size",
			"status", "zip_path", "parse_time", "duration_ms", "error_message",
		}
		if err := writer.Write(headers); err != nil {
			return fmt.Errorf("写入CSV标题失败: %w", err)
		}
	}

	// 写入记录
	recordData := []string{
		record.ID,
		record.TaskID,
		record.FileName,
		record.PDFPath,
		strconv.FormatInt(record.FileSize, 10),
		record.Status,
		record.ZipPath,
		record.ParseTime.Format("2006-01-02 15:04:05"),
		strconv.FormatInt(record.Duration, 10),
		record.ErrorMessage,
	}

	if err := writer.Write(recordData); err != nil {
		return fmt.Errorf("写入CSV记录失败: %w", err)
	}

	log.Printf("Parse record saved to: %s", csvPath)
	return nil
}

// GetParseRecords 获取解析记录
func GetParseRecords(date string) ([]ParseRecord, error) {
	recordsDir := "data/records"

	var parseRecords []ParseRecord

	// 读取成功记录文件
	successCSV := filepath.Join(recordsDir, "mineru_success_records.csv")
	if records, err := readCSVFile(successCSV); err == nil {
		parseRecords = append(parseRecords, records...)
	}

	// 读取失败记录文件
	failedCSV := filepath.Join(recordsDir, "mineru_failed_records.csv")
	if records, err := readCSVFile(failedCSV); err == nil {
		parseRecords = append(parseRecords, records...)
	}

	// 如果两个文件都不存在，尝试读取旧文件（向后兼容）
	legacyCSV := filepath.Join(recordsDir, "mineru_parse_records.csv")
	if len(parseRecords) == 0 {
		if records, err := readCSVFile(legacyCSV); err == nil {
			parseRecords = append(parseRecords, records...)
		}
	}

	return parseRecords, nil
}

// readCSVFile 读取CSV文件并返回记录列表
func readCSVFile(csvPath string) ([]ParseRecord, error) {
	file, err := os.Open(csvPath)
	if err != nil {
		return nil, fmt.Errorf("打开CSV文件失败: %w", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("读取CSV文件失败: %w", err)
	}

	var parseRecords []ParseRecord

	// 跳过标题行
	for i, record := range records {
		if i == 0 {
			continue // 跳过标题
		}

		if len(record) < 10 {
			continue // 跳过格式不正确的行
		}

		parseTime, err := time.Parse("2006-01-02 15:04:05", record[7])
		if err != nil {
			continue // 跳过时间格式错误的行
		}

		fileSize, _ := strconv.ParseInt(record[4], 10, 64)
		duration, _ := strconv.ParseInt(record[8], 10, 64)

		parseRecord := ParseRecord{
			ID:           record[0],
			TaskID:       record[1],
			FileName:     record[2],
			PDFPath:      record[3],
			FileSize:     fileSize,
			Status:       record[5],
			ZipPath:      record[6],
			ParseTime:    parseTime,
			Duration:     duration,
			ErrorMessage: record[9],
		}

		parseRecords = append(parseRecords, parseRecord)
	}

	return parseRecords, nil
}

// ValidateAndRebuildRecords 验证并重建记录，删除不存在的文件条目
func ValidateAndRebuildRecords() error {
	recordsDir := "data/records"

	log.Printf("开始验证解析记录与实际文件的对应关系...")

	// 读取现有记录
	records, err := GetParseRecords("")
	if err != nil {
		return fmt.Errorf("读取现有记录失败: %w", err)
	}

	// 分离有效记录
	var validSuccessRecords []ParseRecord
	var validFailedRecords []ParseRecord
	removedCount := 0

	for _, record := range records {
		isValid := false

		if record.Status == "completed" && record.ZipPath != "" {
			// 检查ZIP文件是否存在
			if _, err := os.Stat(record.ZipPath); err == nil {
				isValid = true
			} else {
				// ZIP文件不存在，检查是否已经组织到results目录
				resultDir := filepath.Join("data/results", extractFolderNameFromZipPath(record.ZipPath))
				if _, err := os.Stat(resultDir); err == nil {
					// 更新路径为组织后的目录
					record.ZipPath = resultDir
					isValid = true
				}
			}
		} else if record.Status == "failed" {
			// 失败的记录保留，但可以添加额外的验证逻辑
			isValid = true
		}

		if isValid {
			if record.Status == "completed" {
				validSuccessRecords = append(validSuccessRecords, record)
			} else {
				validFailedRecords = append(validFailedRecords, record)
			}
		} else {
			removedCount++
			log.Printf("移除无效记录: %s (文件不存在)", record.FileName)
		}
	}

	if removedCount > 0 {
		log.Printf("发现 %d 条无效记录，正在重建CSV文件...", removedCount)

		// 重建成功记录文件
		successCSV := filepath.Join(recordsDir, "mineru_success_records.csv")
		if err := rebuildCSVFile(successCSV, validSuccessRecords); err != nil {
			log.Printf("重建成功记录文件失败: %v", err)
		} else {
			log.Printf("成功记录文件重建完成，保留 %d 条记录", len(validSuccessRecords))
		}

		// 重建失败记录文件
		failedCSV := filepath.Join(recordsDir, "mineru_failed_records.csv")
		if err := rebuildCSVFile(failedCSV, validFailedRecords); err != nil {
			log.Printf("重建失败记录文件失败: %v", err)
		} else {
			log.Printf("失败记录文件重建完成，保留 %d 条记录", len(validFailedRecords))
		}

		// 备份并删除旧的统一CSV文件
		legacyCSV := filepath.Join(recordsDir, "mineru_parse_records.csv")
		if _, err := os.Stat(legacyCSV); err == nil {
			backupPath := legacyCSV + ".backup"
			if err := copyFile(legacyCSV, backupPath); err != nil {
				log.Printf("备份旧CSV文件失败: %v", err)
			} else {
				log.Printf("旧CSV文件已备份到: %s", backupPath)
				os.Remove(legacyCSV)
			}
		}

		log.Printf("CSV文件重建完成，总共保留 %d 条有效记录", len(validSuccessRecords)+len(validFailedRecords))
	} else {
		log.Printf("所有记录都有效，无需重建")
	}

	return nil
}

// extractFolderNameFromZipPath 从ZIP路径提取文件夹名称
func extractFolderNameFromZipPath(zipPath string) string {
	// 从路径中提取文件名（不带扩展名）
	filename := filepath.Base(zipPath)
	folderName := strings.TrimSuffix(filename, filepath.Ext(filename))
	return folderName
}

// rebuildCSVFile 重建CSV文件
func rebuildCSVFile(csvPath string, records []ParseRecord) error {
	file, err := os.Create(csvPath)
	if err != nil {
		return fmt.Errorf("创建CSV文件失败: %w", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// 写入标题行
	headers := []string{
		"id", "task_id", "file_name", "pdf_path", "file_size",
		"status", "zip_path", "parse_time", "duration_ms", "error_message",
	}
	if err := writer.Write(headers); err != nil {
		return fmt.Errorf("写入CSV标题失败: %w", err)
	}

	// 写入所有有效记录
	for _, record := range records {
		recordData := []string{
			record.ID,
			record.TaskID,
			record.FileName,
			record.PDFPath,
			strconv.FormatInt(record.FileSize, 10),
			record.Status,
			record.ZipPath,
			record.ParseTime.Format("2006-01-02 15:04:05"),
			strconv.FormatInt(record.Duration, 10),
			record.ErrorMessage,
		}

		if err := writer.Write(recordData); err != nil {
			return fmt.Errorf("写入CSV记录失败: %w", err)
		}
	}

	return nil
}
