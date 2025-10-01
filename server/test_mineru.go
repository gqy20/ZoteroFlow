//go:build test
// +build test

package main

import (
	"context"
	"log"
	"os"
	"time"

	"zoteroflow2-server/core"
)

func main() {
	log.Println("=== MinerU集成测试 ===")

	// 从环境变量读取配置
	apiURL := os.Getenv("MINERU_API_URL")
	token := os.Getenv("MINERU_TOKEN")

	if apiURL == "" {
		apiURL = "https://api.mineru.io"
	}

	if token == "" {
		log.Println("❌ MINERU_TOKEN 环境变量未设置")
		return
	}

	log.Printf("MinerU API URL: %s", apiURL)
	log.Printf("Token: %s...***", token[:10])

	// 创建MinerU客户端
	client := core.NewMinerUClient(apiURL, token)
	log.Println("✅ MinerU客户端创建成功")

	// 测试CSV记录功能
	testCSVRecords(client)
}

func testCSVRecords(client *core.MinerUClient) {
	log.Println("\n=== 测试CSV记录功能 ===")

	// 获取今天的解析记录
	today := time.Now().Format("2006-01-02")
	records, err := core.GetParseRecords(today)
	if err != nil {
		log.Printf("获取今天的记录失败: %v", err)
	} else {
		log.Printf("✅ 找到 %d 条今天的解析记录", len(records))
		for i, record := range records {
			log.Printf("  [%d] %s - %s (状态: %s)", i+1, record.FileName, record.ParseTime.Format("15:04:05"), record.Status)
		}
	}

	// 测试解析功能（如果有PDF文件）
	pdfPath := findTestPDF()
	if pdfPath != "" {
		log.Printf("找到测试PDF: %s", pdfPath)

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
		defer cancel()

		log.Println("开始MinerU解析测试...")
		result, err := client.ParsePDF(ctx, pdfPath)
		if err != nil {
			log.Printf("❌ MinerU解析失败: %v", err)
			return
		}

		log.Printf("✅ MinerU解析成功!")
		log.Printf("  任务ID: %s", result.TaskID)
		log.Printf("  状态: %s", result.Status)
		log.Printf("  ZIP路径: %s", result.ZipPath)
		log.Printf("  文件大小: %d bytes", result.FileSize)
		log.Printf("  解析耗时: %d ms", result.Duration)

		// 检查结果文件
		if result.ZipPath != "" {
			if _, err := os.Stat(result.ZipPath); err == nil {
				log.Printf("✅ 结果文件存在: %s", result.ZipPath)
			} else {
				log.Printf("❌ 结果文件不存在: %s", result.ZipPath)
			}
		}
	} else {
		log.Println("⚠️  未找到测试PDF文件")
		log.Println("如需测试解析功能，请设置TEST_PDF_PATH环境变量")
	}
}

func findTestPDF() string {
	// 优先查找环境变量指定的PDF
	if pdfPath := os.Getenv("TEST_PDF_PATH"); pdfPath != "" {
		if _, err := os.Stat(pdfPath); err == nil {
			return pdfPath
		}
	}

	// 在当前目录查找PDF文件
	if file, err := os.Open("test.pdf"); err == nil {
		file.Close()
		return "test.pdf"
	}

	// 在data目录查找PDF文件
	if file, err := os.Open("data/test.pdf"); err == nil {
		file.Close()
		return "data/test.pdf"
	}

	return ""
}
