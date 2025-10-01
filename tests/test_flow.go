package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// === 第 1 部分：Zotero 数据读取 (30 行) ===

func testZotero(dbPath string) (int, string, error) {
	// 只读模式打开
	db, err := sql.Open("sqlite3", dbPath+"?mode=ro")
	if err != nil {
		return 0, "", fmt.Errorf("打开数据库失败: %v", err)
	}
	defer db.Close()

	// 查询第一篇文献
	var itemID int
	var title string
	query := `
		SELECT i.itemID, idv.value as title
		FROM items i
		JOIN itemData id ON i.itemID = id.itemID
		JOIN itemDataValues idv ON id.valueID = idv.valueID
		WHERE id.fieldID = 1
		LIMIT 1
	`
	err = db.QueryRow(query).Scan(&itemID, &title)
	if err != nil {
		return 0, "", fmt.Errorf("查询失败: %v", err)
	}

	fmt.Printf("✅ Zotero 读取成功\n")
	fmt.Printf("   - Item ID: %d\n", itemID)
	fmt.Printf("   - Title: %s\n", title)

	// 查询 PDF 路径
	var pdfPath string
	pdfQuery := `
		SELECT ia.path
		FROM itemAttachments ia
		WHERE ia.parentItemID = ? AND ia.contentType = 'application/pdf'
		LIMIT 1
	`
	err = db.QueryRow(pdfQuery, itemID).Scan(&pdfPath)
	if err != nil {
		return itemID, title, fmt.Errorf("未找到 PDF: %v", err)
	}

	// 转换 Zotero 相对路径为绝对路径
	// attachments:Q_生物科学/... -> /home/qy113/workspace/note/zo/articles/Q_生物科学/...
	basePath := os.Getenv("ARTICLE_PATH")
	if basePath == "" {
		basePath = "/home/qy113/workspace/note/zo/articles"
	}

	if len(pdfPath) > 12 && pdfPath[:12] == "attachments:" {
		pdfPath = filepath.Join(basePath, pdfPath[12:])
	}

	fmt.Printf("   - PDF: %s\n", pdfPath)

	// 检查文件是否存在
	if _, err := os.Stat(pdfPath); os.IsNotExist(err) {
		return itemID, pdfPath, fmt.Errorf("PDF 文件不存在: %s", pdfPath)
	}

	return itemID, pdfPath, nil
}

// === 第 2 部分：MinerU API 调用 (50 行) ===

type BatchResponse struct {
	Data struct {
		BatchID  string   `json:"batch_id"`
		FileURLs []string `json:"file_urls"`
	} `json:"data"`
}

type StatusResponse struct {
	Data struct {
		ExtractResult []struct {
			FileName   string `json:"file_name"`
			State      string `json:"state"`
			FullZipURL string `json:"full_zip_url,omitempty"`
		} `json:"extract_result"`
	} `json:"data"`
}

func testMinerU(token, pdfPath string) error {
	baseURL := "https://mineru.net/api/v4"
	fileName := filepath.Base(pdfPath)

	// 1. 提交任务
	fmt.Printf("\n✅ 步骤1: 提交文件任务\n")
	payload := map[string]interface{}{
		"language": "ch",
		"files": []map[string]interface{}{
			{"name": fileName, "is_ocr": true},
		},
	}
	jsonData, _ := json.Marshal(payload)

	req, _ := http.NewRequest("POST", baseURL+"/file-urls/batch", bytes.NewReader(jsonData))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("提交任务失败: %v", err)
	}
	defer resp.Body.Close()

	var batchResp BatchResponse
	json.NewDecoder(resp.Body).Decode(&batchResp)
	batchID := batchResp.Data.BatchID
	uploadURL := batchResp.Data.FileURLs[0]

	fmt.Printf("   - Batch ID: %s\n", batchID)
	fmt.Printf("   - Upload URL: %s...\n", uploadURL[:50])

	// 2. 上传文件
	fmt.Printf("\n✅ 步骤2: 上传 PDF 文件\n")
	file, err := os.Open(pdfPath)
	if err != nil {
		return fmt.Errorf("打开文件失败: %v", err)
	}
	defer file.Close()

	uploadReq, _ := http.NewRequest("PUT", uploadURL, file)
	// 注意：不设置 Content-Type
	uploadResp, err := client.Do(uploadReq)
	if err != nil {
		return fmt.Errorf("上传失败: %v", err)
	}
	uploadResp.Body.Close()

	if uploadResp.StatusCode != 200 {
		return fmt.Errorf("上传失败，状态码: %d", uploadResp.StatusCode)
	}
	fmt.Printf("   - 上传成功\n")

	// 3. 轮询状态
	fmt.Printf("\n✅ 步骤3: 轮询处理状态 (最多3分钟)\n")
	for i := 0; i < 18; i++ { // 3分钟
		time.Sleep(10 * time.Second)

		statusReq, _ := http.NewRequest("GET", baseURL+"/extract-results/batch/"+batchID, nil)
		statusReq.Header.Set("Authorization", "Bearer "+token)
		statusResp, _ := client.Do(statusReq)

		var status StatusResponse
		json.NewDecoder(statusResp.Body).Decode(&status)
		statusResp.Body.Close()

		if len(status.Data.ExtractResult) > 0 {
			result := status.Data.ExtractResult[0]
			fmt.Printf("   [%d秒] 状态: %s\n", (i+1)*10, result.State)

			if result.State == "done" {
				fmt.Printf("\n✅ 步骤4: 下载结果\n")
				fmt.Printf("   - Download URL: %s\n", result.FullZipURL)

				// 下载 ZIP
				dlReq, _ := http.NewRequest("GET", result.FullZipURL, nil)
				dlReq.Header.Set("Authorization", "Bearer "+token)
				dlResp, err := client.Do(dlReq)
				if err != nil {
					return fmt.Errorf("下载失败: %v", err)
				}
				defer dlResp.Body.Close()

				outFile, _ := os.Create("test_result.zip")
				io.Copy(outFile, dlResp.Body)
				outFile.Close()

				fmt.Printf("   ✅ 成功！结果保存到: test_result.zip\n")
				return nil
			} else if result.State == "failed" {
				return fmt.Errorf("处理失败")
			}
		}
	}

	return fmt.Errorf("超时（3分钟内未完成）")
}

// === 主函数 (20 行) ===

func main() {
	fmt.Println("=== ZoteroFlow 极简验证 ===")
	fmt.Println()

	// 从环境变量读取配置
	dbPath := os.Getenv("ZOTERO_DB_PATH")
	token := os.Getenv("MINERU_TOKEN")

	if dbPath == "" || token == "" {
		fmt.Println("❌ 请先设置环境变量:")
		fmt.Println("   export ZOTERO_DB_PATH=/path/to/zotero.sqlite")
		fmt.Println("   export MINERU_TOKEN=your_token")
		os.Exit(1)
	}

	// 测试 Zotero
	itemID, pdfPath, err := testZotero(dbPath)
	if err != nil {
		fmt.Printf("❌ Zotero 测试失败: %v\n", err)
		os.Exit(1)
	}

	// 测试 MinerU
	err = testMinerU(token, pdfPath)
	if err != nil {
		fmt.Printf("❌ MinerU 测试失败: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("\n🎉 验证成功！Item ID: %d 已完成解析\n", itemID)
}
