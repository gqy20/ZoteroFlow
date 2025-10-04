# ZoteroFlow2 MCP 工具列表

## 本地工具

### Zotero 数据库工具

#### zotero_search
搜索Zotero数据库中的文献

**参数:**
- `query` (string, 必需): 搜索关键词
- `limit` (integer, 可选): 返回结果数量限制，默认10

**返回:** 匹配的文献列表，包含标题、作者、年份等信息

#### zotero_list_items
列出Zotero数据库中的文献项目

**参数:**
- `limit` (integer, 可选): 返回结果数量限制，默认20
- `offset` (integer, 可选): 跳过的项目数量，默认0

**返回:** 文献项目列表

#### zotero_find_by_doi
根据DOI查找文献

**参数:**
- `doi` (string, 必需): 文献的DOI标识符

**返回:** 匹配的文献详细信息

#### zotero_get_stats
获取Zotero数据库统计信息

**参数:** 无

**返回:** 数据库中的文献总数、附件数量等统计信息

### MinerU PDF解析工具

#### mineru_parse
使用MinerU API解析PDF文件

**参数:**
- `file_path` (string, 必需): PDF文件路径
- `output_format` (string, 可选): 输出格式，默认"json"

**返回:** PDF解析结果，包含文本内容、表格、图片等结构化数据

### AI对话工具

#### zotero_chat
基于Zotero文献进行AI对话

**参数:**
- `message` (string, 必需): 用户消息
- `document_id` (string, 可选): 特定文献ID，用于基于单篇文献的对话
- `context_length` (integer, 可选): 上下文长度，默认4000

**返回:** AI助手的回复，基于提供的文献内容

## 外部工具（通过配置加载）

### Article MCP 工具集

当配置了article-mcp外部服务器时，可使用以下工具：

#### search_europe_pmc
搜索Europe PMC数据库

**参数:**
- `keyword` (string, 必需): 搜索关键词
- `max_results` (integer, 可选): 最大结果数，默认10
- `start_date` (string, 可选): 开始日期，格式YYYY-MM-DD
- `end_date` (string, 可选): 结束日期，格式YYYY-MM-DD

#### search_arxiv_papers
搜索arXiv预印本论文

**参数:**
- `keyword` (string, 必需): 搜索关键词
- `max_results` (integer, 可选): 最大结果数，默认10

#### get_article_details
获取文献详细信息

**参数:**
- `identifier` (string, 必需): 文献标识符（PMID、DOI或PMCID）
- `id_type` (string, 可选): 标识符类型，默认"pmid"

#### get_similar_articles
获取相似文献

**参数:**
- `identifier` (string, 必需): 文献标识符
- `max_results` (integer, 可选): 最大结果数，默认20

## 工具使用示例

### 基本搜索

```json
{
  "name": "zotero_search",
  "arguments": {
    "query": "machine learning",
    "limit": 5
  }
}
```

### DOI查询

```json
{
  "name": "zotero_find_by_doi",
  "arguments": {
    "doi": "10.1038/s41586-021-03819-2"
  }
}
```

### PDF解析

```json
{
  "name": "mineru_parse",
  "arguments": {
    "file_path": "/path/to/paper.pdf",
    "output_format": "json"
  }
}
```

### AI对话

```json
{
  "name": "zotero_chat",
  "arguments": {
    "message": "这篇论文的主要贡献是什么？",
    "document_id": "12345",
    "context_length": 3000
  }
}
```

### 学术文献搜索（需要article-mcp）

```json
{
  "name": "search_europe_pmc",
  "arguments": {
    "keyword": "COVID-19 vaccine",
    "max_results": 10,
    "start_date": "2020-01-01"
  }
}
```

## 工具调用格式

所有工具调用遵循MCP协议标准格式：

```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "method": "tools/call",
  "params": {
    "name": "工具名称",
    "arguments": {
      "参数名": "参数值"
    }
  }
}
```

## 返回格式

工具成功调用返回：

```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "result": {
    "content": [
      {
        "type": "text",
        "text": "工具执行结果"
      }
    ]
  }
}
```

错误情况返回：

```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "error": {
    "code": -32602,
    "message": "Invalid params",
    "data": "详细错误信息"
  }
}
```

## 错误代码

- `-32600`: Invalid Request - 请求格式错误
- `-32601`: Method not found - 工具不存在
- `-32602`: Invalid params - 参数错误
- `-32603`: Internal error - 内部服务器错误
- `-32000`: Server error - 服务器启动错误
- `-32001`: Tool execution error - 工具执行错误