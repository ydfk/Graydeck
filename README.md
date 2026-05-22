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
- `config/graydeck.yaml`：Graydeck 服务配置，例如 `server.port`、`auth.username`、`auth.password`、`update.prefer-proxy`、`update.proxy-url`、`zashboard.hide-settings`

`config/base.yaml` 在 Docker 场景下至少要保留这几个基础项：

- `bind-address: "0.0.0.0"`
- `allow-lan: true`

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
- 后端：`http://localhost:8080`

默认登录配置在 `config/graydeck.yaml`：

```yaml
server:
  port: 8080
auth:
  enabled: true
  username: admin
  password: admin123
update:
  prefer-proxy: true
  proxy-url: https://ghfast.top/
zashboard:
  hide-settings: true
```

启动后，未登录无法查看控制台、日志、`Zashboard` 和后端 API。

## 可用环境变量

```powershell
$env:GRAYDECK_SECRET="graydeck-secret"
```

说明：

- `GRAYDECK_SECRET` 用来设置控制面密钥
- 管理端口通过 `config/graydeck.yaml` 的 `server.port` 配置，默认值为 `8080`
- 修改 `server.port` 后需要重启 Graydeck 服务
- `GRAYDECK_DATA_DIR` / `GRAYDECK_WEB_ROOT` / `GRAYDECK_CORE_OS` / `GRAYDECK_CORE_ARCH` / `GRAYDECK_CONTROLLER_ADDR` / `GRAYDECK_MIXED_PORT` 已改为程序内固定策略，不再通过环境变量覆盖

## 当前行为

### 核心

- 如果本地没有 `mihomo` 核心，后端启动时会自动拉取最新正式版
- 自动更新会优先尝试 `config/graydeck.yaml` 里的代理地址，失败后再回退到原始地址
- 当前核心版本和最新版本会显示在控制台
- 如果检测到新版本，可以在界面里手动升级
- 核心安装和升级都支持 3 种来源：系统自动更新、手动指定地址、上传文件

### 配置文件

- 当前以远程订阅为主
- 每个配置文件都会记录同步状态
- 常见状态包括：`可用`、`订阅失败`、`格式校验失败`
- 如果没有可用配置，或者当前配置校验失败，核心不会启动，运行状态里会显示原因
- 支持 YAML 在线预览

### Zashboard

- 如果本地没有 `zashboard` 资源，后端启动时会自动拉取最新版本
- 自动更新会优先尝试 `config/graydeck.yaml` 里的代理地址，失败后再回退到原始地址
- `Zashboard` 页面会显示当前版本、最新版本和升级入口
- `Zashboard` 安装和升级也支持系统自动更新、手动指定地址、上传文件
- 页面路由使用 `/zashboard-ui/`
- `config/graydeck.yaml` 中的 `zashboard.hide-settings` 默认为 `true`

## 常用命令

生成单文件可执行程序：

```powershell
pnpm run build:standalone
```

指定 Linux 平台：

```powershell
pnpm run build:standalone -- --target=linux-amd64
```

也可以拆开指定：

```powershell
pnpm run build:standalone -- --os=linux --arch=arm64
```

产物会输出到 `dist/graydeck-<系统>-<架构>`，Windows 目标会自动带 `.exe` 后缀。这个可执行文件已经内置前端管理页，运行时仍会使用当前目录下的 `config/` 和 `data/`。

### Debian 13 服务安装

先生成 Linux 单文件版本：

```powershell
pnpm run build:standalone -- --target=linux-amd64
```

把 `graydeck-linux-amd64` 和仓库里的安装脚本放到 Debian 13 后执行：

```bash
sudo bash ./install-debian-systemd.sh ./graydeck-linux-amd64
```

在仓库目录中也可以直接执行：

```bash
sudo bash ./scripts/install-debian-systemd.sh ./dist/graydeck-linux-amd64
```

脚本只会写入 `/etc/systemd/system/graydeck.service` 并启动服务，不会移动可执行文件，也不会创建用户或修改安装目录。服务会把传入二进制所在目录作为 `WorkingDirectory`，单文件会在该目录下使用 `config/` 和 `data/`。

常用服务命令：

```bash
sudo systemctl status graydeck.service
sudo systemctl restart graydeck.service
sudo journalctl -u graydeck.service -f
```

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

`pnpm run build` 会先构建前端，再把 `web/dist` 同步到 Go 的嵌入目录，最后编译后端。

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

镜像内 `mihomo` 默认监听的是 `17890`、`17891`、`17892`、`17893`，Compose 已经把它们映射成宿主机常见端口 `7890`、`7891`、`7892`、`7893`。

Docker 镜像里的 `managerd` 同样内置前端管理页，不需要额外挂载 Web 静态资源目录。

`cap_add: NET_ADMIN` 与 `/dev/net/tun` 设备挂载仅在你需要 TUN/透明代理时才必须；如果只使用普通 HTTP/SOCKS 代理，可移除这两项。

## 版本发布

Graydeck 使用 `1.0.1` 这类语义化版本。正式发布通过 tag 触发，不会在普通提交时发布：

```bash
git tag v1.0.1
git push origin v1.0.1
```

GitHub Actions 会为同一个版本同时处理：

- GitHub Release 的 standalone 可执行文件
- `ghcr.io/<owner>/graydeck:1.0.1`
- `ghcr.io/<owner>/graydeck:latest`

发布时注入到 Docker 镜像和 standalone 可执行文件内的版本号都使用 tag 去掉 `v` 后的值。

## DockerHub 镜像发布脚本

支持 3 个脚本：`build:docker`、`push:docker`、`buildPush:docker`（同时也提供 `docker:*` 同义命令）。

传参规则：

- 仅支持命令行参数，不再读取环境变量
- 未传必需参数时会直接退出，不会执行打包/推送

设置镜像仓库（示例）：

```powershell
pnpm run build:docker -- --DOCKERHUB_REPO=your-user/graydeck --DOCKER_IMAGE_TAG=1.0.1
```

仅构建：

```powershell
pnpm run build:docker -- your-user/graydeck 1.0.1
```

仅推送指定 tag：

```powershell
pnpm run push:docker -- your-user/graydeck 1.0.1
```

一键构建并推送版本 tag 与 `latest`：

```powershell
pnpm run buildPush:docker -- your-user/graydeck 1.0.1
```
