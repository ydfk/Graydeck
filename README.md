# Graydeck

<p align="center">
  <img src="./web/public/graydeck-logo.svg" alt="Graydeck logo" width="88" />
</p>

<p align="center">
  面向 <code>mihomo</code> 的轻量配置、更新与面板管理工具。
</p>

`Graydeck` 是一个围绕 `mihomo` 日常使用场景设计的轻量管理端。它把远程订阅配置、核心更新、运行状态和 `Zashboard` 面板收进同一个界面，让部署完成后的配置维护更集中，也尽量减少手动整理核心文件、面板资源和运行配置的工作。

项目关注的不是重新实现一套代理面板，而是把 `mihomo` 常见的管理链路串起来：

1. 维护多个远程订阅配置，并按需同步、预览和切换。
2. 在配置应用前生成最终运行配置，并交给 `mihomo` 做校验。
3. 管理 `mihomo` 核心和 `Zashboard` 资源的安装与更新。
4. 在 Graydeck 中查看运行状态、日志和同域接入的 `Zashboard` 页面。

Graydeck 可以作为 Docker 服务运行，也可以打包为内置前端页面的 standalone 单文件程序。无论采用哪种方式，管理界面和运行数据组织方式都保持一致。

## 项目定位

Graydeck 适合希望保留 `mihomo` 配置能力，同时想把日常维护收敛到一个轻量 Web 控制台的场景：

- 以远程订阅为主要配置来源，需要管理多个配置文件
- 希望在切换配置前先看到同步和校验结果
- 希望由管理端处理核心与面板资源的下载、更新和状态展示
- 希望直接在同一入口查看 Graydeck 状态与 `Zashboard` 面板

Graydeck 不负责替代订阅服务，也不把 `mihomo` 的完整配置能力藏进复杂向导。基础配置仍由 YAML 表达，管理端更专注于订阅、校验、应用和更新这些高频动作。

## 特性

- **订阅配置管理**：支持多配置文件、手动同步、切换、状态记录和 YAML 预览
- **配置应用保护**：启动或切换前生成最终运行配置，并使用 `mihomo` 校验
- **更新中心**：自动检查并安装 `mihomo` 核心与 `Zashboard` 资源，也支持指定地址和上传文件
- **面板整合**：内置 `Zashboard` 页面并通过 Graydeck 同域接入，可按配置屏蔽设置菜单
- **基础配置注入**：通过 `base.yaml` 管理端口、`dns`、`tun` 等基础运行配置的覆盖与合并
- **运行可观测性**：提供登录鉴权、运行状态控制、核心日志和 Graydeck 自身版本检查

## 管理流程

1. 添加远程订阅配置。
2. Graydeck 拉取订阅并生成最终运行配置。
3. `mihomo` 校验通过后启动或切换到该配置。
4. 在控制台管理核心版本、配置状态、日志和 `Zashboard` 面板。

## 运行方式

### Docker

仓库提供 [docker-compose.example.yml](./docker-compose.example.yml) 作为启动模板。默认会挂载：

- `./config:/config`
- `./data:/data`

管理页面默认端口为 `8080`。示例还映射了常用 `mihomo` 代理端口：

| 宿主机 | 容器内 | 用途 |
| --- | --- | --- |
| `8080` | `8080` | Graydeck Web |
| `7890` | `17890` | mixed port |
| `7891` | `17891` | SOCKS |
| `7892` | `17892` | redir |
| `7893` | `17893` | tproxy |

示例中保留了 `NET_ADMIN` 和 `/dev/net/tun`。只使用普通 HTTP/SOCKS 代理时可以移除；需要 TUN 或透明代理时保留。

### Standalone

Release 提供 x86_64 单文件产物：

- `graydeck-linux-amd64`
- `graydeck-windows-amd64.exe`

单文件已经内置前端管理页。运行目录中的 `config/` 存放服务与基础运行配置，`data/` 存放核心、面板资源、订阅缓存和运行配置。

## 配置概览

`config/graydeck.yaml` 管理 Graydeck 自身：

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

`config/base.yaml` 会参与最终运行配置生成，适合放 Graydeck 需要统一注入的基础项：

- 代理端口与 `external-controller`
- `allow-lan`、`mode`、`log-level`
- `dns`、`tun` 等运行配置

对于 `dns` 和 `tun`，`base.yaml` 中声明的字段会覆盖订阅配置里的同字段；未声明的字段会保留订阅原值。列表字段在 `base.yaml` 中出现时，以 `base.yaml` 的列表为准。

## 开发

技术栈：

- Go 后端
- React + TypeScript + Vite 管理页
- 内置 `Zashboard` 静态资源集成

项目设计记录见 [DOCKER_IMAGE_DESIGN.md](./DOCKER_IMAGE_DESIGN.md)，前端样式规范见 [web/DESIGN.md](./web/DESIGN.md)。
