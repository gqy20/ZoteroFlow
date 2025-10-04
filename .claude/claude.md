# ZoteroFlow2 开发指南

## 💡 Linus 式 10条核心原则

1. **简单优于复杂** - 优先实现最简单的可行方案，证明功能可用
2. **数据驱动设计** - 根据实际数据流程设计接口，避免过度抽象
3. **只读优先原则** - 绝不在MVP阶段写任何可能破坏用户数据的代码
4. **标准API遵从** - 严格按照官方API规范实现，避免自定义协议
5. **代码量控制** - 总行数控制在2000行以内，保持代码简洁可维护
6. **错误处理渐进** - 从基础错误处理开始，根据实际需要逐步完善
7. **单线程优先** - 避免过早并发，先用最简单的方式跑通流程
8. **内存流式处理** - 大文件使用流式读写，避免一次性加载到内存
9. **功能单一职责** - 每个函数只做一件事，避免功能耦合
10. **渐进式优化** - 先让基础功能跑通，再根据性能需求逐步优化

## 开发规范

### 命令执行规范
- 每次运行命令都直接使用完整路径，而不是相对路径
- 例如：使用 `/home/qy113/workspace/note/zo/ZoteroFlow2/server/bin/zoteroflow2` 而不是 `./bin/zoteroflow2`

### Git 提交信息规范

采用约定式提交格式：

```
<type>(<scope>): <subject>

<body>

<footer>
```

#### 类型 (type)
- `feat`: 新功能
- `fix`: 修复bug
- `docs`: 文档更新
- `style`: 代码格式化（不影响功能）
- `refactor`: 代码重构
- `test`: 测试相关
- `chore`: 构建过程或辅助工具的变动

#### 范围 (scope)
- `parser`: PDF解析相关
- `core`: 核心功能模块
- `mcp`: MCP服务器相关
- `ai`: AI集成相关
- `zotero`: Zotero数据库相关
- `config`: 配置管理
- `cli`: 命令行界面

#### 示例
```
feat(parser): add MinerU PDF parsing support

Implement PDF parsing functionality using MinerU API with support for:
- Single file processing
- Batch processing
- Error handling and retry logic

Closes #123
```

```
fix(core): resolve CSV record encoding issue

Fix UTF-8 encoding problems when writing literature records to CSV files.
Ensure proper character handling for non-English content.
```

```
docs(readme): update installation instructions

Add detailed setup guide for Zotero database configuration and MinerU API setup.
Include troubleshooting section for common issues.
```