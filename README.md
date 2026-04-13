# Graydeck

`Graydeck` 是一个面向 `mihomo` 的轻量管理端，当前已经具备一条真实可运行的基础链路：

- 首次启动时自动检查并下载最新 `mihomo` 核心
- 首次启动时自动检查并下载最新 `zashboard` 静态资源
- 远程订阅配置拉取、状态记录、格式校验、YAML 在线预览
- 当前启用配置切换与运行状态反馈
- 核心与 `zashboard` 版本检查、手动升级
- React + TypeScript + Vite 8 前端
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
├─ cmd/managerd
├─ internal/
├─ web/
├─ data/                 # 运行时数据目录，首次启动后自动生成
├─ DOCKER_IMAGE_DESIGN.md
└─ web/DESIGN.md
```

`data/` 下会生成这些内容：

- `data/core/`：`mihomo` 核心与版本记录
- `data/zashboard/`：`zashboard` 静态资源与版本记录
- `data/subscriptions.json`：配置文件元数据
- `data/subscriptions/*.yaml`：订阅拉取后的 YAML 预览文件
- `data/runtime/current.yaml`：当前生效配置

## 启动

首次安装依赖：

```powershell
pnpm install
```

启动前后端：

```powershell
pnpm run dev
```

只启动后端：

```powershell
pnpm run dev:server
```

只启动前端：

```powershell
pnpm run dev:web
```

默认地址：

- 前端：`http://localhost:5173`
- 后端：`http://localhost:18080`

## 可用环境变量

```powershell
$env:MGR_LISTEN=":18080"
$env:GRAYDECK_DATA_DIR=".\data"
$env:GRAYDECK_CORE_OS="linux"
$env:GRAYDECK_CORE_ARCH="amd64"
```

说明：

- `GRAYDECK_DATA_DIR` 用来指定运行时数据目录
- `GRAYDECK_CORE_OS` / `GRAYDECK_CORE_ARCH` 用来指定自动下载的核心目标平台
- 在 Docker 镜像里通常会使用 `linux` / `amd64` 或 `linux` / `arm64`

## 当前行为

### 核心

- 如果本地没有 `mihomo` 核心，后端启动时会自动拉取最新正式版
- 当前核心版本和最新版本会显示在控制台
- 如果检测到新版本，可以在界面里手动升级

### 配置文件

- 当前以远程订阅为主
- 每个配置文件都会记录同步状态
- 常见状态包括：`可用`、`订阅失败`、`格式校验失败`
- 如果没有可用配置，或者当前配置校验失败，核心不会启动，运行状态里会显示原因
- 支持 YAML 在线预览

### Zashboard

- 如果本地没有 `zashboard` 资源，后端启动时会自动拉取最新版本
- `Zashboard` 页面会显示当前版本、最新版本和升级入口
- 页面路由使用 `/zashboard-ui/`

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
