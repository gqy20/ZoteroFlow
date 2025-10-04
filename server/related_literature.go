package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"zoteroflow2-server/config"
	"zoteroflow2-server/core"
)

// handleRelatedLiterature 处理相关文献分析命令
func handleRelatedLiterature(args []string) {
	if len(args) < 1 {
		log.Fatal("用法: related <文献名称/DOI> [问题]")
	}

	docIdentifier := args[0]
	question := "请分析这篇文献并找到相关研究" // 默认问题
	if len(args) >= 2 {
		question = strings.Join(args[1:], " ")
	}

	fmt.Printf("🔍 正在分析文献: %s\n", docIdentifier)
	fmt.Printf("❓ 用户问题: %s\n", question)
	fmt.Println()

	// 加载配置
	cfg, err := config.Load()
	if err != nil {
		log.Printf("❌ 配置加载失败: %v", err)
		return
	}

	// 1. 本地文献查找
	fmt.Println("📚 步骤1: 查找本地文献...")
	localDocs, err := findLocalDocuments(docIdentifier, cfg)
	if err != nil {
		log.Printf("❌ 本地文献查找失败: %v", err)
	} else {
		fmt.Printf("✅ 找到 %d 篇本地文献\n", len(localDocs))
		for i, doc := range localDocs {
			if i < 3 { // 只显示前3篇
				fmt.Printf("   - %s (%s)\n", doc.Title, doc.Authors)
			}
		}
		if len(localDocs) > 3 {
			fmt.Printf("   ... 还有 %d 篇文献\n", len(localDocs)-3)
		}
	}

	// 2. 全球文献搜索
	fmt.Println("\n🌐 步骤2: 搜索全球相关文献...")
	globalDocs, err := searchGlobalLiterature(docIdentifier, cfg)
	if err != nil {
		log.Printf("❌ 全球文献搜索失败: %v", err)
	} else {
		fmt.Printf("✅ 找到 %d 篇全球文献\n", len(globalDocs))
		for i, doc := range globalDocs {
			if i < 3 { // 只显示前3篇
				fmt.Printf("   - %s (%s)\n", doc.Title, doc.Journal)
			}
		}
		if len(globalDocs) > 3 {
			fmt.Printf("   ... 还有 %d 篇文献\n", len(globalDocs)-3)
		}
	}

	// 3. AI智能分析
	fmt.Println("\n🤖 步骤3: AI智能分析...")
	aiAnalysis, err := performAIAnalysis(docIdentifier, question, localDocs, globalDocs, cfg)
	if err != nil {
		log.Printf("❌ AI分析失败: %v", err)
		fmt.Println("⚠️  AI分析失败，但文献搜索已完成")
	} else {
		fmt.Println("✅ AI分析完成")
	}

	// 4. 结果展示
	fmt.Println("\n=== 分析结果 ===")

	if len(localDocs) > 0 || len(globalDocs) > 0 {
		fmt.Println("\n📋 相关文献详情:")

		// 显示本地文献
		for i, doc := range localDocs {
			if i < 5 { // 显示前5篇
				fmt.Printf("\n%d. %s\n", i+1, doc.Title)
				fmt.Printf("   作者: %s\n", doc.Authors)
				if doc.Journal != "" {
					fmt.Printf("   期刊: %s\n", doc.Journal)
				}
				if doc.Year != 0 {
					fmt.Printf("   年份: %d\n", doc.Year)
				}
				if doc.DOI != "" {
					fmt.Printf("   DOI: %s\n", doc.DOI)
				}
			}
		}

		// 显示全球文献
		startIdx := len(localDocs)
		for i, doc := range globalDocs {
			if i < 5 { // 显示前5篇
				fmt.Printf("\n%d. %s\n", startIdx+i+1, doc.Title)
				fmt.Printf("   作者: %s\n", doc.Authors)
				if doc.Journal != "" {
					fmt.Printf("   期刊: %s\n", doc.Journal)
				}
				if doc.Year != 0 {
					fmt.Printf("   年份: %d\n", doc.Year)
				}
				if doc.DOI != "" {
					fmt.Printf("   DOI: %s\n", doc.DOI)
				}
			}
		}
	}

	// 显示AI分析结果
	if aiAnalysis != "" {
		fmt.Printf("\n=== AI分析结果 ===\n%s\n", aiAnalysis)
	}
}

// DocumentSummary 文档摘要
type DocumentSummary struct {
	Title    string
	Authors  string
	Journal  string
	Year     int
	DOI      string
	Abstract string
}

// findLocalDocuments 查找本地文献
func findLocalDocuments(identifier string, cfg *config.Config) ([]DocumentSummary, error) {
	// 连接Zotero数据库
	zoteroDB, err := core.NewZoteroDB(cfg.ZoteroDBPath, cfg.ZoteroDataDir)
	if err != nil {
		return nil, fmt.Errorf("连接Zotero数据库失败: %w", err)
	}
	defer zoteroDB.Close()

	// 搜索文献
	var docs []DocumentSummary

	// 尝试DOI搜索 (目前使用标题搜索替代)
	if strings.Contains(identifier, "10.") && strings.Contains(identifier, "/") {
		results, err := zoteroDB.SearchByTitle(identifier, 1)
		if err == nil && len(results) > 0 {
			doc := results[0]
			docs = append(docs, DocumentSummary{
				Title:   doc.Title,
				Authors: strings.Join(doc.Authors, "; "),
				Journal: doc.Journal,
				Year:    doc.Year,
				DOI:     doc.DOI,
			})
		}
	}

	// 尝试标题搜索
	results, err := zoteroDB.SearchByTitle(identifier, 5)
	if err == nil {
		for _, result := range results {
			docs = append(docs, DocumentSummary{
				Title:   result.Title,
				Authors: strings.Join(result.Authors, "; "),
				Journal: result.Journal,
				Year:    result.Year,
				DOI:     result.DOI,
			})
		}
	}

	return docs, nil
}

// searchGlobalLiterature 搜索全球文献
func searchGlobalLiterature(identifier string, cfg *config.Config) ([]DocumentSummary, error) {
	// 创建MCP管理器
	manager, err := NewMCPManager("mcp_config.json")
	if err != nil {
		return nil, fmt.Errorf("创建MCP管理器失败: %w", err)
	}
	defer manager.Close()

	// 启动article-mcp服务器
	if err := manager.StartServer("article-mcp"); err != nil {
		return nil, fmt.Errorf("启动article-mcp服务器失败: %w", err)
	}

	// 调用search_europe_pmc工具
	response, err := manager.CallTool("article-mcp", "search_europe_pmc", map[string]interface{}{
		"keyword":     identifier,
		"max_results": 10,
	})
	if err != nil {
		return nil, fmt.Errorf("调用search_europe_pmc失败: %w", err)
	}

	// 解析响应
	docs := parseMCPResponse(response)
	log.Printf("Article MCP搜索找到 %d 篇文献", len(docs))
	return docs, nil
}

// parseMCPResponse 解析MCP响应
func parseMCPResponse(response *MCPResponse) []DocumentSummary {
	if response == nil || response.Result == nil {
		return []DocumentSummary{}
	}

	// 尝试解析结果
	var result struct {
		Articles []DocumentSummary `json:"articles"`
		Results  []DocumentSummary `json:"results"`
	}

	if err := json.Unmarshal(response.Result, &result); err != nil {
		log.Printf("解析MCP响应失败: %v", err)
		return []DocumentSummary{}
	}

	// 返回找到的文章
	if len(result.Articles) > 0 {
		return result.Articles
	}
	if len(result.Results) > 0 {
		return result.Results
	}

	return []DocumentSummary{}
}

// translateToEnglish 简单的中英文关键词翻译
func translateToEnglish(keyword string) string {
	translations := map[string]string{
		"机器学习":  "machine learning",
		"深度学习":  "deep learning",
		"神经网络":  "neural networks",
		"人工智能":  "artificial intelligence",
		"数据科学":  "data science",
		"基因组学":  "genomics",
		"生物信息学": "bioinformatics",
		"遗传学":   "genetics",
		"分子生物学": "molecular biology",
	}

	if english, exists := translations[keyword]; exists {
		return english
	}

	// 如果已经是英文，直接返回
	return keyword
}

// performAIAnalysis 执行AI分析
func performAIAnalysis(identifier, question string, localDocs, globalDocs []DocumentSummary, cfg *config.Config) (string, error) {
	// 检查AI配置
	if cfg.AIAPIKey == "" {
		return "", fmt.Errorf("AI功能未配置")
	}

	// 创建AI客户端
	client := core.NewGLMClient(cfg.AIAPIKey, cfg.AIBaseURL, cfg.AIModel)

	// 构建分析上下文
	analysisContext := buildAnalysisContext(identifier, question, localDocs, globalDocs)

	// 创建对话请求
	messages := []core.ChatMessage{
		{
			Role:    "system",
			Content: "你是一个专业的学术文献分析师，请基于提供的文献信息进行智能分析和回答。",
		},
		{
			Role:    "user",
			Content: analysisContext,
		},
	}

	// 发送请求
	req := &core.AIRequest{
		Model:     cfg.AIModel,
		Messages:  messages,
		MaxTokens: 1000, // 增加输出长度限制
	}

	// 设置100秒超时
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	response, err := client.Chat(ctx, req)
	if err != nil {
		return "", fmt.Errorf("AI分析失败: %w", err)
	}

	if len(response.Choices) > 0 {
		return response.Choices[0].Message.Content, nil
	}

	return "", fmt.Errorf("未收到AI响应")
}

// buildAnalysisContext 构建分析上下文
func buildAnalysisContext(identifier, question string, localDocs, globalDocs []DocumentSummary) string {
	var contextStr strings.Builder

	contextStr.WriteString(fmt.Sprintf("请分析关于'%s'的文献并回答问题: %s\n\n", identifier, question))

	if len(localDocs) > 0 {
		contextStr.WriteString("本地文献:\n")
		for i, doc := range localDocs {
			if i < 3 { // 只使用前3篇构建上下文
				contextStr.WriteString(fmt.Sprintf("%d. %s\n", i+1, doc.Title))
				contextStr.WriteString(fmt.Sprintf("   作者: %s\n", doc.Authors))
				if doc.Journal != "" {
					contextStr.WriteString(fmt.Sprintf("   期刊: %s\n", doc.Journal))
				}
				if doc.Year != 0 {
					contextStr.WriteString(fmt.Sprintf("   年份: %d\n", doc.Year))
				}
				contextStr.WriteString("\n")
			}
		}
	}

	if len(globalDocs) > 0 {
		contextStr.WriteString("全球相关文献:\n")
		for i, doc := range globalDocs {
			if i < 3 { // 只使用前3篇构建上下文
				contextStr.WriteString(fmt.Sprintf("%d. %s\n", i+1, doc.Title))
				contextStr.WriteString(fmt.Sprintf("   作者: %s\n", doc.Authors))
				if doc.Journal != "" {
					contextStr.WriteString(fmt.Sprintf("   期刊: %s\n", doc.Journal))
				}
				if doc.Year != 0 {
					contextStr.WriteString(fmt.Sprintf("   年份: %d\n", doc.Year))
				}
				contextStr.WriteString("\n")
			}
		}
	}

	contextStr.WriteString("请基于以上文献信息，提供专业的分析和回答。")

	return contextStr.String()
}
