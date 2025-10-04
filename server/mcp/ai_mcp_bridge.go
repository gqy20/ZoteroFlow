package mcp

import (
	"context"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"log"
	"regexp"
	"strings"
	"sync"
	"time"

	"zoteroflow2-server/config"
	"zoteroflow2-server/core"
)

// MCPTool MCP工具信息
type MCPTool struct {
	Server    string `json:"server"`
	Name      string `json:"name"`
	Desc      string `json:"description"`
	Arguments map[string]interface{} `json:"arguments"`
}

// ToolCall 工具调用请求
type ToolCall struct {
	Server    string                 `json:"server"`
	Tool      string                 `json:"tool"`
	Arguments map[string]interface{} `json:"arguments"`
}

// CachedResult 缓存结果
type CachedResult struct {
	Response *MCPResponse
	Time     time.Time
	 TTL     time.Duration
}

// ToolCallCache 工具调用缓存
type ToolCallCache struct {
	cache map[string]*CachedResult
	mutex sync.RWMutex
}

// NewToolCallCache 创建工具调用缓存
func NewToolCallCache() *ToolCallCache {
	return &ToolCallCache{
		cache: make(map[string]*CachedResult),
	}
}

// Get 获取缓存结果
func (c *ToolCallCache) Get(key string) (*MCPResponse, bool) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	if result, exists := c.cache[key]; exists {
		if time.Since(result.Time) < result.TTL {
			log.Printf("🎯 缓存命中: %s", key)
			return result.Response, true
		}
		// 过期，删除
		delete(c.cache, key)
	}
	return nil, false
}

// Set 设置缓存结果
func (c *ToolCallCache) Set(key string, response *MCPResponse, ttl time.Duration) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.cache[key] = &CachedResult{
		Response: response,
		Time:     time.Now(),
		TTL:      ttl,
	}
	log.Printf("💾 缓存设置: %s (TTL: %v)", key, ttl)
}

// generateCacheKey 生成缓存键
func (amb *AIMCPBridge) generateCacheKey(toolCall *ToolCall) string {
	// 简单的哈希：工具名+参数
	return fmt.Sprintf("%s.%s:%s", toolCall.Server, toolCall.Tool,
		fmt.Sprintf("%x", md5.Sum([]byte(fmt.Sprintf("%v", toolCall.Arguments)))))
}

// AIMCPBridge AI与MCP的桥接器
type AIMCPBridge struct {
	aiClient  core.AIClient
	config    *config.Config
	mcpManager *MCPManager
	managerOnce sync.Once
	initError  error
	cache     *ToolCallCache
}

// NewAIMCPBridge 创建AI-MCP桥接器
func NewAIMCPBridge(aiClient core.AIClient, config *config.Config) *AIMCPBridge {
	return &AIMCPBridge{
		aiClient: aiClient,
		config:   config,
		cache:    NewToolCallCache(),
	}
}

// GetAvailableTools 获取所有可用的MCP工具
func (amb *AIMCPBridge) GetAvailableTools() ([]MCPTool, error) {
	manager, err := NewMCPManager("mcp/mcp_config.json")
	if err != nil {
		return nil, fmt.Errorf("创建MCP管理器失败: %w", err)
	}
	defer manager.Close()

	var allTools []MCPTool

	// 获取article-mcp工具
	if articleTools, err := amb.getArticleMCPTools(manager); err == nil {
		allTools = append(allTools, articleTools...)
	} else {
		log.Printf("获取article-mcp工具失败: %v", err)
	}

	// 获取context7工具
	if context7Tools, err := amb.getContext7Tools(manager); err == nil {
		allTools = append(allTools, context7Tools...)
	} else {
		log.Printf("获取context7工具失败: %v", err)
	}

	return allTools, nil
}

// getArticleMCPTools 获取article-mcp工具
func (amb *AIMCPBridge) getArticleMCPTools(manager *MCPManager) ([]MCPTool, error) {
	// 启动article-mcp服务器
	if err := manager.StartServer("article-mcp"); err != nil {
		return nil, fmt.Errorf("启动article-mcp失败: %w", err)
	}

	// 预定义article-mcp工具
	return []MCPTool{
		{
			Server: "article-mcp",
			Name:   "search_europe_pmc",
			Desc:   "搜索欧洲生物医学文献数据库",
			Arguments: map[string]interface{}{
				"keyword": map[string]interface{}{
					"type": "string",
					"description": "搜索关键词",
					"required": true,
				},
				"max_results": map[string]interface{}{
					"type": "integer",
					"description": "最大结果数",
					"default": 10,
				},
			},
		},
		{
			Server: "article-mcp",
			Name:   "search_arxiv_papers",
			Desc:   "搜索arXiv预印本论文",
			Arguments: map[string]interface{}{
				"keyword": map[string]interface{}{
					"type": "string",
					"description": "搜索关键词",
					"required": true,
				},
				"max_results": map[string]interface{}{
					"type": "integer",
					"description": "最大结果数",
					"default": 5,
				},
			},
		},
		{
			Server: "article-mcp",
			Name:   "get_article_details",
			Desc:   "获取文献详细信息",
			Arguments: map[string]interface{}{
				"identifier": map[string]interface{}{
					"type": "string",
					"description": "文献标识符(PMID/DOI/PMCID)",
					"required": true,
				},
			},
		},
	}, nil
}

// getContext7Tools 获取context7工具
func (amb *AIMCPBridge) getContext7Tools(manager *MCPManager) ([]MCPTool, error) {
	// 检查context7是否启用
	config, err := manager.GetServerConfig("context7")
	if err != nil {
		return nil, fmt.Errorf("获取context7配置失败: %w", err)
	}

	if !config.Enabled {
		return nil, fmt.Errorf("context7未启用")
	}

	// 启动context7服务器
	if err := manager.StartServer("context7"); err != nil {
		return nil, fmt.Errorf("启动context7失败: %w", err)
	}

	// 预定义context7工具
	return []MCPTool{
		{
			Server: "context7",
			Name:   "get-library-docs",
			Desc:   "获取编程库文档",
			Arguments: map[string]interface{}{
				"context7CompatibleLibraryID": map[string]interface{}{
					"type": "string",
					"description": "Context7兼容的库ID (如: /mongodb/docs, /vercel/next.js)",
					"required": true,
				},
				"topic": map[string]interface{}{
					"type": "string",
					"description": "要关注的主题 (如: hooks, routing)",
					"required": false,
				},
				"tokens": map[string]interface{}{
					"type": "integer",
					"description": "最大令牌数 (默认: 5000)",
					"required": false,
				},
			},
		},
		{
			Server: "context7",
			Name:   "resolve-library-id",
			Desc:   "解析库标识符",
			Arguments: map[string]interface{}{
				"libraryName": map[string]interface{}{
					"type": "string",
					"description": "库名称",
					"required": true,
				},
			},
		},
	}, nil
}

// SelectTool 选择并调用工具（优化版：先快速匹配，再AI分析）
func (amb *AIMCPBridge) SelectTool(message string) (*ToolCall, *string, error) {
	log.Printf("🧠 开始分析用户查询: %s", message)

	// 1. 快速工具选择（基于测试脚本的成功逻辑）
	if toolCall := amb.quickToolSelection(message); toolCall != nil {
		log.Printf("⚡ 快速匹配到工具: %s.%s", toolCall.Server, toolCall.Tool)
		return toolCall, nil, nil
	}

	// 2. AI工具选择（处理复杂情况）
	toolCall, aiResponse, err := amb.aiBasedSelection(message)
	if err != nil {
		log.Printf("❌ AI工具选择失败: %v", err)
		return nil, nil, err
	}

	if toolCall != nil {
		log.Printf("🤖 AI选择工具: %s.%s", toolCall.Server, toolCall.Tool)
		return toolCall, &aiResponse, nil
	}

	// 3. 不需要工具
	log.Printf("💬 AI判断无需工具，直接回答")
	return nil, &aiResponse, nil
}

// quickToolSelection 快速工具选择（基于测试脚本的成功逻辑，增加明确规则）
func (amb *AIMCPBridge) quickToolSelection(message string) *ToolCall {
	message = strings.ToLower(message)

	// 明确的MCP工具调用规则（初级版本）
	switch {
	// Context7工具 - 编程相关查询
	case strings.Contains(message, "库") || strings.Contains(message, "library") ||
		 strings.Contains(message, "文档") || strings.Contains(message, "documentation") ||
		 strings.Contains(message, "代码") || strings.Contains(message, "code") ||
		 strings.Contains(message, "api") || strings.Contains(message, "tutorial") ||
		 strings.Contains(message, "示例") || strings.Contains(message, "example"):

		// 编程语言和框架关键词
		programmingKeywords := []string{
			"react", "vue", "angular", "javascript", "typescript", "python", "java",
			"node", "express", "django", "flask", "spring", "golang", "rust",
			"docker", "kubernetes", "aws", "azure", "firebase", "mongodb",
			"mysql", "postgresql", "redis", "graphql", "rest", "http",
		}

		for _, keyword := range programmingKeywords {
			if strings.Contains(message, keyword) {
				log.Printf("🔧 快速匹配: 使用context7的resolve-library-id工具")
				return &ToolCall{
					Server: "context7",
					Tool:   "resolve-library-id",
					Arguments: map[string]interface{}{
						"libraryName": amb.extractLibraryName(message),
					},
				}
			}
		}

	// Article-MCP工具 - 学术文献查询
	case strings.Contains(message, "搜索") || strings.Contains(message, "search"):
		if strings.Contains(message, "预印本") || strings.Contains(message, "preprint") || strings.Contains(message, "arxiv") {
			log.Printf("🔧 快速匹配: 使用search_arxiv_papers工具")
			return &ToolCall{
				Server: "article-mcp",
				Tool:   "search_arxiv_papers",
				Arguments: map[string]interface{}{
					"keyword":     amb.extractKeyword(message),
					"max_results": 5,
				},
			}
		} else {
			log.Printf("🔧 快速匹配: 使用search_europe_pmc工具")
			return &ToolCall{
				Server: "article-mcp",
				Tool:   "search_europe_pmc",
				Arguments: map[string]interface{}{
					"keyword":     amb.extractKeyword(message),
					"max_results": 5,
				},
			}
		}

	case strings.Contains(message, "论文") || strings.Contains(message, "文献"):
		if strings.Contains(message, "10.") && strings.Contains(message, "/") {
			log.Printf("🔧 快速匹配: 使用get_article_details工具")
			return &ToolCall{
				Server: "article-mcp",
				Tool:   "get_article_details",
				Arguments: map[string]interface{}{
					"identifier": amb.extractDOI(message),
				},
			}
		}

	case strings.Contains(message, "相似") || strings.Contains(message, "类似") || strings.Contains(message, "similar"):
		if strings.Contains(message, "10.") && strings.Contains(message, "/") {
			log.Printf("🔧 快速匹配: 使用get_similar_articles工具")
			return &ToolCall{
				Server: "article-mcp",
				Tool:   "get_similar_articles",
				Arguments: map[string]interface{}{
					"identifier": amb.extractDOI(message),
					"max_results": 3,
				},
			}
		}

	case strings.Contains(message, "引用") || strings.Contains(message, "cite"):
		if strings.Contains(message, "10.") && strings.Contains(message, "/") {
			log.Printf("🔧 快速匹配: 使用get_citing_articles工具")
			return &ToolCall{
				Server: "article-mcp",
				Tool:   "get_citing_articles",
				Arguments: map[string]interface{}{
					"identifier": amb.extractDOI(message),
					"max_results": 3,
				},
			}
		}

	case strings.Contains(message, "期刊") || strings.Contains(message, "影响因子") || strings.Contains(message, "分区"):
		log.Printf("🔧 快速匹配: 使用get_journal_quality工具")
		journalName := amb.extractJournalName(message)
		return &ToolCall{
			Server: "article-mcp",
			Tool:   "get_journal_quality",
			Arguments: map[string]interface{}{
				"journal_name": journalName,
			},
		}
	}

	return nil
}

// aiBasedSelection 基于AI的工具选择（处理复杂情况）
func (amb *AIMCPBridge) aiBasedSelection(message string) (*ToolCall, string, error) {
	// 1. 获取可用工具
	tools, err := amb.GetAvailableTools()
	if err != nil {
		return nil, "", err
	}

	if len(tools) == 0 {
		return nil, "", fmt.Errorf("没有可用的MCP工具")
	}

	// 2. 构建明确的AI提示（初级版本：明确指定工具调用规则）
	toolsPrompt := amb.formatToolsForAI(tools)
	systemPrompt := fmt.Sprintf(`你是一个专业的学术研究助手，可以帮助用户查找和分析学术文献。你有以下MCP工具可以使用：

%s

## 明确的工具调用规则：

**当用户询问以下内容时，必须使用对应的工具：**

1. **搜索论文/文献** → 必须使用：
   - "search_europe_pmc"（用于一般学术文献搜索）
   - "search_arxiv_papers"（专门搜索预印本，包含arxiv、preprint关键词时）

2. **获取论文详情** → 必须使用：
   - "get_article_details"（当用户提供DOI或PMID时）

3. **查找相似研究** → 必须使用：
   - "get_similar_articles"（当用户询问"相似"、"类似"、"similar"研究时）

4. **查找引用文献** → 必须使用：
   - "get_citing_articles"（当用户询问"引用"、"cite"时）

5. **期刊信息查询** → 必须使用：
   - "get_journal_quality"（当用户询问期刊、影响因子、分区时）

**回复格式：**
如果需要使用工具，请严格按照：
TOOL: <工具名>
ARGS: <JSON格式的参数>

如果不需要使用工具（比如询问概念、定义等），请直接回答用户问题。

**重要提示：**
- 不要猜测，严格按照上述规则选择工具
- 参数必须是有效的JSON格式
- keyword参数提取用户查询中的核心关键词
- max_results建议使用3-10

**示例：**
用户：搜索机器学习相关论文
TOOL: search_europe_pmc
ARGS: {"keyword": "machine learning", "max_results": 5}

用户：查找和10.1038/nature12373相似的研究
TOOL: get_similar_articles
ARGS: {"identifier": "10.1038/nature12373", "max_results": 3}

用户：什么是深度学习？
（直接回答，不需要工具）
`, toolsPrompt)

	// 3. AI分析
	aiResponse := amb.callAIForToolSelection(message, systemPrompt)

	// 4. 解析结果
	if toolCall := amb.parseToolSelection(aiResponse); toolCall != nil {
		return toolCall, aiResponse, nil
	}

	return nil, aiResponse, nil
}

// extractKeyword 提取关键词
func (amb *AIMCPBridge) extractKeyword(message string) string {
	keywords := []string{"机器学习", "深度学习", "人工智能", "machine learning", "deep learning", "artificial intelligence", "计算机视觉", "自然语言处理"}
	for _, keyword := range keywords {
		if strings.Contains(strings.ToLower(message), keyword) {
			return keyword
		}
	}

	// 如果没有匹配到关键词，提取查询中的主要词汇
	words := strings.Fields(message)
	for _, word := range words {
		if len(word) > 2 && !strings.Contains(word, "搜索") && !strings.Contains(word, "相关") {
			return word
		}
	}

	return "research"
}

// extractDOI 提取DOI
func (amb *AIMCPBridge) extractDOI(message string) string {
	if strings.Contains(message, "10.") && strings.Contains(message, "/") {
		words := strings.Fields(message)
		for _, word := range words {
			if strings.HasPrefix(word, "10.") && len(strings.Split(word, "/")) >= 2 {
				return word
			}
		}
	}
	return "test-doi"
}

// extractJournalName 从消息中提取期刊名称
func (amb *AIMCPBridge) extractJournalName(message string) string {
	// 简单的期刊名称提取逻辑
	if strings.Contains(strings.ToLower(message), "nature") {
		return "Nature"
	} else if strings.Contains(strings.ToLower(message), "science") {
		return "Science"
	} else if strings.Contains(strings.ToLower(message), "cell") {
		return "Cell"
	} else if strings.Contains(strings.ToLower(message), "lancet") {
		return "The Lancet"
	}

	// 如果没有匹配到已知期刊，尝试提取第一个可能是期刊名称的词
	words := strings.Fields(message)
	for _, word := range words {
		if len(word) > 3 && strings.Contains(word, strings.Title(word)) {
			return word
		}
	}

	return "Journal Name"
}

// extractQuery 从消息中提取查询内容（用于context7）
func (amb *AIMCPBridge) extractQuery(message string) string {
	// 移除常见的引导词
	queries := []string{
		"请帮我", "请帮我查找", "查找", "搜索", "帮我找", "我需要", "我想了解",
		"please help me", "find", "search", "get", "show me",
	}

	lowerMessage := strings.ToLower(message)
	for _, query := range queries {
		if strings.HasPrefix(lowerMessage, query) {
			return strings.TrimSpace(message[len(query):])
		}
	}

	// 提取编程相关的关键词
	programmingKeywords := []string{
		"react", "vue", "angular", "javascript", "typescript", "python", "java",
		"node", "express", "django", "flask", "spring", "golang", "rust",
		"docker", "kubernetes", "aws", "azure", "firebase", "mongodb",
		"mysql", "postgresql", "redis", "graphql", "rest", "http",
	}

	for _, keyword := range programmingKeywords {
		if strings.Contains(lowerMessage, keyword) {
			// 返回包含关键词的整个查询
			return message
		}
	}

	return message
}

// extractLibraryName 从消息中提取库名称（用于context7）
func (amb *AIMCPBridge) extractLibraryName(message string) string {
	// 移除常见的引导词
	queries := []string{
		"请帮我", "请帮我查找", "查找", "搜索", "帮我找", "我需要", "我想了解", "请帮我找",
		"的文档", "的API文档", "的文档", "库", "library", "documentation",
	}

	lowerMessage := strings.ToLower(message)
	result := message

	// 移除引导词
	for _, query := range queries {
		if strings.HasPrefix(lowerMessage, query) {
			result = strings.TrimSpace(result[len(query):])
			lowerMessage = strings.ToLower(result)
		}
	}

	// 移除尾部词汇
	tailWords := []string{"的文档", "的api文档", "的文档", "库", "library", "documentation"}
	for _, word := range tailWords {
		if strings.HasSuffix(lowerMessage, word) && len(result) > len(word) {
			result = strings.TrimSpace(result[:len(result)-len(word)])
		}
	}

	// 提取编程相关的关键词
	programmingKeywords := []string{
		"react", "vue", "angular", "javascript", "typescript", "python", "java",
		"node", "express", "django", "flask", "spring", "golang", "rust",
		"docker", "kubernetes", "aws", "azure", "firebase", "mongodb",
		"mysql", "postgresql", "redis", "graphql", "rest", "http",
	}

	for _, keyword := range programmingKeywords {
		if strings.Contains(lowerMessage, keyword) {
			// 返回找到的关键词
			return keyword
		}
	}

	// 如果没找到，返回清理后的结果
	if result == "" {
		return "react" // 默认值
	}

	return result
}

// formatToolsForAI 格式化工具信息供AI使用
func (amb *AIMCPBridge) formatToolsForAI(tools []MCPTool) string {
	var builder strings.Builder

	for _, tool := range tools {
		builder.WriteString(fmt.Sprintf("- **%s** (来自%s)\n", tool.Name, tool.Server))
		builder.WriteString(fmt.Sprintf("  %s\n", tool.Desc))

		if len(tool.Arguments) > 0 {
			builder.WriteString("  参数:\n")
			for name, arg := range tool.Arguments {
				if argMap, ok := arg.(map[string]interface{}); ok {
					builder.WriteString(fmt.Sprintf("    %s: %s", name, argMap["type"]))
					if desc, ok := argMap["description"].(string); ok {
						builder.WriteString(fmt.Sprintf(" - %s", desc))
					}
					if required, ok := argMap["required"].(bool); ok && required {
						builder.WriteString(" (必需)")
					}
					if defVal, ok := argMap["default"]; ok {
						builder.WriteString(fmt.Sprintf(" (默认: %v)", defVal))
					}
					builder.WriteString("\n")
				}
			}
		}
		builder.WriteString("\n")
	}

	return builder.String()
}

// callAIForToolSelection 调用AI进行工具选择
func (amb *AIMCPBridge) callAIForToolSelection(message, systemPrompt string) string {
	messages := []core.ChatMessage{
		{
			Role:    "system",
			Content: systemPrompt,
		},
		{
			Role:    "user",
			Content: message,
		},
	}

	req := &core.AIRequest{
		Model:    amb.config.AIModel,
		Messages: messages,
		MaxTokens: 300, // 限制长度，避免冗长的回复
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(amb.config.AITimeout)*time.Second)
	defer cancel()

	response, err := amb.aiClient.Chat(ctx, req)
	if err != nil {
		log.Printf("AI工具选择失败: %v", err)
		return ""
	}

	if len(response.Choices) > 0 {
		return response.Choices[0].Message.Content
	}

	return ""
}

// parseToolSelection 解析AI的工具选择
func (amb *AIMCPBridge) parseToolSelection(aiResponse string) *ToolCall {
	// 查找TOOL指令
	toolRegex := regexp.MustCompile(`(?i)TOOL:\s*(\S+)`)
	toolMatch := toolRegex.FindStringSubmatch(aiResponse)
	if toolMatch == nil {
		return nil
	}
	toolName := toolMatch[1]

	// 查找ARGS指令
	argsRegex := regexp.MustCompile(`(?i)ARGS:\s*(\{[\s\S]*\})`)
	argsMatch := argsRegex.FindStringSubmatch(aiResponse)
	if argsMatch == nil {
		return nil
	}
	argsStr := argsMatch[1]

	// 解析JSON参数
	var args map[string]interface{}
	if err := json.Unmarshal([]byte(argsStr), &args); err != nil {
		log.Printf("解析工具参数失败: %v", err)
		return nil
	}

	// 确定服务器
	var server string
	for _, tool := range amb.getAvailableToolList() {
		if tool.Name == toolName {
			server = tool.Server
			break
		}
	}

	if server == "" {
		return nil
	}

	return &ToolCall{
		Server:    server,
		Tool:      toolName,
		Arguments: args,
	}
}

// getAvailableToolList 获取可用工具列表（内部方法）
func (amb *AIMCPBridge) getAvailableToolList() []MCPTool {
	tools, _ := amb.GetAvailableTools()
	return tools
}

// getMCPManager 获取或创建MCP管理器（连接复用）
func (amb *AIMCPBridge) getMCPManager() (*MCPManager, error) {
	amb.managerOnce.Do(func() {
		manager, err := NewMCPManager("mcp/mcp_config.json")
		if err != nil {
			amb.initError = fmt.Errorf("创建MCP管理器失败: %w", err)
			return
		}
		amb.mcpManager = manager
		log.Printf("✅ MCP管理器初始化成功")
	})

	if amb.initError != nil {
		return nil, amb.initError
	}

	return amb.mcpManager, nil
}

// CallTool 调用MCP工具（支持缓存）
func (amb *AIMCPBridge) CallTool(toolCall *ToolCall) (*MCPResponse, error) {
	startTime := time.Now()

	log.Printf("🤖 [AI-MCP] 开始调用工具: %s.%s", toolCall.Server, toolCall.Tool)
	log.Printf("📥 [AI-MCP输入] 工具参数: %s", amb.formatArgumentsForLog(toolCall.Arguments))

	// 生成缓存键
	cacheKey := amb.generateCacheKey(toolCall)
	log.Printf("🔑 [AI-MCP缓存] 缓存键: %s", cacheKey)

	// 检查缓存（对于搜索类工具，缓存5分钟）
	if amb.shouldUseCache(toolCall.Tool) {
		if cachedResponse, found := amb.cache.Get(cacheKey); found {
			duration := time.Since(startTime)
			log.Printf("🎯 [AI-MCP缓存] 缓存命中，节省耗时: %v", duration)
			log.Printf("📤 [AI-MCP输出] 返回缓存结果")
			return cachedResponse, nil
		} else {
			log.Printf("🔍 [AI-MCP缓存] 缓存未命中，执行实际调用")
		}
	} else {
		log.Printf("⚡ [AI-MCP缓存] 工具 %s 不使用缓存", toolCall.Tool)
	}

	// 获取MCP管理器（复用连接）
	log.Printf("🔧 [AI-MCP管理器] 获取MCP管理器...")
	manager, err := amb.getMCPManager()
	if err != nil {
		log.Printf("❌ [AI-MCP错误] 获取MCP管理器失败: %v", err)
		return nil, fmt.Errorf("获取MCP管理器失败: %w", err)
	}
	log.Printf("✅ [AI-MCP管理器] MCP管理器获取成功")

	// 启动对应的服务器（如果未启动）
	log.Printf("🚀 [AI-MCP服务器] 启动服务器: %s", toolCall.Server)
	if err := manager.StartServer(toolCall.Server); err != nil {
		log.Printf("❌ [AI-MCP错误] 启动MCP服务器失败: %v", err)
		return nil, fmt.Errorf("启动MCP服务器失败: %w", err)
	}
	log.Printf("✅ [AI-MCP服务器] 服务器启动成功")

	// 调用工具
	log.Printf("⚡ [AI-MCP执行] 执行工具调用...")
	response, err := manager.CallTool(toolCall.Server, toolCall.Tool, toolCall.Arguments)

	duration := time.Since(startTime)
	log.Printf("⏱️ [AI-MCP性能] 工具调用总耗时: %v", duration)

	if err != nil {
		log.Printf("❌ [AI-MCP错误] 工具调用失败: %v", err)
		log.Printf("📊 [AI-MCP统计] 失败统计 - 服务器: %s, 工具: %s, 耗时: %v",
			toolCall.Server, toolCall.Tool, duration)
		return nil, fmt.Errorf("MCP工具调用失败 (服务器: %s, 工具: %s, 耗时: %v): %w",
			toolCall.Server, toolCall.Tool, duration, err)
	}

	log.Printf("✅ [AI-MCP成功] 工具调用成功")
	log.Printf("📤 [AI-MCP输出] 准备解析结果...")

	// 记录结果摘要
	if response != nil && response.Result != nil {
		resultSize := len(response.Result)
		log.Printf("📊 [AI-MCP统计] 成功统计 - 响应大小: %d 字节, 耗时: %v", resultSize, duration)
	}

	// 设置缓存
	if amb.shouldUseCache(toolCall.Tool) {
		// 搜索类工具缓存5分钟，其他工具缓存30分钟
		ttl := 30 * time.Minute
		if strings.Contains(toolCall.Tool, "search") {
			ttl = 5 * time.Minute
		}
		amb.cache.Set(cacheKey, response, ttl)
		log.Printf("💾 [AI-MCP缓存] 结果已缓存，TTL: %v", ttl)
	}

	log.Printf("🎉 [AI-MCP完成] 工具调用流程完成")
	return response, nil
}

// shouldUseCache 判断是否应该使用缓存
func (amb *AIMCPBridge) shouldUseCache(toolName string) bool {
	// 搜索类工具使用缓存
	cacheableTools := []string{
		"search_europe_pmc",
		"search_arxiv_papers",
		"get_article_details",
		"get_similar_articles",
	}

	for _, cacheable := range cacheableTools {
		if toolName == cacheable {
			return true
		}
	}

	return false
}

// Close 关闭桥接器
func (amb *AIMCPBridge) Close() error {
	if amb.mcpManager != nil {
		return amb.mcpManager.Close()
	}
	return nil
}

// ParseToolResult 解析工具调用结果
func (amb *AIMCPBridge) ParseToolResult(response *MCPResponse) string {
	log.Printf("📋 [AI-MCP解析] 开始解析工具结果...")

	if response == nil {
		log.Printf("❌ [AI-MCP解析] 响应为空")
		return "工具调用未返回结果"
	}

	if response.Result == nil {
		log.Printf("❌ [AI-MCP解析] 响应结果为空")
		return "工具调用未返回结果"
	}

	log.Printf("📊 [AI-MCP解析] 原始结果大小: %d 字节", len(response.Result))

	// 尝试解析为字符串
	var resultStr string
	if err := json.Unmarshal(response.Result, &resultStr); err == nil {
		log.Printf("✅ [AI-MCP解析] 成功解析为字符串，长度: %d", len(resultStr))
		if len(resultStr) > 100 {
			log.Printf("📄 [AI-MCP解析] 字符串内容摘要: %s...", resultStr[:100])
		} else {
			log.Printf("📄 [AI-MCP解析] 字符串内容: %s", resultStr)
		}
		return resultStr
	}

	// 尝试解析为结构化数据
	var resultData interface{}
	if err := json.Unmarshal(response.Result, &resultData); err == nil {
		log.Printf("✅ [AI-MCP解析] 成功解析为结构化数据")
		if formatted := amb.formatStructuredResult(resultData); formatted != "" {
			log.Printf("📄 [AI-MCP解析] 格式化结果长度: %d", len(formatted))
			return formatted
		}
	}

	// 返回原始JSON
	log.Printf("⚠️ [AI-MCP解析] 无法解析，返回原始JSON")
	return string(response.Result)
}

// formatArgumentsForLog 格式化参数用于日志记录
func (amb *AIMCPBridge) formatArgumentsForLog(arguments map[string]interface{}) string {
	if arguments == nil {
		return "{}"
	}

	jsonBytes, err := json.MarshalIndent(arguments, "", "  ")
	if err != nil {
		return fmt.Sprintf("%+v", arguments)
	}
	return string(jsonBytes)
}

// formatStructuredResult 格式化结构化结果
func (amb *AIMCPBridge) formatStructuredResult(data interface{}) string {
	switch v := data.(type) {
	case map[string]interface{}:
		var builder strings.Builder
		for key, value := range v {
			builder.WriteString(fmt.Sprintf("%s: %v\n", key, value))
		}
		return builder.String()

	case []interface{}:
		var builder strings.Builder
		for i, item := range v {
			builder.WriteString(fmt.Sprintf("%d. %v\n", i+1, item))
		}
		return builder.String()

	default:
		return fmt.Sprintf("%v", data)
	}
}

// GenerateFinalAnswer 生成最终答案
func (amb *AIMCPBridge) GenerateFinalAnswer(userMessage, toolResult, aiResponse *string) string {
	if aiResponse != nil && *aiResponse != "" {
		// AI已经给出了基于工具结果的答案
		return *aiResponse
	}

	// 如果AI没有回复，基于工具结果生成答案
	if toolResult != nil {
		return fmt.Sprintf(`根据搜索结果，我为您找到了以下信息：

%s

您需要更详细的信息吗？`, *toolResult)
	}

	return "抱歉，未能获取到相关信息。"
}