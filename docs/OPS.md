# Opsight 运维操作手册

> 版本：v0.2.0 | 最后更新：2026-05-29  
> 目标读者：运维工程师、系统管理员

---

## 目录

1. [启停操作](#1-启停操作)
2. [日志查看](#2-日志查看)
3. [数据库备份与恢复](#3-数据库备份与恢复)
4. [磁盘空间监控](#4-磁盘空间监控)
5. [性能调优](#5-性能调优)
6. [健康检查端点](#6-健康检查端点)
7. [升级流程](#7-升级流程)
8. [紧急回滚](#8-紧急回滚)
9. [日常巡检清单](#9-日常巡检清单)

---

## 1. 启停操作

### 1.1 启动所有服务

```bash
cd /opt/opsight

# 前台启动（调试用）
docker compose up

# 后台启动（生产用）
docker compose up -d

# 强制重新构建镜像
docker compose up -d --build
```

### 1.2 停止服务

```bash
# 停止所有服务（保留容器）
docker compose stop

# 停止所有服务（删除容器，保留数据卷）
docker compose down

# 停止所有服务（删除容器和数据卷，数据全部丢失！）
docker compose down -v
```

> `-v` 参数会删除数据库数据卷，仅在需要完全重置时使用。

### 1.3 重启服务

```bash
# 全部重启
docker compose restart

# 平滑重启（无中断）
docker compose up -d --force-recreate --no-deps nginx

# 仅重启单个服务
docker compose restart api
docker compose restart frontend
docker compose restart nginx
docker compose restart postgres
```

### 1.4 查看运行状态

```bash
# 简洁状态
docker compose ps

# 详细状态（含资源占用）
docker compose ps --format "table {{.Name}}\t{{.Status}}\t{{.Ports}}"

# 实时资源监控
docker stats
```

---

## 2. 日志查看

### 2.1 Docker 容器日志

```bash
# 查看全部容器日志（实时）
docker compose logs -f

# 查看特定服务日志
docker compose logs -f api          # 后端
docker compose logs -f frontend     # 前端
docker compose logs -f nginx        # 反向代理
docker compose logs -f postgres     # 数据库

# 查看最近 N 行
docker compose logs --tail=100 api

# 查看最近 1 小时的日志
docker compose logs --since=1h api
```

### 2.2 Nginx 访问日志

```bash
# 实时查看访问日志
docker compose exec nginx tail -f /var/log/nginx/access.log

# 统计状态码分布
docker compose exec nginx cat /var/log/nginx/access.log | \
  awk '{print $9}' | sort | uniq -c | sort -rn

# 统计请求量 Top 10 IP
docker compose exec nginx cat /var/log/nginx/access.log | \
  awk '{print $1}' | sort | uniq -c | sort -rn | head -10

# 查找响应时间 > 1 秒的慢请求
docker compose exec nginx awk '$NF > 1 {print $0}' /var/log/nginx/access.log
```

### 2.3 Nginx 错误日志

```bash
# 实时查看错误日志
docker compose exec nginx tail -f /var/log/nginx/error.log

# 查看最近 50 条错误
docker compose exec nginx tail -50 /var/log/nginx/error.log

# 搜索特定错误
docker compose exec nginx grep "upstream timed out" /var/log/nginx/error.log
```

### 2.4 应用日志

```bash
# 后端 API 日志（含请求路径、响应码、耗时）
docker compose logs -f api

# 前端 SSR 日志
docker compose logs -f frontend

# 数据库日志
docker compose logs -f postgres
```

### 2.5 日志清理

```bash
# Nginx 日志保留天数（通过 .env 配置）
# NGINX_LOG_RETENTION=30

# Docker 日志大小限制（编辑 /etc/docker/daemon.json）
{
  "log-driver": "json-file",
  "log-opts": {
    "max-size": "100m",
    "max-file": "5"
  }
}

# 重启 Docker 使配置生效
sudo systemctl restart docker
```

---

## 3. 数据库备份与恢复

### 3.1 手动备份

```bash
cd /opt/opsight

# 生成时间戳备份文件
BACKUP_FILE="opsight_$(date +%Y%m%d_%H%M%S).dump"

# 执行备份
docker compose exec -T postgres pg_dump \
  -U opsight \
  -d opsight \
  -F c \
  -f "/tmp/$BACKUP_FILE"

# 复制到宿主机
docker compose cp "postgres:/tmp/$BACKUP_FILE" /opt/backups/opsight/

echo "备份完成: /opt/backups/opsight/$BACKUP_FILE"
```

### 3.2 自动化备份脚本

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

echo "[$(date)] Backup saved: $BACKUP_DIR/$BACKUP_FILE" >> /var/log/opsight-backup.log

# 清理过期备份（保留最近 30 天）
find "$BACKUP_DIR" -name "*.dump" -mtime +"$RETENTION_DAYS" -delete
```

添加可执行权限：
```bash
chmod +x /opt/opsight/backup.sh
```

配置定时任务：
```bash
# 每天凌晨 2:00 执行备份
crontab -e
# 添加：
0 2 * * * /opt/opsight/backup.sh
```

### 3.3 备份验证

```bash
# 列出备份文件
ls -lh /opt/backups/opsight/

# 验证备份文件完整性
docker compose exec -T postgres pg_restore -l /tmp/opsight_20260529_020000.dump | head -20
```

### 3.4 数据库恢复

```bash
# 场景一：数据损坏，需全量恢复
docker compose stop api frontend nginx          # 停止应用
docker compose exec -T postgres psql -U opsight -d opsight \
  -c "DROP SCHEMA public CASCADE; CREATE SCHEMA public;"  # 清空数据库

docker compose exec -T postgres pg_restore \
  -U opsight \
  -d opsight \
  --clean --if-exists \
  < /opt/backups/opsight/opsight_20260529_020000.dump

docker compose up -d                            # 重启服务

# 场景二：恢复到新数据库
docker compose exec -T postgres createdb -U opsight opsight_restore
docker compose exec -T postgres pg_restore \
  -U opsight \
  -d opsight_restore \
  < /opt/backups/opsight/opsight_20260529_020000.dump
```

### 3.5 备份策略建议

| 场景 | 频率 | 保留期限 |
|------|------|---------|
| 全量备份 | 每日 | 30 天 |
| 重要变更前 | 手动 | 永久 |
| 升级前 | 手动 | 90 天 |

---

## 4. 磁盘空间监控

### 4.1 查看磁盘使用

```bash
# 系统磁盘使用
df -h

# 项目目录大小
du -sh /opt/opsight/*
du -sh /opt/backups/opsight/

# Docker 占用空间
docker system df
# TYPE            TOTAL     ACTIVE    SIZE      RECLAIMABLE
# Images          4         4         1.2GB     0B (0%)
# Containers      4         4         50MB      0B (0%)
# Local Volumes   2         2         500MB     0B (0%)
# Build Cache     10        0         300MB     300MB
```

### 4.2 数据库空间

```bash
# 数据库大小
docker compose exec postgres psql -U opsight -d opsight \
  -c "SELECT pg_database_size('opsight')/1024/1024 AS size_mb;"

# 各表大小
docker compose exec postgres psql -U opsight -d opsight \
  -c "SELECT relname AS table, pg_size_pretty(pg_total_relation_size(relid)) AS size
  FROM pg_catalog.pg_statio_user_tables ORDER BY pg_total_relation_size(relid) DESC;"
```

### 4.3 Nginx 日志大小

```bash
# 日志目录大小
docker compose exec nginx du -sh /var/log/nginx/

# 日志文件列表
docker compose exec nginx ls -lh /var/log/nginx/
```

### 4.4 空间清理

```bash
# 清理 Docker 未使用资源
docker system prune -a -f

# 清理构建缓存
docker builder prune -f

# 清理过期备份
find /opt/backups/opsight/ -name "*.dump" -mtime +30 -delete

# 清理旧 Nginx 日志
docker compose exec nginx find /var/log/nginx/ -name "*.log" -mtime +30 -delete
```

### 4.5 监控告警阈值

| 指标 | 告警阈值 | 处理动作 |
|------|---------|---------|
| 系统磁盘使用率 | > 85% | 清理日志、备份、临时文件 |
| Docker 卷大小 | > 10 GB | 检查数据库增长趋势 |
| 备份目录大小 | > 5 GB | 清理过期备份 |
| Nginx 日志 | > 1 GB | 执行日志轮转 |

---

## 5. 性能调优

### 5.1 Nginx 调优

编辑 `nginx/nginx.conf`：

```nginx
# worker 进程数 = CPU 核心数
worker_processes auto;

# 每个 worker 最大连接数（ulimit -n 需 > worker_connections * 2）
events {
    worker_connections 4096;
    multi_accept on;
}

# 静态文件缓存
location /_next/static {
    expires 30d;
    add_header Cache-Control "public, immutable";
}

# Gzip 压缩级别（1-9，5 为性能与压缩率平衡）
gzip_comp_level 5;

# 开启 HTTP/2（需 HTTPS）
listen 443 ssl http2;
```

### 5.2 PostgreSQL 调优

编辑 `docker-compose.yml` 中的数据库环境变量或使用自定义配置：

```yaml
services:
  postgres:
    # ...
    command:
      - "postgres"
      - "-c"
      - "shared_buffers=256MB"        # 共享缓冲区（总内存的 25%）
      - "-c"
      - "effective_cache_size=1GB"    # 有效缓存大小（总内存的 75%）
      - "-c"
      - "max_connections=100"         # 最大连接数
      - "-c"
      - "work_mem=16MB"               # 工作内存
```

### 5.3 容器资源限制

编辑 `docker-compose.yml`：

```yaml
services:
  api:
    deploy:
      resources:
        limits:
          cpus: '2'
          memory: 512M
        reservations:
          cpus: '1'
          memory: 256M

  frontend:
    deploy:
      resources:
        limits:
          cpus: '2'
          memory: 512M

  postgres:
    deploy:
      resources:
        limits:
          cpus: '2'
          memory: 1G
```

### 5.4 性能基准

| 指标 | 目标值 | 测量方法 |
|------|--------|---------|
| API 响应时间 (p95) | < 200ms | Nginx access log `urt` 字段 |
| 页面加载时间 | < 2s | 浏览器 DevTools |
| 并发用户 | 50+ | `ab -n 1000 -c 50 http://localhost/` |
| 数据库查询时间 | < 50ms | `pg_stat_statements` |

---

## 6. 健康检查端点

### 6.1 API 健康检查

```bash
# 基础健康检查
curl http://localhost/healthz
# 响应: {"code":0,"data":{"status":"ok","time":"2026-05-29T12:00:00Z"},"message":"success"}

# 详细健康检查（可添加 readiness 端点）
curl http://localhost/api/v1/services
# 响应: JSON 服务列表
```

### 6.2 数据库健康检查

```bash
# 容器健康状态
docker compose ps postgres
# STATUS 列应为: Up (healthy)

# 手动测试连接
docker compose exec postgres pg_isready -U opsight -d opsight
# 输出: /var/run/postgresql:5432 - accepting connections
```

### 6.3 Nginx 健康检查

```bash
# HTTP 可达性
curl -o /dev/null -s -w "%{http_code}\n" http://localhost/

# Nginx 配置测试
docker compose exec nginx nginx -t
# 输出: nginx: configuration file ... test is successful
```

### 6.4 综合健康检查脚本

创建 `/opt/opsight/healthcheck.sh`：

```bash
#!/usr/bin/env bash
FAIL=0

echo "=== Opsight 健康检查 ==="

# 1. API
if curl -sf http://localhost/healthz > /dev/null; then
    echo "[PASS] API 健康检查"
else
    echo "[FAIL] API 健康检查"
    FAIL=1
fi

# 2. Database
if docker compose exec -T postgres pg_isready -U opsight -d opsight > /dev/null 2>&1; then
    echo "[PASS] 数据库连接检查"
else
    echo "[FAIL] 数据库连接检查"
    FAIL=1
fi

# 3. Nginx
STATUS=$(curl -o /dev/null -s -w "%{http_code}" http://localhost/)
if [ "$STATUS" = "200" ]; then
    echo "[PASS] Nginx 状态检查 (HTTP $STATUS)"
else
    echo "[FAIL] Nginx 状态检查 (HTTP $STATUS)"
    FAIL=1
fi

# 4. 磁盘
DISK_USAGE=$(df /opt | awk 'NR==2 {print $5}' | sed 's/%//')
if [ "$DISK_USAGE" -lt 85 ]; then
    echo "[PASS] 磁盘使用率 (${DISK_USAGE}%)"
else
    echo "[FAIL] 磁盘使用率 (${DISK_USAGE}%)"
    FAIL=1
fi

exit $FAIL
```

---

## 7. 升级流程

### 7.1 标准升级

```bash
cd /opt/opsight

# 1. 备份数据库（必须）
./backup.sh

# 2. 拉取最新代码
git pull origin main
# 或替换为更新的文件包

# 3. 拉取最新基础镜像
docker compose pull

# 4. 重新构建并启动（滚动更新）
docker compose up -d --build

# 5. 验证
curl http://localhost/healthz
docker compose ps
```

### 7.2 零停机升级（推荐）

```bash
# 1. 备份数据库
./backup.sh

# 2. 构建新镜像（不重启）
docker compose build

# 3. 逐个滚动更新（Nginx 对 upstream 有健康检查）
docker compose up -d --no-deps --build api
sleep 5
docker compose up -d --no-deps --build frontend
sleep 5
docker compose up -d --no-deps --build nginx

# 4. 验证
curl http://localhost/healthz
```

### 7.3 升级检查清单

| 步骤 | 检查项 | 命令 |
|------|--------|------|
| 升级前 | 数据库备份成功 | `ls -lh /opt/backups/opsight/` |
| 升级前 | 当前版本确认 | `docker images opsight-*` |
| 升级后 | 所有容器 Healthy | `docker compose ps` |
| 升级后 | API 可达 | `curl http://localhost/healthz` |
| 升级后 | 前端可达 | `curl -I http://localhost` |
| 升级后 | 功能验证 | 登录、查看 Dashboard |
| 升级后 | 无新错误日志 | `docker compose logs --tail=50` |

---

## 8. 紧急回滚

### 8.1 镜像回滚

```bash
# 1. 查看历史镜像
docker images opsight-api --format "{{.Tag}}\t{{.CreatedAt}}"

# 2. 使用旧标签重新部署
# 假设旧版本标签为 v0.1.0
# 编辑 docker-compose.yml，指定镜像版本：
#   api:
#     image: opsight-api:v0.1.0

docker compose up -d

# 3. 验证
curl http://localhost/healthz
```

### 8.2 数据库回滚

```bash
# 1. 停止应用
docker compose stop api frontend nginx

# 2. 恢复数据库
docker compose exec -T postgres pg_restore \
  -U opsight -d opsight --clean --if-exists \
  < /opt/backups/opsight/opsight_<升级前备份>.dump

# 3. 重启服务
docker compose up -d
```

### 8.3 完整回滚流程

```bash
#!/usr/bin/env bash
# rollback.sh - 完整回滚脚本
set -euo pipefail

BACKUP_FILE="$1"  # 传入备份文件路径

if [ -z "$BACKUP_FILE" ]; then
    echo "用法: ./rollback.sh <备份文件路径>"
    echo "可用备份:"
    ls -lh /opt/backups/opsight/
    exit 1
fi

cd /opt/opsight

# 1. 停止服务
echo "停止服务..."
docker compose down

# 2. 恢复数据库
echo "恢复数据库..."
docker compose up -d postgres
sleep 5
docker compose exec -T postgres pg_restore -U opsight -d opsight --clean --if-exists < "$BACKUP_FILE"

# 3. 启动全部服务
echo "启动服务..."
docker compose up -d

# 4. 验证
sleep 5
if curl -sf http://localhost/healthz > /dev/null; then
    echo "[SUCCESS] 回滚完成"
else
    echo "[ERROR] 回滚后服务异常，请检查日志"
    docker compose logs --tail=50
    exit 1
fi
```

---

## 9. 日常巡检清单

### 每日检查

| 检查项 | 命令 | 正常标准 |
|--------|------|---------|
| 容器运行状态 | `docker compose ps` | 全部 Up (healthy) |
| API 健康检查 | `curl http://localhost/healthz` | HTTP 200 |
| 数据库连接 | `docker compose exec postgres pg_isready` | accepting connections |
| 磁盘使用率 | `df -h /opt` | < 80% |
| 错误日志 | `docker compose logs --tail=50 \| grep -i error` | 无新增 |

### 每周检查

| 检查项 | 命令 | 正常标准 |
|--------|------|---------|
| 备份完整性 | `ls -lh /opt/backups/opsight/` | 最近 7 天有备份 |
| Nginx 访问量 | `docker compose exec nginx cat /var/log/nginx/access.log \| wc -l` | 与预期一致 |
| 数据库大小 | 见 4.2 | 增长趋势正常 |
| SSL 证书有效期 | `openssl x509 -in ssl/fullchain.pem -noout -dates` | > 30 天 |

### 每月检查

| 检查项 | 操作 |
|--------|------|
| 安全更新 | 检查 Docker 镜像漏洞：`docker scout cves` |
| 性能回顾 | 分析 Nginx 慢请求日志 |
| 资源规划 | 评估 CPU/内存/磁盘增长趋势 |
| 备份恢复演练 | 在测试环境验证备份可恢复性 |
| 权限审计 | 检查 `.env` 和证书文件权限 |

---

## 附录 A：常用命令速查

```bash
# 启停
docker compose up -d            # 启动
docker compose down             # 停止并删除容器
docker compose restart          # 重启
docker compose ps               # 状态

# 日志
docker compose logs -f [服务名]  # 实时日志
docker compose logs --tail=100  # 最近 100 行

# 进入容器
docker compose exec api sh       # 进入后端
docker compose exec postgres psql -U opsight -d opsight  # 进入数据库

# 数据
docker compose cp "postgres:/tmp/backup.dump" ./   # 复制文件

# 清理
docker system prune -a -f       # 清理全部未使用资源
docker volume prune -f          # 清理未使用卷
```

## 附录 B：告警联系人配置

建议将以下运维告警接入企业通讯工具：

| 告警级别 | 通知方式 | 响应时间 |
|---------|---------|---------|
| Critical（严重） | 电话 + 即时通讯 | 5 分钟内 |
| Warning（警告） | 即时通讯 + 邮件 | 30 分钟内 |
| Info（信息） | 邮件 | 次日处理 |
