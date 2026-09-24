#!/usr/bin/env bash
set -euo pipefail

SERVICE_NAME="aprs-thursday"
CONFIG_FILE="/etc/${SERVICE_NAME}/${SERVICE_NAME}.env"

fail() {
  printf 'Error: %s\n' "$1" >&2
  exit 1
}

if [[ ${EUID} -ne 0 ]]; then
  exec sudo bash "$0" "$@"
fi

FORCE=false
if [[ "${1:-}" == "--force" ]]; then
  FORCE=true
elif [[ $# -ne 0 ]]; then
  printf 'Usage: %s [--force]\n' "$0" >&2
  exit 2
fi

[[ -f "$CONFIG_FILE" ]] || fail "${CONFIG_FILE} is missing; install the service first."
command -v curl >/dev/null 2>&1 || fail "curl is required."

# The file is root-owned and contains the APRS-IS passcode.
# shellcheck disable=SC1090
source "$CONFIG_FILE"
: "${APRS_CALLSIGN:?APRS_CALLSIGN is missing from ${CONFIG_FILE}}"
: "${APRS_PASSCODE:?APRS_PASSCODE is missing from ${CONFIG_FILE}}"
: "${APRS_IS_SERVER:?APRS_IS_SERVER is missing from ${CONFIG_FILE}}"
[[ "$APRS_CALLSIGN" =~ ^[A-Za-z0-9]{1,6}(-[0-9]{1,2})?$ ]] || fail "Invalid APRS_CALLSIGN in ${CONFIG_FILE}."
[[ "$APRS_PASSCODE" =~ ^[0-9]+$ ]] || fail "Invalid APRS_PASSCODE in ${CONFIG_FILE}."
[[ "$APRS_IS_SERVER" =~ ^[A-Za-z0-9.-]+:[0-9]{1,5}$ ]] || fail "Invalid APRS_IS_SERVER in ${CONFIG_FILE}."

SCHEDULE_TIMEZONE="${SCHEDULE_TIMEZONE:-UTC}"
if [[ "$FORCE" != true ]] && [[ "$(TZ="$SCHEDULE_TIMEZONE" date +%u)" != 4 ]]; then
  fail "This script is restricted to Thursdays; use --force for a deliberate test."
fi

{
  printf 'user %s pass %s vers aprs-thursday-signout 1\r\n' "$APRS_CALLSIGN" "$APRS_PASSCODE"
  printf '%s>APRS,TCPIP*::%-9s:U HOTG\r\n' "$APRS_CALLSIGN" "ANSRVR"
} | curl --fail --silent --show-error --no-buffer --connect-timeout 10 --max-time 15 "telnet://${APRS_IS_SERVER}"

printf 'Sent ANSRVR sign-out request for %s from HOTG.\n' "$APRS_CALLSIGN"
