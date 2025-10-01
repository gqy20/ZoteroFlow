package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"zoteroflow2-server/config"
	"zoteroflow2-server/core"
)

func main() {
	if len(os.Args) > 1 {
		// 处理CLI命令
		handleCommand(os.Args[1:])
		return
	}

	// 默认行为：运行基础测试
	runBasicTest()
}

// handleCommand 处理CLI命令
func handleCommand(args []string) {
	switch args[0] {
	case "list":
		listResults()
	case "open":
		if len(args) < 2 {
			log.Fatal("用法: open <文献名称>")
		}
		openResult(args[1])
	case "search":
		if len(args) < 2 {
			log.Fatal("用法: search <标题关键词>")
		}
		searchAndParse(strings.Join(args[1:], " "), "title")
	case "doi":
		if len(args) < 2 {
			log.Fatal("用法: doi <DOI号>")
		}
		searchAndParse(args[1], "doi")
	case "clean":
		cleanResults()
	case "help":
		showHelp()
	default:
		fmt.Printf("未知命令: %s\n", args[0])
		showHelp()
	}
}

// showHelp 显示帮助信息
func showHelp() {
	fmt.Println("ZoteroFlow2 - PDF文献管理工具")
	fmt.Println()
	fmt.Println("命令:")
	fmt.Println("  list                    - 列出所有解析结果")
	fmt.Println("  open <名称>             - 打开指定文献文件夹")
	fmt.Println("  search <关键词>         - 按标题搜索并解析文献")
	fmt.Println("  doi <DOI号>             - 按DOI搜索并解析文献")
	fmt.Println("  clean                   - 清理重复/损坏文件")
	fmt.Println("  help                    - 显示此帮助信息")
	fmt.Println()
	fmt.Println("示例:")
	fmt.Println("  ./zoteroflow2 list")
	fmt.Println("  ./zoteroflow2 open 丛枝菌根")
	fmt.Println("  ./zoteroflow2 search \"solanum chromosome\"")
	fmt.Println("  ./zoteroflow2 doi \"10.1111/j.1469-8137.2012.04195.x\"")
}

// listResults 列出所有解析结果
func listResults() {
	resultsDir := "data/results"

	entries, err := os.ReadDir(resultsDir)
	if err != nil {
		log.Printf("读取结果目录失败: %v", err)
		return
	}

	fmt.Printf("找到 %d 个解析结果:\n\n", len(entries))

	for i, entry := range entries {
		if entry.IsDir() && entry.Name() != "latest" {
			metaFile := filepath.Join(resultsDir, entry.Name(), "meta.json")
			if info := readMeta(metaFile); info != nil {
				fmt.Printf("[%d] %s\n", i+1, entry.Name())
				fmt.Printf("     标题: %s\n", info.Title)
				fmt.Printf("     作者: %s\n", info.Authors)
				fmt.Printf("     大小: %.1f MB\n", float64(info.Size)/1024/1024)
				fmt.Printf("     日期: %s\n", info.Date)
				fmt.Println()
			}
		}
	}
}

// openResult 打开指定文献
func openResult(name string) {
	resultsDir := "data/results"

	entries, err := os.ReadDir(resultsDir)
	if err != nil {
		log.Printf("读取结果目录失败: %v", err)
		return
	}

	for _, entry := range entries {
		if entry.IsDir() && entry.Name() != "latest" {
			if strings.Contains(entry.Name(), name) {
				folderPath := filepath.Join(resultsDir, entry.Name())
				fmt.Printf("打开文献文件夹: %s\n", folderPath)
				fmt.Printf("文件内容:\n")

				listFiles(folderPath)
				return
			}
		}
	}

	fmt.Printf("未找到包含 '%s' 的文献\n", name)
}

// listFiles 列出文件夹内容
func listFiles(folderPath string) {
	entries, err := os.ReadDir(folderPath)
	if err != nil {
		return
	}

	for _, entry := range entries {
		if entry.IsDir() {
			subEntries, _ := os.ReadDir(filepath.Join(folderPath, entry.Name()))
			fmt.Printf("  📁 %s/ (%d 个文件)\n", entry.Name(), len(subEntries))
		} else {
			info, _ := entry.Info()
			fmt.Printf("  📄 %s (%.1f KB)\n", entry.Name(), float64(info.Size())/1024)
		}
	}
}

// cleanResults 清理重复和损坏文件
func cleanResults() {
	fmt.Println("清理功能待实现")
	// TODO: 实现清理逻辑
}

// readMeta 读取元数据文件
func readMeta(metaFile string) *core.ParsedFileInfo {
	data, err := os.ReadFile(metaFile)
	if err != nil {
		return nil
	}

	var info core.ParsedFileInfo
	if err := json.Unmarshal(data, &info); err != nil {
		return nil
	}

	return &info
}

// runBasicTest 运行基础测试
func runBasicTest() {
	log.Println("=== ZoteroFlow2 MinerU Integration Test ===")

	// 1. 加载配置
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("配置加载失败: %v", err)
	}

	log.Printf("Zotero数据库路径: %s", cfg.ZoteroDBPath)
	log.Printf("Zotero数据目录: %s", cfg.ZoteroDataDir)
	log.Printf("MinerU API URL: %s", cfg.MineruAPIURL)
	log.Printf("缓存目录: %s", cfg.CacheDir)

	// 2. 连接Zotero数据库
	zoteroDB, err := core.NewZoteroDB(cfg.ZoteroDBPath, cfg.ZoteroDataDir)
	if err != nil {
		log.Fatalf("连接Zotero数据库失败: %v", err)
	}
	defer zoteroDB.Close()

	// 3. 创建MinerU客户端
	mineruClient := core.NewMinerUClient(cfg.MineruAPIURL, cfg.MineruToken)
	log.Println("MinerU client created successfully")

	// 4. 创建PDF解析器
	parser, err := core.NewPDFParser(zoteroDB, mineruClient, cfg.CacheDir)
	if err != nil {
		log.Fatalf("创建PDF解析器失败: %v", err)
	}
	log.Println("PDF parser created successfully")

	// 5. 测试基础功能
	testBasicFunctions(zoteroDB, mineruClient, parser)

	log.Println("\n=== Test Completed ===")
	log.Println("Tip: 使用 './zoteroflow2 help' 查看CLI命令")
}

// searchAndParse 搜索并解析文献 - 核心函数
func searchAndParse(query, searchType string) {
	cfg, err := config.Load()
	if err != nil {
		log.Printf("配置加载失败: %v", err)
		return
	}

	if cfg.MineruToken == "" {
		log.Println("❌ MinerU Token 未配置")
		return
	}

	// 连接数据库
	zoteroDB, err := core.NewZoteroDB(cfg.ZoteroDBPath, cfg.ZoteroDataDir)
	if err != nil {
		log.Printf("连接数据库失败: %v", err)
		return
	}
	defer zoteroDB.Close()

	// 使用搜索功能 - 速度优先！
	results, err := zoteroDB.SearchByTitle(query, 5)
	if err != nil {
		log.Printf("搜索失败: %v", err)
		return
	}

	if len(results) == 0 {
		log.Printf("❌ 未找到匹配的文献")
		return
	}

	// 显示搜索结果
	fmt.Printf("\n📄 找到 %d 篇文献:\n", len(results))
	for i, result := range results {
		fmt.Printf("  %d. 标题: %s (评分: %.1f)\n", i+1, result.Title, result.Score)
		fmt.Printf("     作者: %s\n", result.Authors)
		if result.Journal != "" {
			fmt.Printf("     期刊: %s\n", result.Journal)
		}
		if result.Year != "" {
			fmt.Printf("     年份: %s\n", result.Year)
		}
		if result.DOI != "" {
			fmt.Printf("     DOI: %s\n", result.DOI)
		}
		fmt.Printf("     PDF路径: %s\n", result.PDFPath)
		fmt.Println()
	}

	// 解析第一篇文献
	if len(results) > 0 {
		parseDocument(results[0].PDFPath, cfg)
	}
}

// parseDocument 解析文档
func parseDocument(pdfPath string, cfg *config.Config) {
	if pdfPath == "" {
		log.Printf("❌ PDF路径为空")
		return
	}

	// 检查文件存在
	if _, err := os.Stat(pdfPath); os.IsNotExist(err) {
		log.Printf("❌ PDF文件不存在: %s", pdfPath)
		return
	}

	// 创建MinerU客户端
	mineruClient := core.NewMinerUClient(cfg.MineruAPIURL, cfg.MineruToken)

	log.Println("🚀 开始解析PDF...")
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	startTime := time.Now()
	result, err := mineruClient.ParsePDF(ctx, pdfPath)
	if err != nil {
		log.Printf("❌ PDF解析失败: %v", err)
		return
	}

	duration := time.Since(startTime)
	log.Printf("✅ PDF解析成功! 耗时: %v", duration)

	// 显示结果
	fmt.Printf("\n📊 解析结果:\n")
	fmt.Printf("  任务ID: %s\n", result.TaskID)
	fmt.Printf("  处理耗时: %d ms\n", result.Duration)

	if result.ZipPath != "" {
		fmt.Printf("  ZIP文件: %s\n", result.ZipPath)
		fmt.Printf("\n📁 文件已自动组织到: data/results/\n")
		fmt.Printf("使用 './zoteroflow2 list' 查看所有结果\n")
	}
}

func testBasicFunctions(zoteroDB *core.ZoteroDB, mineruClient *core.MinerUClient, parser *core.PDFParser) {
	log.Println("\n=== Testing Basic Functions ===")

	// 测试数据库查询
	items, err := zoteroDB.GetItemsWithPDF(5)
	if err != nil {
		log.Printf("Database query failed: %v", err)
		return
	}

	log.Printf("Database query successful, found %d documents", len(items))
	for i, item := range items {
		fmt.Printf("  [%d] %s (类型: %s)\n", i+1, item.PDFName, item.ItemType)
	}

	// 测试MinerU客户端连接
	if mineruClient.Token != "" {
		log.Println("MinerU client configured correctly")
	} else {
		log.Println("Warning: MinerU Token not configured")
	}
}
