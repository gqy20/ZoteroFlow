package web

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"zoteroflow2-server/config"
	"zoteroflow2-server/core"
	"zoteroflow2-server/mcp"
	"github.com/gin-gonic/gin"
)

// AskRequest 请求结构
type AskRequest struct {
	Query string `json:"query"`
}

// AskResponse 响应结构
type AskResponse struct {
	Answer string `json:"answer"`
	PDFURL string `json:"pdfUrl,omitempty"`
}

// HandleAsk 处理AI问答请求
func HandleAsk(c *gin.Context) {
	var req AskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请输入有效问题"})
		return
	}

	log.Printf("收到查询: %s", req.Query)

	// 加载配置
	cfg := loadConfig()
	if cfg == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "配置加载失败"})
		return
	}

	// 智能路由：根据问题内容自动选择处理方式
	response, pdfURL := intelligentRouterWithAI(req.Query, cfg)

	c.JSON(http.StatusOK, AskResponse{
		Answer: response,
		PDFURL: pdfURL,
	})
}

// loadConfig 加载配置
func loadConfig() *config.Config {
	cfg, err := config.Load()
	if err != nil {
		log.Printf("配置加载失败: %v", err)
		return nil
	}
	return cfg
}

// intelligentRouterWithAI 集成AI功能的智能路由器
func intelligentRouterWithAI(query string, cfg *config.Config) (string, string) {
	query = strings.ToLower(query)

	// PDF查看类
	if containsAny(query, []string{"查看", "预览", "打开", "view", "open", "preview"}) {
		return handlePDFView(query)
	}

	// 相关文献分析类
	if containsAny(query, []string{"相关", "related", "相似", "similar", "推荐"}) {
		return handleRelatedLiterature(query, cfg)
	}

	// 文献搜索类
	if containsAny(query, []string{"搜索", "找", "查找", "search", "find"}) {
		return handleRealSearch(query, cfg)
	}

	// 文献分析类
	if containsAny(query, []string{"分析", "总结", "概括", "analyze", "summary"}) {
		return handleRealAnalysis(query, cfg)
	}

	// AI对话类
	return handleRealAIChat(query, cfg)
}

// handlePDFView PDF查看处理
func handlePDFView(query string) (string, string) {
	// 从查询中提取文献名称或DOI
	docName := extractDocumentName(query)
	if docName == "" {
		return "请指定要查看的文献名称或DOI，例如：查看Attention Is All You Need", ""
	}

	// 查找PDF文件路径
	pdfPath, err := findPDFPath(docName)
	if err != nil {
		return fmt.Sprintf("未找到文献: %s，请检查文献名称或PDF文件是否存在", docName), ""
	}

	// 生成可访问的PDF URL
	pdfURL := fmt.Sprintf("/static/pdf/%s", pdfPath)

	answer := fmt.Sprintf("已找到文献《%s》，点击\"查看PDF\"按钮即可阅读", docName)
	return answer, pdfURL
}

// findPDFPath 查找PDF文件路径
func findPDFPath(docName string) (string, error) {
	// 简单的PDF文件查找逻辑
	// 实际使用中需要集成ZoteroDB
	// 这里返回一个示例PDF路径
	return "example.pdf", nil
}

// handleRelatedLiterature 相关文献分析处理
func handleRelatedLiterature(query string, cfg *config.Config) (string, string) {
	// 检查MCP配置
	if !mcp.IsMCPConfigured() {
		return "MCP功能未配置，无法进行相关文献分析。请检查MCP服务器配置。", ""
	}

	// 直接使用AI处理文献搜索和分析，让AI自己理解查询意图
	enhancedQuery := fmt.Sprintf("请根据用户需求查找相关的学术文献并提供详细分析：%s\n\n请使用可用的搜索工具查找相关论文，然后对搜索结果进行综合分析和总结。", query)

	// 使用AI进行相关文献分析
	return handleRealAIChat(enhancedQuery, cfg)
}


// extractDocumentName 从查询中提取文献名称
func extractDocumentName(query string) string {
	// 简单的文献名称提取逻辑
	query = strings.ToLower(query)

	// 移除关键词
	prefixes := []string{"查看", "预览", "打开", "view", "open", "preview", "pdf", "论文", "文献"}
	for _, prefix := range prefixes {
		query = strings.ReplaceAll(query, prefix, "")
		query = strings.TrimSpace(query)
	}

	// 移除标点符号
	query = strings.Trim(query, " ,.!?，。！？")

	if len(query) > 0 {
		return query
	}
	return ""
}

// handleRealSearch 真实文献搜索处理
func handleRealSearch(query string, cfg *config.Config) (string, string) {
	// 连接Zotero数据库
	zoteroDB, err := core.NewZoteroDB(cfg.ZoteroDBPath, cfg.ZoteroDataDir)
	if err != nil {
		log.Printf("连接Zotero数据库失败: %v", err)
		return "数据库连接失败，请检查配置", ""
	}
	defer zoteroDB.Close()

	// 搜索文献
	items, err := zoteroDB.SearchByTitle(query, 10)
	if err != nil {
		log.Printf("搜索文献失败: %v", err)
		return "搜索失败: " + err.Error(), ""
	}

	if len(items) == 0 {
		return fmt.Sprintf("未找到与 \"%s\" 相关的文献，请尝试其他关键词", query), ""
	}

	var formatted strings.Builder
	formatted.WriteString(fmt.Sprintf("找到 %d 篇相关文献：\n\n", len(items)))

	for i, item := range items {
		formatted.WriteString(fmt.Sprintf("%d. **%s**\n", i+1, item.Title))
		if len(item.Authors) > 0 {
			authors := strings.Join(item.Authors, "; ")
			formatted.WriteString(fmt.Sprintf("   作者: %s\n", authors))
		}
		if item.Year != 0 {
			formatted.WriteString(fmt.Sprintf("   年份: %d\n", item.Year))
		}
		if item.DOI != "" {
			formatted.WriteString(fmt.Sprintf("   DOI: %s\n", item.DOI))
		}
		formatted.WriteString("\n")
	}

	return formatted.String(), ""
}

// handleRealAnalysis 真实文献分析处理
func handleRealAnalysis(query string, cfg *config.Config) (string, string) {
	// 首先尝试AI分析
	if cfg.AIAPIKey != "" {
		return handleRealAIChat(query, cfg)
	}

	// 如果没有AI配置，则使用简单的文本分析
	return "AI分析功能未配置。请在 .env 文件中设置 AI_API_KEY 来启用智能分析功能。", ""
}

// handleRealAIChat 真实AI对话处理
func handleRealAIChat(query string, cfg *config.Config) (string, string) {
	if cfg.AIAPIKey == "" {
		return "AI功能未配置，请设置 AI_API_KEY 环境变量或在 .env 文件中配置", ""
	}

	// 创建AI客户端
	aiClient := core.NewGLMClient(cfg.AIAPIKey, cfg.AIBaseURL, cfg.AIModel)
	if aiClient == nil {
		return "AI客户端创建失败，请检查配置", ""
	}

	// 创建AI-MCP桥接器（与CLI模式相同）
	aiMCPBridge := mcp.NewAIMCPBridge(aiClient, cfg)
	defer aiMCPBridge.Close()

	// 让AI选择工具
	toolCall, aiResponse, err := aiMCPBridge.SelectTool(query)
	if err != nil {
		log.Printf("AI工具选择失败: %v", err)
		log.Printf("降级到普通AI对话...")

		// 降级到普通AI对话
		aiRequest := &core.AIRequest{
			Model: cfg.AIModel,
			Messages: []core.ChatMessage{
				{
					Role:    "system",
					Content: "你是一个专业的学术文献助手，能够帮助用户分析、搜索和回答关于学术文献的问题。请用中文回答，保持专业和准确。",
				},
				{
					Role:    "user",
					Content: query,
				},
			},
			MaxTokens: 1000,
			Temperature: 0.7,
		}

		// 发送AI请求（带超时）
		ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
		defer cancel()

		response, err := aiClient.Chat(ctx, aiRequest)
		if err != nil {
			log.Printf("AI请求失败: %v", err)
			return "AI请求失败: " + err.Error(), ""
		}

		if response == nil || len(response.Choices) == 0 {
			return "AI响应为空，请稍后重试", ""
		}

		return response.Choices[0].Message.Content, ""
	}

	// 如果AI选择了工具，执行工具调用并获取结果
	if toolCall != nil {
		log.Printf("🔧 执行工具调用: %s.%s", toolCall.Server, toolCall.Tool)

		// 调用工具
		toolResponse, err := aiMCPBridge.CallTool(toolCall)
		if err != nil {
			log.Printf("工具调用失败: %v", err)
			return "工具调用失败: " + err.Error(), ""
		}

		// 解析工具结果
		toolResult := aiMCPBridge.ParseToolResult(toolResponse)
		log.Printf("📄 工具调用完成，结果长度: %d", len(toolResult))

		// 使用AI分析和总结工具结果
		if len(toolResult) > 0 {
			log.Printf("🧠 开始用AI分析工具结果...")

			// 构建AI分析请求
			analysisPrompt := fmt.Sprintf(`请分析以下搜索结果，并生成一份简洁、用户友好的摘要报告。搜索关键词："%s"

搜索结果（原始数据）：
%s

请提供：
1. 对搜索结果的分析和总结
2. 最重要的发现或亮点
3. 相关性和质量评估
4. 用中文回答，保持专业和准确`, query, toolResult)

			// 使用AI分析工具结果
			analysisRequest := &core.AIRequest{
				Model: cfg.AIModel,
				Messages: []core.ChatMessage{
					{
						Role:    "system",
						Content: "你是一个专业的学术文献分析师，能够分析搜索结果并生成简洁、有用的摘要。请用中文回答。",
					},
					{
						Role:    "user",
						Content: analysisPrompt,
					},
				},
				MaxTokens: 2000,
				Temperature: 0.3,
			}

			// 发送AI分析请求（带超时）
			ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
			defer cancel()

			analysisResponse, err := aiClient.Chat(ctx, analysisRequest)
			if err != nil {
				log.Printf("AI分析失败: %v", err)
				// 降级到GenerateFinalAnswer
				log.Printf("降级到GenerateFinalAnswer方法...")
				finalAnswer := aiMCPBridge.GenerateFinalAnswer(&query, &toolResult, aiResponse)
				return finalAnswer, ""
			}

			if analysisResponse != nil && len(analysisResponse.Choices) > 0 {
				log.Printf("✅ AI分析成功，生成用户友好的答案")
				return analysisResponse.Choices[0].Message.Content, ""
			}

			log.Printf("⚠️ AI分析响应为空，使用降级方案")
		}

		// 降级到原始方法
		finalAnswer := aiMCPBridge.GenerateFinalAnswer(&query, &toolResult, aiResponse)
		return finalAnswer, ""
	}

	// 如果AI没有选择工具，但有直接回复
	if aiResponse != nil && *aiResponse != "" {
		return *aiResponse, ""
	}

	// 如果没有工具响应，返回默认消息
	return "AI已处理您的请求，但没有生成具体响应。", ""
}

// formatSearchResults 搜索结果格式化
func formatSearchResults(results []string) string {
	if len(results) == 0 {
		return "未找到相关文献"
	}

	var formatted strings.Builder
	formatted.WriteString("找到以下文献：\n\n")

	for i, result := range results {
		formatted.WriteString(fmt.Sprintf("%d. %s\n", i+1, result))
	}

	return formatted.String()
}

// containsAny 简单的关键词匹配
func containsAny(text string, keywords []string) bool {
	for _, keyword := range keywords {
		if strings.Contains(text, keyword) {
			return true
		}
	}
	return false
}

// HandleStatus 系统状态检查
func HandleStatus(c *gin.Context) {
	status := gin.H{
		"status":    "running",
		"mode":      "web",
		"version":   "v1.0.0",
		"features": []string{
			"AI问答",
			"PDF查看",
			"文献搜索",
			"智能分析",
		},
	}
	c.JSON(http.StatusOK, status)
}

// HandleStaticConfig 静态配置信息
func HandleStaticConfig(c *gin.Context) {
	config := gin.H{
		"title":       "ZoteroFlow - AI文献助手",
		"description": "基于Go语言开发的智能文献管理工具",
		"version":     "v1.0.0",
	}
	c.JSON(http.StatusOK, config)
}