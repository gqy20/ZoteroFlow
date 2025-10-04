# CLI 命令行接口文档

## 概述

ZoteroFlow2 提供了丰富的命令行接口，支持文献管理、PDF 解析、AI 对话等功能。

## 基本用法

```bash
./zoteroflow2 [command] [arguments...]
```

## 命令列表

### 📚 文献管理命令

#### `list` - 列出解析结果

列出所有已解析的文献及其基本信息。

```bash
./zoteroflow2 list
```

**输出示例**:
```
找到 3 个解析结果:

[1] 机器学习基础_20241201
     标题: 机器学习基础教程
     作者: 张三; 李四
     大小: 2.3 MB
     日期: 2024-12-01

[2] 深度学习研究_20241130
     标题: 深度学习在医疗诊断中的应用
     作者: 王五; 赵六
     大小: 1.8 MB
     日期: 2024-11-30
```

#### `open <名称>` - 打开文献文件夹

打开指定文献的文件夹，显示文件结构。

```bash
./zoteroflow2 open "机器学习"
./zoteroflow2 open "机器学习基础_20241201"
```

**参数**:
- `名称`: 文献名称或文件夹名（支持模糊匹配）

**输出示例**:
```
打开文献文件夹: data/results/机器学习基础_20241201
文件内容:
  📁 images/ (12 个文件)
  📄 full.md (45.2 KB)
  📄 meta.json (1.2 KB)
  📄 raw.zip (2.3 MB)
  📄 source.pdf (2.1 MB)
```

#### `search <关键词>` - 按标题搜索文献

在 Zotero 数据库中搜索匹配标题关键词的文献。

```bash
./zoteroflow2 search "机器学习"
./zoteroflow2 search "深度学习"
```

**参数**:
- `关键词`: 搜索关键词，支持中文和英文

**输出示例**:
```
📄 找到 2 篇文献:
  1. 标题: 机器学习基础教程 (评分: 95.0)
     作者: 张三; 李四
     期刊: 计算机科学
     年份: 2024
     DOI: 10.1234/ml.2024.001
     PDF路径: /home/user/Zotero/storage/ABC123/机器学习基础.pdf

  2. 标题: 深度学习入门指南 (评分: 85.0)
     作者: 王五
     期刊: 人工智能
     年份: 2023
     DOI: 10.5678/dl.2023.002
     PDF路径: /home/user/Zotero/storage/DEF456/深度学习入门.pdf
```

#### `doi <DOI号>` - 按 DOI 搜索文献

根据 DOI 号码搜索并解析特定文献。

```bash
./zoteroflow2 doi "10.1234/ml.2024.001"
./zoteroflow2 doi "10.5678/dl.2023.002"
```

**参数**:
- `DOI号`: 文献的 DOI 标识符

### 🤖 AI 对话命令

#### `chat` - 交互式 AI 对话

进入交互式 AI 对话模式，支持多轮对话和上下文记忆。

```bash
./zoteroflow2 chat
```

**交互式命令**:
- `help` / `帮助` - 显示帮助信息
- `new` / `新建` - 开始新对话
- `clear` / `清屏` - 清空屏幕
- `quit` / `exit` / `退出` - 退出对话

**对话示例**:
```
🤖 ZoteroFlow2 AI学术助手
输入 'help' 查看帮助，输入 'quit' 或 'exit' 退出
--------------------------------------------------
📚 您: 什么是机器学习？
🤖 助手: 机器学习是人工智能的一个分支，它使计算机能够从数据中学习...

📚 您: 能详细解释一下监督学习吗？
🤖 助手: 监督学习是机器学习的一种主要方法，它需要标记的训练数据...

📚 您: new
🆕 开始新对话

📚 您: 深度学习和机器学习有什么区别？
🤖 助手: 深度学习是机器学习的一个子集，主要使用神经网络...
```

#### `chat <问题>` - 单次 AI 问答

直接向 AI 提出问题，获得单次回答。

```bash
./zoteroflow2 chat "什么是深度学习？"
./zoteroflow2 chat "请解释一下机器学习的基本概念"
```

**参数**:
- `问题`: 要询问 AI 的问题

**输出示例**:
```
🤖 正在向AI发送问题: 什么是深度学习？
🤖 助手: 深度学习是机器学习的一个分支，它使用多层神经网络来学习数据的复杂模式...

📊 Token使用: 45 (输入) + 128 (输出) = 173 (总计)
```

#### `chat --doc=文献名 <问题>` - 基于文献的 AI 对话

基于指定文献内容进行 AI 对话，提供更精准的回答。

```bash
./zoteroflow2 chat --doc=机器学习 "介绍一下主要内容"
./zoteroflow2 chat -d=深度学习 "解释一下第三章的核心观点"
```

**参数**:
- `--doc=文献名` 或 `-d=文献名`: 指定文献名称
- `问题`: 要询问的问题

**输出示例**:
```
📚 基于文献 '机器学习基础教程' 进行对话
📝 作者: 张三; 李四
📄 摘要: 本文介绍了机器学习的基本概念、主要算法和应用场景...
--------------------------------------------------
🤖 助手: 基于您提供的《机器学习基础教程》，这篇文献主要内容包括...

📊 Token使用: 156 (输入) + 234 (输出) = 390 (总计)
```

### 🔧 维护命令

#### `clean` - 清理重复/损坏文件

清理系统中的重复文件、损坏文件和无用缓存。

```bash
./zoteroflow2 clean
```

**功能**:
- 删除重复的解析结果
- 清理损坏的 ZIP 文件
- 移除无效的缓存文件
- 压缩过大的日志文件

#### `help` - 显示帮助信息

显示所有可用命令的详细说明。

```bash
./zoteroflow2 help
```

**输出示例**:
```
ZoteroFlow2 - PDF文献管理工具

📚 文献管理:
  list                    - 列出所有解析结果
  open <名称>             - 打开指定文献文件夹
  search <关键词>         - 按标题搜索并解析文献
  doi <DOI号>             - 按DOI搜索并解析文献

🤖 AI助手对话:
  chat                    - 进入交互式AI对话模式
  chat <问题>             - 单次AI问答
  chat --doc=文献名 <问题> - 基于指定文献的AI对话

🔧 维护命令:
  clean                   - 清理重复/损坏文件
  help                    - 显示此帮助信息

💡 使用示例:
  ./zoteroflow2 list                                    # 列出文献
  ./zoteroflow2 search "机器学习"                      # 搜索文献
  ./zoteroflow2 chat "什么是深度学习？"                # AI问答
  ./zoteroflow2 chat --doc=基因组 "介绍一下CRISPR"        # 基于文献的AI对话

🎯 AI功能特性:
  • 支持学术文献分析和解释
  • 可基于特定文献内容进行对话
  • 交互式对话模式支持上下文记忆
  • 单次问答模式，适合快速查询
```

## 默认行为

当不提供任何命令时，系统会运行基础集成测试：

```bash
./zoteroflow2
```

**测试内容**:
- 配置加载验证
- Zotero 数据库连接测试
- MinerU 客户端创建测试
- 基础功能验证

## 退出码

- `0`: 成功执行
- `1`: 一般错误
- `2`: 配置错误
- `3`: 数据库连接失败
- `4`: API 调用失败

## 环境变量

可以通过环境变量自定义行为：

```bash
export ZOTERO_DB_PATH="/path/to/zotero.sqlite"
export ZOTERO_DATA_DIR="/path/to/storage"
export MINERU_TOKEN="your_token_here"
export AI_API_KEY="your_ai_key"

./zoteroflow2 list
```

## 配置文件

支持 `.env` 文件配置：

```bash
# .env
ZOTERO_DB_PATH=~/Zotero/zotero.sqlite
ZOTERO_DATA_DIR=~/Zotero/storage
MINERU_API_URL=https://mineru.net/api/v4
MINERU_TOKEN=your_token
AI_API_KEY=your_ai_key
AI_BASE_URL=https://open.bigmodel.cn/api/coding/paas/v4
AI_MODEL=glm-4.6