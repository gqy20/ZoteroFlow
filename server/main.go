package main

import (
	"bufio"
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
	case "chat":
		if len(args) < 2 {
			startInteractiveChat()
		} else {
			// 检查是否是文献指定格式
			if strings.HasPrefix(args[1], "--doc=") || strings.HasPrefix(args[1], "-d=") {
				docName := strings.TrimPrefix(strings.TrimPrefix(args[1], "--doc="), "-d=")
				message := strings.Join(args[2:], " ")
				chatWithDocument(docName, message)
			} else {
				chatWithAI(strings.Join(args[1:], " "))
			}
		}
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
	fmt.Println("  chat [消息]             - AI学术助手对话")
	fmt.Println("  chat --doc=文献名 [消息] - 基于指定文献的AI对话")
	fmt.Println("  clean                   - 清理重复/损坏文件")
	fmt.Println("  help                    - 显示此帮助信息")
	fmt.Println()
	fmt.Println("示例:")
	fmt.Println("  ./zoteroflow2 list")
	fmt.Println("  ./zoteroflow2 open 丛枝菌根")
	fmt.Println("  ./zoteroflow2 search \"solanum chromosome\"")
	fmt.Println("  ./zoteroflow2 doi \"10.1111/j.1469-8137.2012.04195.x\"")
	fmt.Println("  ./zoteroflow2 chat")
	fmt.Println("  ./zoteroflow2 chat \"请解释一下什么是基因组学\"")
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

// startInteractiveChat 启动交互式AI对话
func startInteractiveChat() {
	// 加载配置
	cfg, err := config.Load()
	if err != nil {
		log.Printf("加载配置失败: %v", err)
		return
	}

	// 检查AI配置
	if cfg.AIAPIKey == "" {
		fmt.Println("❌ AI功能未配置，请设置 AI_API_KEY 环境变量")
		fmt.Println("示例: export AI_API_KEY=your_api_key_here")
		return
	}

	// 创建AI客户端
	client := core.NewGLMClient(cfg.AIAPIKey, cfg.AIBaseURL, cfg.AIModel)

	// 连接Zotero数据库
	zoteroDB, err := core.NewZoteroDB(cfg.ZoteroDBPath, cfg.ZoteroDataDir)
	if err != nil {
		log.Printf("连接Zotero数据库失败: %v", err)
		return
	}
	defer zoteroDB.Close()

	// 创建对话管理器
	chatManager := core.NewAIConversationManager(client, zoteroDB)

	fmt.Println("🤖 ZoteroFlow2 AI学术助手")
	fmt.Println("输入 'help' 查看帮助，输入 'quit' 或 'exit' 退出")
	fmt.Println(strings.Repeat("-", 50))

	// 开始对话循环
	scanner := bufio.NewScanner(os.Stdin)
	var currentConvID string

	for {
		fmt.Print("📚 您: ")
		if !scanner.Scan() {
			break
		}

		input := strings.TrimSpace(scanner.Text())
		if input == "" {
			continue
		}

		switch input {
		case "quit", "exit", "退出":
			fmt.Println("👋 再见!")
			return
		case "help", "帮助":
			showChatHelp()
			continue
		case "new", "新建":
			currentConvID = ""
			fmt.Println("🆕 开始新对话")
			continue
		case "clear", "清屏":
			fmt.Print("\033[H\033[2J")
			continue
		}

		// 如果没有当前对话，创建新对话
		if currentConvID == "" {
			conv, err := chatManager.StartConversation(context.Background(), input, nil)
			if err != nil {
				fmt.Printf("❌ 对话失败: %v\n", err)
				continue
			}

			currentConvID = conv.ID
			if len(conv.Messages) >= 3 {
				aiResponse := conv.Messages[2]
				fmt.Printf("🤖 助手: %s\n", aiResponse.Content)
			}
		} else {
			// 继续当前对话
			conv, err := chatManager.ContinueConversation(context.Background(), currentConvID, input)
			if err != nil {
				fmt.Printf("❌ 对话失败: %v\n", err)
				continue
			}

			if len(conv.Messages) >= 2 {
				lastMsg := conv.Messages[len(conv.Messages)-1]
				if lastMsg.Role == "assistant" {
					fmt.Printf("🤖 助手: %s\n", lastMsg.Content)
				}
			}
		}

		fmt.Println()
	}
}

// chatWithDocument 基于指定文献的AI对话
func chatWithDocument(docName, message string) {
	// 加载配置
	cfg, err := config.Load()
	if err != nil {
		log.Printf("加载配置失败: %v", err)
		return
	}

	// 检查AI配置
	if cfg.AIAPIKey == "" {
		fmt.Println("❌ AI功能未配置，请设置 AI_API_KEY 环境变量")
		return
	}

	// 创建AI客户端
	client := core.NewGLMClient(cfg.AIAPIKey, cfg.AIBaseURL, cfg.AIModel)

	// 连接Zotero数据库
	zoteroDB, err := core.NewZoteroDB(cfg.ZoteroDBPath, cfg.ZoteroDataDir)
	if err != nil {
		log.Printf("连接Zotero数据库失败: %v", err)
		return
	}
	defer zoteroDB.Close()

	// 创建对话管理器
	chatManager := core.NewAIConversationManager(client, zoteroDB)

	// 查找指定的文献
	docContext, err := findDocumentContext(docName)
	if err != nil {
		fmt.Printf("❌ 未找到文献 '%s': %v\n", docName, err)
		fmt.Println("💡 使用 './zoteroflow2 list' 查看可用文献")
		return
	}

	fmt.Printf("📚 基于文献 '%s' 进行对话\n", docContext.Documents[0].Title)
	fmt.Printf("📝 作者: %s\n", docContext.Documents[0].Authors)
	if docContext.Documents[0].Abstract != "" {
		fmt.Printf("📄 摘要: %s\n", docContext.Documents[0].Abstract[:min(100, len(docContext.Documents[0].Abstract))]+"...")
	}
	fmt.Println(strings.Repeat("-", 50))

	// 创建基于指定文献的对话
	conv, err := chatManager.StartConversationWithDocument(context.Background(), message, docContext)
	if err != nil {
		fmt.Printf("❌ 对话失败: %v\n", err)
		return
	}

	if len(conv.Messages) >= 3 {
		aiResponse := conv.Messages[2]
		fmt.Printf("🤖 助手: %s\n", aiResponse.Content)

		if len(conv.Context.Documents) > 0 {
			fmt.Printf("\n📊 Token使用: %d (输入) + %d (输出) = %d (总计)\n",
				len(conv.Messages[1].Content), // 简化的输入计数
				len(aiResponse.Content)/3,     // 简化的输出计数
				len(aiResponse.Content)/3+len(conv.Messages[1].Content))
		}
	}
}

// findDocumentContext 查找指定文献的上下文
func findDocumentContext(docName string) (*core.DocumentContext, error) {
	resultsDir := "data/results"

	// 首先尝试精确匹配文件名
	entries, err := os.ReadDir(resultsDir)
	if err != nil {
		return nil, fmt.Errorf("读取结果目录失败: %w", err)
	}

	targetDocName := ""

	// 查找匹配的文献
	for _, entry := range entries {
		if !entry.IsDir() || entry.Name() == "latest" {
			continue
		}

		// 检查文件名匹配
		entryName := entry.Name()
		if strings.Contains(entryName, docName) || docName == entryName {
			targetDocName = entryName
			break
		}
	}

	if targetDocName == "" {
		return nil, fmt.Errorf("未找到匹配的文献")
	}

	// 读取文献内容
	mdFile := filepath.Join(resultsDir, targetDocName, "full.md")
	content, err := os.ReadFile(mdFile)
	if err != nil {
		return nil, fmt.Errorf("读取文献内容失败: %w", err)
	}

	// 读取元数据
	metaFile := filepath.Join(resultsDir, targetDocName, "meta.json")
	info := readMeta(metaFile)
	if info == nil {
		return nil, fmt.Errorf("读取文献元数据失败")
	}

	// 创建文档摘要
	doc := &core.DocumentSummary{
		ID:       1,
		Title:    info.Title,
		Authors:  "见论文内容",
		Abstract: extractSimpleAbstract(string(content)),
		Keywords: extractSimpleKeywords(string(content)),
	}

	return &core.DocumentContext{
		Documents: []core.DocumentSummary{*doc},
		Query:     "",
		Relevance: 1.0, // 指定文献的相关性最高
	}, nil
}

// chatWithAI 单次AI对话
func chatWithAI(message string) {
	// 加载配置
	cfg, err := config.Load()
	if err != nil {
		log.Printf("加载配置失败: %v", err)
		return
	}

	// 检查AI配置
	if cfg.AIAPIKey == "" {
		fmt.Println("❌ AI功能未配置，请设置 AI_API_KEY 环境变量")
		return
	}

	// 创建AI客户端
	client := core.NewGLMClient(cfg.AIAPIKey, cfg.AIBaseURL, cfg.AIModel)

	// 创建对话请求
	messages := []core.ChatMessage{
		{
			Role:    "system",
			Content: "你是一个专业的学术文献助手，请用中文提供准确、有用的信息。",
		},
		{
			Role:    "user",
			Content: message,
		},
	}

	// 发送请求
	req := &core.AIRequest{
		Model:    cfg.AIModel,
		Messages: messages,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	response, err := client.Chat(ctx, req)
	if err != nil {
		fmt.Printf("❌ AI响应失败: %v\n", err)
		return
	}

	if len(response.Choices) > 0 {
		fmt.Printf("🤖 助手: %s\n", response.Choices[0].Message.Content)
		fmt.Printf("\n📊 Token使用: %d (输入) + %d (输出) = %d (总计)\n",
			response.Usage.PromptTokens,
			response.Usage.CompletionTokens,
			response.Usage.TotalTokens)
	} else {
		fmt.Println("❌ 未收到AI响应")
	}
}

// showChatHelp 显示对话帮助
func showChatHelp() {
	fmt.Println("\n📖 对话帮助:")
	fmt.Println("  help/帮助   - 显示此帮助信息")
	fmt.Println("  new/新建   - 开始新对话")
	fmt.Println("  clear/清屏  - 清空屏幕")
	fmt.Println("  quit/exit/退出 - 退出对话")
	fmt.Println("\n💡 使用建议:")
	fmt.Println("  - 可以询问学术概念、研究方法、论文分析等")
	fmt.Println("  - 支持中文对话，推荐使用学术相关问题")
	fmt.Println("  - 每次新对话会保留上下文，便于深入讨论")
	fmt.Println()
}

// extractSimpleAbstract 简化版摘要提取
func extractSimpleAbstract(content string) string {
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "摘要：") || strings.HasPrefix(line, "Abstract:") {
			// 返回摘要的第一部分
			abstract := strings.TrimPrefix(line, "摘要：")
			abstract = strings.TrimPrefix(abstract, "Abstract:")
			if len(abstract) > 200 {
				return abstract[:200] + "..."
			}
			return abstract
		}
	}
	return "无摘要信息"
}

// extractSimpleKeywords 简化版关键词提取
func extractSimpleKeywords(content string) []string {
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "关键词：") {
			keywordsStr := strings.TrimPrefix(line, "关键词：")
			keywordsStr = strings.TrimPrefix(keywordsStr, "Key words:")

			// 简单分割
			kwList := strings.FieldsFunc(keywordsStr, func(r rune) bool {
				return r == '；' || r == ';' || r == ' ' || r == ','
			})

			var keywords []string
			for _, kw := range kwList {
				kw = strings.TrimSpace(kw)
				if len(kw) > 1 && len(keywords) < 5 {
					keywords = append(keywords, kw)
				}
			}
			return keywords
		}
	}
	return []string{"未找到关键词"}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
