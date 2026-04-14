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
├─ config/               # 服务配置与基础运行配置
├─ internal/
├─ web/
├─ data/                 # 运行时数据目录，首次启动后自动生成
├─ DOCKER_IMAGE_DESIGN.md
└─ web/DESIGN.md
```

`config/` 下包含这些内容：

- `config/base.yaml`：基础运行配置，启动时会注入到最终运行配置
- `config/graydeck.yaml`：Graydeck 服务配置，例如 `zashboard.hide-settings`

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
$env:GRAYDECK_SECRET="graydeck-secret"
```

说明：

- `GRAYDECK_SECRET` 用来设置控制面密钥
- `MGR_LISTEN` 也已固定为 `:18080`，不再通过环境变量覆盖
- `GRAYDECK_DATA_DIR` / `GRAYDECK_WEB_ROOT` / `GRAYDECK_CORE_OS` / `GRAYDECK_CORE_ARCH` / `GRAYDECK_CONTROLLER_ADDR` / `GRAYDECK_MIXED_PORT` 已改为程序内固定策略，不再通过环境变量覆盖

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

## Docker

仓库已提供：

- `Dockerfile`
- `.dockerignore`
- `docker-compose.example.yml`

示例（本地 compose 启动）：

```powershell
docker compose -f docker-compose.example.yml up -d --build
```

访问地址：

- `http://localhost:8080`

Compose 示例会同时挂载：

- `./config:/config`
- `./data:/data`

Compose 示例还额外映射了常用 mihomo 端口：`7890`、`7891`、`7892`、`7893`（含必要 UDP），便于直接在宿主机使用代理能力。

`cap_add: NET_ADMIN` 与 `/dev/net/tun` 设备挂载仅在你需要 TUN/透明代理时才必须；如果只使用普通 HTTP/SOCKS 代理，可移除这两项。

## DockerHub 镜像发布脚本

支持 3 个脚本：`buil`、`push`、`buildPush`（同时也提供 `docker:*` 同义命令）。

传参规则：

- 仅支持命令行参数，不再读取环境变量
- 未传必需参数时会直接退出，不会执行打包/推送

设置镜像仓库（示例）：

```powershell
pnpm run build:docker -- --DOCKERHUB_REPO=your-user/graydeck
```

仅构建（自动按时间生成版本号，如 `202604132146`）：

```powershell
pnpm run buil -- your-user/graydeck
```

仅推送指定 tag：

```powershell
pnpm run push -- your-user/graydeck 202604132146
```

一键构建并推送（自动时间版本 + `latest`）：

```powershell
pnpm run buildPush -- your-user/graydeck
```
