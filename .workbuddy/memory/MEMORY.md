# Opsight 项目记忆

## 项目定位
- Opsight: AIOps 智能运维监控平台
- 目标：政府私有化部署，IT 验收交付
- 仓库：zhangsan1989707/opsight (main branch)

## 技术栈（已落地）
- 后端：Go + Gin + GORM + PostgreSQL + zerolog + golang-jwt
- 前端：Next.js 14 + React 18 + Tailwind CSS + Chart.js（字体离线化）
- 部署：Docker Compose（Nginx + api + frontend + postgres）

## 2026-05-29 全量改造完成 (两期 7 Agent × 14 任务)
### 一期（基础设施）
- 部署标准化：deploy.sh/deploy.bat/.env.example
- Nginx：反向代理 + Gzip + 安全头 + WebSocket + SSL就绪
- Docker Compose：Nginx + 健康检查 + 命名卷 + opsight-net
- 前端离线化：本地 woff2 字体（零外部依赖）
- zerolog 结构化日志 + 请求ID追踪
- API 统一 {code, message, data} 格式
- README.md

### 二期（业务核心）
- PostgreSQL + GORM：11 个模型 + AutoMigrate + Seed 数据
- JWT 认证 + RBAC（admin/editor/viewer 三级权限）
- 审计日志系统（async 写入 + 查询过滤 + 统计）
- 告警通知系统（6 个 API + 邮件/企微通道 + 事件集成）
- API 限流（滑动窗口）+ 安全头 + 输入验证
- 前端登录页 + AuthContext + 路由守卫
- 3 份专业文档：DEPLOY.md / OPS.md / TEST.md

## 关键约定
- 政府私有化部署要求（离线/审计/RBAC/等保）已基本覆盖
- API 响应格式：{ code: int, message: string, data: any }
- 部署：cp .env.example .env → docker compose up -d → http://localhost
- 默认管理员：admin@opsight.io / admin123
