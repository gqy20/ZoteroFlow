# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## 项目概述

ZoteroFlow2 是一个基于 Go 的智能文献分析 MCP 服务器，通过 MinerU 解析本地 PDF，结合 Article MCP 扩展搜索，为 AI 提供结构化文献访问能力。

**核心架构**：
- **本地文献解析层**: MinerU PDF 解析 + Zotero 数据库访问
- **MCP 兼容服务层**: 支持 Article MCP 等多个 MCP 服务
- **AI 智能分析引擎**: GLM-4.6 模型集成

## 常用命令

### 服务器端 (server/)

**构建和运行**：
```bash
cd server/
make build              # 构建二进制到 bin/zoteroflow2
make run                # 构建并运行
make dev                # 直接运行 go run .
```

**测试**：
```bash
make test               # 运行测试（含竞态检测）
make test-coverage      # 生成覆盖率报告 (coverage.html)
go test -v ./...        # 详细测试输出
```

**代码质量**：
```bash
make fmt                # 格式化代码
make lint               # 运行 golangci-lint
make vet                # 运行 go vet
make check              # 依次执行 fmt, lint, test
make quick              # 快速检查 (fmt + vet)
```

**依赖管理**：
```bash
make deps               # 下载依赖并整理 go.mod
make mod-upgrade        # 升级所有依赖
```

**特殊测试**：
```bash
go run test_mineru.go   # MinerU API 集成测试
```

### 测试端 (tests/)

**MCP 集成测试**：
```bash
cd tests/
python3 test_article_mcp.py    # Article MCP 完整集成测试
go run test_complete.go        # 完整工作流测试
go run test_flow.go           # 流程测试
```

## 核心架构

### 主要组件

1. **ZoteroDB** (`server/core/zotero.go`)
   - 只读访问 Zotero SQLite 数据库
   - 提取文献元数据和 PDF 附件路径
   - 处理 Zotero 存储系统 (storage:XXXXXX.pdf 格式)

2. **MinerUClient** (`server/core/mineru.go`)
   - HTTP 客户端，支持单个和批量 PDF 解析
   - 文件上传和结果轮询机制
   - 使用命名类型替代匿名结构体

3. **PDFParser** (`server/core/parser.go`)
   - 协调 Zotero 和 MinerU 集成
   - 管理解析结果缓存
   - 处理 PDF 文件发现流程

4. **配置管理** (`server/config/config.go`)
   - 环境变量和 `.env` 文件支持
   - 路径展开和验证功能

### 配置要求

通过环境变量或 `.env` 文件配置：
- `ZOTERO_DB_PATH` - Zotero SQLite 数据库路径
- `ZOTERO_DATA_DIR` - Zotero 存储目录路径
- `MINERU_API_URL` - MinerU API 端点
- `MINERU_TOKEN` - MinerU 认证令牌
- `AI_*` 变量 - AI 模型配置

### 数据流程

1. 从 `.env` 和环境变量加载配置
2. 连接 Zotero 数据库（只读模式）
3. 创建 MinerU 客户端
4. 查询 Zotero 获取 PDF 项目
5. 对每个 PDF：查找文件 → 上传到 MinerU → 获取解析结果
6. 在 `CACHE_DIR` 中缓存结果

## 开发规范

### 代码标准
- 使用英文日志（无表情符号或中文字符）
- 优先使用命名类型而非匿名结构体
- 遵循约定式提交格式：`<type>(<scope>): <description>`
- Pre-commit 钩子强制执行格式化和基础检查
- Pre-push 钩子要求 80% 测试覆盖率

### Git 钩子
项目使用自动化 git 钩子进行质量控制：
- **Pre-commit**: 格式化、go vet、基础检查
- **Pre-push**: 完整测试套件和覆盖率
- **Commit-msg**: 约定式提交格式验证

### 测试策略
- 当前单元测试较少（覆盖率显示 0%）
- `test_mineru.go` 中的集成测试验证 MinerU API 连接性
- 使用 `make test-coverage` 生成覆盖率报告

## MCP 集成

### Article MCP 已验证功能
- Europe PMC + arXiv 文献搜索
- 10 个文献搜索分析工具
- 完全兼容 MCP v2024-11-05 标准
- 支持与 Claude Desktop、Continue.dev 等 AI 客户端集成

### 测试验证
- Article MCP 集成已通过完整测试 (`tests/test_article_mcp.py`)
- MCP 协议通信正常
- 工具发现和调用功能正常
- 与 MCP 生态完全兼容

## 项目状态

### ✅ 已完成 (v0.8)
- Article MCP 集成 (~300行)
- Go MCP 客户端 (~200行)
- AI 智能分析 (~100行)

### 🔧 待实现
- MinerU PDF 解析 (~200行)
- Zotero 数据库访问 (~100行)
- 整合工作流 (~100行)

**总代码约束**: 不超过 1000 行（核心功能优先）

## 开发提示

- 使用 `make help` 查看所有可用命令
- 开发前先运行 `make quick` 进行快速检查
- 提交前确保 `make check` 通过
- MinerU 测试需要有效的 `MINERU_TOKEN` 环境变量
- Zotero 数据库路径需要正确配置才能运行完整测试