# APRS Thursday

This project is for one small job: send an APRS Thursday message once a week, then listen for replies and make a QSL style image for the messages that come in.

The service is meant to run beside an existing iGate. It has its own APRS-IS identity, Linux account, configuration, and systemd unit. It does not change or restart the other APRS services on the machine.

## How it behaves

At the configured time on Thursday, the service sends one message to `ANSRVR` in the `HOTG` group. It checks that the #APRSThursday net is open first. A dated message from `specialMessage.sh` can replace the regular message for one Thursday; the regular message is used again after that date.

The service also listens for direct messages to the operator and replies delivered through the `HOTG` group. Incoming messages do not cause an APRS reply or acknowledgement. Instead, each new message is saved and turned into a QSL style card with the sender, message, timestamp, and the configured image and greeting.

APRS cannot carry the card image. Card delivery is therefore a separate piece of the project and is still open for implementation. The wording on a card should say that an APRS message was received; it should not claim a confirmed radio QSO unless the operator has verified one.

## Current state

The repository has tested Go packages for message selection, schedule calculation, APRS message parsing, direct-message filtering, inbox deduplication, QSL card generation, and a receive-only service loop. The weekly scheduler and some operational wiring are still being built.

## Configuration and credentials

The service will use these settings:

| Setting | What it controls |
| --- | --- |
| `OPERATOR_CALLSIGN` | Callsign addressed by incoming messages and shown on cards |
| `APRS_CALLSIGN` | Unique callsign-SSID used by this APRS-IS client |
| `APRS_PASSCODE` | APRS-IS passcode used to send |
| `APRS_IS_SERVER` | APRS-IS host and port |
| `SCHEDULE_TIME` | Local Thursday send time in `HH:MM` format |
| `SCHEDULE_TIMEZONE` | IANA timezone for the schedule |
| `WEEKLY_MESSAGE_FILE` | File containing the regular message |
| `MESSAGE_OVERRIDES_DIR` | Directory containing dated one-time messages |

Use a callsign-SSID that is not already connected to APRS-IS by the iGate or another client. Keep the passcode in the protected environment file used by systemd. Do not put credentials in Git or in shell history.

## Installation

`install.sh` is intended for a Linux machine running systemd. It asks for the callsigns, APRS-IS connection details, schedule, timezone, and regular message. It creates a dedicated account and unit, enables the unit at boot, and starts it. It does not touch other services.

The installer builds `cmd/aprs-thursday`, creates the service account and systemd unit, and starts the receive-only daemon:

```sh
./install.sh
```

You can also pass a prebuilt executable:

```sh
./install.sh ./path/to/aprs-thursday
```

To inspect cards waiting for delivery, run the read-only status command. It uses `/var/lib/aprs-thursday/outbox.json` by default:

```sh
go run ./cmd/aprs-thursday-status
```

For a different outbox location, use `-outbox /path/to/outbox.json`.

The regular message is stored in `/etc/aprs-thursday/weekly-message.txt`. Edit that file with `sudoedit`; the service will read it immediately before the next send.

To prepare a one-time holiday message, run:

```sh
./specialMessage.sh
```

The script chooses the next Thursday according to the installed schedule and writes a dated override. The service uses that override throughout the seven-day period anchored by that Thursday, including a delayed send later in the period, and then goes back to the regular message when the next Thursday period begins.

## Development checks

The GitHub Actions workflow runs Bash syntax checks and ShellCheck. When the Go module is present, it also runs formatting checks, race-enabled tests, `go build`, `go vet`, and `staticcheck`.

Run the main checks locally with:

```sh
go test -race ./...
go vet ./...
go build ./...
bash -n install.sh specialMessage.sh
shellcheck -s bash install.sh specialMessage.sh
```

The APRS sandbox and fake QSL delivery target will be used for end-to-end tests. They must remain opt-in and disconnected from the live APRS-IS network. The first version keeps cards in the local outbox; it does not connect to email or callsign-directory services.

## References

- [APRS-IS connection guide](https://www.aprs-is.net/Connecting.aspx)
- [APRS-IS filter guide](https://www.aprs-is.net/javAPRSFilter.aspx)
- [#APRSThursday instructions](https://aprsph.net/aprsthursday/)
