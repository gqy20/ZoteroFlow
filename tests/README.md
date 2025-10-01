# ZoteroFlow 测试目录

本目录包含ZoteroFlow项目的所有测试文件和测试结果报告。

## 📁 目录结构

### 🔬 核心测试文件

#### Go语言测试
- **`test_complete.go`** - 🔬 **主要测试文件** - 完整的Go MCP客户端 + AI集成测试
  - 支持从.env文件读取AI配置（GLM-4.6模型）
  - 实现完整的MCP协议（v2024-11-05）
  - 集成Article MCP服务器进行智能文献搜索和分析
  - 异步消息处理和错误恢复机制
  - 编译命令：`go build -o test_complete test_complete.go`

- **`test_complete`** - 编译后的可执行文件
- **`test_flow.go`** - Zotero工作流测试（MinerU API集成）
  - 测试Zotero数据库读取
  - PDF解析和MinerU API调用
  - 完整的文献处理工作流

#### Python测试
- **`test_article_mcp.py`** - Article MCP Python客户端测试
  - 验证与Article MCP服务器的完整通信
  - 支持多线程处理避免日志干扰
  - 测试所有10个可用工具
  - 可作为Go版本的独立验证

### 📊 测试报告

- **`ARTICLE_MCP_SUCCESS.md`** - Article MCP测试成功报告
- **`GO_AI_MCP_SUCCESS.md`** - Go + AI + MCP完整集成成功报告
- **`TEST_SUMMARY.md`** - 测试总结和项目状态报告

## 🚀 快速开始

### 运行Go + AI完整测试
```bash
cd tests
# 确保上级目录有正确的.env配置
cat ../.env | grep -E "(AI_API_KEY|AI_BASE_URL|AI_MODEL)"

# 编译并运行完整测试
go build -o test_complete test_complete.go
./test_complete
```

### 运行Python Article MCP测试
```bash
cd tests
python3 test_article_mcp.py
```

### 运行Zotero工作流测试
```bash
cd tests
# 需要Zotero数据库路径和MinerU token
export ZOTERO_DB_PATH=/path/to/zotero.sqlite
export MINERU_TOKEN=your_token

go run test_flow.go
```

## 📋 测试覆盖范围

### ✅ 已验证功能
- [x] MCP协议通信（JSON-RPC 2.0）
- [x] Article MCP服务器连接和工具调用
- [x] Go MCP客户端实现
- [x] AI智能分析集成（GLM-4.6模型）
- [x] 多数据库文献搜索（Europe PMC、arXiv）
- [x] 文献质量评估和引用分析
- [x] 配置文件管理（.env支持）

### 🔧 支持的MCP工具
1. `search_europe_pmc` - Europe PMC文献搜索
2. `search_arxiv_papers` - arXiv预印本搜索
3. `get_article_details` - 获取文献详细信息
4. `get_references_by_doi` - 通过DOI获取参考文献
5. `batch_enrich_references_by_dois` - 批量DOI参考文献补全
6. `get_similar_articles` - 获取相似文章
7. `get_citing_articles` - 获取引用文献
8. `get_literature_relations` - 文献关联信息
9. `get_journal_quality` - 期刊质量评估
10. `evaluate_articles_quality` - 批量文献质量评估

## 🛠️ 环境要求

### Go环境
- Go 1.19+
- article-mcp工具（通过uv安装）

### Python环境
- Python 3.8+
- uv工具（用于运行article-mcp）

### 配置文件
确保项目根目录有正确的`.env`文件：
```bash
AI_API_KEY=your_api_key_here
AI_BASE_URL=https://open.bigmodel.cn/api/coding/paas/v4
AI_MODEL=glm-4.6
```

## 📝 测试结果

所有测试均已通过验证：
- ✅ Go MCP客户端封装工作正常
- ✅ Article MCP服务器连接成功
- ✅ AI智能分析集成成功
- ✅ 端到端流程验证通过

系统已准备好用于生产环境！

## 🔗 相关文档

- 项目主目录：`../`
- 配置文件：`../.env`
- Go模块配置：`../go.mod`, `../go.sum`
- 文档目录：`../docs/`