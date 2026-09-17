#!/usr/bin/env bash
set -euo pipefail

PROJECT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SERVICE_NAME="aprs-thursday"
INSTALL_DIR="/usr/local/bin"
CONFIG_FILE="/etc/${SERVICE_NAME}/${SERVICE_NAME}.env"
SEND_NOW_UNIT_FILE="/etc/systemd/system/${SERVICE_NAME}-send-now.service"
BUILD_DIR=""
STAGED_UNIT=""

cleanup() {
  [[ -z "$BUILD_DIR" ]] || rm -rf "$BUILD_DIR"
  [[ -z "$STAGED_UNIT" ]] || rm -f "$STAGED_UNIT"
}
trap cleanup EXIT

fail() {
  printf 'Error: %s\n' "$1" >&2
  exit 1
}

[[ "$(uname -s)" == "Linux" ]] || fail "This script targets Linux with systemd."
command -v go >/dev/null 2>&1 || fail "Go is required to rebuild the service."
command -v sudo >/dev/null 2>&1 || fail "sudo is required to reinstall the service."
command -v systemctl >/dev/null 2>&1 || fail "systemctl is required."
[[ -f "$CONFIG_FILE" ]] || fail "${CONFIG_FILE} is missing; use install.sh for a first installation."

BUILD_DIR="$(mktemp -d)"
(cd "$PROJECT_DIR" && go build -o "${BUILD_DIR}/${SERVICE_NAME}" ./cmd/aprs-thursday && go build -o "${BUILD_DIR}/${SERVICE_NAME}-send-now" ./cmd/aprs-thursday-send-now && go build -o "${BUILD_DIR}/${SERVICE_NAME}-status" ./cmd/aprs-thursday-status)

STAGED_UNIT="$(mktemp)"
cat > "$STAGED_UNIT" <<EOF
[Unit]
Description=Send one APRS Thursday message immediately
After=network-online.target
Wants=network-online.target

[Service]
Type=oneshot
User=${SERVICE_NAME}
Group=${SERVICE_NAME}
EnvironmentFile=${CONFIG_FILE}
ExecStart=${INSTALL_DIR}/${SERVICE_NAME}-send-now --confirm
WorkingDirectory=/var/lib/${SERVICE_NAME}
NoNewPrivileges=true
PrivateTmp=true
ProtectSystem=strict
ProtectHome=true
ReadWritePaths=/var/lib/${SERVICE_NAME}
EOF

printf 'Rebuilding and restarting %s without changing its configuration or state.\n' "$SERVICE_NAME"
sudo -v
sudo install -o root -g root -m 0755 "${BUILD_DIR}/${SERVICE_NAME}" "${INSTALL_DIR}/${SERVICE_NAME}"
sudo install -o root -g root -m 0755 "${BUILD_DIR}/${SERVICE_NAME}-send-now" "${INSTALL_DIR}/${SERVICE_NAME}-send-now"
sudo install -o root -g root -m 0755 "${BUILD_DIR}/${SERVICE_NAME}-status" "${INSTALL_DIR}/${SERVICE_NAME}-status"
sudo install -o root -g root -m 0644 "$STAGED_UNIT" "$SEND_NOW_UNIT_FILE"
sudo systemctl daemon-reload
sudo systemctl restart "${SERVICE_NAME}.service"

printf 'Reinstalled and restarted %s.service.\n' "$SERVICE_NAME"
printf 'Manual send: sudo systemctl start %s-send-now.service\n' "$SERVICE_NAME"
