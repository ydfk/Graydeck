# Graydeck

`Graydeck` 是一个面向 `mihomo` 的管理端项目，当前包含：

- Go 编写的 `managerd` 后端
- React + TypeScript + Vite 前端
- 简体中文 / English i18n
- 以远程订阅为主的配置管理界面
- `zashboard` safe 模式接入外壳
- `pnpm workspace` 根脚本
- Go `air` 热更新开发流程

## 环境要求

- Go `1.24+`
- Node.js `22+`
- `pnpm`
- `air`

安装 `air`：

```powershell
go install github.com/air-verse/air@latest
```

如果本机还没有 `pnpm`，先安装：

```powershell
npm install -g pnpm
```

## 目录

```text
.
├─ cmd/managerd         # 后端入口
├─ internal/            # 后端内部实现
├─ web/                 # 前端
├─ DOCKER_IMAGE_DESIGN.md
└─ web/DESIGN.md
```

## 启动后端

项目根目录执行：

```powershell
pnpm run dev:server
```

根脚本会自动把 `GOCACHE` 指向仓库内的 `.gocache/`。

默认监听地址：

```text
http://localhost:18080
```

可用环境变量：

```powershell
$env:MGR_LISTEN=":18080"
$env:ZASHBOARD_MODE="safe"
```

## 启动前端

项目根目录执行：

```powershell
pnpm run dev:web
```

默认地址：

```text
http://localhost:5173
```

前端开发服务器会把 `/api/*` 代理到后端 `http://localhost:18080`。

## 同时启动前后端

项目根目录执行：

```powershell
pnpm run dev
```

这条命令会同时启动：

- `pnpm run dev:web`
- `pnpm run dev:server`

## 首次安装说明

首次拉项目建议先在根目录执行：

```powershell
pnpm install
```

当前仓库里如果已经存在旧的 `node_modules` 或旧锁文件，实际运行的 `vite` 版本可能不是 `package.json` 里声明的版本。为了确保前端真正使用 `pnpm + Vite 8`，建议删除旧依赖后重新安装。

## 常用命令

后端构建：

```powershell
pnpm run build:server
```

前端类型检查：

```powershell
pnpm run check:web
```

前端构建：

```powershell
pnpm run build:web
```

前后端一起检查：

```powershell
pnpm run check
```

前后端一起构建：

```powershell
pnpm run build
```

## 当前开发重点

- 多订阅源管理
- 当前启用订阅自动校验与应用
- 配置草稿与历史版本
- `zashboard` safe 模式控制层
- `mihomo` 核心更新
