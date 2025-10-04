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
	"zoteroflow2-server/mcp"
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

// loadConfigWithCheck 加载配置并进行错误检查的公共函数
func loadConfigWithCheck() *config.Config {
	cfg, err := config.Load()
	if err != nil {
		log.Printf("配置加载失败: %v", err)
		return nil
	}
	return cfg
}

// createClients 根据配置创建AI和Zotero客户端的公共函数
func createClients(cfg *config.Config) (*core.ZoteroDB, core.AIClient, error) {
	if cfg == nil {
		return nil, nil, fmt.Errorf("配置为空")
	}

	// 连接Zotero数据库
	zoteroDB, err := core.NewZoteroDB(cfg.ZoteroDBPath, cfg.ZoteroDataDir)
	if err != nil {
		return nil, nil, fmt.Errorf("连接Zotero数据库失败: %w", err)
	}

	// 检查是否需要AI客户端
	var aiClient core.AIClient
	if cfg.AIAPIKey != "" {
		aiClient = core.NewGLMClient(cfg.AIAPIKey, cfg.AIBaseURL, cfg.AIModel)
	}

	return zoteroDB, aiClient, nil
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
	case "related":
		mcp.HandleRelatedLiterature(args[1:])
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
	fmt.Println("📚 文献管理:")
	fmt.Println("  list                    - 列出所有解析结果")
	fmt.Println("  open <名称>             - 打开指定文献文件夹")
	fmt.Println("  search <关键词>         - 按标题搜索并解析文献")
	fmt.Println("  doi <DOI号>             - 按DOI搜索并解析文献")
	fmt.Println()
	fmt.Println("🤖 AI助手对话:")
	fmt.Println("  chat                    - 进入交互式AI对话模式")
	fmt.Println("  chat <问题>             - 单次AI问答")
	fmt.Println("  chat --doc=文献名 <问题> - 基于指定文献的AI对话")
	fmt.Println()
	fmt.Println("🔍 智能文献分析:")
	fmt.Println("  related <文献名/DOI> <问题> - 查找相关文献并AI分析")
	fmt.Println()
	fmt.Println("🔧 帮助命令:")
	fmt.Println("  help                    - 显示此帮助信息")
	fmt.Println()
	fmt.Println("💡 使用示例:")
	fmt.Println("  ./zoteroflow2 list                                    # 列出文献")
	fmt.Println("  ./zoteroflow2 search \"机器学习\"                      # 搜索文献")
	fmt.Println("  ./zoteroflow2 chat \"什么是深度学习？\"                # AI问答")
	fmt.Println("  ./zoteroflow2 chat --doc=基因组 \"介绍一下CRISPR\"        # 基于文献的AI对话")
	fmt.Println("  ./zoteroflow2 related \"机器学习教程\" \"这篇论文的主要贡献是什么？\" # 智能文献分析")
	fmt.Println("  ./zoteroflow2 related \"10.1038/nature12373\" \"找到相似的研究\" # 相关文献查找")
	fmt.Println()
	fmt.Println("🎯 AI功能特性:")
	fmt.Println("  • 支持学术文献分析和解释")
	fmt.Println("  • 可基于特定文献内容进行对话")
	fmt.Println("  • 交互式对话模式支持上下文记忆")
	fmt.Println("  • 单次问答模式，适合快速查询")
}

// listResults 列出所有解析结果
func listResults() {
	cfg := loadConfigWithCheck()
	if cfg == nil {
		return
	}

	resultsDir := cfg.ResultsDir

	entries, err := os.ReadDir(resultsDir)
	if err != nil {
		log.Printf("读取结果目录失败: %v", err)
		return
	}

	// 先过滤出有效的结果文件夹
	var validResults []os.DirEntry
	for _, entry := range entries {
		if entry.IsDir() && !strings.HasPrefix(entry.Name(), ".") {
			metaFile := filepath.Join(resultsDir, entry.Name(), "meta.json")
			if info := readMeta(metaFile); info != nil {
				validResults = append(validResults, entry)
			}
		}
	}

	if len(validResults) == 0 {
		fmt.Println("暂无解析结果")
		return
	}

	fmt.Printf("找到 %d 个解析结果:\n\n", len(validResults))

	for i, entry := range validResults {
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

// openResult 打开指定文献
func openResult(name string) {
	cfg := loadConfigWithCheck()
	if cfg == nil {
		return
	}

	resultsDir := cfg.ResultsDir

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

	// 0. 验证并重建解析记录（确保数据一致性）
	log.Println("验证解析记录与实际文件的对应关系...")
	if err := core.ValidateAndRebuildRecords(); err != nil {
		log.Printf("验证记录失败: %v", err)
	} else {
		log.Println("✅ 记录验证完成")
	}

	// 0.5. 重新生成缺失或无效的meta.json文件
	log.Println("检查并重新生成缺失的meta.json文件...")
	if err := core.RegenerateMissingMeta(); err != nil {
		log.Printf("重新生成meta.json失败: %v", err)
	} else {
		log.Println("✅ meta.json检查完成")
	}

	// 0.6. 清理冗余的ZIP文件
	log.Println("清理冗余的ZIP文件...")
	if err := core.CleanupRedundantZIPs(); err != nil {
		log.Printf("清理ZIP文件失败: %v", err)
	} else {
		log.Println("✅ ZIP文件清理完成")
	}

	// 1. 加载配置
	cfg := loadConfigWithCheck()
	if cfg == nil {
		return
	}

	log.Printf("Zotero数据库路径: %s", cfg.ZoteroDBPath)
	log.Printf("Zotero数据目录: %s", cfg.ZoteroDataDir)
	log.Printf("MinerU API URL: %s", cfg.MineruAPIURL)
	log.Printf("缓存目录: %s", cfg.CacheDir)

	// 连接Zotero数据库
	zoteroDB, _, err := createClients(cfg)
	if err != nil {
		log.Fatalf("创建客户端失败: %v", err)
	}
	defer zoteroDB.Close()

	// 创建MinerU客户端
	mineruClient := core.NewMinerUClientWithResultsDir(cfg.MineruAPIURL, cfg.MineruToken, cfg.ResultsDir)
	log.Println("MinerU client created successfully")

	// 创建PDF解析器
	parser, err := core.NewPDFParser(zoteroDB, mineruClient, cfg.CacheDir)
	if err != nil {
		log.Fatalf("创建PDF解析器失败: %v", err)
	}
	log.Println("PDF parser created successfully")

	// 测试基础功能
	testBasicFunctions(zoteroDB, mineruClient, parser)

	log.Println("\n=== Test Completed ===")
	log.Println("Tip: 使用 './zoteroflow2 help' 查看CLI命令")
}

// searchAndParse 搜索并解析文献 - 核心函数
func searchAndParse(query, _ string) {
	cfg := loadConfigWithCheck()
	if cfg == nil || cfg.MineruToken == "" {
		log.Println("❌ 配置加载失败或MinerU Token 未配置")
		return
	}

	// 连接数据库
	log.Printf("配置数据目录: %s", cfg.ZoteroDataDir)
	zoteroDB, _, err := createClients(cfg)
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
		// 显示作者列表
		authorsStr := "未知作者"
		if len(result.Authors) > 0 {
			authorsStr = strings.Join(result.Authors, "; ")
		}
		fmt.Printf("     作者: %s\n", authorsStr)
		if result.Journal != "" {
			fmt.Printf("     期刊: %s\n", result.Journal)
		}
		if result.Year != 0 {
			fmt.Printf("     年份: %d\n", result.Year)
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
	mineruClient := core.NewMinerUClientWithResultsDir(cfg.MineruAPIURL, cfg.MineruToken, cfg.ResultsDir)

	log.Println("🚀 开始解析PDF...")
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(cfg.MineruTimeout)*time.Second)
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
		fmt.Printf("\n📁 文件已自动组织到: %s/\n", cfg.ResultsDir)
		fmt.Printf("使用 './zoteroflow2 list' 查看所有结果\n")
	}
}

func testBasicFunctions(zoteroDB *core.ZoteroDB, mineruClient *core.MinerUClient, _ *core.PDFParser) {
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
	cfg := loadConfigWithCheck()
	if cfg == nil || cfg.AIAPIKey == "" {
		fmt.Println("❌ AI功能未配置，请设置 AI_API_KEY 环境变量")
		fmt.Println("示例: export AI_API_KEY=your_api_key_here")
		return
	}

	// 创建客户端
	zoteroDB, aiClient, err := createClients(cfg)
	if err != nil {
		log.Printf("创建客户端失败: %v", err)
		return
	}
	defer zoteroDB.Close()

	// 创建对话管理器
	chatManager := core.NewAIConversationManager(aiClient, zoteroDB)

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

	// 创建客户端
	zoteroDB, aiClient, err := createClients(cfg)
	if err != nil {
		log.Printf("创建客户端失败: %v", err)
		return
	}
	defer zoteroDB.Close()

	// 创建对话管理器
	chatManager := core.NewAIConversationManager(aiClient, zoteroDB)

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
	cfg := loadConfigWithCheck()
	if cfg == nil {
		return nil, fmt.Errorf("配置加载失败")
	}

	resultsDir := cfg.ResultsDir

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

// chatWithAI 单次AI对话（优化版）
func chatWithAI(message string) {
	if strings.TrimSpace(message) == "" {
		fmt.Println("❌ 请输入有效的消息内容")
		return
	}

	fmt.Printf("🤖 正在分析您的问题: %s\n", message)

	cfg := loadConfigWithCheck()
	if cfg == nil || cfg.AIAPIKey == "" {
		fmt.Println("❌ AI功能未配置，请设置 AI_API_KEY 环境变量")
		fmt.Println("示例: export AI_API_KEY=your_api_key_here")
		return
	}

	// 创建AI客户端
	aiClient := core.NewGLMClient(cfg.AIAPIKey, cfg.AIBaseURL, cfg.AIModel)

	// 创建AI-MCP桥接器
	aiMCPBridge := mcp.NewAIMCPBridge(aiClient, cfg)
	defer aiMCPBridge.Close()

	// 记录开始时间
	startTime := time.Now()

	// 让AI选择并调用工具
	fmt.Printf("🧠 AI正在分析并选择合适的工具...\n")
	toolCall, aiResponse, err := aiMCPBridge.SelectTool(message)
	if err != nil {
		fmt.Printf("❌ AI工具选择失败: %v\n", err)
		fmt.Printf("💡 降级到普通AI对话...\n")
		callAIWithoutTools(aiClient, message)
		return
	}

	var finalResponse string
	if toolCall != nil {
		// 调用MCP工具
		fmt.Printf("🔧 正在调用工具: %s (来自 %s)\n", toolCall.Tool, toolCall.Server)

		response, err := aiMCPBridge.CallTool(toolCall)
		if err != nil {
			fmt.Printf("❌ 工具调用失败: %v\n", err)
			fmt.Printf("💡 可能的原因:\n")
			fmt.Printf("   - MCP服务器启动失败\n")
			fmt.Printf("   - 网络连接问题\n")
			fmt.Printf("   - 工具参数格式错误\n")
			fmt.Printf("💡 降级到普通AI对话...\n")
			callAIWithoutTools(aiClient, message)
			return
		}

		fmt.Printf("✅ 工具调用成功，正在生成回答...\n")

		// 解析工具结果
		toolResult := aiMCPBridge.ParseToolResult(response)

		// 生成最终答案
		finalResponse = aiMCPBridge.GenerateFinalAnswer(&message, &toolResult, aiResponse)

	} else {
		// 不需要工具，使用AI的直接回复
		if aiResponse != nil && *aiResponse != "" {
			finalResponse = *aiResponse
		} else {
			fmt.Printf("⚠️ AI未生成回复，降级到普通对话...\n")
			callAIWithoutTools(aiClient, message)
			return
		}
	}

	// 显示结果
	totalTime := time.Since(startTime)
	fmt.Printf("🤖 助手: %s\n", finalResponse)
	fmt.Printf("⏱️ 总耗时: %v\n", totalTime)
}

// callAIWithoutTools 不使用MCP工具的普通AI对话
func callAIWithoutTools(aiClient core.AIClient, message string) {
	// 加载配置获取模型信息
	cfg := loadConfigWithCheck()
	if cfg == nil {
		fmt.Printf("❌ 配置加载失败\n")
		return
	}

	messages := []core.ChatMessage{
		{
			Role:    "system",
			Content: "你是一个专业的学术文献助手，请用中文提供准确、有用的信息。回答要简洁明了。",
		},
		{
			Role:    "user",
			Content: message,
		},
	}

	req := &core.AIRequest{
		Model:     cfg.AIModel,
		Messages:  messages,
		MaxTokens: 500,
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(cfg.AITimeout)*time.Second)
	defer cancel()

	response, err := aiClient.Chat(ctx, req)
	if err != nil {
		fmt.Printf("❌ AI响应失败: %v\n", err)
		fmt.Println("💡 可能的原因:")
		fmt.Println("   - 网络连接问题")
		fmt.Println("   - API密钥无效")
		fmt.Println("   - 请求超时")
		return
	}

	if len(response.Choices) > 0 {
		aiResponse := response.Choices[0].Message.Content
		if aiResponse == "" {
			// 如果content为空，检查是否有思考过程
			fmt.Printf("🤖 助手: 正在思考...\n")
			fmt.Printf("💡 AI正在处理您的问题，请稍等片刻\n")
			fmt.Printf("   或使用交互模式进行更详细的对话: ./zoteroflow2 chat\n")
		} else {
			fmt.Printf("🤖 助手: %s\n", aiResponse)
			fmt.Printf("\n📊 Token使用: %d (输入) + %d (输出) = %d (总计)\n",
				response.Usage.PromptTokens,
				response.Usage.CompletionTokens,
				response.Usage.TotalTokens)
		}
	} else {
		fmt.Println("❌ 未收到AI响应")
	}

	// 显示调试信息
	if response.Usage.TotalTokens > 0 {
		log.Printf("✅ AI响应成功，Token使用: %d", response.Usage.TotalTokens)
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
		var abstract string
		var ok bool

		if abstract, ok = strings.CutPrefix(line, "摘要："); ok {
			// 中文摘要
		} else if abstract, ok = strings.CutPrefix(line, "Abstract:"); ok {
			// 英文摘要
		} else {
			continue
		}

		// 使用配置的长度限制
		cfg := loadConfigWithCheck()
		if cfg != nil {
			if len(abstract) > cfg.AbstractLength {
				return abstract[:cfg.AbstractLength] + "..."
			}
		} else {
			// 如果配置加载失败，使用默认值
			if len(abstract) > 200 {
				return abstract[:200] + "..."
			}
		}
		return abstract
	}
	return "无摘要信息"
}

// extractSimpleKeywords 简化版关键词提取
func extractSimpleKeywords(content string) []string {
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		var keywordsStr string
		var ok bool

		if keywordsStr, ok = strings.CutPrefix(line, "关键词："); ok {
			// 中文关键词
		} else if keywordsStr, ok = strings.CutPrefix(line, "Key words:"); ok {
			// 英文关键词
		} else {
			continue
		}

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
	return []string{"未找到关键词"}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
