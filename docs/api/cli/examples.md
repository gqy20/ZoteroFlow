# CLI 使用示例

## 基础使用流程

### 1. 环境准备

首先确保配置了必要的环境变量：

```bash
# 创建 .env 文件
cat > .env << EOF
ZOTERO_DB_PATH=~/Zotero/zotero.sqlite
ZOTERO_DATA_DIR=~/Zotero/storage
MINERU_API_URL=https://mineru.net/api/v4
MINERU_TOKEN=your_mineru_token_here
AI_API_KEY=your_ai_api_key_here
AI_BASE_URL=https://open.bigmodel.cn/api/coding/paas/v4
AI_MODEL=glm-4.6
EOF

# 构建项目
cd server/
make build
```

### 2. 基础文献管理

#### 列出所有文献

```bash
./bin/zoteroflow2 list
```

**预期输出**:
```
找到 5 个解析结果:

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

#### 搜索特定主题文献

```bash
# 搜索机器学习相关文献
./bin/zoteroflow2 search "机器学习"

# 搜索深度学习相关文献
./bin/zoteroflow2 search "深度学习"

# 搜索特定作者
./bin/zoteroflow2 search "张三"
```

#### 通过DOI查找文献

```bash
./bin/zoteroflow2 doi "10.1234/ml.2024.001"
```

**预期输出**:
```
📄 找到 1 篇文献:
  1. 标题: 机器学习基础教程 (评分: 95.0)
     作者: 张三; 李四
     期刊: 计算机科学
     年份: 2024
     DOI: 10.1234/ml.2024.001
     PDF路径: /home/user/Zotero/storage/ABC123/机器学习基础.pdf

🚀 开始解析PDF...
[2024-12-01 10:30:15] 步骤1: 提交批量任务
[2024-12-01 10:30:16] 任务ID: task_123456, 上传URL: https://mineru.net/upload/...
[2024-12-01 10:30:20] 步骤2: 上传PDF文件
[2024-12-01 10:30:25] 步骤3: 轮询处理状态
[2024-12-01 10:31:30] 步骤4: 下载解析结果
✅ PDF解析成功! 耗时: 1分15秒

📊 解析结果:
  任务ID: task_123456
  处理耗时: 75000 ms
  ZIP文件: data/results/机器学习基础_20241201/raw.zip

📁 文件已自动组织到: data/results/
使用 './bin/zoteroflow2 list' 查看所有结果
```

### 3. 查看解析结果

#### 打开文献文件夹

```bash
# 打开最新的解析结果
./bin/zoteroflow2 open "机器学习基础"

# 或者使用完整文件夹名
./bin/zoteroflow2 open "机器学习基础_20241201"
```

**预期输出**:
```
打开文献文件夹: data/results/机器学习基础_20241201
文件内容:
  📁 images/ (12 个文件)
  📄 full.md (45.2 KB)
  📄 meta.json (1.2 KB)
  📄 raw.zip (2.3 MB)
  📄 source.pdf (2.1 MB)
```

#### 查看解析的Markdown内容

```bash
# 查看解析后的markdown文件
cat data/results/机器学习基础_20241201/full.md | head -20
```

**预期输出**:
```
# 机器学习基础教程

## 摘要
本文介绍了机器学习的基本概念、主要算法和应用场景...

## 1. 引言
机器学习是人工智能的一个重要分支，它使计算机能够从数据中学习...

## 2. 基本概念

### 2.1 监督学习
监督学习是机器学习的一种主要方法...
```

## AI 对话示例

### 1. 基础AI问答

#### 单次问答

```bash
./bin/zoteroflow2 chat "什么是机器学习？"
```

**预期输出**:
```
🤖 正在向AI发送问题: 什么是机器学习？
🤖 助手: 机器学习是人工智能的一个分支，它使计算机能够从数据中学习模式和规律，而无需明确编程。主要包括监督学习、无监督学习和强化学习等方法。

📊 Token使用: 45 (输入) + 128 (输出) = 173 (总计)
```

#### 交互式对话

```bash
./bin/zoteroflow2 chat
```

**对话示例**:
```
🤖 ZoteroFlow2 AI学术助手
输入 'help' 查看帮助，输入 'quit' 或 'exit' 退出
--------------------------------------------------
📚 您: 什么是深度学习？
🤖 助手: 深度学习是机器学习的一个分支，它使用多层神经网络来学习数据的复杂模式。与传统机器学习方法相比，深度学习能够自动提取特征，在大数据环境下表现优异。

📚 您: 深度学习和传统机器学习有什么区别？
🤖 助手: 主要区别包括：
1. 特征工程：传统ML需要手动设计特征，DL自动学习特征
2. 数据需求：DL通常需要更多数据
3. 计算资源：DL需要更强的计算能力
4. 可解释性：传统ML通常更易解释

📚 您: 能推荐一些深度学习的入门资源吗？
🤖 助手: 当然可以！以下是一些优质的深度学习入门资源：
1. 《深度学习》- Ian Goodfellow 等
2. Coursera 上的 "Deep Learning Specialization"
3. fast.ai 的实用深度学习课程
4. TensorFlow 和 PyTorch 官方教程

📚 您: quit
👋 再见!
```

### 2. 基于文献的AI对话

#### 基于特定文献提问

```bash
# 基于机器学习基础教程提问
./bin/zoteroflow2 chat --doc="机器学习基础" "这篇文章的主要观点是什么？"

# 基于深度学习研究提问
./bin/zoteroflow2 chat --doc="深度学习研究" "这篇文章使用了什么方法？"
```

**预期输出**:
```
📚 基于文献 '机器学习基础_20241201' 进行对话
📝 作者: 张三; 李四
📄 摘要: 本文介绍了机器学习的基本概念、主要算法和应用场景...
--------------------------------------------------
🤖 助手: 基于《机器学习基础教程》这篇文章，主要观点包括：

1. **机器学习的定义和重要性**：文章强调了机器学习在现代人工智能中的核心地位，以及它如何改变数据处理和决策制定的方式。

2. **三大主要方法**：
   - 监督学习：使用标记数据进行预测
   - 无监督学习：发现数据中的隐藏模式
   - 强化学习：通过试错学习最优策略

3. **实际应用场景**：文章详细介绍了机器学习在医疗、金融、交通等领域的具体应用案例。

4. **未来发展趋势**：作者讨论了深度学习、联邦学习等前沿技术的发展方向。

📊 Token使用: 156 (输入) + 234 (输出) = 390 (总计)
```

#### 连续对话

```bash
./bin/zoteroflow2 chat --doc="机器学习基础"
```

**对话示例**:
```
📚 基于文献 '机器学习基础_20241201' 进行对话
📝 作者: 张三; 李四
📄 摘要: 本文介绍了机器学习的基本概念...
--------------------------------------------------
📚 您: 这篇文章提到了哪些算法？
🤖 助手: 根据文章内容，主要提到了以下算法：
1. 线性回归
2. 决策树
3. 支持向量机（SVM）
4. 神经网络
5. K-means聚类

📚 您: 能详细解释一下支持向量机吗？
🤖 助手: 基于文章第三章的内容，支持向量机（SVM）是一种监督学习算法，主要特点包括：
- 寻找最优分离超平面
- 使用核技巧处理非线性问题
- 在小样本情况下表现良好
- 具有较好的泛化能力

📚 您: 这些算法有什么优缺点？
🤖 助手: 根据文章第四章的比较分析：
- 线性回归：简单易懂，但只能处理线性关系
- 决策树：可解释性强，但容易过拟合
- SVM：在高维空间表现好，但计算复杂度高
- 神经网络：拟合能力强，但需要大量数据和计算资源
- K-means：简单快速，但需要预先指定聚类数量
```

## 高级使用场景

### 1. 批量处理文献

```bash
# 搜索并批量解析多篇文献
./bin/zoteroflow2 search "机器学习" | head -5

# 然后逐个解析
./bin/zoteroflow2 doi "10.1234/ml.2024.001"
./bin/zoteroflow2 doi "10.5678/dl.2023.002"
./bin/zoteroflow2 doi "10.9012/ai.2023.003"
```

### 2. 研究主题分析

```bash
# 搜索特定主题
./bin/zoteroflow2 search "深度学习"

# 基于搜索结果进行AI分析
./bin/zoteroflow2 chat "基于我库中的深度学习文献，总结一下这个领域的主要研究方向"
```

### 3. 文献综述辅助

```bash
# 搜索相关文献
./bin/zoteroflow2 search "机器学习 医疗"

# 基于文献进行综述分析
./bin/zoteroflow2 chat --doc="机器学习医疗" "这篇文章在医疗领域的应用有哪些创新点？"

# 综合分析
./bin/zoteroflow2 chat "请帮我分析机器学习在医疗诊断中的应用现状和发展趋势"
```

## 故障排除

### 常见问题及解决方案

#### 1. 配置问题

```bash
# 检查配置
./bin/zoteroflow2

# 如果出现配置错误，检查 .env 文件
cat .env
```

#### 2. 数据库连接问题

```bash
# 检查Zotero数据库路径
ls -la ~/Zotero/zotero.sqlite

# 检查存储目录
ls -la ~/Zotero/storage
```

#### 3. MinerU API问题

```bash
# 检查网络连接
curl -I https://mineru.net/api/v4

# 验证token
curl -H "Authorization: Bearer your_token" https://mineru.net/api/v4/file-urls/batch
```

#### 4. AI服务问题

```bash
# 检查AI API连接
curl -H "Authorization: Bearer your_ai_key" \
     -H "Content-Type: application/json" \
     -d '{"model":"glm-4.6","messages":[{"role":"user","content":"test"}]}' \
     https://open.bigmodel.cn/api/coding/paas/v4/chat/completions
```

## 性能优化建议

### 1. 缓存管理

```bash
# 清理缓存
./bin/zoteroflow2 clean

# 查看缓存大小
du -sh ~/.zoteroflow/cache/
```

### 2. 并发处理

```bash
# 使用并行处理提高效率
for doi in $(cat doi_list.txt); do
    ./bin/zoteroflow2 doi "$doi" &
done
wait
```

### 3. 监控和日志

```bash
# 启用详细日志
export LOG_LEVEL=debug
./bin/zoteroflow2 list

# 查看解析记录
ls -la data/records/
cat data/records/mineru_parse_records_$(date +%Y-%m-%d).csv