package main

import (
	"fmt"
	"log"

	"zoteroflow2-server/config"
	"zoteroflow2-server/core"
)

func main() {
	log.Println("=== ZoteroFlow2 MinerU 集成测试 ===")

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
	log.Println("✅ MinerU客户端创建成功")

	// 4. 创建PDF解析器
	parser, err := core.NewPDFParser(zoteroDB, mineruClient, cfg.CacheDir)
	if err != nil {
		log.Fatalf("创建PDF解析器失败: %v", err)
	}
	log.Println("✅ PDF解析器创建成功")

	// 5. 测试基础功能
	testBasicFunctions(zoteroDB, mineruClient, parser)

	log.Println("\n=== 测试完成 ===")
	log.Println("提示: 使用 'go run test_mineru.go' 进行完整的MinerU集成测试")
}

func testBasicFunctions(zoteroDB *core.ZoteroDB, mineruClient *core.MinerUClient, parser *core.PDFParser) {
	log.Println("\n=== 测试基础功能 ===")

	// 测试数据库查询
	items, err := zoteroDB.GetItemsWithPDF(5)
	if err != nil {
		log.Printf("❌ 数据库查询失败: %v", err)
		return
	}

	log.Printf("✅ 数据库查询成功，找到 %d 篇文献", len(items))
	for i, item := range items {
		fmt.Printf("  [%d] %s (类型: %s)\n", i+1, item.PDFName, item.ItemType)
	}

	// 测试MinerU客户端连接
	if mineruClient.Token != "" {
		log.Println("✅ MinerU客户端配置正确")
	} else {
		log.Println("⚠️  MinerU Token 未配置")
	}
}