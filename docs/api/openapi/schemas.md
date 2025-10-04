# OpenAPI 数据模型文档

## 概述

本文档详细说明了 ZoteroFlow2 API 中使用的所有数据模型，包括请求和响应的结构、字段含义、验证规则等。

## 基础数据模型

### BaseResponse

基础响应结构，所有API响应都包含此结构。

```yaml
BaseResponse:
  type: object
  properties:
    success:
      type: boolean
      description: 请求是否成功
    message:
      type: string
      description: 响应消息
    timestamp:
      type: string
      format: date-time
      description: 响应时间
  required:
    - success
    - message
    - timestamp
```

**字段说明**:
- `success`: 布尔值，表示请求是否成功处理
- `message`: 字符串，提供响应的描述信息
- `timestamp`: ISO 8601格式的日期时间，表示响应生成时间

**示例**:
```json
{
  "success": true,
  "message": "操作成功",
  "timestamp": "2024-12-01T10:30:00Z"
}
```

### ErrorResponse

错误响应结构，用于返回错误信息。

```yaml
ErrorResponse:
  allOf:
    - $ref: '#/components/schemas/BaseResponse'
  - type: object
  properties:
    error:
      type: object
      properties:
        code:
          type: string
          description: 错误代码
        message:
          type: string
          description: 错误消息
        details:
          type: object
          description: 错误详情
      required:
        - code
        - message
  required:
    - error
```

**字段说明**:
- `error.code`: 错误代码，用于程序化处理
- `error.message`: 错误消息，提供用户友好的错误描述
- `error.details`: 错误详情对象，包含额外的错误信息

**示例**:
```json
{
  "success": false,
  "message": "请求失败",
  "timestamp": "2024-12-01T10:30:00Z",
  "error": {
    "code": "INVALID_PARAMETER",
    "message": "参数无效",
    "details": {
      "field": "limit",
      "reason": "值超出范围"
    }
  }
}
```

### Pagination

分页信息结构，用于列表响应。

```yaml
Pagination:
  type: object
  properties:
    page:
      type: integer
      minimum: 1
      description: 当前页码
    limit:
      type: integer
      minimum: 1
      maximum: 100
      description: 每页记录数
    total:
      type: integer
      description: 总记录数
    total_pages:
      type: integer
      description: 总页数
    has_next:
      type: boolean
      description: 是否有下一页
    has_prev:
      type: boolean
      description: 是否有上一页
  required:
    - page
    - limit
    - total
    - total_pages
```

**字段说明**:
- `page`: 当前页码，从1开始
- `limit`: 每页记录数，范围1-100
- `total`: 总记录数
- `total_pages`: 总页数
- `has_next`: 是否有下一页
- `has_prev`: 是否有上一页

**示例**:
```json
{
  "page": 1,
  "limit": 20,
  "total": 156,
  "total_pages": 8,
  "has_next": true,
  "has_prev": false
}
```

## 文献相关模型

### LiteratureItem

文献项目结构，表示单个文献的基本信息。

```yaml
LiteratureItem:
  type: object
  properties:
    id:
      type: integer
      format: int64
      description: 文献ID
    title:
      type: string
      description: 文献标题
    authors:
      type: array
      items:
        type: string
      description: 作者列表
    year:
      type: integer
      description: 发表年份
    journal:
      type: string
      description: 期刊名称
    doi:
      type: string
      description: DOI标识符
    abstract:
      type: string
      description: 摘要
    keywords:
      type: array
      items:
        type: string
      description: 关键词列表
    has_pdf:
      type: boolean
      description: 是否有PDF附件
    pdf_path:
      type: string
      description: PDF文件路径
    created_at:
      type: string
      format: date-time
      description: 创建时间
    updated_at:
      type: string
      format: date-time
      description: 更新时间
  required:
    - id
    - title
    - authors
    - created_at
    - updated_at
```

**字段说明**:
- `id`: 唯一标识符，64位整数
- `title`: 文献标题，字符串
- `authors`: 作者列表，字符串数组
- `year`: 发表年份，整数
- `journal`: 期刊名称，字符串
- `doi`: DOI标识符，字符串
- `abstract`: 摘要，字符串
- `keywords`: 关键词列表，字符串数组
- `has_pdf`: 是否有PDF附件，布尔值
- `pdf_path`: PDF文件路径，字符串
- `created_at`: 创建时间，ISO 8601格式
- `updated_at`: 更新时间，ISO 8601格式

**示例**:
```json
{
  "id": 12345,
  "title": "机器学习在医疗诊断中的应用研究",
  "authors": ["张三", "李四", "王五"],
  "year": 2024,
  "journal": "人工智能医学",
  "doi": "10.1234/ai.med.2024.001",
  "abstract": "本文研究了机器学习技术在医疗诊断中的应用...",
  "keywords": ["机器学习", "医疗诊断", "人工智能"],
  "has_pdf": true,
  "pdf_path": "/path/to/paper.pdf",
  "created_at": "2024-12-01T10:30:00Z",
  "updated_at": "2024-12-01T10:30:00Z"
}
```

### LiteratureListResponse

文献列表响应结构。

```yaml
LiteratureListResponse:
  allOf:
    - $ref: '#/components/schemas/BaseResponse'
  - type: object
  properties:
    data:
      type: object
      properties:
        items:
          type: array
          items:
            $ref: '#/components/schemas/LiteratureItem'
        pagination:
          $ref: '#/components/schemas/Pagination'
      required:
        - items
        - pagination
  required:
    - data
```

**示例**:
```json
{
  "success": true,
  "message": "获取文献列表成功",
  "timestamp": "2024-12-01T10:30:00Z",
  "data": {
    "items": [
      {
        "id": 12345,
        "title": "机器学习在医疗诊断中的应用研究",
        "authors": ["张三", "李四"],
        "year": 2024,
        "created_at": "2024-12-01T10:30:00Z",
        "updated_at": "2024-12-01T10:30:00Z"
      }
    ],
    "pagination": {
      "page": 1,
      "limit": 20,
      "total": 156,
      "total_pages": 8,
      "has_next": true,
      "has_prev": false
    }
  }
}
```

### SearchRequest

搜索请求结构。

```yaml
SearchRequest:
  type: object
  properties:
    query:
      type: string
      description: 搜索查询
    filters:
      type: object
      properties:
        authors:
          type: array
          items:
            type: string
          description: 作者过滤
        year_from:
          type: integer
          description: 起始年份
        year_to:
          type: integer
          description: 结束年份
        journal:
          type: string
          description: 期刊过滤
        has_pdf:
          type: boolean
          description: 是否有PDF
      description: 搜索过滤器
    sort:
      type: object
      properties:
        field:
          type: string
          enum: [title, year, journal, created_at]
          description: 排序字段
        order:
          type: string
          enum: [asc, desc]
          description: 排序方向
      description: 排序设置
    pagination:
      $ref: '#/components/schemas/Pagination'
  required:
    - query
```

**字段说明**:
- `query`: 搜索查询字符串，必需
- `filters`: 搜索过滤器对象
- `filters.authors`: 作者过滤，字符串数组
- `filters.year_from`: 起始年份，整数
- `filters.year_to`: 结束年份，整数
- `filters.journal`: 期刊过滤，字符串
- `filters.has_pdf`: 是否有PDF，布尔值
- `sort.field`: 排序字段，枚举值
- `sort.order`: 排序方向，枚举值
- `pagination`: 分页信息

**示例**:
```json
{
  "query": "机器学习",
  "filters": {
    "authors": ["张三"],
    "year_from": 2020,
    "year_to": 2024,
    "has_pdf": true
  },
  "sort": {
    "field": "year",
    "order": "desc"
  },
  "pagination": {
    "page": 1,
    "limit": 20
  }
}
```

### SearchResponse

搜索响应结构。

```yaml
SearchResponse:
  allOf:
    - $ref: '#/components/schemas/BaseResponse'
  - type: object
  properties:
    data:
      type: object
      properties:
        items:
          type: array
          items:
            $ref: '#/components/schemas/LiteratureItem'
        pagination:
          $ref: '#/components/schemas/Pagination'
        total:
          type: integer
          description: 总结果数
        search_time:
          type: number
          format: float
          description: 搜索耗时（秒）
      required:
        - items
        - pagination
        - total
  required:
    - data
```

**示例**:
```json
{
  "success": true,
  "message": "搜索完成",
  "timestamp": "2024-12-01T10:30:00Z",
  "data": {
    "items": [
      {
        "id": 12345,
        "title": "机器学习在医疗诊断中的应用研究",
        "authors": ["张三", "李四"],
        "year": 2024
      }
    ],
    "pagination": {
      "page": 1,
      "limit": 20,
      "total": 45,
      "total_pages": 3,
      "has_next": true,
      "has_prev": false
    },
    "total": 45,
    "search_time": 0.123
  }
}
```

## PDF解析相关模型

### ParseRequest

PDF解析请求结构。

```yaml
ParseRequest:
  type: object
  properties:
    file:
      type: string
      format: binary
      description: PDF文件
    options:
      type: object
      properties:
        language:
          type: string
          description: 处理语言
          default: ch
        ocr:
          type: boolean
          description: 是否启用OCR
          default: true
        extract_images:
          type: boolean
          description: 是否提取图片
          default: true
        extract_tables:
          type: boolean
          description: 是否提取表格
          default: true
      description: 解析选项
  required:
    - file
```

**字段说明**:
- `file`: PDF文件，二进制数据，必需
- `options`: 解析选项对象
- `options.language`: 处理语言，字符串，默认"ch"
- `options.ocr`: 是否启用OCR，布尔值，默认true
- `options.extract_images`: 是否提取图片，布尔值，默认true
- `options.extract_tables`: 是否提取表格，布尔值，默认true

**示例**:
```json
{
  "file": "<binary data>",
  "options": {
    "language": "ch",
    "ocr": true,
    "extract_images": true,
    "extract_tables": true
  }
}
```

### ParseResult

PDF解析结果结构。

```yaml
ParseResult:
  type: object
  properties:
    content:
      type: string
      description: 解析后的内容（Markdown格式）
    metadata:
      type: object
      properties:
        title:
          type: string
          description: 文档标题
        authors:
          type: array
          items:
            type: string
          description: 作者列表
        pages:
          type: integer
          description: 页数
        language:
          type: string
          description: 文档语言
      description: 文档元数据
    files:
      type: array
      items:
        type: object
        properties:
          type:
            type: string
            enum: [image, table, text]
            description: 文件类型
          name:
            type: string
            description: 文件名
          path:
            type: string
            description: 文件路径
          size:
            type: integer
            description: 文件大小（字节）
        description: 解析出的文件列表
    duration:
      type: number
      format: float
      description: 解析耗时（秒）
  required:
    - content
    - metadata
```

**字段说明**:
- `content`: 解析后的Markdown内容，字符串
- `metadata`: 文档元数据对象
- `metadata.title`: 文档标题，字符串
- `metadata.authors`: 作者列表，字符串数组
- `metadata.pages`: 页数，整数
- `metadata.language`: 文档语言，字符串
- `files`: 解析出的文件列表
- `files[].type`: 文件类型，枚举值
- `files[].name`: 文件名，字符串
- `files[].path`: 文件路径，字符串
- `files[].size`: 文件大小，整数
- `duration`: 解析耗时，浮点数

**示例**:
```json
{
  "content": "# 机器学习基础\n\n## 摘要\n本文介绍了机器学习的基本概念...",
  "metadata": {
    "title": "机器学习基础",
    "authors": ["张三", "李四"],
    "pages": 45,
    "language": "zh"
  },
  "files": [
    {
      "type": "image",
      "name": "figure1.png",
      "path": "/path/to/figure1.png",
      "size": 1024
    }
  ],
  "duration": 12.5
}
```

### ParseResponse

PDF解析响应结构。

```yaml
ParseResponse:
  allOf:
    - $ref: '#/components/schemas/BaseResponse'
  - type: object
  properties:
    data:
      type: object
      properties:
        task_id:
          type: string
          format: uuid
          description: 任务ID
        status:
          type: string
          enum: [pending, processing, completed, failed]
          description: 解析状态
        progress:
          type: integer
          minimum: 0
          maximum: 100
          description: 解析进度百分比
        result:
          $ref: '#/components/schemas/ParseResult'
        error:
          type: object
          properties:
            code:
              type: string
              description: 错误代码
            message:
              type: string
              description: 错误消息
          description: 错误信息
      required:
        - task_id
        - status
  required:
    - data
```

**字段说明**:
- `task_id`: 任务ID，UUID格式
- `status`: 解析状态，枚举值
- `progress`: 解析进度百分比，0-100
- `result`: 解析结果对象
- `error`: 错误信息对象

**示例**:
```json
{
  "success": true,
  "message": "PDF解析完成",
  "timestamp": "2024-12-01T10:30:00Z",
  "data": {
    "task_id": "123e4567-e89b-12d3-a456-426614174000",
    "status": "completed",
    "progress": 100,
    "result": {
      "content": "# 机器学习基础\n\n## 摘要\n本文介绍了机器学习的基本概念...",
      "metadata": {
        "title": "机器学习基础",
        "authors": ["张三", "李四"],
        "pages": 45,
        "language": "zh"
      },
      "files": [],
      "duration": 12.5
    }
  }
}
```

## AI对话相关模型

### ChatMessage

聊天消息结构。

```yaml
ChatMessage:
  type: object
  properties:
    role:
      type: string
      enum: [system, user, assistant]
      description: 消息角色
    content:
      type: string
      description: 消息内容
    timestamp:
      type: string
      format: date-time
      description: 消息时间
    metadata:
      type: object
      properties:
        document_ids:
          type: array
          items:
            type: integer
          description: 关联的文献ID
        query_type:
          type: string
          enum: [search, analysis, summary]
          description: 查询类型
      description: 消息元数据
  required:
    - role
    - content
    - timestamp
```

**字段说明**:
- `role`: 消息角色，枚举值
- `content`: 消息内容，字符串
- `timestamp`: 消息时间，ISO 8601格式
- `metadata`: 消息元数据对象
- `metadata.document_ids`: 关联的文献ID，整数数组
- `metadata.query_type`: 查询类型，枚举值

**示例**:
```json
{
  "role": "user",
  "content": "什么是机器学习？",
  "timestamp": "2024-12-01T10:30:00Z",
  "metadata": {
    "document_ids": [12345, 12346],
    "query_type": "analysis"
  }
}
```

### ChatRequest

聊天请求结构。

```yaml
ChatRequest:
  type: object
  properties:
    message:
      type: string
      description: 用户消息
    conversation_id:
      type: string
      format: uuid
      description: 对话ID（可选，用于继续对话）
    context:
      type: object
      properties:
        document_ids:
          type: array
          items:
            type: integer
          description: 关联的文献ID
        query_type:
          type: string
          enum: [search, analysis, summary]
          description: 查询类型
      description: 对话上下文
    options:
      type: object
      properties:
        model:
          type: string
          description: AI模型
          default: glm-4.6
        temperature:
          type: number
          minimum: 0
          maximum: 2
          default: 0.7
          description: 温度参数
        max_tokens:
          type: integer
          minimum: 1
          maximum: 4000
          default: 2000
          description: 最大Token数
      description: 对话选项
  required:
    - message
```

**字段说明**:
- `message`: 用户消息，字符串，必需
- `conversation_id`: 对话ID，UUID格式，可选
- `context`: 对话上下文对象
- `context.document_ids`: 关联的文献ID，整数数组
- `context.query_type`: 查询类型，枚举值
- `options`: 对话选项对象
- `options.model`: AI模型，字符串，默认"glm-4.6"
- `options.temperature`: 温度参数，浮点数，0-2，默认0.7
- `options.max_tokens`: 最大Token数，整数，1-4000，默认2000

**示例**:
```json
{
  "message": "什么是机器学习？",
  "context": {
    "document_ids": [12345, 12346],
    "query_type": "analysis"
  },
  "options": {
    "model": "glm-4.6",
    "temperature": 0.7,
    "max_tokens": 1000
  }
}
```

### ChatResponse

聊天响应结构。

```yaml
ChatResponse:
  allOf:
    - $ref: '#/components/schemas/BaseResponse'
  - type: object
  properties:
    data:
      type: object
      properties:
        conversation_id:
          type: string
          format: uuid
          description: 对话ID
        message:
          $ref: '#/components/schemas/ChatMessage'
        usage:
          type: object
          properties:
            prompt_tokens:
              type: integer
              description: 输入Token数
            completion_tokens:
              type: integer
              description: 输出Token数
            total_tokens:
              type: integer
              description: 总Token数
          description: Token使用情况
      required:
        - conversation_id
        - message
        - usage
  required:
    - data
```

**字段说明**:
- `conversation_id`: 对话ID，UUID格式
- `message`: AI回复消息对象
- `usage`: Token使用情况对象
- `usage.prompt_tokens`: 输入Token数，整数
- `usage.completion_tokens`: 输出Token数，整数
- `usage.total_tokens`: 总Token数，整数

**示例**:
```json
{
  "success": true,
  "message": "对话成功",
  "timestamp": "2024-12-01T10:30:00Z",
  "data": {
    "conversation_id": "123e4567-e89b-12d3-a456-426614174000",
    "message": {
      "role": "assistant",
      "content": "机器学习是人工智能的一个重要分支...",
      "timestamp": "2024-12-01T10:30:00Z"
    },
    "usage": {
      "prompt_tokens": 45,
      "completion_tokens": 128,
      "total_tokens": 173
    }
  }
}
```

### Conversation

对话结构。

```yaml
Conversation:
  type: object
  properties:
    id:
      type: string
      format: uuid
      description: 对话ID
    title:
      type: string
      description: 对话标题
    messages:
      type: array
      items:
        $ref: '#/components/schemas/ChatMessage'
      description: 消息列表
    created_at:
      type: string
      format: date-time
      description: 创建时间
    updated_at:
      type: string
      format: date-time
      description: 更新时间
  required:
    - id
    - messages
    - created_at
    - updated_at
```

**字段说明**:
- `id`: 对话ID，UUID格式
- `title`: 对话标题，字符串
- `messages`: 消息列表，ChatMessage数组
- `created_at`: 创建时间，ISO 8601格式
- `updated_at`: 更新时间，ISO 8601格式

**示例**:
```json
{
  "id": "123e4567-e89b-12d3-a456-426614174000",
  "title": "机器学习讨论",
  "messages": [
    {
      "role": "user",
      "content": "什么是机器学习？",
      "timestamp": "2024-12-01T10:30:00Z"
    },
    {
      "role": "assistant",
      "content": "机器学习是人工智能的一个重要分支...",
      "timestamp": "2024-12-01T10:30:05Z"
    }
  ],
  "created_at": "2024-12-01T10:30:00Z",
  "updated_at": "2024-12-01T10:30:05Z"
}
```

## MCP相关模型

### MCPTool

MCP工具结构。

```yaml
MCPTool:
  type: object
  properties:
    name:
      type: string
      description: 工具名称
    description:
      type: string
      description: 工具描述
    input_schema:
      type: object
      description: 输入参数模式（JSON Schema）
  required:
    - name
    - description
    - input_schema
```

**字段说明**:
- `name`: 工具名称，字符串
- `description`: 工具描述，字符串
- `input_schema`: 输入参数模式，JSON Schema对象

**示例**:
```json
{
  "name": "search_europe_pmc",
  "description": "搜索 Europe PMC 数据库中的文献",
  "input_schema": {
    "type": "object",
    "properties": {
      "keyword": {
        "type": "string",
        "description": "搜索关键词"
      },
      "max_results": {
        "type": "integer",
        "description": "最大结果数量",
        "default": 10
      }
    },
    "required": ["keyword"]
  }
}
```

### MCPCallRequest

MCP工具调用请求结构。

```yaml
MCPCallRequest:
  type: object
  properties:
    arguments:
      type: object
      description: 工具参数
      additionalProperties: true
    options:
      type: object
      properties:
        timeout:
          type: integer
          description: 超时时间（秒）
          default: 30
      description: 调用选项
  required:
    - arguments
```

**字段说明**:
- `arguments`: 工具参数对象，键值对
- `options`: 调用选项对象
- `options.timeout`: 超时时间，整数，默认30秒

**示例**:
```json
{
  "arguments": {
    "keyword": "machine learning",
    "max_results": 10
  },
  "options": {
    "timeout": 30
  }
}
```

### MCPCallResponse

MCP工具调用响应结构。

```yaml
MCPCallResponse:
  allOf:
    - $ref: '#/components/schemas/BaseResponse'
  - type: object
  properties:
    data:
      type: object
      properties:
        result:
          type: object
          description: 工具执行结果
        content:
          type: array
          items:
            type: object
            properties:
              type:
                type: string
                enum: [text, image, resource]
                description: 内容类型
              text:
                type: string
                description: 文本内容
              data:
                type: object
                description: 数据内容
              mime_type:
                type: string
                description: MIME类型
              uri:
                type: string
                description: 资源URI
            description: 工具输出内容
        duration:
          type: number
          format: float
          description: 执行耗时（秒）
      required:
        - result
        - content
  required:
    - data
```

**字段说明**:
- `result`: 工具执行结果对象
- `content`: 工具输出内容数组
- `content[].type`: 内容类型，枚举值
- `content[].text`: 文本内容，字符串
- `content[].data`: 数据内容，对象
- `content[].mime_type`: MIME类型，字符串
- `content[].uri`: 资源URI，字符串
- `duration`: 执行耗时，浮点数

**示例**:
```json
{
  "success": true,
  "message": "工具调用成功",
  "timestamp": "2024-12-01T10:30:00Z",
  "data": {
    "result": {
      "status": "success",
      "count": 10
    },
    "content": [
      {
        "type": "text",
        "text": "找到10篇相关文献"
      }
    ],
    "duration": 1.23
  }
}
```

## 系统相关模型

### HealthResponse

健康检查响应结构。

```yaml
HealthResponse:
  allOf:
    - $ref: '#/components/schemas/BaseResponse'
  - type: object
  properties:
    data:
      type: object
      properties:
        status:
          type: string
          enum: [healthy, unhealthy]
          description: 系统状态
        version:
          type: string
          description: 系统版本
        uptime:
          type: number
          format: float
          description: 运行时间（秒）
        services:
          type: object
          properties:
            database:
              type: object
              properties:
                status:
                  type: string
                  enum: [up, down]
                  description: 数据库状态
                response_time:
                  type: number
                  format: float
                  description: 响应时间（毫秒）
              description: 数据库服务状态
            mineru:
              type: object
              properties:
                status:
                  type: string
                  enum: [up, down]
                  description: MinerU服务状态
                response_time:
                  type: number
                  format: float
                  description: 响应时间（毫秒）
              description: MinerU服务状态
            ai:
              type: object
              properties:
                status:
                  type: string
                  enum: [up, down]
                  description: AI服务状态
                response_time:
                  type: number
                  format: float
                  description: 响应时间（毫秒）
              description: AI服务状态
            mcp:
              type: object
              properties:
                status:
                  type: string
                  enum: [up, down]
                  description: MCP服务状态
                response_time:
                  type: number
                  format: float
                  description: 响应时间（毫秒）
              description: MCP服务状态
          description: 服务状态
      required:
        - status
        - version
        - uptime
        - services
  required:
    - data
```

**字段说明**:
- `status`: 系统状态，枚举值
- `version`: 系统版本，字符串
- `uptime`: 运行时间，浮点数（秒）
- `services`: 服务状态对象
- `services.database`: 数据库服务状态
- `services.mineru`: MinerU服务状态
- `services.ai`: AI服务状态
- `services.mcp`: MCP服务状态

**示例**:
```json
{
  "success": true,
  "message": "系统健康",
  "timestamp": "2024-12-01T10:30:00Z",
  "data": {
    "status": "healthy",
    "version": "1.0.0",
    "uptime": 86400,
    "services": {
      "database": {
        "status": "up",
        "response_time": 15.5
      },
      "mineru": {
        "status": "up",
        "response_time": 123.4
      },
      "ai": {
        "status": "up",
        "response_time": 45.2
      },
      "mcp": {
        "status": "up",
        "response_time": 25.8
      }
    }
  }
}
```

### StatsResponse

系统统计响应结构。

```yaml
StatsResponse:
  allOf:
    - $ref: '#/components/schemas/BaseResponse'
  - type: object
  properties:
    data:
      type: object
      properties:
        literature:
          type: object
          properties:
            total:
              type: integer
              description: 总文献数
            with_pdf:
              type: integer
              description: 有PDF的文献数
            parsed:
              type: integer
              description: 已解析的文献数
          description: 文献统计
        pdf_parsing:
          type: object
          properties:
            total:
              type: integer
              description: 总解析任务数
            completed:
              type: integer
              description: 已完成任务数
            failed:
              type: integer
              description: 失败任务数
            avg_duration:
              type: number
              format: float
              description: 平均解析耗时（秒）
          description: PDF解析统计
        ai_chat:
          type: object
          properties:
            total_conversations:
              type: integer
              description: 总对话数
            total_messages:
              type: integer
              description: 总消息数
            avg_tokens_per_message:
              type: number
              format: float
              description: 平均每条消息Token数
          description: AI对话统计
        mcp:
          type: object
          properties:
            total_calls:
              type: integer
              description: 总MCP调用数
            successful_calls:
              type: integer
              description: 成功调用数
            failed_calls:
              type: integer
              description: 失败调用数
            avg_response_time:
              type: number
              format: float
              description: 平均响应时间（秒）
          description: MCP调用统计
        system:
          type: object
          properties:
            requests_per_minute:
              type: number
              format: float
              description: 每分钟请求数
            avg_response_time:
              type: number
              format: float
              description: 平均响应时间（毫秒）
            error_rate:
              type: number
              format: float
              description: 错误率（百分比）
          description: 系统统计
      required:
        - literature
        - pdf_parsing
        - ai_chat
        - mcp
        - system
  required:
    - data
```

**字段说明**:
- `literature`: 文献统计对象
- `pdf_parsing`: PDF解析统计对象
- `ai_chat`: AI对话统计对象
- `mcp`: MCP调用统计对象
- `system`: 系统统计对象

**示例**:
```json
{
  "success": true,
  "message": "统计信息",
  "timestamp": "2024-12-01T10:30:00Z",
  "data": {
    "literature": {
      "total": 156,
      "with_pdf": 120,
      "parsed": 85
    },
    "pdf_parsing": {
      "total": 100,
      "completed": 85,
      "failed": 15,
      "avg_duration": 12.5
    },
    "ai_chat": {
      "total_conversations": 50,
      "total_messages": 250,
      "avg_tokens_per_message": 85.5
    },
    "mcp": {
      "total_calls": 200,
      "successful_calls": 190,
      "failed_calls": 10,
      "avg_response_time": 1.2
    },
    "system": {
      "requests_per_minute": 15.5,
      "avg_response_time": 125.3,
      "error_rate": 2.1
    }
  }
}
```

## 数据验证规则

### 字符串验证

- **非空字符串**: 必须包含至少一个字符
- **长度限制**: 通常限制在1-2048字符之间
- **特殊字符**: 根据字段不同，可能有特殊字符限制

### 数值验证

- **整数范围**: 根据字段不同，有不同的最小值和最大值限制
- **浮点数精度**: 通常保留2位小数
- **格式验证**: 如日期时间必须符合ISO 8601格式

### 数组验证

- **非空数组**: 必须包含至少一个元素
- **元素类型**: 数组中所有元素必须是相同类型
- **长度限制**: 通常限制在1-1000个元素之间

### 对象验证

- **必需字段**: 必须包含所有必需字段
- **类型匹配**: 字段值必须符合指定的类型
- **嵌套验证**: 嵌套对象也必须符合相应的验证规则

## 错误代码

### 通用错误代码

- `INVALID_PARAMETER`: 参数无效
- `UNAUTHORIZED`: 未授权
- `FORBIDDEN`: 禁止访问
- `NOT_FOUND`: 资源不存在
- `INTERNAL_ERROR`: 内部错误
- `SERVICE_UNAVAILABLE`: 服务不可用

### 业务错误代码

- `LITERATURE_NOT_FOUND`: 文献不存在
- `PDF_PARSE_FAILED`: PDF解析失败
- `AI_CHAT_FAILED`: AI对话失败
- `MCP_TOOL_NOT_FOUND`: MCP工具不存在
- `MCP_CALL_FAILED`: MCP调用失败

### HTTP状态码映射

- `200`: 成功
- `400`: 客户端错误（参数无效等）
- `401`: 未授权
- `403`: 禁止访问
- `404`: 资源不存在
- `429`: 请求频率超限
- `500`: 服务器内部错误
- `503`: 服务不可用

这个数据模型文档详细说明了 ZoteroFlow2 API 中使用的所有数据结构，为开发者提供了完整的接口规范参考。