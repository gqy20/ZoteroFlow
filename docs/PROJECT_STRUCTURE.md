# ZoteroFlow2 项目结构说明

## 📁 目录结构

```
ZoteroFlow2/
├── 📚 docs/                    # 文档目录
│   ├── api/                   # API文档
│   ├── reports/               # 测试报告和技术文档
│   ├── mcp-integration-plan.md # MCP集成方案
│   ├── external-mcp-configuration.md # 外部MCP配置
│   └── README-MCP.md          # MCP使用指南
│
├── 🔧 tools/                   # 开发工具脚本
│   └── MCP_STATUS_CHECK.go    # 项目状态检查工具
│
├── 🧪 tests/                   # 测试目录
│   ├── core/                  # 核心功能测试
│   ├── integration/           # 集成测试
│   ├── mcp/                   # MCP协议测试
│   └── test_article_mcp.py    # Python集成测试
│
├── 🖥️ server/                  # 主服务器代码
│   ├── bin/                   # 编译后的二进制文件
│   ├── config/                # 配置管理
│   ├── core/                  # 核心功能模块
│   ├── mcp/                   # MCP服务器实现
│   └── external-mcp-servers.json # 外部MCP配置
│
└── 📜 scripts/                 # 构建和部署脚本
```

## 🗂️ 文件分类说明

### 📚 文档 (docs/)

**技术文档**:
- `mcp-integration-plan.md` - 完整的MCP集成架构设计
- `external-mcp-configuration.md` - 外部MCP服务器配置指南
- `api/mcp-tools-list.md` - MCP工具详细说明

**报告文档** (`docs/reports/`):
- `MCP-INTEGRATION-SUMMARY.md` - MCP集成项目总结
- `MAIN_WORKFLOW_TEST_RESULTS.md` - 主流程测试报告
- `ARTICLE-MCP-TEST-RESULTS.md` - Article MCP测试结果
- `MCP-USAGE-GUIDE.md` - 完整使用指南

### 🔧 工具 (tools/)

- `MCP_STATUS_CHECK.go` - 项目状态全面检查工具
  - 检查项目结构完整性
  - 验证编译状态
  - 测试MCP服务器功能
  - 验证外部MCP集成

### 🧪 测试 (tests/)

**核心功能测试** (`tests/core/`):
- `test_complete.go` - 完整功能测试
- `test_flow.go` - 工作流程测试

**MCP协议测试** (`tests/mcp/`):
- `test_mcp_basic.go` - MCP服务器基础测试
- `test_article_mcp_debug.go` - Article MCP调试测试

**集成测试** (`tests/integration/`):
- `test_article_mcp_integration.go` - Article MCP集成测试
- `test_external_mcp_integration.go` - 外部MCP集成测试
- `test_main_article_mcp.go` - 主流程测试
- `test_article_mcp_main.go` - Article MCP主流程测试

**Python测试**:
- `test_article_mcp.py` - Article MCP Python集成测试

### 🖥️ 服务器 (server/)

**核心代码**:
- `main.go` - 主程序入口
- `config/` - 配置管理模块
- `core/` - 核心功能（ZoteroDB、MinerU、AI等）
- `mcp/` - MCP服务器实现

**配置文件**:
- `external-mcp-servers.json` - 外部MCP服务器配置
- `.env` - 环境变量配置

**二进制文件** (`bin/`):
- `zoteroflow2` - 主程序二进制文件
- `zoteroflow2-prod` - 生产环境版本

## 🚀 使用指南

### 1. 开发环境设置

```bash
# 检查项目状态
go run tools/MCP_STATUS_CHECK.go

# 编译项目
cd server && make build

# 运行基础测试
go run tests/mcp/test_mcp_basic.go
```

### 2. MCP服务器使用

```bash
# 启动主MCP服务器
./server/bin/zoteroflow2 mcp

# 启动Article MCP（独立终端）
uvx article-mcp server

# 配置验证
go run server/tests/test-mcp-ai-config.go
```

### 3. 测试执行

```bash
# MCP基础测试
go run tests/mcp/test_mcp_basic.go

# 集成测试
go run tests/integration/test_main_article_mcp.go

# Python集成测试
cd tests && python3 test_article_mcp.py
```

## 📊 项目状态

### ✅ 已完成功能

- **MCP服务器**: 6个本地工具 + 外部MCP支持框架
- **Article MCP集成**: 10个学术工具，完全测试通过
- **配置管理**: 自动配置生成和验证
- **测试覆盖**: 基础功能 + 集成测试 + 工作流测试
- **文档完整**: 从使用指南到技术架构的完整文档

### 🔧 开发工具

- **状态检查**: `tools/MCP_STATUS_CHECK.go`
- **配置验证**: `server/tests/test-mcp-ai-config.go`
- **基础测试**: `tests/mcp/test_mcp_basic.go`
- **集成测试**: `tests/integration/` 目录下的所有测试

### 📈 性能指标

- **本地工具响应**: < 1秒
- **Article MCP搜索**: 1-5秒
- **数据库容量**: 986个文献项目
- **工具总数**: 16个专业工具

## 🎯 最佳实践

### 开发流程

1. **开发前**: 运行 `go run tools/MCP_STATUS_CHECK.go` 检查状态
2. **编码后**: 运行 `make quick` 进行快速检查
3. **提交前**: 运行 `make check` 进行完整检查
4. **发布前**: 运行完整测试套件

### 测试策略

- **单元测试**: `tests/core/` 目录
- **MCP协议测试**: `tests/mcp/` 目录
- **集成测试**: `tests/integration/` 目录
- **端到端测试**: Python测试脚本

### 文档维护

- **API文档**: `docs/api/` 目录
- **技术文档**: `docs/` 根目录
- **测试报告**: `docs/reports/` 目录
- **使用指南**: `docs/README-MCP.md`

## 🔄 项目维护

### 定期检查

```bash
# 每周执行一次完整状态检查
go run tools/MCP_STATUS_CHECK.go

# 每月执行一次完整测试
cd tests && python3 test_article_mcp.py
go run tests/integration/test_main_article_mcp.go
```

### 清理策略

- **临时文件**: `git clean -fd` 清理未跟踪文件
- **编译文件**: `make clean` 清理编译产物
- **缓存文件**: 定期清理 `data/cache/` 目录

---

**文档版本**: v1.0
**创建日期**: 2025-10-04
**维护者**: ZoteroFlow2 开发团队
**更新频率**: 根据项目变化更新