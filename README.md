# 考古发掘出土文物编目系统（DigCatalog）

面向考古工地出土文物登记与编目的全栈演示项目：支持发掘工地、探方/发掘单位、出土文物、材质字典的 CRUD，以及概览统计。

## 技术栈

- **前端**: Vue 3 + Vite + Pinia + Vue Router（Composition API + `<script setup>`）
- **后端**: Go 1.21+ + Gin + GORM
- **数据库**: MySQL 8.0
- **认证**: JWT + bcrypt

## 一键启动

```bash
docker compose up --build
```

启动完成后访问：

| 服务 | 地址 |
|------|------|
| 前端 | http://localhost:3200 |
| 后端 API | http://localhost:8200/api |
| MySQL | localhost:3307（用户 `root` / 密码 `root`，库名 `digcatalog`） |

停止服务：

```bash
docker compose down
```

清除数据卷后重建：

```bash
docker compose down -v
docker compose up --build
```

## 测试账号

| 用户名 | 密码 | 角色 |
|--------|------|------|
| `admin` | `123456` | 管理员 |
| `recorder` | `123456` | 记录员 |

## 功能模块

1. **登录认证** — 管理员 / 记录员角色，JWT 鉴权
2. **发掘工地 Site** — 名称、时代、经纬度、负责人
3. **探方/发掘单位 Unit** — 所属工地、编号、平面长宽（厘米，可选）、深度区间、地层简述
4. **出土文物 Find** — 所属探方、登记号、器物类型、材质、完整度、**探方局部三维坐标 (xCm/yCm/zCm)**、出土日期、描述、存放位置
5. **出土点示意** — 选定探方后以纯 SVG 平面网格标出全部带坐标文物，点击点位查看登记号与坐标，无重型地图 SDK
6. **材质分类 Material** — 名称、描述（字典表）
7. **概览页** — 工地数、探方数、文物总数、按器物类型统计

### 出土点坐标规则

- 每件 Find 的坐标 `xCm/yCm/zCm`（整数厘米，相对探方原点的局部坐标）均为**可选**：
  - 三个值要么**全部留空**（旧记录、未测点允许保留），要么**同时填写且均为非负整数**；
  - **同一探方内** `(xCm, yCm, zCm)` 组合唯一，重复提交返回 **409 Conflict**。
- Unit 的 `lengthCm/widthCm` 描述探方平面尺寸，可选；未配置时示意图按 **1000 × 1000 cm** 占位并提示去补填实际尺寸。
- 演示数据中「二里头遗址发掘区A / T1」配置了 1000 × 1000 cm 尺寸，并挂有 3 件带坐标文物（EL-2024-0001/0002/0004）。

## API 前缀

所有接口以 `/api` 开头：

- `POST /api/auth/login`
- `GET|POST|PUT|DELETE /api/sites`
- `GET|POST|PUT|DELETE /api/units`
- `GET /api/units/:id/spot-map` — 返回探方平面尺寸与该探方全部带坐标 Find 的登记号、器物类型、坐标
- `GET|POST|PUT|DELETE /api/finds` — Find 载荷含可选 `xCm/yCm/zCm`
- `GET|POST|PUT|DELETE /api/materials`
- `GET /api/overview`

坐标接口约定：

- `POST/PUT /api/finds` 中 `xCm/yCm/zCm` 要么都省略（`null`），要么三者同为非负整数；否则 `400`。
- 同 Unit 内坐标组合已被占用时返回 `409`，响应体含 `{"conflict":"coord","registerNo":"..."}`。

前端经 Nginx 将 `/api` 反代至后端容器 `http://backend:8080`。

## 端口映射

| 服务 | 宿主机 | 容器内 |
|------|--------|--------|
| Frontend | 3200 | 80 |
| Backend | 8200 | 8080 |
| MySQL | 3307 | 3306 |

## 目录结构

```
DigCatalog/
├── docker-compose.yml
├── README.md
├── .gitignore
├── backend/
│   ├── Dockerfile
│   ├── go.mod
│   ├── go.sum
│   ├── main.go
│   └── internal/
│       ├── config/
│       ├── models/
│       ├── handlers/
│       ├── middleware/
│       └── seed/
└── frontend/
    ├── Dockerfile
    ├── nginx.conf
    ├── package.json
    ├── vite.config.js
    ├── index.html
    └── src/
```

## 本地开发（可选）

### 后端

```bash
cd backend
go mod tidy
# 确保 MySQL 已启动且环境变量正确
go run .
```

### 前端

```bash
cd frontend
npm install --registry=https://registry.npmmirror.com
npm run dev
```
