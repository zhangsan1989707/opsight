# Opsight Agent

Opsight 服务器监控探针 —— 部署在目标服务器上，采集系统指标并推送至 Opsight 后端。

## 支持的指标

| 指标类别 | 采集项 | 说明 |
|---------|--------|------|
| CPU | usage_percent, cores | CPU 使用率（%），逻辑核心数 |
| 内存 | total_mb, used_mb, usage_percent | 内存总量、已用量、使用率 |
| 磁盘 | mount, total_mb, used_mb, usage_percent | 各挂载点磁盘使用情况 |
| 网络 | bytes_recv_per_sec, bytes_sent_per_sec | 每秒收/发字节数 |
| 负载 | load1, load5, load15 | 系统 1/5/15 分钟平均负载 |

## 安装

### 方式一：预编译二进制（推荐）

```bash
# 下载对应平台的二进制文件
# Linux amd64:
wget https://your-repo/opsight-agent-linux-amd64 -O /usr/local/bin/opsight-agent
chmod +x /usr/local/bin/opsight-agent

# 创建配置
cat > /usr/local/bin/agent.yaml <<EOF
server:
  url: "http://your-opsight-server:80"
  api_key: "your-api-key"
collector:
  interval_seconds: 10
  cpu: true
  memory: true
  disk: true
  network: true
  load: true
logging:
  level: "info"
EOF

# 启动
/usr/local/bin/opsight-agent
```

### 方式二：源码编译

```bash
# 静态编译（不需要 CGO）
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o opsight-agent .
CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -o opsight-agent.exe .
```

## agent.yaml 配置说明

| 配置项 | 类型 | 默认值 | 说明 |
|--------|------|--------|------|
| server.url | string | http://localhost:80 | Opsight 后端地址 |
| server.api_key | string | (必填) | Agent 认证密钥 |
| collector.interval_seconds | int | 10 | 采集间隔（秒） |
| collector.cpu | bool | true | 是否采集 CPU |
| collector.memory | bool | true | 是否采集内存 |
| collector.disk | bool | true | 是否采集磁盘 |
| collector.network | bool | true | 是否采集网络 |
| collector.load | bool | true | 是否采集负载 |
| logging.level | string | info | 日志级别（info/debug） |

### 环境变量覆盖

配置项可通过环境变量覆盖（优先级高于 agent.yaml）：

| 环境变量 | 对应配置项 |
|----------|-----------|
| OPSIGHT_SERVER_URL | server.url |
| OPSIGHT_API_KEY | server.api_key |
| OPSIGHT_INTERVAL | collector.interval_seconds |

## systemd 服务配置

将 `opsight-agent.service` 复制到 `/etc/systemd/system/`：

```bash
cp opsight-agent.service /etc/systemd/system/
systemctl daemon-reload
systemctl enable opsight-agent
systemctl start opsight-agent
systemctl status opsight-agent
```

## 离线部署

在无法联网的服务器上部署时：

1. 在开发机上编译静态二进制：
   ```bash
   CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o opsight-agent .
   ```

2. 将 `opsight-agent` 二进制和 `agent.yaml` 配置文件一起复制到目标服务器：
   ```bash
   scp opsight-agent agent.yaml user@target-server:/usr/local/bin/
   ```

3. 在目标服务器上修改 `agent.yaml` 中的 `server.url` 和 `api_key`

4. 配置 systemd 服务并启动（见上方）

## 安全说明

- Agent 为**只读模式**，不会修改目标服务器上的任何文件或配置
- 所有采集均使用操作系统标准接口（Linux `/proc` 文件系统，Windows WMI/gopsutil）
- 与服务端通信使用 HTTPS 时建议配置 TLS 证书校验
- API Key 建议通过环境变量 `OPSIGHT_API_KEY` 传入，避免写入配置文件

## 故障排查

### Agent 无法启动

```bash
# 检查配置是否正确
/usr/local/bin/opsight-agent  # 直接运行，查看输出

# 常见问题：
# - agent.yaml 格式错误（YAML 缩进）
# - server.api_key 未配置
# - 端口被占用
```

### 上报失败

```bash
# 检查网络连通性
curl -X POST http://your-server/api/v1/agents/report \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer your-api-key" \
  -d '{"agent_version":"1.0.0"}'

# 检查日志
journalctl -u opsight-agent -f
```

### 采集数据为空

```bash
# Linux：确认 /proc 文件系统可读
cat /proc/stat
cat /proc/meminfo

# 检查 systemd 服务日志
journalctl -u opsight-agent --since "5 minutes ago"
```
