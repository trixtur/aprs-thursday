#!/usr/bin/env bash
set -euo pipefail

SERVICE_NAME="aprs-thursday"
CONFIG_DIR="/etc/${SERVICE_NAME}"
CONFIG_FILE="${CONFIG_DIR}/${SERVICE_NAME}.env"
OVERRIDES_DIR="${CONFIG_DIR}/message-overrides"

fail() {
  printf 'Error: %s\n' "$1" >&2
  exit 1
}

read_config_value() {
  local key="$1"
  local value
  value="$(sed -n "s/^${key}=//p" "$CONFIG_FILE" | head -n 1)"
  [[ -n "$value" ]] || fail "Missing ${key} in ${CONFIG_FILE}. Reinstall or correct the service configuration."
  printf '%s' "$value"
}

[[ "$(uname -s)" == "Linux" ]] || fail "This script targets the Linux system running the service."
command -v sudo >/dev/null 2>&1 || fail "sudo is required to install the dated message override."
command -v systemctl >/dev/null 2>&1 || fail "systemctl is required. Install the service before setting a special message."
[[ -r "$CONFIG_FILE" ]] || fail "Cannot read ${CONFIG_FILE}. Install the service first."

SCHEDULE_TIME="$(read_config_value SCHEDULE_TIME)"
SCHEDULE_TIMEZONE="$(read_config_value SCHEDULE_TIMEZONE)"
[[ "$SCHEDULE_TIME" =~ ^([01][0-9]|2[0-3]):[0-5][0-9]$ ]] || fail "Invalid SCHEDULE_TIME in ${CONFIG_FILE}."
[[ -f "/usr/share/zoneinfo/${SCHEDULE_TIMEZONE}" ]] || fail "Timezone ${SCHEDULE_TIMEZONE} is not installed on this machine."

read -r -p 'Special APRS message body (must begin with CQ HOTG): ' SPECIAL_MESSAGE
SPECIAL_MESSAGE_UPPER="${SPECIAL_MESSAGE^^}"
[[ "$SPECIAL_MESSAGE_UPPER" == "CQ HOTG "* ]] || fail "The message must start with 'CQ HOTG ' to post to the APRSThursday group."
SPECIAL_MESSAGE_BYTES="$(LC_ALL=C printf '%s' "$SPECIAL_MESSAGE" | wc -c)"
SPECIAL_MESSAGE_BYTES="${SPECIAL_MESSAGE_BYTES//[[:space:]]/}"
[[ "$SPECIAL_MESSAGE_BYTES" =~ ^[0-9]+$ ]] || fail "Could not measure the special message."
(( SPECIAL_MESSAGE_BYTES <= 67 )) || fail "The APRS message body must be 67 bytes or fewer."

LOCAL_WEEKDAY="$(TZ="$SCHEDULE_TIMEZONE" date +%u)"
LOCAL_DATE="$(TZ="$SCHEDULE_TIMEZONE" date +%F)"
LOCAL_HHMM="$(TZ="$SCHEDULE_TIMEZONE" date +%H%M)"
SCHEDULE_HHMM="${SCHEDULE_TIME//:/}"
DAYS_UNTIL_THURSDAY=$(((4 - LOCAL_WEEKDAY + 7) % 7))
if (( DAYS_UNTIL_THURSDAY == 0 )) && [[ "$LOCAL_HHMM" > "$SCHEDULE_HHMM" || "$LOCAL_HHMM" == "$SCHEDULE_HHMM" ]]; then
  DAYS_UNTIL_THURSDAY=7
fi

if (( DAYS_UNTIL_THURSDAY == 0 )); then
  TARGET_DATE="$LOCAL_DATE"
else
  TARGET_DATE="$(TZ="$SCHEDULE_TIMEZONE" date --date="${LOCAL_DATE} +${DAYS_UNTIL_THURSDAY} days" +%F)"
fi

TEMP_MESSAGE="$(mktemp)"
trap 'rm -f "$TEMP_MESSAGE"' EXIT
chmod 600 "$TEMP_MESSAGE"
printf '%s\n' "$SPECIAL_MESSAGE" > "$TEMP_MESSAGE"

sudo -v
sudo install -d -o root -g "$SERVICE_NAME" -m 0750 "$OVERRIDES_DIR"
sudo install -o root -g "$SERVICE_NAME" -m 0640 "$TEMP_MESSAGE" "${OVERRIDES_DIR}/${TARGET_DATE}.txt"

printf 'Special message set for %s (%s %s).\n' "$TARGET_DATE" "$SCHEDULE_TIME" "$SCHEDULE_TIMEZONE"
printf 'The regular weekly message resumes automatically the following Thursday.\n'
