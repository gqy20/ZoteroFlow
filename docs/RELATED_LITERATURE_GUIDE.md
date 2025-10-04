# ZoteroFlow2 智能文献分析命令使用指南

## 🎯 功能概述

`related` 命令是 ZoteroFlow2 的一个强大功能，它能够：
- 查找本地文献库中的相关文献
- 调用 Article MCP 搜索全球学术数据库
- 使用 AI 进行智能分析和综合
- 提供完整的研究工作流支持

## 📋 命令语法

```bash
./server/bin/zoteroflow2 related <文献名称/DOI> [问题]
```

### 参数说明

- **文献名称/DOI**: 要分析的文献标识符
  - 可以是文献标题关键词（如 "机器学习"）
  - 可以是 DOI 号码（如 "10.1038/nature12373"）
  - 可以是文献的任何可识别信息

- **问题** (可选): 针对文献的具体问题
  - 如果不提供问题，将使用默认问题："请分析这篇文献并找到相关研究"
  - 可以是任何关于文献的问题

## 🚀 使用示例

### 基本用法

```bash
# 仅查找相关文献
./server/bin/zoteroflow2 related "机器学习"

# 带问题进行分析
./server/bin/zoteroflow2 related "机器学习" "这篇论文的主要贡献是什么？"

# 使用DOI查找文献
./server/bin/zoteroflow2 related "10.1038/nature12373" "找到相似的研究"

# 更具体的问题
./server/bin/zoteroflow2 related "深度学习" "这个研究方法可以应用到医学影像吗？"
```

## 🔧 工作流程详解

### 步骤1: 本地文献查找
- 在 Zotero 数据库中搜索匹配的文献
- 支持标题关键词搜索
- 支持 DOI 精确匹配
- 返回文献的详细信息

### 步骤2: 全球文献搜索
- 启动 Article MCP 服务
- 使用自动提取的关键词搜索 Europe PMC
- 获取相关的全球学术文献
- 支持高性能批量搜索

### 步骤3: AI智能分析
- 结合本地文献和全球文献信息
- 使用 GLM-4.6 模型进行智能分析
- 回答用户的特定问题
- 提供研究建议和见解

### 步骤4: 结果展示
- 显示找到的文献详细信息
- 提供AI分析结果
- 列出相关文献的完整引用信息

## 📊 输出示例

### 成功执行示例

```
🔍 正在分析文献: 机器学习
❓ 用户问题: 这篇论文的主要贡献是什么？

📚 步骤1: 查找本地文献...
✅ 找到本地文献: Machine Learning Fundamentals
   作者: 张三, 李四, 王五
   期刊: Journal of Artificial Intelligence
   年份: 2023
   DOI: 10.1000/ml.example

🌐 步骤2: 搜索全球相关文献...
   🔍 搜索关键词: machine learning fundamentals
✅ 找到 5 篇相关文献

🤖 步骤3: AI智能分析...
✅ AI分析完成

=== AI分析结果 ===
这篇关于机器学习基础的研究论文做出了以下主要贡献：

1. 理论框架创新：提出了新的机器学习理论框架...
2. 实验验证：通过大量实验验证了方法的有效性...
3. 应用拓展：展示了在多个领域的应用可能性...

📋 相关文献详情:

1. Deep Learning Approaches in Pattern Recognition
   作者: Research Team A
   期刊: Neural Networks Journal
   年份: 2024
   DOI: 10.1000/example1

2. Statistical Learning Theory and Applications
   作者: Research Team B
   期刊: Statistics Journal
   年份: 2023
   DOI: 10.1000/example2
```

## ⚡ 性能特点

### 搜索速度
- **本地文献查找**: 通常 < 1 秒
- **全球文献搜索**: 2-5 秒（取决于网络）
- **AI分析**: 5-15 秒（取决于问题复杂度）

### 搜索覆盖
- **本地文献**: 基于 Zotero 数据库（当前986篇）
- **Europe PMC**: 超过3900万篇文献
- **arXiv**: 超过200万篇预印本

## 🔧 故障排除

### 常见问题及解决方案

#### 1. 未找到本地文献
**问题**: 找不到匹配的文献
**解决**:
- 检查文献名称拼写
- 尝试使用更通用的关键词
- 确认文献已添加到 Zotero 数据库

#### 2. Article MCP 服务失败
**问题**: 无法搜索全球文献
**解决**:
- 确保网络连接正常
- 检查 `uvx` 命令是否可用
- 可能需要手动安装：`pip install article-mcp`

#### 3. AI分析超时
**问题**: AI分析过程超时
**解决**:
- 检查网络连接
- 简化问题描述
- 确认 AI API Key 配置正确

#### 4. 命令执行失败
**问题**: 命令无法执行
**解决**:
- 确保已编译最新版本：`cd server && make build`
- 检查配置文件：`server/.env`
- 查看错误日志获取详细信息

## 📝 高级用法

### 批量分析
```bash
# 可以编写脚本批量分析多篇文献
for topic in "machine learning" "deep learning" "neural networks"; do
    ./server/bin/zoteroflow2 related "$topic" "总结主要研究方向和趋势"
    echo "========================"
done
```

### 结果保存
```bash
# 将分析结果保存到文件
./server/bin/zoteroflow2 related "机器学习" "研究趋势分析" > analysis_result.txt
```

### 与其他命令结合
```bash
# 先列出所有文献，再对特定文献进行相关分析
./server/bin/zoteroflow2 list | grep "machine learning"
./server/bin/zoteroflow2 related "机器学习" "这个领域有哪些重要突破？"
```

## 🎯 最佳实践

### 1. 提问技巧
- **具体化问题**: 避免过于宽泛的问题
- **上下文明确**: 提供足够的背景信息
- **目标导向**: 明确分析的目的

### 2. 搜索优化
- **关键词选择**: 使用文献的核心概念
- **DOI 优先**: 如果有 DOI，优先使用
- **迭代搜索**: 根据结果调整搜索策略

### 3. 结果利用
- **记录重要发现**: 保存有价值的分析结果
- **追踪引用链**: 根据相关文献进一步研究
- **跨领域对比**: 比较不同领域的研究方法

## 🔮 未来功能

### 计划中的改进
- [ ] 支持更多学术数据库（Web of Science, Scopus）
- [ ] 改进 AI 分析的上下文理解
- [ ] 支持批量文献对比分析
- [ ] 添加引用网络分析功能
- [ ] 支持研究趋势预测

### 用户反馈
如有功能建议或问题报告，请：
1. 检查本文档的故障排除部分
2. 查看项目的 GitHub Issues
3. 提交新的 Issue 或 Pull Request

---

**最后更新**: 2025-10-04
**版本**: v1.0
**状态**: ✅ 生产就绪