# ZoteroFlow2 安装指南

## 概述

本文档提供了 ZoteroFlow2 的完整安装指南，包括系统要求、依赖安装、配置设置和部署验证等步骤。

## 系统要求

### 最低系统要求

- **操作系统**: Linux (Ubuntu 20.04+), macOS 10.15+, Windows 10+
- **Go 版本**: 1.21 或更高版本
- **内存**: 最少 2GB RAM，推荐 4GB+
- **存储空间**: 最少 1GB 可用空间
- **网络**: 稳定的互联网连接

### 推荐系统配置

- **CPU**: 4核心或更多
- **内存**: 8GB RAM 或更多
- **存储**: SSD，至少 10GB 可用空间
- **网络**: 宽带连接，支持 HTTPS

## 依赖安装

### 1. Go 语言环境

#### Linux (Ubuntu/Debian)

```bash
# 更新包管理器
sudo apt update

# 安装 Go
sudo apt install golang-go

# 或者从官方源安装最新版本
wget https://go.dev/dl/go1.21.0.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.21.0.linux-amd64.tar.gz

# 设置环境变量
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
echo 'export GOPATH=$HOME/go' >> ~/.bashrc
source ~/.bashrc

# 验证安装
go version
```

#### macOS

```bash
# 使用 Homebrew 安装
brew install go

# 或者从官方源安装
wget https://go.dev/dl/go1.21.0.darwin-amd64.pkg
open go1.21.0.darwin-amd64.pkg

# 验证安装
go version
```

#### Windows

```powershell
# 使用 Chocolatey 安装
choco install golang

# 或者从官方下载安装包
# 访问 https://go.dev/dl/ 下载 Windows 安装包

# 验证安装
go version
```

### 2. Python 和 uv (用于 Article MCP)

#### Linux (Ubuntu/Debian)

```bash
# 安装 Python 3.8+
sudo apt install python3 python3-pip python3-venv

# 安装 uv
curl -LsSf https://astral.sh/uv/install.sh | sh

# 或者使用 pip 安装
pip3 install uv

# 验证安装
python3 --version
uv --version
```

#### macOS

```bash
# 使用 Homebrew 安装 Python
brew install python@3.8

# 安装 uv
brew install uv

# 验证安装
python3 --version
uv --version
```

#### Windows

```powershell
# 使用 Chocolatey 安装 Python
choco install python

# 安装 uv
pip install uv

# 验证安装
python --version
uv --version
```

### 3. SQLite 数据库

#### Linux (Ubuntu/Debian)

```bash
# 安装 SQLite 开发包
sudo apt install sqlite3 libsqlite3-dev

# 验证安装
sqlite3 --version
```

#### macOS

```bash
# 使用 Homebrew 安装
brew install sqlite

# 验证安装
sqlite3 --version
```

#### Windows

```powershell
# 使用 Chocolatey 安装
choco install sqlite

# 验证安装
sqlite3 --version
```

## 项目安装

### 1. 获取源代码

```bash
# 克隆仓库
git clone https://github.com/your-org/zoteroflow2.git
cd zoteroflow2

# 或者下载发布版本
wget https://github.com/your-org/zoteroflow2/releases/latest/download/zoteroflow2.tar.gz
tar -xzf zoteroflow2.tar.gz
cd zoteroflow2
```

### 2. 构建项目

```bash
# 进入服务器目录
cd server/

# 下载依赖
make deps

# 构建项目
make build

# 验证构建
ls -la bin/
```

### 3. 运行基础测试

```bash
# 运行基础集成测试
./bin/zoteroflow2

# 检查帮助信息
./bin/zoteroflow2 help

# 测试基本功能
./bin/zoteroflow2 list
```

## 配置设置

### 1. 环境变量配置

#### 创建 .env 文件

```bash
# 在项目根目录创建 .env 文件
cat > .env << EOF
# Zotero 配置
ZOTERO_DB_PATH=~/Zotero/zotero.sqlite
ZOTERO_DATA_DIR=~/Zotero/storage

# MinerU 配置
MINERU_API_URL=https://mineru.net/api/v4
MINERU_TOKEN=your_mineru_token_here

# AI 配置
AI_API_KEY=your_ai_api_key_here
AI_BASE_URL=https://open.bigmodel.cn/api/coding/paas/v4
AI_MODEL=glm-4.6
AI_TIMEOUT=20

# 缓存配置
CACHE_DIR=~/.zoteroflow/cache
RESULTS_DIR=data/results
RECORDS_DIR=data/records

# 性能配置
MINERU_TIMEOUT=60
MAX_CONCURRENT_PARSING=3
EOF
```

#### 设置环境变量

```bash
# 加载环境变量
source .env

# 或者手动设置
export ZOTERO_DB_PATH=~/Zotero/zotero.sqlite
export ZOTERO_DATA_DIR=~/Zotero/storage
export MINERU_TOKEN=your_token
export AI_API_KEY=your_ai_key
```

### 2. Zotero 数据库配置

#### 定位 Zotero 数据库

```bash
# 常见 Zotero 数据库位置
# Linux: ~/Zotero/zotero.sqlite
# macOS: ~/Library/Application Support/Zotero/Zotero/Zotero.sqlite
# Windows: %APPDATA%\Zotero\Zotero\Zotero.sqlite

# 查找数据库文件
find ~ -name "zotero.sqlite" 2>/dev/null
```

#### 验证数据库访问

```bash
# 检查数据库文件
ls -la ~/Zotero/zotero.sqlite

# 测试数据库连接
sqlite3 ~/Zotero/zotero.sqlite "SELECT COUNT(*) FROM items;"
```

### 3. MinerU API 配置

#### 获取 MinerU Token

1. 访问 [MinerU 官网](https://mineru.net/)
2. 注册账号并登录
3. 在控制台中获取 API Token
4. 将 Token 配置到环境变量中

#### 验证 MinerU 连接

```bash
# 测试 API 连接
curl -H "Authorization: Bearer your_token" \
     https://mineru.net/api/v4/file-urls/batch
```

### 4. AI 服务配置

#### 获取智谱 AI API Key

1. 访问 [智谱 AI 开放平台](https://open.bigmodel.cn/)
2. 注册账号并创建 API Key
3. 将 API Key 配置到环境变量中

#### 验证 AI 连接

```bash
# 测试 AI API 连接
curl -H "Authorization: Bearer your_api_key" \
     -H "Content-Type: application/json" \
     -d '{"model":"glm-4.6","messages":[{"role":"user","content":"test"}]}' \
     https://open.bigmodel.cn/api/coding/paas/v4/chat/completions
```

## 部署选项

### 1. 开发环境部署

#### 本地开发

```bash
# 克隆代码
git clone https://github.com/your-org/zoteroflow2.git
cd zoteroflow2/server

# 配置环境
cp .env.example .env
# 编辑 .env 文件，填入必要的配置

# 构建和运行
make build
make dev

# 或者直接运行
make run
```

#### Docker 开发环境

```bash
# 创建 Dockerfile
cat > Dockerfile << EOF
FROM golang:1.21-alpine AS builder

WORKDIR /app
COPY . .
RUN go mod download
RUN go build -o bin/zoteroflow2 .

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/bin/zoteroflow2 .
COPY --from=builder /app/.env .env

CMD ["./zoteroflow2"]
EOF

# 构建镜像
docker build -t zoteroflow2:dev .

# 运行容器
docker run --rm -v ~/.zoteroflow:/root/.zoteroflow zoteroflow2:dev
```

### 2. 生产环境部署

#### 系统服务部署 (Linux)

```bash
# 创建系统用户
sudo useradd -r -s /bin/false zoteroflow
sudo mkdir -p /opt/zoteroflow2
sudo chown zoteroflow:zoteroflow /opt/zoteroflow2

# 复制文件
sudo cp -r . /opt/zoteroflow2/
sudo chown -R zoteroflow:zoteroflow /opt/zoteroflow2/

# 创建 systemd 服务
sudo tee /etc/systemd/system/zoteroflow2.service > /dev/null << EOF
[Unit]
Description=ZoteroFlow2 Literature Analysis Service
After=network.target

[Service]
Type=simple
User=zoteroflow
WorkingDirectory=/opt/zoteroflow2/server
ExecStart=/opt/zoteroflow2/server/bin/zoteroflow2
Restart=always
RestartSec=10

[Install]
WantedBy=multi-user.target
EOF

# 启用并启动服务
sudo systemctl daemon-reload
sudo systemctl enable zoteroflow2
sudo systemctl start zoteroflow2

# 检查状态
sudo systemctl status zoteroflow2
sudo journalctl -u zoteroflow2 -f
```

#### Docker 生产部署

```bash
# 创建生产 Dockerfile
cat > Dockerfile.prod << EOF
FROM golang:1.21-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=1 go build -ldflags="-w -s" -o bin/zoteroflow2 .

FROM alpine:latest
RUN apk --no-cache add ca-certificates tzdata
RUN addgroup -g 1000 zoteroflow && \
    adduser -D -s /bin/sh -u 1000 zoteroflow

WORKDIR /app
COPY --from=builder /app/bin/zoteroflow2 .
COPY --from=builder /app/.env.production .env

USER zoteroflow
EXPOSE 8080
CMD ["./zoteroflow2"]
EOF

# 创建 docker-compose.yml
cat > docker-compose.yml << EOF
version: '3.8'

services:
  zoteroflow2:
    build:
      context: .
      dockerfile: Dockerfile.prod
    ports:
      - "8080:8080"
    volumes:
      - ./data:/app/data
      - /home/user/.zoteroflow:/app/.zoteroflow
    environment:
      - ZOTERO_DB_PATH=/app/.zoteroflow/zotero.sqlite
      - ZOTERO_DATA_DIR=/app/.zoteroflow/storage
    restart: unless-stopped
    healthcheck:
      test: ["CMD", "./bin/zoteroflow2", "help"]
      interval: 30s
      timeout: 10s
      retries: 3
EOF

# 构建和运行
docker-compose up -d --build

# 查看日志
docker-compose logs -f zoteroflow2
```

#### Kubernetes 部署

```yaml
# k8s-deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: zoteroflow2
  labels:
    app: zoteroflow2
spec:
  replicas: 3
  selector:
    matchLabels:
      app: zoteroflow2
  template:
    metadata:
      labels:
        app: zoteroflow2
    spec:
      containers:
      - name: zoteroflow2
        image: zoteroflow2:latest
        ports:
        - containerPort: 8080
        env:
        - name: ZOTERO_DB_PATH
          value: "/app/data/zotero.sqlite"
        - name: MINERU_TOKEN
          valueFrom:
            secretKeyRef:
              name: zoteroflow2-secrets
              key: mineru-token
        - name: AI_API_KEY
          valueFrom:
            secretKeyRef:
              name: zoteroflow2-secrets
              key: ai-api-key
        volumeMounts:
        - name: data-volume
          mountPath: /app/data
        - name: config-volume
          mountPath: /app/.env
          subPath: .env
        livenessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 30
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /ready
            port: 8080
          initialDelaySeconds: 5
          periodSeconds: 5
      volumes:
      - name: data-volume
        persistentVolumeClaim:
          claimName: zoteroflow2-data
      - name: config-volume
        configMap:
          name: zoteroflow2-config
---
apiVersion: v1
kind: Service
metadata:
  name: zoteroflow2-service
spec:
  selector:
    app: zoteroflow2
  ports:
  - protocol: TCP
    port: 80
    targetPort: 8080
  type: LoadBalancer
```

## 验证安装

### 1. 基础功能验证

```bash
# 检查版本信息
./bin/zoteroflow2 help

# 测试数据库连接
./bin/zoteroflow2 list

# 测试 AI 功能
./bin/zoteroflow2 chat "测试连接"
```

### 2. 集成测试

```bash
# 运行完整测试套件
make test

# 运行覆盖率测试
make test-coverage

# 检查覆盖率报告
open coverage.html
```

### 3. 性能测试

```bash
# 测试数据库查询性能
time ./bin/zoteroflow2 list

# 测试 PDF 解析性能
time ./bin/zoteroflow2 search "test"  # 如果有测试PDF

# 测试 AI 响应性能
time ./bin/zoteroflow2 chat "简单测试"
```

## 故障排除

### 常见问题及解决方案

#### 1. 数据库连接失败

**问题**: `连接数据库失败: database is locked`

**解决方案**:
```bash
# 检查 Zotero 是否正在运行
ps aux | grep zotero

# 关闭 Zotero 后重试
# 或者使用只读模式
export ZOTERO_DB_PATH="file:$HOME/Zotero/zotero.sqlite?mode=ro"
```

#### 2. MinerU API 调用失败

**问题**: `MinerU解析失败: authentication failed`

**解决方案**:
```bash
# 验证 Token
curl -H "Authorization: Bearer your_token" \
     https://mineru.net/api/v4/file-urls/batch

# 检查 Token 是否过期
# 重新获取 Token 并更新配置
```

#### 3. AI 服务连接失败

**问题**: `AI调用失败: authentication failed`

**解决方案**:
```bash
# 验证 API Key
curl -H "Authorization: Bearer your_api_key" \
     -H "Content-Type: application/json" \
     -d '{"model":"glm-4.6","messages":[{"role":"user","content":"test"}]}' \
     https://open.bigmodel.cn/api/coding/paas/v4/chat/completions

# 检查 API Key 是否有效
# 重新获取 API Key 并更新配置
```

#### 4. 权限问题

**问题**: `权限不足: permission denied`

**解决方案**:
```bash
# 检查文件权限
ls -la ~/Zotero/zotero.sqlite
ls -la ~/.zoteroflow/

# 修复权限
chmod 644 ~/Zotero/zotero.sqlite
chmod 755 ~/.zoteroflow
```

#### 5. 端口占用

**问题**: `端口被占用: address already in use`

**解决方案**:
```bash
# 查看端口占用
netstat -tlnp | grep :8080

# 杀死占用进程
sudo kill -9 <PID>

# 或者更改端口
export ZOTEROFLOW_PORT=8081
```

### 日志分析

#### 查看应用日志

```bash
# 系统服务日志
sudo journalctl -u zoteroflow2 -f

# Docker 日志
docker logs zoteroflow2

# 应用日志
tail -f ~/.zoteroflow/logs/app.log
```

#### 调试模式

```bash
# 启用调试模式
export LOG_LEVEL=debug
./bin/zoteroflow2

# 查看详细错误信息
RUST_BACKTRACE=1 ./bin/zoteroflow2
```

## 维护和更新

### 1. 定期维护

```bash
# 清理缓存
./bin/zoteroflow2 clean

# 更新依赖
make mod-upgrade

# 重新构建
make clean && make build
```

### 2. 备份和恢复

```bash
# 备份配置和数据
tar -czf zoteroflow2-backup-$(date +%Y%m%d).tar.gz \
    ~/.zoteroflow \
    ~/Zotero/zotero.sqlite \
    .env

# 恢复配置
tar -xzf zoteroflow2-backup-20241201.tar.gz
```

### 3. 监控和告警

```bash
# 创建监控脚本
cat > monitor.sh << 'EOF'
#!/bin/bash

# 检查服务状态
if ! systemctl is-active --quiet zoteroflow2; then
    echo "ZoteroFlow2 服务未运行" | mail -s "服务告警" admin@example.com
fi

# 检查磁盘空间
df -h | grep -E "9[0-9]%" | mail -s "磁盘空间告警" admin@example.com

# 检查内存使用
free -m | grep "Mem:" | awk '{if($3/$2*100 > 90) print "内存使用告警"}' | \
    mail -s "内存告警" admin@example.com
EOF

chmod +x monitor.sh

# 添加到 crontab
echo "*/5 * * * * /path/to/monitor.sh" | crontab -
```

## 安全配置

### 1. 文件权限

```bash
# 设置适当的文件权限
chmod 600 .env
chmod 700 ~/.zoteroflow
chmod 644 ~/Zotero/zotero.sqlite
```

### 2. 网络安全

```bash
# 配置防火墙
sudo ufw allow 8080
sudo ufw enable

# 使用 HTTPS
# 配置反向代理 (Nginx/Apache)
```

### 3. API 密钥管理

```bash
# 使用密钥管理工具
# 例如 HashiCorp Vault, AWS Secrets Manager

# 定期轮换 API 密钥
# 设置密钥过期策略
```

## 性能优化

### 1. 系统优化

```bash
# 调整系统参数
echo 'vm.swappiness=10' | sudo tee -a /etc/sysctl.conf
echo 'fs.file-max=65536' | sudo tee -a /etc/sysctl.conf
sudo sysctl -p

# 优化数据库
sqlite3 ~/Zotero/zotero.sqlite "PRAGMA journal_mode=WAL;"
sqlite3 ~/Zotero/zotero.sqlite "PRAGMA synchronous=NORMAL;"
```

### 2. 应用优化

```bash
# 调整并发参数
export MAX_CONCURRENT_PARSING=5
export CACHE_SIZE=1000

# 优化缓存策略
export CACHE_TTL=3600
export CLEANUP_INTERVAL=86400
```

这个安装指南涵盖了从基础环境准备到生产部署的完整流程，包括故障排除、维护更新和安全配置等重要方面。按照这个指南，您可以成功部署和运行 ZoteroFlow2 系统。