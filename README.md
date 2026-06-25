<p align="center">
  <img src="./web/public/graydeck-logo.svg" alt="Graydeck logo" width="120" />
</p>

<h1 align="center">Graydeck</h1>

<p align="center">
  <code>mihomo</code> 的轻量配置、更新与面板管理工具
</p>

<p align="center">
  <img src="https://img.shields.io/badge/Go-1.24-00ADD8?logo=go&logoColor=white" alt="Go 1.24" />
  <img src="https://img.shields.io/badge/React-19-61DAFB?logo=react&logoColor=white" alt="React 19" />
  <img src="https://img.shields.io/badge/Vite-8-646CFF?logo=vite&logoColor=white" alt="Vite 8" />
  <img src="https://img.shields.io/badge/Docker-blue?logo=docker&logoColor=white" alt="Docker" />
  <img src="https://img.shields.io/badge/license-MIT-green" alt="MIT" />
</p>

---

## 这是什么

Graydeck 把 `mihomo` 日常使用中最频繁的几个操作收进同一个 Web 控制台：管理远程订阅配置、更新核心和面板资源、查看运行状态与日志。它不做代理面板本身——而是让配置维护、版本更新和状态监控这些"运维动作"更顺手。

**适合谁用**：保留了 `mihomo` YAML 配置能力、以远程订阅为主要配置来源、希望用轻量 Web 界面替代手动维护的场景。

## 核心功能

<p align="center">
  <table>
    <tr>
      <td><strong>📡 订阅配置管理</strong></td>
      <td>多配置、手动同步与切换、YAML 预览、状态记录</td>
    </tr>
    <tr>
      <td><strong>🛡 配置校验</strong></td>
      <td>启动或切换前生成最终运行配置，交由 <code>mihomo</code> 做合法性校验</td>
    </tr>
    <tr>
      <td><strong>🔄 更新中心</strong></td>
      <td>自动检查并安装 <code>mihomo</code> 核心与 <code>Zashboard</code> 资源，支持自定义地址和上传本地文件</td>
    </tr>
    <tr>
      <td><strong>🧩 面板整合</strong></td>
      <td>内置 <code>Zashboard</code>，通过 Graydeck 同域接入，可按配置屏蔽设置菜单</td>
    </tr>
    <tr>
      <td><strong>⚙ 基础配置注入</strong></td>
      <td>通过 <code>base.yaml</code> 管理端口、<code>dns</code>、<code>tun</code> 等基础运行配置的覆盖与合并</td>
    </tr>
    <tr>
      <td><strong>📊 运行可观测</strong></td>
      <td>登录鉴权、运行状态控制、核心日志、自身版本检查</td>
    </tr>
  </table>
</p>

## 快速开始

### Docker（推荐）

```bash
# 1. 准备目录和配置
mkdir -p config data
cp docker-compose.example.yml docker-compose.yml

# 2. 修改 docker-compose.yml 中的 GRAYDECK_SECRET
# 3. （可选）在 config/ 下放置 base.yaml 和 graylevel.yaml

# 4. 启动
docker compose up -d
```

管理页默认监听 `8080` 端口。示例 compose 文件还映射了常用代理端口：

| 宿主机 | 容器内 | 用途 |
| --- | --- | --- |
| `8080` | `8080` | Graydeck Web |
| `7890` | `17890` | mixed port |
| `7891` | `17891` | SOCKS |
| `7892` | `17892` | redir |
| `7893` | `17893` | tproxy |

> 只使用 HTTP/SOCKS 代理时可以移除 `NET_ADMIN` 和 `/dev/net/tun`；需要 TUN 或透明代理时保留。

### Standalone

Release 页面提供 x86_64 单文件产物（`graydeck-linux-amd64` / `graydeck-windows-amd64.exe`），内置前端管理页，直接运行即可。程序会在运行目录下创建 `config/` 和 `data/` 目录。

## 配置

**`config/graydeck.yaml`** — Graydeck 自身配置：

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

**`config/base.yaml`** — 参与运行配置生成，适合声明端口、`external-controller`、`dns`、`tun` 等基础项。`base.yaml` 中的字段会覆盖订阅配置中的同名字段（列表字段以 `base.yaml` 为准，未声明的保留订阅原值）。

## 管理流程

1. 添加远程订阅配置
2. Graydeck 拉取订阅并生成最终运行配置
3. `mihomo` 校验通过后启动或切换到该配置
4. 在控制台管理核心版本、配置状态、日志和 `Zashboard` 面板

## 技术栈

| 层 | 技术 |
| --- | --- |
| 后端 | Go 1.24 |
| 前端 | React 19 + TypeScript + Vite 8 |
| 面板 | 内置 Zashboard 静态资源 |
| 部署 | Docker / 单文件 standalone |

## 开发

```bash
# 后端
go run ./cmd/graydeck

# 前端（开发模式）
cd web && npm run dev
```

项目设计记录见 [DOCKER_IMAGE_DESIGN.md](./DOCKER_IMAGE_DESIGN.md)，前端样式规范见 [web/DESIGN.md](./web/DESIGN.md)。
