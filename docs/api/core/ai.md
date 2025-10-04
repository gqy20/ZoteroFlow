# AI 对话接口 API

## 概述

AI 对话接口提供了与大型语言模型的集成功能，支持智能文献分析、对话管理、上下文理解等高级 AI 功能。

## 数据结构

### ChatMessage

```go
type ChatMessage struct {
    Role      string           `json:"role"`       // 消息角色：user, assistant, system
    Content   string           `json:"content"`    // 消息内容
    Timestamp time.Time        `json:"timestamp"`  // 时间戳
    Metadata  *MessageMetadata `json:"metadata"`   // 消息元数据
}
```

### MessageMetadata

```go
type MessageMetadata struct {
    DocumentIDs []int  `json:"document_ids,omitempty"` // 关联的文献ID
    QueryType   string `json:"query_type,omitempty"`   // 查询类型：search, analysis, summary
}
```

### DocumentContext

```go
type DocumentContext struct {
    Documents []DocumentSummary `json:"documents"` // 文档列表
    Query     string            `json:"query"`     // 查询内容
    Relevance float64           `json:"relevance"`  // 相关性评分
}
```

### DocumentSummary

```go
type DocumentSummary struct {
    ID       int      `json:"id"`        // 文档ID
    Title    string   `json:"title"`     // 文档标题
    Authors  string   `json:"authors"`   // 作者信息
    Abstract string   `json:"abstract"`  // 摘要内容
    Keywords []string `json:"keywords"`  // 关键词列表
}
```

### Conversation

```go
type Conversation struct {
    ID        string           `json:"id"`         // 对话ID
    Messages  []ChatMessage    `json:"messages"`   // 消息列表
    Context   *DocumentContext `json:"context"`    // 文档上下文
    CreatedAt time.Time        `json:"created_at"` // 创建时间
    UpdatedAt time.Time        `json:"updated_at"` // 更新时间
}
```

### AIRequest

```go
type AIRequest struct {
    Model       string        `json:"model"`        // 模型名称
    Messages    []ChatMessage `json:"messages"`    // 消息列表
    Stream      bool          `json:"stream"`      // 是否流式响应
    Temperature float64       `json:"temperature"` // 温度参数
    MaxTokens   int           `json:"max_tokens"`  // 最大Token数
    TopP        float64       `json:"top_p"`       // Top-P采样参数
}
```

### AIResponse

```go
type AIResponse struct {
    ID      string    `json:"id"`      // 响应ID
    Object  string    `json:"object"`  // 对象类型
    Created int64     `json:"created"` // 创建时间戳
    Model   string    `json:"model"`   // 使用的模型
    Choices []Choice  `json:"choices"` // 选择列表
    Usage   UsageInfo `json:"usage"`   // 使用情况
}
```

### Choice

```go
type Choice struct {
    Index        int           `json:"index"`        // 选择索引
    Message      ChatMessage   `json:"message"`      // 消息内容
    FinishReason string        `json:"finish_reason"` // 结束原因
    Delta        *MessageDelta `json:"delta"`       // 流式响应增量
}
```

### UsageInfo

```go
type UsageInfo struct {
    PromptTokens     int `json:"prompt_tokens"`     // 输入Token数
    CompletionTokens int `json:"completion_tokens"` // 输出Token数
    TotalTokens      int `json:"total_tokens"`      // 总Token数
}
```

### AIClient 接口

```go
type AIClient interface {
    Chat(ctx context.Context, req *AIRequest) (*AIResponse, error)
    ChatStream(ctx context.Context, req *AIRequest) (<-chan *Choice, error)
}
```

### GLMClient

```go
type GLMClient struct {
    apiKey     string
    baseURL    string
    model      string
    httpClient *http.Client
}
```

### AIConversationManager

```go
type AIConversationManager struct {
    client        AIClient
    zoteroDB      *ZoteroDB
    conversations map[string]*Conversation
}
```

## 核心接口

### NewGLMClient

创建新的 GLM 客户端实例。

```go
func NewGLMClient(apiKey, baseURL, model string) *GLMClient
```

**参数**:
- `apiKey` (string): API 密钥
- `baseURL` (string): API 基础URL
- `model` (string): 模型名称

**返回值**:
- `*GLMClient`: GLM 客户端实例

**示例**:
```go
client := core.NewGLMClient(
    "your_api_key_here",
    "https://open.bigmodel.cn/api/coding/paas/v4",
    "glm-4.6",
)
```

**默认配置**:
- HTTP超时: 60秒
- 默认模型: glm-4.6

### Chat

同步对话接口。

```go
func (c *GLMClient) Chat(ctx context.Context, req *AIRequest) (*AIResponse, error)
```

**参数**:
- `ctx` (context.Context): 上下文对象
- `req` (*AIRequest): AI 请求对象

**返回值**:
- `*AIResponse`: AI 响应对象
- `error`: 错误信息

**示例**:
```go
messages := []core.ChatMessage{
    {
        Role:    "system",
        Content: "你是一个专业的学术文献助手。",
    },
    {
        Role:    "user",
        Content: "什么是机器学习？",
    },
}

req := &core.AIRequest{
    Model:    "glm-4.6",
    Messages: messages,
    MaxTokens: 500,
}

ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

response, err := client.Chat(ctx, req)
if err != nil {
    log.Printf("AI调用失败: %v", err)
    return
}

if len(response.Choices) > 0 {
    fmt.Printf("AI回答: %s\n", response.Choices[0].Message.Content)
    fmt.Printf("Token使用: %d\n", response.Usage.TotalTokens)
}
```

### ChatStream

流式对话接口。

```go
func (c *GLMClient) ChatStream(ctx context.Context, req *AIRequest) (<-chan *Choice, error)
```

**参数**:
- `ctx` (context.Context): 上下文对象
- `req` (*AIRequest): AI 请求对象

**返回值**:
- `<-chan *Choice`: 流式响应通道
- `error`: 错误信息

**示例**:
```go
req := &core.AIRequest{
    Model:    "glm-4.6",
    Messages: messages,
    Stream:   true,
}

ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

choiceChan, err := client.ChatStream(ctx, req)
if err != nil {
    log.Printf("流式调用失败: %v", err)
    return
}

fmt.Print("AI回答: ")
for choice := range choiceChan {
    if choice.Delta != nil && choice.Delta.Content != "" {
        fmt.Print(choice.Delta.Content)
    }
}
fmt.Println()
```

### NewAIConversationManager

创建 AI 对话管理器。

```go
func NewAIConversationManager(client AIClient, zoteroDB *ZoteroDB) *AIConversationManager
```

**参数**:
- `client` (AIClient): AI 客户端
- `zoteroDB` (*ZoteroDB): Zotero 数据库实例

**返回值**:
- `*AIConversationManager`: 对话管理器实例

**示例**:
```go
aiClient := core.NewGLMClient(apiKey, baseURL, model)
zoteroDB, _ := core.NewZoteroDB(dbPath, dataDir)

chatManager := core.NewAIConversationManager(aiClient, zoteroDB)
```

### StartConversation

开始新对话。

```go
func (m *AIConversationManager) StartConversation(ctx context.Context, message string, documentIDs []int) (*Conversation, error)
```

**参数**:
- `ctx` (context.Context): 上下文对象
- `message` (string): 用户消息
- `documentIDs` ([]int): 关联的文档ID列表

**返回值**:
- `*Conversation`: 对话对象
- `error`: 错误信息

**示例**:
```go
conv, err := chatManager.StartConversation(
    context.Background(),
    "请分析一下机器学习的基本概念",
    []int{123, 456}, // 指定相关文献
)
if err != nil {
    log.Printf("开始对话失败: %v", err)
    return
}

fmt.Printf("对话ID: %s\n", conv.ID)
if len(conv.Messages) >= 3 {
    aiResponse := conv.Messages[2]
    fmt.Printf("AI回答: %s\n", aiResponse.Content)
}
```

### StartConversationWithDocument

基于指定文档开始对话。

```go
func (m *AIConversationManager) StartConversationWithDocument(ctx context.Context, message string, docContext *DocumentContext) (*Conversation, error)
```

**参数**:
- `ctx` (context.Context): 上下文对象
- `message` (string): 用户消息
- `docContext` (*DocumentContext): 文档上下文

**返回值**:
- `*Conversation`: 对话对象
- `error`: 错误信息

**示例**:
```go
docContext := &core.DocumentContext{
    Documents: []core.DocumentSummary{
        {
            ID:       1,
            Title:    "机器学习基础教程",
            Authors:  "张三; 李四",
            Abstract: "本文介绍了机器学习的基本概念...",
            Keywords: []string{"机器学习", "人工智能", "算法"},
        },
    },
    Query:     "机器学习的基本概念",
    Relevance: 0.9,
}

conv, err := chatManager.StartConversationWithDocument(
    context.Background(),
    "请总结这篇文章的主要内容",
    docContext,
)
if err != nil {
    log.Printf("基于文档的对话失败: %v", err)
    return
}
```

### ContinueConversation

继续现有对话。

```go
func (m *AIConversationManager) ContinueConversation(ctx context.Context, convID, message string) (*Conversation, error)
```

**参数**:
- `ctx` (context.Context): 上下文对象
- `convID` (string): 对话ID
- `message` (string): 用户消息

**返回值**:
- `*Conversation`: 更新后的对话对象
- `error`: 错误信息

**示例**:
```go
conv, err := chatManager.ContinueConversation(
    context.Background(),
    "conv_123456789",
    "能详细解释一下监督学习吗？",
)
if err != nil {
    log.Printf("继续对话失败: %v", err)
    return
}

if len(conv.Messages) >= 2 {
    lastMsg := conv.Messages[len(conv.Messages)-1]
    if lastMsg.Role == "assistant" {
        fmt.Printf("AI回答: %s\n", lastMsg.Content)
    }
}
```

## 辅助功能

### 文档上下文构建

#### buildDocumentContext

构建文档上下文。

```go
func (m *AIConversationManager) buildDocumentContext(ctx context.Context, query string, documentIDs []int) (*DocumentContext, error)
```

**处理逻辑**:
1. 优先从解析结果中获取文档
2. 降级到数据库查询
3. 计算相关性评分

#### getDocumentsFromResults

从解析结果获取文档。

```go
func (m *AIConversationManager) getDocumentsFromResults(query string) ([]DocumentSummary, error)
```

**搜索策略**:
- 标题匹配
- 内容匹配
- 关键词匹配

#### extractAuthorsFromContent

从内容中提取作者信息。

```go
func extractAuthorsFromContent(content string) string
```

#### extractAbstractFromContent

从内容中提取摘要。

```go
func extractAbstractFromContent(content string) string
```

#### extractKeywordsFromContent

从内容中提取关键词。

```go
func extractKeywordsFromContent(content string) []string
```

### 系统提示构建

#### buildSystemPrompt

构建系统提示。

```go
func (m *AIConversationManager) buildSystemPrompt(context *DocumentContext) string
```

**基础提示内容**:
```
你是一个专业的学术文献助手，专门帮助用户分析和理解学术论文。请用中文回答，并提供准确、有用的信息。

你的主要功能包括：
1. 分析论文内容和方法
2. 解释专业术语和概念
3. 比较不同研究方法的优缺点
4. 提供研究建议和未来方向
5. 协助文献综述和总结

请基于提供的文献内容进行回答，保持专业、客观、有帮助的态度。
```

### 对话管理

#### GetConversation

获取对话历史。

```go
func (m *AIConversationManager) GetConversation(convID string) (*Conversation, error)
```

#### ListConversations

列出所有对话。

```go
func (m *AIConversationManager) ListConversations() []string
```

#### DeleteConversation

删除对话。

```go
func (m *AIConversationManager) DeleteConversation(convID string) error
```

## 使用示例

### 基础对话

```go
package main

import (
    "context"
    "fmt"
    "log"
    "time"
    "zoteroflow2-server/core"
)

func main() {
    // 创建AI客户端
    client := core.NewGLMClient(
        "your_api_key",
        "https://open.bigmodel.cn/api/coding/paas/v4",
        "glm-4.6",
    )
    
    // 创建对话请求
    messages := []core.ChatMessage{
        {
            Role:    "system",
            Content: "你是一个专业的学术文献助手。",
        },
        {
            Role:    "user",
            Content: "什么是深度学习？",
        },
    }
    
    req := &core.AIRequest{
        Model:    "glm-4.6",
        Messages: messages,
        MaxTokens: 500,
    }
    
    // 发送请求
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()
    
    response, err := client.Chat(ctx, req)
    if err != nil {
        log.Fatalf("AI调用失败: %v", err)
    }
    
    // 处理响应
    if len(response.Choices) > 0 {
        fmt.Printf("AI回答: %s\n", response.Choices[0].Message.Content)
        fmt.Printf("Token使用: %d\n", response.Usage.TotalTokens)
    }
}
```

### 对话管理

```go
func conversationExample() {
    // 创建客户端和数据库连接
    aiClient := core.NewGLMClient(apiKey, baseURL, model)
    zoteroDB, _ := core.NewZoteroDB(dbPath, dataDir)
    
    // 创建对话管理器
    chatManager := core.NewAIConversationManager(aiClient, zoteroDB)
    
    // 开始新对话
    conv, err := chatManager.StartConversation(
        context.Background(),
        "请分析一下机器学习在医疗领域的应用",
        nil, // 不指定特定文档
    )
    if err != nil {
        log.Printf("开始对话失败: %v", err)
        return
    }
    
    fmt.Printf("对话ID: %s\n", conv.ID)
    
    // 继续对话
    conv, err = chatManager.ContinueConversation(
        context.Background(),
        conv.ID,
        "能具体介绍一下深度学习在医学影像诊断中的应用吗？",
    )
    if err != nil {
        log.Printf("继续对话失败: %v", err)
        return
    }
    
    if len(conv.Messages) >= 2 {
        lastMsg := conv.Messages[len(conv.Messages)-1]
        fmt.Printf("AI回答: %s\n", lastMsg.Content)
    }
}
```

### 基于文档的对话

```go
func documentBasedChat() {
    // 创建客户端和数据库连接
    aiClient := core.NewGLMClient(apiKey, baseURL, model)
    zoteroDB, _ := core.NewZoteroDB(dbPath, dataDir)
    
    // 创建对话管理器
    chatManager := core.NewAIConversationManager(aiClient, zoteroDB)
    
    // 构建文档上下文
    docContext := &core.DocumentContext{
        Documents: []core.DocumentSummary{
            {
                ID:       1,
                Title:    "深度学习在医疗诊断中的应用研究",
                Authors:  "张三; 李四; 王五",
                Abstract: "本文研究了深度学习技术在医疗影像诊断中的应用...",
                Keywords: []string{"深度学习", "医疗诊断", "神经网络", "医学影像"},
            },
        },
        Query:     "深度学习在医疗诊断中的应用",
        Relevance: 0.95,
    }
    
    // 基于文档开始对话
    conv, err := chatManager.StartConversationWithDocument(
        context.Background(),
        "请总结这篇文章的主要贡献和创新点",
        docContext,
    )
    if err != nil {
        log.Printf("基于文档的对话失败: %v", err)
        return
    }
    
    if len(conv.Messages) >= 3 {
        aiResponse := conv.Messages[2]
        fmt.Printf("AI回答: %s\n", aiResponse.Content)
    }
}
```

### 流式对话

```go
func streamingChat() {
    client := core.NewGLMClient(apiKey, baseURL, model)
    
    messages := []core.ChatMessage{
        {
            Role:    "user",
            Content: "请详细介绍机器学习的主要算法类型",
        },
    }
    
    req := &core.AIRequest{
        Model:    "glm-4.6",
        Messages: messages,
        Stream:   true,
    }
    
    ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
    defer cancel()
    
    choiceChan, err := client.ChatStream(ctx, req)
    if err != nil {
        log.Printf("流式调用失败: %v", err)
        return
    }
    
    fmt.Print("AI回答: ")
    for choice := range choiceChan {
        if choice.Delta != nil && choice.Delta.Content != "" {
            fmt.Print(choice.Delta.Content)
        }
    }
    fmt.Println()
}
```

## 错误处理

### 常见错误类型

1. **认证错误**
   ```go
   if strings.Contains(err.Error(), "authentication") {
       // API密钥无效或过期
   }
   ```

2. **配额错误**
   ```go
   if strings.Contains(err.Error(), "quota") {
       // API调用配额不足
   }
   ```

3. **模型错误**
   ```go
   if strings.Contains(err.Error(), "model") {
       // 模型不存在或不支持
   }
   ```

4. **超时错误**
   ```go
   if strings.Contains(err.Error(), "timeout") {
       // 请求超时
   }
   ```

### 错误处理示例

```go
func safeChat(client *core.GLMClient, req *core.AIRequest) (*core.AIResponse, error) {
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()
    
    response, err := client.Chat(ctx, req)
    if err != nil {
        switch {
        case strings.Contains(err.Error(), "authentication"):
            return nil, fmt.Errorf("API认证失败，请检查密钥")
        case strings.Contains(err.Error(), "quota"):
            return nil, fmt.Errorf("API配额不足，请稍后重试")
        case strings.Contains(err.Error(), "timeout"):
            return nil, fmt.Errorf("请求超时，请重试")
        case strings.Contains(err.Error(), "model"):
            return nil, fmt.Errorf("模型不支持，请检查模型名称")
        default:
            return nil, fmt.Errorf("AI调用失败: %w", err)
        }
    }
    
    return response, nil
}
```

## 性能优化

### 连接池

```go
type PooledGLMClient struct {
    *GLMClient
    pool chan *http.Client
}

func NewPooledGLMClient(apiKey, baseURL, model string, poolSize int) *PooledGLMClient {
    pool := make(chan *http.Client, poolSize)
    for i := 0; i < poolSize; i++ {
        pool <- &http.Client{Timeout: 60 * time.Second}
    }
    
    return &PooledGLMClient{
        GLMClient: NewGLMClient(apiKey, baseURL, model),
        pool:      pool,
    }
}

func (c *PooledGLMClient) Chat(ctx context.Context, req *core.AIRequest) (*core.AIResponse, error) {
    httpClient := <-c.pool
    defer func() { c.pool <- httpClient }()
    
    // 使用池中的HTTP客户端
    c.httpClient = httpClient
    return c.GLMClient.Chat(ctx, req)
}
```

### 缓存策略

```go
type CachedConversationManager struct {
    *AIConversationManager
    cache map[string]*Conversation
    mutex sync.RWMutex
    ttl   time.Duration
}

func (c *CachedConversationManager) GetConversation(convID string) (*Conversation, error) {
    c.mutex.RLock()
    if conv, exists := c.cache[convID]; exists {
        c.mutex.RUnlock()
        
        // 检查是否过期
        if time.Since(conv.UpdatedAt) < c.ttl {
            return conv, nil
        }
    }
    c.mutex.RUnlock()
    
    // 调用原始方法
    return c.AIConversationManager.GetConversation(convID)
}
```

## 配置建议

### 环境变量

```bash
# AI API 配置
AI_API_KEY=your_api_key_here
AI_BASE_URL=https://open.bigmodel.cn/api/coding/paas/v4
AI_MODEL=glm-4.6

# 性能配置
AI_TIMEOUT=60              # 超时时间（秒）
AI_MAX_TOKENS=2000         # 最大Token数
AI_TEMPERATURE=0.7         # 温度参数
AI_TOP_P=0.9               # Top-P参数

# 对话配置
AI_MAX_CONVERSATIONS=100   # 最大对话数量
AI_CONTEXT_LIMIT=10        # 上下文文档数量限制
```

### 模型选择

```go
func selectModel(taskType string) string {
    switch taskType {
    case "analysis":
        return "glm-4.6"  // 复杂分析任务
    case "summary":
        return "glm-4-flash"  // 快速摘要任务
    case "translation":
        return "glm-4.6"  // 翻译任务
    default:
        return "glm-4.6"  // 默认模型
    }
}
```

## 监控和日志

### Token使用监控

```go
func monitorTokenUsage(response *core.AIResponse) {
    usage := response.Usage
    log.Printf("Token使用统计:")
    log.Printf("  输入Token: %d", usage.PromptTokens)
    log.Printf("  输出Token: %d", usage.CompletionTokens)
    log.Printf("  总Token: %d", usage.TotalTokens)
    
    // 计算成本（假设价格）
    inputCost := float64(usage.PromptTokens) * 0.001  // $0.001 per 1K tokens
    outputCost := float64(usage.CompletionTokens) * 0.002  // $0.002 per 1K tokens
    totalCost := inputCost + outputCost
    
    log.Printf("  预估成本: $%.4f", totalCost)
}
```

### 对话质量监控

```go
func monitorConversationQuality(conv *core.Conversation) {
    if len(conv.Messages) < 3 {
        return
    }
    
    userMsg := conv.Messages[len(conv.Messages)-2]
    aiMsg := conv.Messages[len(conv.Messages)-1]
    
    // 简单的质量评估
    responseLength := len(aiMsg.Content)
    if responseLength < 50 {
        log.Printf("警告: 回答过短 (%d 字符)", responseLength)
    }
    
    if strings.Contains(aiMsg.Content, "我不知道") || 
       strings.Contains(aiMsg.Content, "无法回答") {
        log.Printf("警告: AI表示无法回答问题")
    }
    
    log.Printf("对话质量: 用户问题长度=%d, AI回答长度=%d", 
        len(userMsg.Content), responseLength)
}
```

## 最佳实践

### 1. 提示工程
- 使用清晰、具体的系统提示
- 提供充分的上下文信息
- 合理设置温度和Token限制

### 2. 对话管理
- 定期清理过期对话
- 限制对话长度避免Token溢出
- 保存重要对话历史

### 3. 错误处理
- 实现重试机制
- 提供用户友好的错误信息
- 监控API调用状态

### 4. 性能优化
- 使用连接池减少连接开销
- 实现响应缓存
- 合理设置超时时间

## 版本兼容性

- **GLM-4.6**: 完全支持
- **GLM-4-flash**: 支持
- **GLM-3-turbo**: 支持
- **API v1**: 支持

## 注意事项

1. **API密钥安全**: 不要在代码中硬编码API密钥
2. **Token限制**: 注意模型的Token限制和费用
3. **内容过滤**: 遵免生成不当内容
4. **并发控制**: 避免过多的并发请求
5. **数据隐私**: 不要在对话中包含敏感信息