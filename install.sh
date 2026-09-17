#!/usr/bin/env bash
set -euo pipefail

PROJECT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SERVICE_NAME="aprs-thursday"
INSTALL_DIR="/usr/local/bin"
CONFIG_DIR="/etc/${SERVICE_NAME}"
CONFIG_FILE="${CONFIG_DIR}/${SERVICE_NAME}.env"
MESSAGE_FILE="${CONFIG_DIR}/weekly-message.txt"
MESSAGE_OVERRIDES_DIR="${CONFIG_DIR}/message-overrides"
UNIT_FILE="/etc/systemd/system/${SERVICE_NAME}.service"
SEND_NOW_UNIT_FILE="/etc/systemd/system/${SERVICE_NAME}-send-now.service"
BUILD_DIR=""
STAGED_ENV=""
STAGED_UNIT=""
STAGED_SEND_NOW_UNIT=""
STAGED_MESSAGE=""

cleanup() {
  [[ -z "$BUILD_DIR" ]] || rm -rf "$BUILD_DIR"
  [[ -z "$STAGED_ENV" ]] || rm -f "$STAGED_ENV"
  [[ -z "$STAGED_UNIT" ]] || rm -f "$STAGED_UNIT"
  [[ -z "$STAGED_SEND_NOW_UNIT" ]] || rm -f "$STAGED_SEND_NOW_UNIT"
  [[ -z "$STAGED_MESSAGE" ]] || rm -f "$STAGED_MESSAGE"
}
trap cleanup EXIT

fail() {
  printf 'Error: %s\n' "$1" >&2
  exit 1
}

usage() {
  cat <<'EOF'
Usage: ./install.sh [path/to/aprs-thursday]

Builds the service from ./cmd/aprs-thursday when available, or installs the
provided executable. Installs a systemd unit on Linux and prompts for APRS-IS
credentials and scheduling settings. The service executable must read
OPERATOR_CALLSIGN, APRS_CALLSIGN, APRS_PASSCODE, APRS_IS_SERVER, SCHEDULE_TIME,
SCHEDULE_TIMEZONE, WEEKLY_MESSAGE_FILE, and MESSAGE_OVERRIDES_DIR from its
environment.
EOF
}

if [[ "${1:-}" == "-h" || "${1:-}" == "--help" ]]; then
  usage
  exit 0
fi

[[ "$(uname -s)" == "Linux" ]] || fail "This installer targets Linux with systemd."
command -v systemctl >/dev/null 2>&1 || fail "systemctl is required."
command -v sudo >/dev/null 2>&1 || fail "sudo is required to install the service."

if [[ $# -gt 1 ]]; then
  usage >&2
  exit 2
fi

SERVICE_BINARY=""
SEND_NOW_BINARY=""
STATUS_BINARY=""
if [[ $# -eq 1 ]]; then
  SERVICE_BINARY="$1"
  [[ "$SERVICE_BINARY" = /* ]] || SERVICE_BINARY="${PROJECT_DIR}/${SERVICE_BINARY}"
  [[ -f "$SERVICE_BINARY" && -x "$SERVICE_BINARY" ]] || fail "The supplied binary is missing or not executable: $SERVICE_BINARY"
  command -v go >/dev/null 2>&1 || fail "Go is required to build the manual send command."
  BUILD_DIR="$(mktemp -d)"
  SEND_NOW_BINARY="${BUILD_DIR}/${SERVICE_NAME}-send-now"
  STATUS_BINARY="${BUILD_DIR}/${SERVICE_NAME}-status"
  (cd "$PROJECT_DIR" && go build -o "$SEND_NOW_BINARY" ./cmd/aprs-thursday-send-now && go build -o "$STATUS_BINARY" ./cmd/aprs-thursday-status)
elif [[ -f "${PROJECT_DIR}/cmd/aprs-thursday/main.go" ]]; then
  command -v go >/dev/null 2>&1 || fail "Go is required to build the service."
  BUILD_DIR="$(mktemp -d)"
  SERVICE_BINARY="${BUILD_DIR}/${SERVICE_NAME}"
  SEND_NOW_BINARY="${BUILD_DIR}/${SERVICE_NAME}-send-now"
  STATUS_BINARY="${BUILD_DIR}/${SERVICE_NAME}-status"
  (cd "$PROJECT_DIR" && go build -o "$SERVICE_BINARY" ./cmd/aprs-thursday && go build -o "$SEND_NOW_BINARY" ./cmd/aprs-thursday-send-now && go build -o "$STATUS_BINARY" ./cmd/aprs-thursday-status)
else
  fail "The Go service is not implemented yet and no binary was supplied. Once available, add ./cmd/aprs-thursday or pass a built executable to this installer."
fi

printf 'APRS-IS credentials for the weekly message to ANSRVR\n'
read -r -p 'Operator callsign (used in the weekly message and cards): ' OPERATOR_CALLSIGN
OPERATOR_CALLSIGN="${OPERATOR_CALLSIGN^^}"
[[ "$OPERATOR_CALLSIGN" =~ ^[A-Z0-9]{3,7}$ ]] || fail "Enter your station callsign using 3 to 7 letters or digits."

read -r -p 'Unique APRS-IS client callsign-SSID (not the iGate identity): ' APRS_CALLSIGN
[[ -n "$APRS_CALLSIGN" ]] || fail "A unique callsign-SSID is required; do not reuse the iGate identity."
APRS_CALLSIGN="${APRS_CALLSIGN^^}"
[[ "$APRS_CALLSIGN" =~ ^[A-Z0-9]{1,6}(-[0-9]{1,2})?$ ]] || fail "Enter a valid APRS callsign-SSID, such as W1ABC or W1ABC-1."

read -r -s -p 'APRS-IS passcode (input hidden): ' APRS_PASSCODE
printf '\n'
[[ "$APRS_PASSCODE" =~ ^[0-9]+$ ]] || fail "A numeric APRS-IS passcode is required for sending; the receive-only passcode -1 cannot send."

read -r -p 'APRS-IS host and port: ' APRS_IS_SERVER
[[ "$APRS_IS_SERVER" =~ ^[A-Za-z0-9.-]+:[0-9]{1,5}$ ]] || fail "Enter an APRS-IS host and port, for example rotate.aprs2.net:14580."

read -r -p 'Thursday send time (24-hour HH:MM): ' SCHEDULE_TIME
[[ "$SCHEDULE_TIME" =~ ^([01][0-9]|2[0-3]):[0-5][0-9]$ ]] || fail "Enter the send time in 24-hour HH:MM format, such as 09:00."

read -r -p 'IANA timezone for the schedule (for example UTC): ' SCHEDULE_TIMEZONE
[[ "$SCHEDULE_TIMEZONE" =~ ^[A-Za-z0-9_+-]+(/[A-Za-z0-9_+-]+)*$ && -f "/usr/share/zoneinfo/${SCHEDULE_TIMEZONE}" ]] || fail "Enter a valid installed IANA timezone, such as UTC."

DEFAULT_WEEKLY_MESSAGE="CQ HOTG Happy #APRSTHURSDAY 73! ${OPERATOR_CALLSIGN}"
read -r -p "Weekly APRS message body [${DEFAULT_WEEKLY_MESSAGE}]: " WEEKLY_MESSAGE
WEEKLY_MESSAGE="${WEEKLY_MESSAGE:-$DEFAULT_WEEKLY_MESSAGE}"
WEEKLY_MESSAGE_UPPER="${WEEKLY_MESSAGE^^}"
[[ "$WEEKLY_MESSAGE_UPPER" == "CQ HOTG "* ]] || fail "The message must start with 'CQ HOTG ' to post to the APRSThursday group."
WEEKLY_MESSAGE_BYTES="$(LC_ALL=C printf '%s' "$WEEKLY_MESSAGE" | wc -c)"
WEEKLY_MESSAGE_BYTES="${WEEKLY_MESSAGE_BYTES//[[:space:]]/}"
[[ "$WEEKLY_MESSAGE_BYTES" =~ ^[0-9]+$ ]] || fail "Could not measure the weekly message."
(( WEEKLY_MESSAGE_BYTES <= 67 )) || fail "The APRS message body must be 67 bytes or fewer."

STAGED_ENV="$(mktemp)"
STAGED_UNIT="$(mktemp)"
STAGED_SEND_NOW_UNIT="$(mktemp)"
STAGED_MESSAGE="$(mktemp)"
chmod 600 "$STAGED_ENV"
chmod 600 "$STAGED_MESSAGE"
cat > "$STAGED_ENV" <<EOF
OPERATOR_CALLSIGN=${OPERATOR_CALLSIGN}
APRS_CALLSIGN=${APRS_CALLSIGN}
APRS_PASSCODE=${APRS_PASSCODE}
APRS_IS_SERVER=${APRS_IS_SERVER}
SCHEDULE_TIME=${SCHEDULE_TIME}
SCHEDULE_TIMEZONE=${SCHEDULE_TIMEZONE}
WEEKLY_MESSAGE_FILE=${MESSAGE_FILE}
MESSAGE_OVERRIDES_DIR=${MESSAGE_OVERRIDES_DIR}
EOF
printf '%s\n' "$WEEKLY_MESSAGE" > "$STAGED_MESSAGE"
unset APRS_PASSCODE

cat > "$STAGED_UNIT" <<EOF
[Unit]
Description=APRS Thursday weekly message and QSL card service
After=network-online.target
Wants=network-online.target

[Service]
Type=simple
User=${SERVICE_NAME}
Group=${SERVICE_NAME}
EnvironmentFile=${CONFIG_FILE}
ExecStart=${INSTALL_DIR}/${SERVICE_NAME}
Restart=on-failure
RestartSec=10s
StateDirectory=${SERVICE_NAME}
WorkingDirectory=/var/lib/${SERVICE_NAME}
NoNewPrivileges=true
PrivateTmp=true
ProtectSystem=strict
ProtectHome=true
ReadWritePaths=/var/lib/${SERVICE_NAME}

[Install]
WantedBy=multi-user.target
EOF

cat > "$STAGED_SEND_NOW_UNIT" <<EOF
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

printf 'Installing %s with sudo. The service will start after installation.\n' "$SERVICE_NAME"
sudo -v
if ! getent group "$SERVICE_NAME" >/dev/null; then
  sudo groupadd --system "$SERVICE_NAME"
fi
if ! id -u "$SERVICE_NAME" >/dev/null 2>&1; then
  sudo useradd --system --gid "$SERVICE_NAME" --home-dir "/var/lib/${SERVICE_NAME}" --no-create-home --shell /usr/sbin/nologin "$SERVICE_NAME"
fi
sudo install -d -o root -g "$SERVICE_NAME" -m 0750 "$CONFIG_DIR" "$MESSAGE_OVERRIDES_DIR"
sudo install -o root -g root -m 0755 "$SERVICE_BINARY" "${INSTALL_DIR}/${SERVICE_NAME}"
sudo install -o root -g root -m 0755 "$SEND_NOW_BINARY" "${INSTALL_DIR}/${SERVICE_NAME}-send-now"
sudo install -o root -g root -m 0755 "$STATUS_BINARY" "${INSTALL_DIR}/${SERVICE_NAME}-status"
sudo install -o root -g "$SERVICE_NAME" -m 0640 "$STAGED_ENV" "$CONFIG_FILE"
sudo install -o root -g "$SERVICE_NAME" -m 0640 "$STAGED_MESSAGE" "$MESSAGE_FILE"
sudo install -o root -g root -m 0644 "$STAGED_UNIT" "$UNIT_FILE"
sudo install -o root -g root -m 0644 "$STAGED_SEND_NOW_UNIT" "$SEND_NOW_UNIT_FILE"
sudo systemctl daemon-reload
sudo systemctl enable --now "${SERVICE_NAME}.service"

printf '\nInstalled and started %s.service.\n' "$SERVICE_NAME"
printf 'Check status with: sudo systemctl status %s.service\n' "$SERVICE_NAME"
printf 'View logs with: sudo journalctl -u %s.service -f\n' "$SERVICE_NAME"
