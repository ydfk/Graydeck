#!/usr/bin/env bash

set -euo pipefail

SERVICE_NAME="graydeck"
UNIT_PATH="/etc/systemd/system/${SERVICE_NAME}.service"
SOURCE_BINARY="${1:-}"
BINARY_PATH=""
WORK_DIR=""
SERVICE_TIMEZONE=""

main() {
  require_root
  require_command systemctl
  require_binary
  detect_timezone

  install_unit
  start_service
  print_result
}

require_root() {
  if [[ "${EUID}" -ne 0 ]]; then
    echo "请使用 root 权限执行，例如：sudo bash $0 <graydeck-binary>" >&2
    exit 1
  fi
}

require_command() {
  if ! command -v "$1" >/dev/null 2>&1; then
    echo "缺少命令：$1" >&2
    exit 1
  fi
}

require_binary() {
  if [[ -z "${SOURCE_BINARY}" ]]; then
    echo "用法：sudo bash $0 <graydeck-binary>" >&2
    exit 1
  fi

  if [[ ! -f "${SOURCE_BINARY}" ]]; then
    echo "找不到 Graydeck 可执行文件：${SOURCE_BINARY}" >&2
    exit 1
  fi

  BINARY_PATH="$(readlink -f "${SOURCE_BINARY}")"
  WORK_DIR="$(dirname "${BINARY_PATH}")"
}

detect_timezone() {
  if command -v timedatectl >/dev/null 2>&1; then
    SERVICE_TIMEZONE="$(timedatectl show -p Timezone --value 2>/dev/null || true)"
  fi

  if [[ -z "${SERVICE_TIMEZONE}" && -f /etc/timezone ]]; then
    SERVICE_TIMEZONE="$(tr -d '[:space:]' </etc/timezone)"
  fi

  if [[ -z "${SERVICE_TIMEZONE}" ]]; then
    SERVICE_TIMEZONE="UTC"
  fi
}

install_unit() {
  cat >"${UNIT_PATH}" <<EOF
[Unit]
Description=Graydeck mihomo manager
After=network-online.target
Wants=network-online.target

[Service]
Type=simple
WorkingDirectory=${WORK_DIR}
ExecStart=${BINARY_PATH}
Environment=TZ=${SERVICE_TIMEZONE}
Environment=GRAYDECK_TIMEZONE=${SERVICE_TIMEZONE}
Restart=on-failure
RestartSec=3s

[Install]
WantedBy=multi-user.target
EOF
}

start_service() {
  systemctl daemon-reload
  systemctl enable "${SERVICE_NAME}.service"
  systemctl restart "${SERVICE_NAME}.service"
}

print_result() {
  echo "graydeck.service 已写入并启动。"
  echo "服务状态：systemctl status ${SERVICE_NAME}.service"
  echo "实时日志：journalctl -u ${SERVICE_NAME}.service -f"
}

main "$@"
