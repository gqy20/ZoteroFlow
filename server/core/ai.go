package core

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Conversation 对话结构
type Conversation struct {
	ID        string           `json:"id"`
	Messages  []ChatMessage    `json:"messages"`
	Context   *DocumentContext `json:"context,omitempty"`
	CreatedAt time.Time        `json:"created_at"`
	UpdatedAt time.Time        `json:"updated_at"`
}

// ChatMessage 聊天消息
type ChatMessage struct {
	Role      string           `json:"role"` // user, assistant, system
	Content   string           `json:"content"`
	Timestamp time.Time        `json:"timestamp"`
	Metadata  *MessageMetadata `json:"metadata,omitempty"`
}

// MessageMetadata 消息元数据
type MessageMetadata struct {
	DocumentIDs []int  `json:"document_ids,omitempty"` // 关联的文献ID
	QueryType   string `json:"query_type,omitempty"`   // search, analysis, summary
}

// DocumentContext 文档上下文
type DocumentContext struct {
	Documents []DocumentSummary `json:"documents"`
	Query     string            `json:"query"`
	Relevance float64           `json:"relevance"`
}

// DocumentSummary 文档摘要
type DocumentSummary struct {
	ID       int      `json:"id"`
	Title    string   `json:"title"`
	Authors  string   `json:"authors"`
	Abstract string   `json:"abstract"`
	Keywords []string `json:"keywords"`
}

// AIRequest AI 请求结构
type AIRequest struct {
	Model       string        `json:"model"`
	Messages    []ChatMessage `json:"messages"`
	Stream      bool          `json:"stream,omitempty"`
	Temperature float64       `json:"temperature,omitempty"`
	MaxTokens   int           `json:"max_tokens,omitempty"`
	TopP        float64       `json:"top_p,omitempty"`
}

// AIResponse AI 响应结构
type AIResponse struct {
	ID      string    `json:"id"`
	Object  string    `json:"object"`
	Created int64     `json:"created"`
	Model   string    `json:"model"`
	Choices []Choice  `json:"choices"`
	Usage   UsageInfo `json:"usage"`
}

// Choice 选择项
type Choice struct {
	Index        int           `json:"index"`
	Message      ChatMessage   `json:"message"`
	FinishReason string        `json:"finish_reason"`
	Delta        *MessageDelta `json:"delta,omitempty"`
}

// MessageDelta 流式响应增量
type MessageDelta struct {
	Role    string `json:"role,omitempty"`
	Content string `json:"content,omitempty"`
}

// UsageInfo 使用情况
type UsageInfo struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// AIClient AI 客户端接口
type AIClient interface {
	Chat(ctx context.Context, req *AIRequest) (*AIResponse, error)
	ChatStream(ctx context.Context, req *AIRequest) (<-chan *Choice, error)
}

// GLMClient 智谱 GLM 客户端
type GLMClient struct {
	apiKey     string
	baseURL    string
	model      string
	httpClient *http.Client
}

// NewGLMClient 创建 GLM 客户端
func NewGLMClient(apiKey, baseURL, model string) *GLMClient {
	return &GLMClient{
		apiKey:  apiKey,
		baseURL: baseURL,
		model:   model,
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

// Chat 同步对话
func (c *GLMClient) Chat(ctx context.Context, req *AIRequest) (*AIResponse, error) {
	if req.Model == "" {
		req.Model = c.model
	}

	// 构建智谱API格式的请求体
	// 转换ChatMessage到API格式（只包含role和content字段）
	apiMessages := make([]map[string]interface{}, len(req.Messages))
	for i, msg := range req.Messages {
		apiMessages[i] = map[string]interface{}{
			"role":    msg.Role,
			"content": msg.Content,
		}
	}

	glmReq := map[string]interface{}{
		"model":    req.Model,
		"messages": apiMessages,
	}

	if req.Stream {
		glmReq["stream"] = true
	}
	if req.Temperature > 0 {
		glmReq["temperature"] = req.Temperature
	}
	if req.MaxTokens > 0 {
		glmReq["max_tokens"] = req.MaxTokens
	}
	if req.TopP > 0 {
		glmReq["top_p"] = req.TopP
	}

	reqBody, err := json.Marshal(glmReq)
	if err != nil {
		return nil, fmt.Errorf("序列化请求失败: %w", err)
	}

	// 构建完整的API URL
	apiURL := c.baseURL
	if !strings.HasSuffix(apiURL, "/chat/completions") {
		if strings.HasSuffix(apiURL, "/") {
			apiURL += "chat/completions"
		} else {
			apiURL += "/chat/completions"
		}
	}

	// 添加调试信息
	log.Printf("请求URL: %s", apiURL)

	httpReq, err := http.NewRequestWithContext(ctx, "POST", apiURL, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API 错误 %d: %s", resp.StatusCode, string(body))
	}

	var aiResp AIResponse
	if err := json.Unmarshal(body, &aiResp); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	return &aiResp, nil
}

// ChatStream 流式对话
func (c *GLMClient) ChatStream(ctx context.Context, req *AIRequest) (<-chan *Choice, error) {
	if req.Model == "" {
		req.Model = c.model
	}

	req.Stream = true

	reqBody, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("序列化请求失败: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.baseURL, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+c.apiKey)
	httpReq.Header.Set("Accept", "text/event-stream")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("请求失败: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		defer resp.Body.Close()
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API 错误 %d: %s", resp.StatusCode, string(body))
	}

	choiceChan := make(chan *Choice, 10)

	go func() {
		defer resp.Body.Close()
		defer close(choiceChan)

		decoder := json.NewDecoder(resp.Body)
		for {
			var event struct {
				Data string `json:"data"`
			}

			if err := decoder.Decode(&event); err != nil {
				if err == io.EOF {
					break
				}
				log.Printf("解析流式响应失败: %v", err)
				continue
			}

			if strings.HasPrefix(event.Data, "[DONE]") {
				break
			}

			var aiResp AIResponse
			if err := json.Unmarshal([]byte(event.Data), &aiResp); err != nil {
				continue
			}

			if len(aiResp.Choices) > 0 {
				select {
				case choiceChan <- &aiResp.Choices[0]:
				case <-ctx.Done():
					return
				}
			}
		}
	}()

	return choiceChan, nil
}

// AIConversationManager AI 对话管理器
type AIConversationManager struct {
	client        AIClient
	zoteroDB      *ZoteroDB
	conversations map[string]*Conversation
}

// NewAIConversationManager 创建对话管理器
func NewAIConversationManager(client AIClient, zoteroDB *ZoteroDB) *AIConversationManager {
	return &AIConversationManager{
		client:        client,
		zoteroDB:      zoteroDB,
		conversations: make(map[string]*Conversation),
	}
}

// StartConversation 开始新对话
func (m *AIConversationManager) StartConversation(ctx context.Context, message string, documentIDs []int) (*Conversation, error) {
	convID := fmt.Sprintf("conv_%d", time.Now().Unix())

	conv := &Conversation{
		ID:        convID,
		Messages:  []ChatMessage{},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// 自动查找相关文献上下文（即使没有指定documentIDs）
	if len(documentIDs) == 0 {
		// 自动从解析结果中查找相关文献
		context, err := m.buildDocumentContext(ctx, message, []int{})
		if err != nil {
			log.Printf("自动构建文档上下文失败: %v", err)
		} else {
			conv.Context = context
		}
	} else {
		// 使用指定的文档ID构建上下文
		context, err := m.buildDocumentContext(ctx, message, documentIDs)
		if err != nil {
			log.Printf("构建文档上下文失败: %v", err)
		} else {
			conv.Context = context
		}
	}

	// 添加系统消息
	systemMsg := ChatMessage{
		Role:      "system",
		Content:   m.buildSystemPrompt(conv.Context),
		Timestamp: time.Now(),
	}
	conv.Messages = append(conv.Messages, systemMsg)

	// 添加用户消息
	userMsg := ChatMessage{
		Role:      "user",
		Content:   message,
		Timestamp: time.Now(),
		Metadata: &MessageMetadata{
			DocumentIDs: documentIDs,
			QueryType:   "general",
		},
	}
	conv.Messages = append(conv.Messages, userMsg)

	// 获取 AI 响应
	aiResp, err := m.client.Chat(ctx, &AIRequest{
		Model:    "", // 使用默认模型
		Messages: conv.Messages,
	})
	if err != nil {
		return nil, fmt.Errorf("AI 响应失败: %w", err)
	}

	if len(aiResp.Choices) > 0 {
		assistantMsg := aiResp.Choices[0].Message
		assistantMsg.Timestamp = time.Now()
		conv.Messages = append(conv.Messages, assistantMsg)
	}

	conv.UpdatedAt = time.Now()
	m.conversations[convID] = conv

	return conv, nil
}

// buildDocumentContext 构建文档上下文
func (m *AIConversationManager) buildDocumentContext(ctx context.Context, query string, documentIDs []int) (*DocumentContext, error) {
	// 首先尝试从解析结果中获取文档
	documents, err := m.getDocumentsFromResults(query)
	if err != nil {
		log.Printf("从解析结果获取文档失败: %v", err)
		// 降级到数据库查询
		return m.buildDocumentContextFromDB(ctx, query, documentIDs)
	}

	return &DocumentContext{
		Documents: documents,
		Query:     query,
		Relevance: 0.9, // 解析结果的相关性更高
	}, nil
}

// getDocumentsFromResults 从解析结果获取文档
func (m *AIConversationManager) getDocumentsFromResults(query string) ([]DocumentSummary, error) {
	resultsDir := "data/results"

	entries, err := os.ReadDir(resultsDir)
	if err != nil {
		return nil, fmt.Errorf("读取结果目录失败: %w", err)
	}

	var documents []DocumentSummary
	queryLower := strings.ToLower(query)

	for _, entry := range entries {
		if !entry.IsDir() || entry.Name() == "latest" {
			continue
		}

		// 读取元数据
		metaFile := filepath.Join(resultsDir, entry.Name(), "meta.json")
		if info := readMeta(metaFile); info != nil {
			// 检查是否与查询相关
			titleMatch := strings.Contains(strings.ToLower(info.Title), queryLower)
			contentMatch := false

			// 检查markdown内容是否相关
			mdFile := filepath.Join(resultsDir, entry.Name(), "full.md")
			if content, err := os.ReadFile(mdFile); err == nil {
				contentStr := string(content)
				if strings.Contains(strings.ToLower(contentStr), queryLower) {
					contentMatch = true
				}

				// 如果相关，创建文档摘要
				if titleMatch || contentMatch {
					doc := DocumentSummary{
						ID:       len(documents), // 简化ID
						Title:    info.Title,
						Authors:  extractAuthorsFromContent(string(content)),
						Abstract: extractAbstractFromContent(string(content)),
						Keywords: extractKeywordsFromContent(string(content)),
					}
					documents = append(documents, doc)
				}
			}
		}
	}

	return documents, nil
}

// buildDocumentContextFromDB 从数据库构建文档上下文（降级方案）
func (m *AIConversationManager) buildDocumentContextFromDB(ctx context.Context, query string, documentIDs []int) (*DocumentContext, error) {
	var documents []DocumentSummary

	for _, docID := range documentIDs {
		// 从数据库获取文档信息
		// 这里可以实现更复杂的数据库查询逻辑
		doc := DocumentSummary{
			ID:      docID,
			Title:   fmt.Sprintf("Document %d", docID),
			Authors: "Unknown Authors",
		}
		documents = append(documents, doc)
	}

	return &DocumentContext{
		Documents: documents,
		Query:     query,
		Relevance: 0.7, // 数据库结果相关性较低
	}, nil
}

// readMeta 读取元数据（从main.go复制过来的函数）
func readMeta(metaFile string) *ParsedFileInfo {
	data, err := os.ReadFile(metaFile)
	if err != nil {
		return nil
	}

	var info ParsedFileInfo
	if err := json.Unmarshal(data, &info); err != nil {
		return nil
	}

	return &info
}

// extractAuthorsFromContent 从内容中提取作者
func extractAuthorsFromContent(content string) string {
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.Contains(line, "摘要") || strings.Contains(line, "Abstract") {
			// 查找作者行，通常在摘要前面
			for i := range lines {
				if strings.Contains(lines[i], line) && i > 0 {
					prevLine := strings.TrimSpace(lines[i-1])
					if len(prevLine) > 10 && len(prevLine) < 200 {
						// 可能是作者行
						return prevLine
					}
				}
			}
		}
	}
	return "未知作者"
}

// extractAbstractFromContent 从内容中提取摘要
func extractAbstractFromContent(content string) string {
	lines := strings.Split(content, "\n")
	var abstract strings.Builder
	inAbstract := false

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "摘要：") || strings.HasPrefix(line, "Abstract:") {
			inAbstract = true
			abstract.WriteString(line)
			continue
		}

		if inAbstract {
			if strings.HasPrefix(line, "关键词") || strings.HasPrefix(line, "Key words") ||
				strings.HasPrefix(line, "#") && !strings.HasPrefix(line, "# ") {
				break
			}
			if abstract.Len() > 0 {
				abstract.WriteString(" ")
			}
			abstract.WriteString(line)
		}
	}

	abstractStr := abstract.String()
	if len(abstractStr) > 500 {
		abstractStr = abstractStr[:500] + "..."
	}

	return abstractStr
}

// extractKeywordsFromContent 从内容中提取关键词
func extractKeywordsFromContent(content string) []string {
	var keywords []string

	// 查找关键词部分
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "关键词：") {
			keywordsStr := strings.TrimPrefix(line, "关键词：")
			keywordsStr = strings.TrimPrefix(keywordsStr, "Key words:")

			// 分割关键词
			kwList := strings.FieldsFunc(keywordsStr, func(r rune) bool {
				return r == '；' || r == ';' || r == ' ' || r == ','
			})

			for _, kw := range kwList {
				kw = strings.TrimSpace(kw)
				if len(kw) > 1 && len(keywords) < 10 {
					keywords = append(keywords, kw)
				}
			}
			break
		}
	}

	return keywords
}

// buildSystemPrompt 构建系统提示
func (m *AIConversationManager) buildSystemPrompt(context *DocumentContext) string {
	basePrompt := `你是一个专业的学术文献助手，专门帮助用户分析和理解学术论文。请用中文回答，并提供准确、有用的信息。

你的主要功能包括：
1. 分析论文内容和方法
2. 解释专业术语和概念
3. 比较不同研究方法的优缺点
4. 提供研究建议和未来方向
5. 协助文献综述和总结

请基于提供的文献内容进行回答，保持专业、客观、有帮助的态度。`

	if context != nil && len(context.Documents) > 0 {
		contextInfo := "\n\n=== 相关文献信息 ===\n"
		for i, doc := range context.Documents {
			contextInfo += fmt.Sprintf("\n%d. %s\n", i+1, doc.Title)
			contextInfo += fmt.Sprintf("   作者: %s\n", doc.Authors)

			if doc.Abstract != "" {
				contextInfo += fmt.Sprintf("   摘要: %s\n", doc.Abstract)
			}

			if len(doc.Keywords) > 0 {
				contextInfo += fmt.Sprintf("   关键词: %s\n", strings.Join(doc.Keywords, ", "))
			}

			contextInfo += "   ---\n"
		}

		contextInfo += fmt.Sprintf("\n💡 请基于上述文献内容回答用户的问题: %s", context.Query)
		basePrompt += contextInfo
	}

	return basePrompt
}

// StartConversationWithDocument 基于指定文献开始对话
func (m *AIConversationManager) StartConversationWithDocument(ctx context.Context, message string, docContext *DocumentContext) (*Conversation, error) {
	convID := fmt.Sprintf("conv_%d", time.Now().Unix())

	conv := &Conversation{
		ID:        convID,
		Messages:  []ChatMessage{},
		Context:   docContext,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// 添加系统消息
	systemMsg := ChatMessage{
		Role:      "system",
		Content:   m.buildSystemPrompt(docContext),
		Timestamp: time.Now(),
	}
	conv.Messages = append(conv.Messages, systemMsg)

	// 添加用户消息
	userMsg := ChatMessage{
		Role:      "user",
		Content:   message,
		Timestamp: time.Now(),
		Metadata: &MessageMetadata{
			DocumentIDs: []int{1}, // 指定文献
			QueryType:   "document_specific",
		},
	}
	conv.Messages = append(conv.Messages, userMsg)

	// 获取 AI 响应
	aiResp, err := m.client.Chat(ctx, &AIRequest{
		Model:    "", // 使用默认模型
		Messages: conv.Messages,
	})
	if err != nil {
		return nil, fmt.Errorf("AI 响应失败: %w", err)
	}

	if len(aiResp.Choices) > 0 {
		assistantMsg := aiResp.Choices[0].Message
		assistantMsg.Timestamp = time.Now()
		conv.Messages = append(conv.Messages, assistantMsg)
	}

	conv.UpdatedAt = time.Now()
	m.conversations[convID] = conv

	return conv, nil
}

// ContinueConversation 继续对话
func (m *AIConversationManager) ContinueConversation(ctx context.Context, convID, message string) (*Conversation, error) {
	conv, exists := m.conversations[convID]
	if !exists {
		return nil, fmt.Errorf("对话不存在: %s", convID)
	}

	// 添加用户消息
	userMsg := ChatMessage{
		Role:      "user",
		Content:   message,
		Timestamp: time.Now(),
	}
	conv.Messages = append(conv.Messages, userMsg)

	// 获取 AI 响应
	aiResp, err := m.client.Chat(ctx, &AIRequest{
		Model:    "", // 使用默认模型
		Messages: conv.Messages,
	})
	if err != nil {
		return nil, fmt.Errorf("AI 响应失败: %w", err)
	}

	if len(aiResp.Choices) > 0 {
		assistantMsg := aiResp.Choices[0].Message
		assistantMsg.Timestamp = time.Now()
		conv.Messages = append(conv.Messages, assistantMsg)
	}

	conv.UpdatedAt = time.Now()
	return conv, nil
}

// GetConversation 获取对话历史
func (m *AIConversationManager) GetConversation(convID string) (*Conversation, error) {
	conv, exists := m.conversations[convID]
	if !exists {
		return nil, fmt.Errorf("对话不存在: %s", convID)
	}
	return conv, nil
}

// ListConversations 列出所有对话
func (m *AIConversationManager) ListConversations() []string {
	var convIDs []string
	for id := range m.conversations {
		convIDs = append(convIDs, id)
	}
	return convIDs
}

// DeleteConversation 删除对话
func (m *AIConversationManager) DeleteConversation(convID string) error {
	if _, exists := m.conversations[convID]; !exists {
		return fmt.Errorf("对话不存在: %s", convID)
	}
	delete(m.conversations, convID)
	return nil
}
