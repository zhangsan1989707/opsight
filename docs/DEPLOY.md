# Opsight 政企私有化部署指南

> 版本：v0.2.0 | 最后更新：2026-05-29  
> 适用场景：政企内网环境、信创服务器、私有云交付

---

## 1. 系统要求

### 1.1 最低配置

| 项目 | 最低要求 | 推荐配置 |
|------|---------|---------|
| CPU | 4 核 | 8 核及以上 |
| 内存 | 8 GB | 16 GB 及以上 |
| 磁盘 | 50 GB（不含日志存储） | 100 GB SSD |
| 操作系统 | Linux (x86_64) | Ubuntu 22.04+ / CentOS 8+ / 麒麟 V10 / openEuler 22.03 |
| Docker | 20.10+ | 24.0+ |
| Docker Compose | v2.0+ | v2.20+ |

### 1.2 网络要求

| 端口 | 协议 | 用途 | 方向 |
|------|------|------|------|
| 80 | TCP | HTTP 入口（Nginx） | 入站 |
| 443 | TCP | HTTPS 入口（Nginx） | 入站 |
| 5432 | TCP | PostgreSQL 数据库 | 容器内部 |
| 3800 | TCP | Next.js 前端 | 容器内部 |
| 8800 | TCP | Go API 服务 | 容器内部 |

> 容器内部端口无需对外开放，仅 80/443 需要从客户端可达。

---

## 2. 前置条件

### 2.1 安装 Docker

```bash
# Ubuntu / Debian
curl -fsSL https://get.docker.com | sudo bash
sudo usermod -aG docker $USER

# CentOS / RHEL
sudo yum install -y yum-utils
sudo yum-config-manager --add-repo https://download.docker.com/linux/centos/docker-ce.repo
sudo yum install -y docker-ce docker-ce-cli containerd.io
sudo systemctl enable --now docker
```

**离线环境**：如目标服务器无外网访问，请提前在有网环境下载 Docker 离线包和 Opsight 镜像，使用 `docker save/load` 导入。

### 2.2 安装 Docker Compose

```bash
# Docker Compose v2（推荐，已内置在 Docker CLI 插件中）
docker compose version

# 如未安装，可通过包管理器安装
sudo apt install docker-compose-plugin    # Ubuntu
sudo yum install docker-compose-plugin    # CentOS/RHEL
```

### 2.3 验证安装

```bash
docker --version          # >= 20.10
docker compose version    # >= v2.0
```

---

## 3. 架构概览

```
                        ┌──────────────────────────────────────┐
                        │            Client Browser            │
                        │         (Chrome / Firefox / Edge)    │
                        └─────────────────┬────────────────────┘
                                          │ HTTP/HTTPS
                                          ▼
                        ┌──────────────────────────────────────┐
                        │          Nginx Reverse Proxy          │
                        │         Port 80 / 443                 │
                        │         SSL 终止 / Gzip / 限流        │
                        └──────┬───────────────────┬───────────┘
                               │                   │
                 ┌─────────────▼──────┐  ┌─────────▼──────────┐
                 │  Frontend :3800    │  │   Backend :8800     │
                 │  Next.js 14        │  │   Go 1.22+ / Gin    │
                 │  React + Tailwind   │  │   REST API + WS     │
                 └────────────────────┘  └─────────┬───────────┘
                                                   │
                                        ┌──────────▼───────────┐
                                        │  PostgreSQL 16        │
                                        │  :5432 (容器内部)     │
                                        │  Data: named volume   │
                                        └──────────────────────┘
```

**组件说明**：
- **Nginx**：反向代理，统一入口，WebSocket 代理，SSL 终止，安全头注入
- **Frontend**：Next.js 14 生产模式运行，SSR + 静态资源
- **Backend**：Go/Gin REST API，WebSocket 实时推送，指标数据生成
- **PostgreSQL**：持久化数据库，存储用户、配置、审计日志等

---

## 4. 部署步骤

### 步骤 1：传输文件到服务器

```bash
# 方式一：Git 克隆（有外网环境）
git clone <your-repo-url> /opt/opsight
cd /opt/opsight

# 方式二：打包上传（无外网环境）
# 在本地执行：
tar -czf opsight.tar.gz opsight/
scp opsight.tar.gz root@<服务器IP>:/opt/
# 在服务器上执行：
cd /opt && tar -xzf opsight.tar.gz && cd opsight
```

### 步骤 2：配置环境变量

```bash
# 从模板生成配置文件
cp .env.example .env

# 编辑配置（必须修改数据库密码等敏感项）
vim .env
```

**必须修改的配置项**：

| 变量 | 说明 | 默认值 | 建议 |
|------|------|--------|------|
| `DB_PASSWORD` | 数据库密码 | `opsight_secret` | 修改为强密码 |
| `JWT_SECRET` | JWT 签名密钥 | 空 | 生成长随机字符串 |
| `GIN_MODE` | 运行模式 | `release` | 生产环境保持 `release` |

**生成随机密钥**：
```bash
# Linux
openssl rand -base64 32

# 或使用 /dev/urandom
cat /dev/urandom | tr -dc 'a-zA-Z0-9' | head -c 32
```

### 步骤 3：启动服务

```bash
# 构建镜像并启动所有服务
docker compose up -d --build

# 查看启动状态
docker compose ps

# 查看实时日志（确认无报错）
docker compose logs -f
```

**预期输出**（`docker compose ps`）：
```
NAME                   STATUS                    PORTS
opsight-postgres-1     Up (healthy)             5432/tcp
opsight-api-1          Up (healthy)             8800/tcp
opsight-frontend-1     Up (healthy)             3800/tcp
opsight-nginx-1        Up (healthy)             0.0.0.0:80->80/tcp, 0.0.0.0:443->443/tcp
```

### 步骤 4：验证部署

```bash
# 1. 健康检查
curl http://localhost/healthz
# 预期输出: {"code":0,"data":{"status":"ok","time":"..."},"message":"success"}

# 2. 前端页面可达
curl -I http://localhost
# 预期输出: HTTP/1.1 200 OK

# 3. API 接口可达
curl http://localhost/api/v1/services
# 预期输出: JSON 格式的服务列表

# 4. 数据库连接
docker compose exec postgres pg_isready -U opsight -d opsight
# 预期输出: /var/run/postgresql:5432 - accepting connections
```

### 步骤 5：访问平台

浏览器访问（根据实际部署 IP 调整）：

```
http://<服务器IP>
```

---

## 5. 端口列表

| 服务 | 端口 | 协议 | 对外开放 | 用途 |
|------|------|------|---------|------|
| Nginx HTTP | 80 | TCP | 是 | Web 入口 |
| Nginx HTTPS | 443 | TCP | 是 | 加密 Web 入口 |
| Frontend | 3800 | TCP | 否（仅容器内） | Next.js 服务 |
| Backend | 8800 | TCP | 否（仅容器内） | API 服务 |
| PostgreSQL | 5432 | TCP | 否（仅容器内） | 数据库 |

---

## 6. 环境变量参考

### 6.1 服务端口

| 变量 | 默认值 | 说明 |
|------|--------|------|
| `API_PORT` | `8800` | 后端 API 服务端口 |
| `FRONTEND_PORT` | `3800` | 前端服务端口 |
| `NGINX_HTTP_PORT` | `80` | Nginx HTTP 监听端口 |
| `NGINX_HTTPS_PORT` | `443` | Nginx HTTPS 监听端口 |

### 6.2 数据库

| 变量 | 默认值 | 说明 |
|------|--------|------|
| `DB_HOST` | `postgres` | 数据库主机（容器名） |
| `DB_PORT` | `5432` | 数据库端口 |
| `DB_USER` | `opsight` | 数据库用户名 |
| `DB_PASSWORD` | `opsight_secret` | 数据库密码 |
| `DB_NAME` | `opsight` | 数据库名称 |

### 6.3 运行模式

| 变量 | 默认值 | 说明 |
|------|--------|------|
| `GIN_MODE` | `release` | Gin 模式（`debug`/`release`） |

### 6.4 安全配置

| 变量 | 默认值 | 说明 |
|------|--------|------|
| `JWT_SECRET` | - | JWT 签名密钥（必填） |
| `JWT_EXPIRY_HOURS` | `24` | Token 过期时间（小时） |
| `RATE_LIMIT_ENABLED` | `true` | 启用 API 速率限制 |
| `RATE_LIMIT_RPS` | `100` | 每秒最大请求数 |
| `LOGIN_MAX_ATTEMPTS` | `5` | 登录失败最大次数 |
| `LOGIN_LOCK_MINUTES` | `15` | 登录锁定时间（分钟） |
| `AUDIT_ENABLED` | `true` | 启用审计日志 |
| `AUDIT_RETENTION_DAYS` | `90` | 审计日志保留天数 |
| `SESSION_TIMEOUT_MINUTES` | `60` | 会话超时时间（分钟） |
| `CORS_ALLOWED_ORIGINS` | - | 允许的跨域来源（逗号分隔） |

### 6.5 超时与并发

| 变量 | 默认值 | 说明 |
|------|--------|------|
| `API_PROXY_READ_TIMEOUT` | `60` | API 代理读超时（秒） |
| `CLIENT_MAX_BODY_SIZE` | `10m` | 客户端请求体最大大小 |

### 6.6 SSL/TLS

| 变量 | 默认值 | 说明 |
|------|--------|------|
| `SSL_CERT_PATH` | - | SSL 证书路径 |
| `SSL_KEY_PATH` | - | SSL 私钥路径 |
| `ENABLE_HTTPS_REDIRECT` | `false` | 启用 HTTP→HTTPS 重定向 |

### 6.7 日志

| 变量 | 默认值 | 说明 |
|------|--------|------|
| `LOG_LEVEL` | `info` | 日志级别（`debug`/`info`/`warn`/`error`） |
| `NGINX_LOG_RETENTION` | `30` | Nginx 日志保留天数 |

---

## 7. SSL/HTTPS 配置

### 7.1 准备证书

```bash
# 创建证书目录
mkdir -p ssl

# 将证书文件放入 ssl 目录
# - ssl/fullchain.pem  (证书链)
# - ssl/privkey.pem    (私钥)

# 确保证书权限安全
chmod 600 ssl/privkey.pem
chmod 644 ssl/fullchain.pem
```

### 7.2 配置 .env

```bash
# 编辑 .env，启用 HTTPS
ENABLE_HTTPS_REDIRECT=true
SSL_CERT_PATH=./ssl/fullchain.pem
SSL_KEY_PATH=./ssl/privkey.pem
```

### 7.3 修改 docker-compose.yml

取消 `nginx` 服务中 SSL 证书卷挂载的注释：

```yaml
services:
  nginx:
    # ...
    volumes:
      - nginx_logs:/var/log/nginx
      - ./ssl:/etc/nginx/ssl:ro    # ← 取消此行注释
```

### 7.4 启用 Nginx HTTPS 配置

编辑 `nginx/nginx.conf`，取消 HTTPS server 块和 HTTP 重定向块的注释。

### 7.5 重启服务

```bash
docker compose down
docker compose up -d
```

### 7.6 验证 HTTPS

```bash
curl -I https://<服务器IP>
# 预期: HTTP/2 200

# 测试 HSTS 头
curl -I https://<服务器IP> 2>/dev/null | grep Strict-Transport-Security
```

---

## 8. 数据备份与恢复

### 8.1 数据库备份

```bash
# 创建备份目录
mkdir -p /opt/backups/opsight

# 执行备份
docker compose exec -T postgres pg_dump \
  -U opsight \
  -d opsight \
  -F c \
  -f /tmp/opsight_$(date +%Y%m%d_%H%M%S).dump

# 从容器中复制备份文件
docker compose cp postgres:/tmp/opsight_*.dump /opt/backups/opsight/
```

### 8.2 自动化备份脚本

创建 `/opt/opsight/backup.sh`：

```bash
#!/usr/bin/env bash
set -euo pipefail

BACKUP_DIR="/opt/backups/opsight"
RETENTION_DAYS=30
TIMESTAMP=$(date +%Y%m%d_%H%M%S)
BACKUP_FILE="opsight_${TIMESTAMP}.dump"

mkdir -p "$BACKUP_DIR"

cd /opt/opsight
docker compose exec -T postgres pg_dump -U opsight -d opsight -F c -f "/tmp/$BACKUP_FILE"
docker compose cp "postgres:/tmp/$BACKUP_FILE" "$BACKUP_DIR/"

echo "Backup saved: $BACKUP_DIR/$BACKUP_FILE"

# 清理过期备份
find "$BACKUP_DIR" -name "*.dump" -mtime +"$RETENTION_DAYS" -delete
```

配置 crontab 每日凌晨备份：
```bash
crontab -e
# 添加以下行：
0 2 * * * /opt/opsight/backup.sh >> /var/log/opsight-backup.log 2>&1
```

### 8.3 数据库恢复

```bash
# 停止应用（保留数据库）
docker compose stop api frontend nginx

# 执行恢复
docker compose exec -T postgres pg_restore \
  -U opsight \
  -d opsight \
  --clean \
  --if-exists \
  < /opt/backups/opsight/opsight_20260529_020000.dump

# 重启服务
docker compose up -d
```

> 恢复前建议先对当前数据库做一次备份，以防恢复失败。

### 8.4 全量备份

```bash
# 备份项目目录（含 .env、证书、nginx 配置等）
tar -czf /opt/backups/opsight/opsight_config_$(date +%Y%m%d).tar.gz \
  /opt/opsight/.env \
  /opt/opsight/nginx/ \
  /opt/opsight/ssl/ \
  /opt/opsight/docker-compose.yml
```

---

## 9. 故障排查

### 9.1 服务无法启动

```bash
# 查看所有服务状态
docker compose ps

# 查看全部日志
docker compose logs

# 查看特定服务日志
docker compose logs api
docker compose logs frontend
docker compose logs postgres
docker compose logs nginx
```

**常见问题**：

| 现象 | 可能原因 | 解决方案 |
|------|---------|---------|
| `api` 容器反复重启 | 端口被占用 | `lsof -i :8800` 检查端口占用 |
| `postgres` 不健康 | 数据目录损坏 | 删除 volume 重建：`docker compose down -v && docker compose up -d` |
| `frontend` 无法连接 API | API 未就绪 | 等待 API 健康检查通过，自动重试 |
| 镜像构建失败 | Dockerfile 错误 | 检查 `opsight-backend/Dockerfile` 和 `opsight-frontend/Dockerfile` |

### 9.2 数据库连接失败

```bash
# 检查数据库容器运行状态
docker compose ps postgres

# 测试数据库连接
docker compose exec postgres pg_isready -U opsight -d opsight

# 检查 .env 中的数据库配置
grep DB_ .env

# 重置数据库（会丢失数据）
docker compose down -v
docker compose up -d
```

**排查清单**：
1. `.env` 中 `DB_PASSWORD` 是否正确
2. `DB_USER`、`DB_NAME` 是否正确
3. PostgreSQL 容器健康检查是否通过
4. 防火墙是否阻止了容器间通信

### 9.3 端口冲突

```bash
# 检查端口占用
sudo ss -tlnp | grep -E ':80 |:443 |:5432 '

# 修改 .env 中的端口映射
NGINX_HTTP_PORT=8080    # 改为非冲突端口
NGINX_HTTPS_PORT=8443   # 改为非冲突端口

# 重启服务
docker compose up -d
```

### 9.4 磁盘空间不足

```bash
# 检查磁盘使用
df -h

# 清理 Docker 资源
docker system prune -a -f   # 清理未使用的镜像、容器、网络
docker volume prune -f       # 清理未使用的卷

# 清理旧日志
sudo find /var/log -name "*.log" -mtime +30 -delete
```

### 9.5 HTTPS 证书问题

```bash
# 验证证书文件
openssl x509 -in ssl/fullchain.pem -text -noout | grep -E "Subject:|Not After"

# 测试 SSL 连接
openssl s_client -connect localhost:443 -servername <域名>

# 检查证书挂载
docker compose exec nginx ls -la /etc/nginx/ssl/
```

### 9.6 访问 502/504 错误

```bash
# 查看 Nginx 错误日志
docker compose exec nginx tail -100 /var/log/nginx/error.log

# 检查后端服务健康状态
curl http://localhost/healthz

# 检查 Nginx upstream 配置
docker compose exec nginx nginx -t
```

### 9.7 性能问题

```bash
# 查看容器资源使用
docker stats

# 查看慢查询（数据库）
docker compose exec postgres psql -U opsight -d opsight \
  -c "SELECT query, mean_exec_time FROM pg_stat_statements ORDER BY mean_exec_time DESC LIMIT 10;"

# 查看 Nginx 请求时间（慢请求）
docker compose exec nginx tail -100 /var/log/nginx/access.log | awk '$NF > 1 {print}'
```

---

## 10. 信创环境补充说明

### 10.1 ARM 架构（鲲鹏 / 飞腾）

如需在 ARM 架构服务器运行，Dockerfile 需使用多架构构建：

```bash
# 构建 ARM 镜像
docker buildx build --platform linux/arm64 -t opsight-api:arm64 ./opsight-backend/

# 或在 docker-compose.yml 中指定平台
services:
  api:
    platform: linux/arm64
```

### 10.2 离线部署包制作

在有网机器上：

```bash
# 拉取基础镜像
docker pull golang:1.22-alpine
docker pull node:20-alpine
docker pull nginx:1.25-alpine
docker pull postgres:16-alpine

# 构建 Opsight 镜像
docker compose build

# 导出镜像
docker save golang:1.22-alpine node:20-alpine nginx:1.25-alpine postgres:16-alpine \
  opsight-api opsight-frontend opsight-nginx -o opsight-images.tar

# 传输到离线服务器后导入
docker load -i opsight-images.tar
```
