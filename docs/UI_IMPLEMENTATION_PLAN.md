# ZoteroFlow UI实现方案 (基于现有后端优化)

> **核心理念**: AI无处不在，无需寻找；尊重知识产权，不优化PDF下载
> **架构优势**: 复用成熟后端代码，最小化开发工作量

## 📋 UI实现概览

基于现有2300行成熟Go后端代码，采用"最小改动、最大价值"的实现策略。现有的ZoteroDB、MinerUClient和AIConversationManager已经提供了完整的核心功能，只需增加Web API层和前端界面。

## 🏗️ 前后端架构设计

### 现有后端架构 (已实现2300行)
```
server/
├── main.go                 # 878行 - 完整CLI应用
├── config/config.go        # 配置管理
└── core/
    ├── zotero.go           # 784行 - Zotero数据库访问
    ├── mineru.go           # 716行 - MinerU PDF解析
    ├── ai.go               # 707行 - AI对话系统
    └── parser.go           # 198行 - PDF解析器
```

### 新增Web API层 (200行)
```
server/
├── api/
│   ├── handlers.go         # Web API处理器
│   ├── middleware.go       # CORS和中间件
│   └── routes.go           # 路由配置
```

### 前端架构 (Next.js 15 - 600行)
```
frontend/
├── app/
│   ├── layout.tsx          # 50行 - 根布局
│   ├── page.tsx            # 80行 - 文献列表页
│   ├── components/
│   │   ├── LiteratureCard.tsx  # 50行 - 文献卡片
│   │   ├── AIChat.tsx          # 80行 - AI对话组件
│   │   ├── SearchBar.tsx       # 60行 - 智能搜索
│   │   └── PDFViewer.tsx       # 110行 - PDF阅读器
│   └── lib/
│       ├── api.ts              # 50行 - API客户端
│       └── types.ts            # 30行 - 类型定义
└── package.json
```

## 🚀 UI实现步骤

### Phase 1: Web API层开发 (第1-2天)

#### 1.1 添加Gin依赖和路由
```bash
# 在现有server目录中添加依赖
go get -u github.com/gin-gonic/gin
go get -u github.com/gin-contrib/cors
```

#### 1.2 创建API处理器 (100行)
```go
// server/api/handlers.go
package api

import (
    "net/http"
    "strconv"
    "github.com/gin-gonic/gin"
    "zoteroflow2/core"
)

// 文献相关API
func getLiteratureList(c *gin.Context) {
    zoteroDB := c.MustGet("zoteroDB").(*core.ZoteroDB)
    items, err := zoteroDB.GetItemsWithPDF(100) // 获取前100篇文献
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    c.JSON(http.StatusOK, gin.H{"data": items})
}

func getLiteratureDetail(c *gin.Context) {
    itemID, err := strconv.Atoi(c.Param("id"))
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "无效的文献ID"})
        return
    }

    zoteroDB := c.MustGet("zoteroDB").(*core.ZoteroDB)
    items, err := zoteroDB.GetItemsWithPDF(1)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    if len(items) == 0 {
        c.JSON(http.StatusNotFound, gin.H{"error": "文献不存在"})
        return
    }

    c.JSON(http.StatusOK, gin.H{"data": items[0]})
}

// AI对话API
func startAIChat(c *gin.Context) {
    var req struct {
        Message string `json:"message"`
        DocIDs  []int  `json:"doc_ids,omitempty"`
    }

    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    aiManager := c.MustGet("aiManager").(*core.AIConversationManager)
    conv, err := aiManager.StartConversation(c.Request.Context(), req.Message, req.DocIDs)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    c.JSON(http.StatusOK, gin.H{"data": conv})
}

func continueAIChat(c *gin.Context) {
    convID := c.Param("id")
    var req struct {
        Message string `json:"message"`
    }

    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    aiManager := c.MustGet("aiManager").(*core.AIConversationManager)
    conv, err := aiManager.ContinueConversation(c.Request.Context(), convID, req.Message)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    c.JSON(http.StatusOK, gin.H{"data": conv})
}
```

#### 1.3 创建路由配置 (50行)
```go
// server/api/routes.go
package api

import (
    "github.com/gin-contrib/cors"
    "github.com/gin-gonic/gin"
)

func SetupRoutes(r *gin.Engine) {
    // CORS配置
    r.Use(cors.New(cors.Config{
        AllowOrigins:     []string{"http://localhost:3000"},
        AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
        AllowHeaders:     []string{"Content-Type", "Authorization"},
        ExposeHeaders:    []string{"Content-Length"},
        AllowCredentials: true,
    }))

    // API路由组
    api := r.Group("/api")
    {
        // 文献相关
        literature := api.Group("/literature")
        {
            literature.GET("", getLiteratureList)
            literature.GET("/:id", getLiteratureDetail)
            literature.POST("/:id/analyze", analyzeLiterature)
        }

        // AI对话相关
        ai := api.Group("/ai")
        {
            ai.POST("/chat", startAIChat)
            ai.GET("/chat/:id", getChatHistory)
            ai.POST("/chat/:id", continueAIChat)
        }

        // 搜索相关
        search := api.Group("/search")
        {
            search.GET("", searchLiterature)
            search.POST("/smart", smartSearch)
        }
    }
}
```

#### 1.4 修改main.go启动Web服务 (50行)
```go
// 在现有main.go中添加Web服务器启动代码
func startWebServer() {
    // 初始化核心组件
    config := config.LoadConfig()
    zoteroDB := core.NewZoteroDB(config.ZoteroDBPath)
    defer zoteroDB.Close()

    // 初始化AI客户端
    aiClient := core.NewGLMClient(config.AIAPIKey, config.AIBaseURL, config.AIModel)
    aiManager := core.NewAIConversationManager(aiClient, zoteroDB)

    // 创建Gin引擎
    r := gin.Default()

    // 设置中间件
    r.Use(func(c *gin.Context) {
        c.Set("zoteroDB", zoteroDB)
        c.Set("aiManager", aiManager)
        c.Next()
    })

    // 设置路由
    api.SetupRoutes(r)

    // 启动服务器
    log.Println("Web服务器启动在 http://localhost:8080")
    if err := r.Run(":8080"); err != nil {
        log.Fatal("启动Web服务器失败:", err)
    }
}
```

### Phase 2: 前端项目初始化 (第3天)

#### 2.1 创建Next.js项目
```bash
# 创建Next.js 15项目
npx create-next-app@latest frontend --typescript --tailwind --eslint --app
cd frontend

# 安装依赖
npm install lucide-react
npm install @radix-ui/react-dialog
npm install @radix-ui/react-avatar
```

#### 2.2 基础配置 (30行)
```typescript
// frontend/app/lib/config.ts
export const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080';

// frontend/app/lib/types.ts
export interface Literature {
  itemID: number;
  title: string;
  authors: string;
  year: number;
  pdfPath?: string;
  abstract?: string;
}

export interface ChatMessage {
  id: string;
  role: 'user' | 'assistant';
  content: string;
  timestamp: string;
}
```

#### 2.3 API客户端 (50行)
```typescript
// frontend/app/lib/api.ts
import { API_BASE_URL } from './config';
import { Literature, ChatMessage } from './types';

export const api = {
  // 文献相关
  async getLiteratureList(): Promise<Literature[]> {
    const response = await fetch(`${API_BASE_URL}/api/literature`);
    const data = await response.json();
    return data.data;
  },

  async getLiteratureDetail(id: number): Promise<Literature> {
    const response = await fetch(`${API_BASE_URL}/api/literature/${id}`);
    const data = await response.json();
    return data.data;
  },

  // AI对话相关
  async startChat(message: string, docIds?: number[]): Promise<{ id: string; messages: ChatMessage[] }> {
    const response = await fetch(`${API_BASE_URL}/api/ai/chat`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ message, doc_ids: docIds }),
    });
    const data = await response.json();
    return data.data;
  },

  async continueChat(chatId: string, message: string): Promise<{ messages: ChatMessage[] }> {
    const response = await fetch(`${API_BASE_URL}/api/ai/chat/${chatId}`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ message }),
    });
    const data = await response.json();
    return data.data;
  },
};
```

### Phase 3: 核心UI组件实现 (第4-7天)

#### 3.1 文献卡片组件 (50行)
```typescript
// frontend/app/components/LiteratureCard.tsx
'use client';

import { Literature } from '@/lib/types';
import { api } from '@/lib/api';
import { useState } from 'react';
import { FileText, MessageCircle, Brain } from 'lucide-react';

interface LiteratureCardProps {
  literature: Literature;
  onChatStart?: (docId: number) => void;
}

export function LiteratureCard({ literature, onChatStart }: LiteratureCardProps) {
  const [isLoading, setIsLoading] = useState(false);

  const handleStartChat = async () => {
    setIsLoading(true);
    try {
      onChatStart?.(literature.itemID);
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <div className="bg-white rounded-lg shadow-md p-6 hover:shadow-lg transition-shadow">
      <div className="flex items-start justify-between mb-4">
        <div className="flex-1">
          <h3 className="text-lg font-semibold text-gray-900 mb-2 line-clamp-2">
            {literature.title}
          </h3>
          <p className="text-sm text-gray-600 mb-1">
            {literature.authors} • {literature.year}
          </p>
        </div>
        <FileText className="w-5 h-5 text-blue-500 ml-4 flex-shrink-0" />
      </div>

      {literature.abstract && (
        <p className="text-sm text-gray-700 mb-4 line-clamp-3">
          {literature.abstract}
        </p>
      )}

      <div className="flex gap-2">
        <button
          onClick={handleStartChat}
          disabled={isLoading}
          className="flex items-center gap-2 px-3 py-1.5 bg-blue-500 text-white text-sm rounded hover:bg-blue-600 disabled:opacity-50"
        >
          {isLoading ? (
            <>处理中...</>
          ) : (
            <>
              <Brain className="w-4 h-4" />
              AI分析
            </>
          )}
        </button>

        {literature.pdfPath && (
          <button className="flex items-center gap-2 px-3 py-1.5 border border-gray-300 text-sm rounded hover:bg-gray-50">
            <MessageCircle className="w-4 h-4" />
            阅读
          </button>
        )}
      </div>
    </div>
  );
}
```

#### 3.2 AI对话组件 (80行)
```typescript
// frontend/app/components/AIChat.tsx
'use client';

import { useState, useRef, useEffect } from 'react';
import { ChatMessage } from '@/lib/types';
import { api } from '@/lib/api';
import { Send, Bot, User } from 'lucide-react';

interface AIChatProps {
  initialDocId?: number;
  onMessage?: (message: string) => void;
}

export function AIChat({ initialDocId, onMessage }: AIChatProps) {
  const [messages, setMessages] = useState<ChatMessage[]>([]);
  const [input, setInput] = useState('');
  const [isLoading, setIsLoading] = useState(false);
  const [chatId, setChatId] = useState<string>('');
  const messagesEndRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    scrollToBottom();
  }, [messages]);

  const scrollToBottom = () => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' });
  };

  const handleSend = async () => {
    if (!input.trim() || isLoading) return;

    const userMessage: ChatMessage = {
      id: Date.now().toString(),
      role: 'user',
      content: input,
      timestamp: new Date().toISOString(),
    };

    setMessages(prev => [...prev, userMessage]);
    setInput('');
    setIsLoading(true);

    try {
      let response;
      if (chatId) {
        response = await api.continueChat(chatId, input);
      } else {
        response = await api.startChat(input, initialDocId ? [initialDocId] : undefined);
        setChatId(response.id);
      }

      const assistantMessage: ChatMessage = {
        id: Date.now().toString(),
        role: 'assistant',
        content: response.messages[response.messages.length - 1].content,
        timestamp: new Date().toISOString(),
      };

      setMessages(prev => [...prev, assistantMessage]);
      onMessage?.(assistantMessage.content);
    } catch (error) {
      console.error('AI对话失败:', error);
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <div className="bg-white rounded-lg shadow-md h-[600px] flex flex-col">
      <div className="p-4 border-b">
        <h3 className="font-semibold text-gray-900 flex items-center gap-2">
          <Bot className="w-5 h-5 text-blue-500" />
          AI学术助手
        </h3>
      </div>

      <div className="flex-1 overflow-y-auto p-4 space-y-4">
        {messages.length === 0 ? (
          <div className="text-center text-gray-500 py-8">
            <Bot className="w-12 h-12 mx-auto mb-4 text-gray-300" />
            <p>开始与AI助手对话，我可以帮您分析文献、回答问题</p>
          </div>
        ) : (
          messages.map((message) => (
            <div
              key={message.id}
              className={`flex gap-3 ${message.role === 'user' ? 'flex-row-reverse' : ''}`}
            >
              <div className={`w-8 h-8 rounded-full flex items-center justify-center flex-shrink-0 ${
                message.role === 'user' ? 'bg-blue-500' : 'bg-gray-200'
              }`}>
                {message.role === 'user' ? (
                  <User className="w-4 h-4 text-white" />
                ) : (
                  <Bot className="w-4 h-4 text-gray-600" />
                )}
              </div>
              <div className={`max-w-[70%] px-4 py-2 rounded-lg ${
                message.role === 'user'
                  ? 'bg-blue-500 text-white'
                  : 'bg-gray-100 text-gray-900'
              }`}>
                <p className="text-sm">{message.content}</p>
              </div>
            </div>
          ))
        )}

        {isLoading && (
          <div className="flex gap-3">
            <div className="w-8 h-8 rounded-full bg-gray-200 flex items-center justify-center">
              <Bot className="w-4 h-4 text-gray-600" />
            </div>
            <div className="bg-gray-100 px-4 py-2 rounded-lg">
              <div className="flex gap-1">
                <div className="w-2 h-2 bg-gray-400 rounded-full animate-bounce"></div>
                <div className="w-2 h-2 bg-gray-400 rounded-full animate-bounce" style={{ animationDelay: '0.1s' }}></div>
                <div className="w-2 h-2 bg-gray-400 rounded-full animate-bounce" style={{ animationDelay: '0.2s' }}></div>
              </div>
            </div>
          </div>
        )}
        <div ref={messagesEndRef} />
      </div>

      <div className="p-4 border-t">
        <div className="flex gap-2">
          <input
            type="text"
            value={input}
            onChange={(e) => setInput(e.target.value)}
            onKeyPress={(e) => e.key === 'Enter' && handleSend()}
            placeholder="输入您的问题..."
            className="flex-1 px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
            disabled={isLoading}
          />
          <button
            onClick={handleSend}
            disabled={isLoading || !input.trim()}
            className="px-4 py-2 bg-blue-500 text-white rounded-lg hover:bg-blue-600 disabled:opacity-50 disabled:cursor-not-allowed"
          >
            <Send className="w-4 h-4" />
          </button>
        </div>
      </div>
    </div>
  );
}
```

#### 3.3 智能搜索组件 (60行)
```typescript
// frontend/app/components/SearchBar.tsx
'use client';

import { useState, useEffect } from 'react';
import { Search, Sparkles } from 'lucide-react';
import { api } from '@/lib/api';
import { Literature } from '@/lib/types';

interface SearchBarProps {
  onResults: (results: Literature[]) => void;
  onSearch?: (query: string) => void;
}

export function SearchBar({ onResults, onSearch }: SearchBarProps) {
  const [query, setQuery] = useState('');
  const [suggestions, setSuggestions] = useState<string[]>([]);
  const [showSuggestions, setShowSuggestions] = useState(false);
  const [isLoading, setIsLoading] = useState(false);

  const quickSuggestions = [
    "显示最近添加的文献",
    "推荐高影响力论文",
    "查找机器学习相关研究",
    "按年份分组显示",
  ];

  useEffect(() => {
    if (query.length > 2) {
      setIsLoading(true);
      const timer = setTimeout(() => {
        // 这里可以实现智能搜索建议
        setSuggestions(quickSuggestions.filter(s => s.includes(query)));
        setIsLoading(false);
      }, 300);
      return () => clearTimeout(timer);
    } else {
      setSuggestions([]);
      setShowSuggestions(false);
    }
  }, [query]);

  const handleSearch = async (searchQuery?: string) => {
    const finalQuery = searchQuery || query;
    if (!finalQuery.trim()) return;

    setIsLoading(true);
    try {
      const literature = await api.getLiteratureList();
      const filtered = literature.filter(item =>
        item.title.toLowerCase().includes(finalQuery.toLowerCase()) ||
        item.authors.toLowerCase().includes(finalQuery.toLowerCase())
      );
      onResults(filtered);
      onSearch?.(finalQuery);
      setShowSuggestions(false);
    } catch (error) {
      console.error('搜索失败:', error);
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <div className="relative">
      <div className="relative">
        <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 w-5 h-5 text-gray-400" />
        <input
          type="text"
          value={query}
          onChange={(e) => {
            setQuery(e.target.value);
            setShowSuggestions(true);
          }}
          onFocus={() => setShowSuggestions(true)}
          onKeyPress={(e) => e.key === 'Enter' && handleSearch()}
          placeholder="搜索文献标题、作者，或让AI推荐..."
          className="w-full pl-10 pr-4 py-3 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
        />
        {isLoading && (
          <div className="absolute right-3 top-1/2 transform -translate-y-1/2">
            <div className="w-4 h-4 border-2 border-blue-500 border-t-transparent rounded-full animate-spin"></div>
          </div>
        )}
      </div>

      {showSuggestions && suggestions.length > 0 && (
        <div className="absolute top-full left-0 right-0 mt-2 bg-white border border-gray-200 rounded-lg shadow-lg z-10">
          <div className="p-3 border-b border-gray-100">
            <div className="flex items-center gap-2 text-sm text-gray-600">
              <Sparkles className="w-4 h-4 text-blue-500" />
              <span>AI搜索建议</span>
            </div>
          </div>
          <div className="max-h-60 overflow-y-auto">
            {suggestions.map((suggestion, index) => (
              <button
                key={index}
                onClick={() => handleSearch(suggestion)}
                className="w-full px-4 py-3 text-left hover:bg-gray-50 transition-colors text-sm"
              >
                {suggestion}
              </button>
            ))}
          </div>
        </div>
      )}
    </div>
  );
}
```

### Phase 4: 主页面集成 (第8-10天)

#### 4.1 主页面组件 (80行)
```typescript
// frontend/app/page.tsx
'use client';

import { useState, useEffect } from 'react';
import { Literature } from '@/lib/types';
import { api } from '@/lib/api';
import { LiteratureCard } from '@/components/LiteratureCard';
import { AIChat } from '@/components/AIChat';
import { SearchBar } from '@/components/SearchBar';
import { BookOpen, Brain, Sparkles } from 'lucide-react';

export default function HomePage() {
  const [literature, setLiterature] = useState<Literature[]>([]);
  const [filteredLiterature, setFilteredLiterature] = useState<Literature[]>([]);
  const [selectedDocId, setSelectedDocId] = useState<number>();
  const [isLoading, setIsLoading] = useState(true);
  const [showAIChat, setShowAIChat] = useState(false);

  useEffect(() => {
    loadLiterature();
  }, []);

  const loadLiterature = async () => {
    try {
      setIsLoading(true);
      const data = await api.getLiteratureList();
      setLiterature(data);
      setFilteredLiterature(data);
    } catch (error) {
      console.error('加载文献失败:', error);
    } finally {
      setIsLoading(false);
    }
  };

  const handleSearch = (results: Literature[]) => {
    setFilteredLiterature(results);
  };

  const handleChatStart = (docId: number) => {
    setSelectedDocId(docId);
    setShowAIChat(true);
  };

  return (
    <div className="min-h-screen bg-gray-50">
      {/* 顶部导航 */}
      <header className="bg-white shadow-sm border-b">
        <div className="max-w-7xl mx-auto px-4 py-4">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-3">
              <BookOpen className="w-8 h-8 text-blue-500" />
              <div>
                <h1 className="text-2xl font-bold text-gray-900">ZoteroFlow</h1>
                <p className="text-sm text-gray-600">AI驱动的智能文献分析</p>
              </div>
            </div>
            <button
              onClick={() => setShowAIChat(!showAIChat)}
              className="flex items-center gap-2 px-4 py-2 bg-blue-500 text-white rounded-lg hover:bg-blue-600"
            >
              <Brain className="w-4 h-4" />
              AI助手
            </button>
          </div>
        </div>
      </header>

      <div className="max-w-7xl mx-auto px-4 py-8">
        {/* 统计信息 */}
        <div className="grid grid-cols-1 md:grid-cols-3 gap-6 mb-8">
          <div className="bg-white p-6 rounded-lg shadow-sm border">
            <div className="flex items-center gap-3">
              <BookOpen className="w-8 h-8 text-blue-500" />
              <div>
                <p className="text-2xl font-bold text-gray-900">{literature.length}</p>
                <p className="text-sm text-gray-600">文献总数</p>
              </div>
            </div>
          </div>
          <div className="bg-white p-6 rounded-lg shadow-sm border">
            <div className="flex items-center gap-3">
              <Brain className="w-8 h-8 text-green-500" />
              <div>
                <p className="text-2xl font-bold text-gray-900">AI分析</p>
                <p className="text-sm text-gray-600">深度理解</p>
              </div>
            </div>
          </div>
          <div className="bg-white p-6 rounded-lg shadow-sm border">
            <div className="flex items-center gap-3">
              <Sparkles className="w-8 h-8 text-orange-500" />
              <div>
                <p className="text-2xl font-bold text-gray-900">智能推荐</p>
                <p className="text-sm text-gray-600">个性化</p>
              </div>
            </div>
          </div>
        </div>

        {/* 搜索栏 */}
        <div className="mb-8">
          <SearchBar onResults={handleSearch} />
        </div>

        {/* 主内容区 */}
        <div className="grid grid-cols-1 lg:grid-cols-3 gap-8">
          {/* 文献列表 */}
          <div className="lg:col-span-2">
            <h2 className="text-xl font-semibold text-gray-900 mb-6">文献列表</h2>

            {isLoading ? (
              <div className="text-center py-12">
                <div className="w-8 h-8 border-2 border-blue-500 border-t-transparent rounded-full animate-spin mx-auto mb-4"></div>
                <p className="text-gray-600">正在加载文献...</p>
              </div>
            ) : filteredLiterature.length === 0 ? (
              <div className="text-center py-12">
                <BookOpen className="w-12 h-12 text-gray-300 mx-auto mb-4" />
                <p className="text-gray-600">暂无文献</p>
              </div>
            ) : (
              <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                {filteredLiterature.map((item) => (
                  <LiteratureCard
                    key={item.itemID}
                    literature={item}
                    onChatStart={handleChatStart}
                  />
                ))}
              </div>
            )}
          </div>

          {/* AI对话面板 */}
          {showAIChat && (
            <div className="lg:col-span-1">
              <AIChat initialDocId={selectedDocId} />
            </div>
          )}
        </div>
      </div>
    </div>
  );
}
```

#### 4.2 根布局 (50行)
```typescript
// frontend/app/layout.tsx
import './globals.css';
import type { Metadata } from 'next';

export const metadata: Metadata = {
  title: 'ZoteroFlow - AI驱动的智能文献分析',
  description: '让AI成为你的学术研究伙伴',
};

export default function RootLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <html lang="zh-CN">
      <body className="antialiased">
        {children}
      </body>
    </html>
  );
}
```

## 🔧 前后端对接关键点

### 1. API端点映射
```go
// 现有核心服务 → Web API端点
ZoteroDB.GetItemsWithPDF()     → GET /api/literature
AIConversationManager.Start() → POST /api/ai/chat
MinerUClient.ParsePDF()        → POST /api/literature/:id/analyze
```

### 2. 数据流设计
```
前端组件 → API调用 → 现有核心服务 → 数据库/外部API → 返回结果
```

### 3. 错误处理策略
- 前端：用户友好的错误提示
- 后端：结构化错误响应
- 现有组件：保持原有错误处理逻辑

## 📊 优化后的实现指标

### 代码量分配
- **现有后端**: 2300行 (无需修改)
- **Web API层**: 200行 (新增)
- **前端界面**: 600行 (优化后)
- **总代码量**: 3100行

### 开发时间安排
- **Week 1**: Web API层 + 前端基础架构
- **Week 2**: UI组件开发 + 集成测试

### 技术优势
1. **最大化复用** - 95%的后端代码直接复用
2. **最小化改动** - 只增加必要的Web API层
3. **快速开发** - 2周完成完整UI实现
4. **稳定可靠** - 基于成熟的核心组件

## 🎯 部署配置

### 开发环境
```bash
# 后端 (端口8080)
cd server
go run . -web

# 前端 (端口3000)
cd frontend
npm run dev
```

### 生产环境
- **后端**: Railway + Docker
- **前端**: Vercel
- **数据库**: 现有Zotero SQLite

---

**文档版本**: v2.0 (基于现有后端优化)
**创建时间**: 2025-10-04
**核心策略**: 复用现有2300行成熟代码，只新增800行实现完整UI