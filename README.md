# ZoteroFlow - 智能文献分析助手

> **核心理念**: AI无处不在，无需寻找；尊重知识产权，不优化PDF下载

## 📋 项目概述

ZoteroFlow是基于2025年最新技术的智能文献分析平台，为学术研究者提供AI驱动的文献理解和分析服务。

### 🎯 核心特性
- **智能文献列表**: AI驱动的文献推荐和管理
- **深度AI分析**: 流式AI分析，实时获得论文洞察
- **智能搜索**: 支持语义搜索和智能纠错
- **无缝PDF阅读**: AI标注和智能摘要

## 🚀 技术栈

- **后端**: Go + Gin + SQLite
- **前端**: Next.js 15 + TypeScript + Tailwind CSS
- **AI**: Vercel AI SDK + OpenAI GPT-4o
- **UI**: Park UI 2025
- **搜索**: Typesense
- **PDF**: Nutrient Web SDK 2025

## 📊 项目指标

- **代码量**: 800行
- **开发时间**: 2周
- **首屏加载**: < 1秒
- **AI响应**: < 3秒

## 🏗️ 快速开始

### 环境要求
- Node.js 18+
- Go 1.21+
- OpenAI API密钥

### 配置环境变量
```env
# .env.local
OPENAI_API_KEY=your_openai_key
NEXT_PUBLIC_API_URL=http://localhost:8080
```

### 启动服务
```bash
# 后端
go run main.go

# 前端
npm run dev
```

访问 http://localhost:3000 开始使用。

## 📚 文档

- [📋 实现方案](docs/IMPLEMENTATION.md) - 详细的技术实现计划
- [💭 设计理念](docs/think.md) - 核心设计理念

## 🤝 贡献

欢迎提交Issue和Pull Request来改进项目。

## 📄 许可证

MIT License

---

**ZoteroFlow** - 让AI成为你的学术研究伙伴 🚀