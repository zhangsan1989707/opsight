# Opsight — AIOps 智能运维监控平台

![Version](https://img.shields.io/badge/version-0.1.0-blue)
![Go](https://img.shields.io/badge/Go-1.22+-00ADD8?logo=go)
![Next.js](https://img.shields.io/badge/Next.js-14-black?logo=next.js)
![License](https://img.shields.io/badge/license-MIT-green)

## 概述

Opsight 是一套面向企业 IT 运维团队的 AIOps 智能监控平台。它提供统一的服务监控、告警管理、故障分析和 AI 辅助洞察能力，帮助团队快速发现异常、定位根因、高效协作。平台采用前后端分离架构，支持 Docker Compose 一键部署，适用于私有化交付场景。

## 核心功能

- **Dashboard** — 全局服务健康概览，包含 QPS、延迟、错误率等关键指标卡片与趋势图
- **Metrics** — 多维度指标查询，支持自定义时间范围和指标选择
- **Incidents** — 故障事件全生命周期管理：告警生成、状态跟踪、解决关闭
- **Services** — 服务目录，展示各服务实时状态、响应时间、错误率及依赖关系
- **AI Insights** — 基于历史数据的智能运维建议，异常检测和趋势预测
- **Topology** — 服务拓扑可视化，支持根因分析（RCA）链路下钻
- **Alert Rules** — 告警规则管理，支持阈值配置、告警级别、启用/禁用
- **Integrations** — 外部系统集成（Prometheus、Webhook 等接入层）
- **Team** — 团队成员管理与角色分配

## 架构

```
                    ┌─────────────────────────────────┐
                    │          Browser / 客户端         │
                    └──────────────┬──────────────────┘
                                   │
                                   ▼
                    ┌─────────────────────────────────┐
                    │       Nginx (反向代理 :80)        │
                    └──────┬──────────────────┬───────┘
                           │                  │
               ┌───────────▼──────┐   ┌──────▼───────────┐
               │  Frontend :3800  │   │   Backend :8800   │
               │   Next.js 14     │   │    Go / Gin       │
               │   React + Tailwind│  │    REST + WS      │
               └──────────────────┘   └──────┬───────────┘
                                             │
                              ┌──────────────┼──────────────┐
                              │              │              │
                     ┌────────▼────┐  ┌──────▼──────┐  ┌───▼──────────┐
                     │ PostgreSQL  │  │   Redis     │  │  Prometheus  │
                     │   (规划中)   │  │  (规划中)    │  │  (规划中)     │
                     └─────────────┘  └─────────────┘  └──────────────┘
```

## 技术栈

| 层级     | 技术                                    | 说明                     |
| -------- | --------------------------------------- | ------------------------ |
| 后端语言 | Go 1.22+                                | 高性能并发支持           |
| Web 框架 | Gin                                     | 轻量级 HTTP 路由与中间件 |
| WebSocket | gorilla/websocket                      | 实时数据推送             |
| 前端框架 | Next.js 14 (App Router)                 | React SSR/SSG            |
| UI 组件  | React 18                                | 函数组件 + Hooks         |
| 样式方案 | Tailwind CSS 3                          | 暗色主题、响应式         |
| 图表库   | Chart.js + react-chartjs-2              | 折线图、柱状图、饼图     |
| 语言     | TypeScript                              | 类型安全                 |
| 容器化   | Docker + Docker Compose                 | 一键构建与编排           |
| 反向代理 | Nginx                                   | 统一入口、SSL 终止       |
| 日志     | Zerolog                                 | 结构化高性能日志         |

> 数据库层（PostgreSQL + Redis）处于规划阶段，当前版本使用内存数据模拟。

## 快速开始

### 1. 克隆项目

```bash
git clone <your-repo-url> opsight
cd opsight
```

### 2. 配置环境变量

```bash
cp .env.example .env
```

根据需要编辑 `.env` 中的配置项。

### 3. 启动服务

```bash
docker compose up -d
```

### 4. 访问平台

打开浏览器访问：

```
http://localhost
```

默认端口分配：

| 服务     | 端口  | 说明          |
| -------- | ----- | ------------- |
| Nginx    | 80    | 统一入口      |
| Frontend | 3800  | Next.js 前端  |
| Backend  | 8800  | Go API 服务   |

## 项目结构

```
opsight/
├── opsight-backend/               # Go 后端
│   ├── main.go                    # 主入口（路由、模型、处理函数）
│   ├── pkg/
│   │   └── logger/                # Zerolog 结构化日志封装
│   ├── Dockerfile
│   ├── go.mod
│   └── go.sum
├── opsight-frontend/              # Next.js 前端
│   ├── app/
│   │   ├── layout.tsx             # 根布局（侧边栏、导航）
│   │   ├── page.tsx               # 首页重定向
│   │   ├── globals.css            # 全局样式 + Tailwind
│   │   ├── lib/
│   │   │   └── api.ts             # API 请求封装
│   │   ├── components/            # 公共组件
│   │   │   ├── UI.tsx             # 通用 UI 组件
│   │   │   ├── Notification.tsx   # 通知系统
│   │   │   └── ErrorBoundary.tsx  # 错误边界
│   │   └── {page}/page.tsx        # 各页面路由
│   │       ├── dashboard/         # 仪表盘
│   │       ├── metrics/           # 指标查询
│   │       ├── incidents/         # 故障管理
│   │       ├── services/          # 服务目录
│   │       ├── insights/          # AI 洞察
│   │       ├── topology/          # 服务拓扑
│   │       ├── alerts/            # 告警规则
│   │       ├── integrations/      # 集成管理
│   │       └── team/              # 团队管理
│   ├── public/                    # 静态资源
│   ├── Dockerfile
│   ├── package.json
│   └── tailwind.config.ts
├── ui-template/                   # 原始 UI 模板（HTML 原型）
├── docker-compose.yml             # Docker Compose 编排
├── .env.example                   # 环境变量模板
├── deploy.sh                      # Linux/macOS 部署脚本
└── deploy.bat                     # Windows 部署脚本
```

## API 接口

### 基础接口

| 方法   | 路径            | 说明         |
| ------ | --------------- | ------------ |
| GET    | `/healthz`      | 健康检查     |
| GET    | `/api/v1/ws`    | WebSocket 连接 |

### 认证

| 方法 | 路径                  | 说明     |
| ---- | --------------------- | -------- |
| POST | `/api/v1/auth/login`  | 用户登录 |

### 仪表盘

| 方法 | 路径                             | 说明           |
| ---- | -------------------------------- | -------------- |
| GET  | `/api/v1/dashboard/summary`      | 服务健康摘要   |
| GET  | `/api/v1/dashboard/error-rate`   | 错误率趋势     |
| GET  | `/api/v1/dashboard/latency`      | 延迟趋势       |
| GET  | `/api/v1/dashboard/top-errors`   | Top 错误统计   |

### 故障管理

| 方法 | 路径                                | 说明           |
| ---- | ----------------------------------- | -------------- |
| GET  | `/api/v1/incidents`                 | 故障事件列表   |
| GET  | `/api/v1/incidents/:id`             | 故障事件详情   |
| POST | `/api/v1/incidents/:id/resolve`     | 解决故障事件   |

### 服务管理

| 方法 | 路径                        | 说明           |
| ---- | --------------------------- | -------------- |
| GET  | `/api/v1/services`          | 服务列表       |
| GET  | `/api/v1/services/:name`    | 服务详情       |

### 告警规则

| 方法  | 路径                               | 说明           |
| ----- | ---------------------------------- | -------------- |
| GET   | `/api/v1/alert-rules`              | 告警规则列表   |
| PATCH | `/api/v1/alert-rules/:id/toggle`   | 启用/禁用规则  |

### 指标查询

| 方法 | 路径                        | 说明           |
| ---- | --------------------------- | -------------- |
| GET  | `/api/v1/metrics/query`     | 指标数据查询   |
| GET  | `/api/v1/metrics/names`     | 可用指标列表   |

### 拓扑与洞察

| 方法 | 路径                               | 说明                 |
| ---- | ---------------------------------- | -------------------- |
| GET  | `/api/v1/topology`                 | 服务拓扑数据         |
| GET  | `/api/v1/topology/:serviceId/rca`  | 根因分析              |
| GET  | `/api/v1/insights`                 | AI 运维洞察          |

### 集成与团队

| 方法 | 路径                        | 说明           |
| ---- | --------------------------- | -------------- |
| GET  | `/api/v1/integrations`      | 集成列表       |
| GET  | `/api/v1/team`              | 团队成员列表   |

## 部署

项目支持统一的部署脚本进行一键部署：

- **Linux / macOS**: 执行 `./deploy.sh`
- **Windows**: 执行 `deploy.bat`

部署脚本会自动完成以下步骤：

1. 检查 Docker 和 Docker Compose 是否安装
2. 从 `.env.example` 生成 `.env` 配置（如不存在）
3. 构建 Docker 镜像
4. 启动所有服务容器
5. 输出访问地址和服务状态

> 部署前请确保已安装 Docker 20.10+ 和 Docker Compose v2。

## 本地开发

如果不使用 Docker，也可以直接在本地启动前后端服务。

### 后端

```bash
cd opsight-backend

# 安装依赖
go mod download

# 启动（开发模式）
go run main.go

# 服务运行在 http://localhost:8800
```

### 前端

```bash
cd opsight-frontend

# 安装依赖
npm install

# 启动开发服务器
npm run dev

# 前端运行在 http://localhost:3800
```

> 开发模式下前端直连后端 `http://localhost:8800`，无需 Nginx。

## 许可证

本项目基于 MIT 许可证开源。详见 [LICENSE](LICENSE) 文件。
