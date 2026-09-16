// Package config loads the daemon's environment-based configuration.
package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type Config struct {
	OperatorCallsign string
	APRSCallsign     string
	APRSPasscode     string
	APRSISServer     string
	WeeklyMessage    string
	MessageOverrides string
	InboxPath        string
	OutboxPath       string
	CardOutputDir    string
	CardGreeting     string
	CardLocation     string
	ScheduleTime     string
	ScheduleTimezone string
}

func FromEnv(getenv func(string) string) (Config, error) {
	if getenv == nil {
		getenv = os.Getenv
	}
	c := Config{
		OperatorCallsign: strings.ToUpper(strings.TrimSpace(getenv("OPERATOR_CALLSIGN"))),
		APRSCallsign:     strings.ToUpper(strings.TrimSpace(getenv("APRS_CALLSIGN"))),
		APRSPasscode:     getenv("APRS_PASSCODE"),
		APRSISServer:     strings.TrimSpace(getenv("APRS_IS_SERVER")),
		WeeklyMessage:    getenv("WEEKLY_MESSAGE_FILE"),
		MessageOverrides: getenv("MESSAGE_OVERRIDES_DIR"),
		InboxPath:        getenvOr(getenv, "INBOX_PATH", "/var/lib/aprs-thursday/inbox.json"),
		OutboxPath:       getenvOr(getenv, "OUTBOX_PATH", "/var/lib/aprs-thursday/outbox.json"),
		CardOutputDir:    getenvOr(getenv, "CARD_OUTPUT_DIR", "/var/lib/aprs-thursday/cards"),
		CardGreeting:     getenvOr(getenv, "CARD_GREETING", "Greetings from the station"),
		CardLocation:     getenvOr(getenv, "CARD_LOCATION", ""),
		ScheduleTime:     getenvOr(getenv, "SCHEDULE_TIME", "09:00"),
		ScheduleTimezone: getenvOr(getenv, "SCHEDULE_TIMEZONE", "UTC"),
	}
	missing := make([]string, 0)
	for key, value := range map[string]string{"OPERATOR_CALLSIGN": c.OperatorCallsign, "APRS_CALLSIGN": c.APRSCallsign, "APRS_PASSCODE": c.APRSPasscode, "APRS_IS_SERVER": c.APRSISServer} {
		if value == "" {
			missing = append(missing, key)
		}
	}
	if len(missing) != 0 {
		return Config{}, fmt.Errorf("missing required configuration: %s", strings.Join(missing, ", "))
	}
	return c, nil
}

func (c Config) Schedule() (int, int, *time.Location, error) {
	parsed, err := time.Parse("15:04", c.ScheduleTime)
	if err != nil {
		return 0, 0, nil, fmt.Errorf("invalid SCHEDULE_TIME: %w", err)
	}
	location, err := time.LoadLocation(c.ScheduleTimezone)
	if err != nil {
		return 0, 0, nil, fmt.Errorf("invalid SCHEDULE_TIMEZONE: %w", err)
	}
	return parsed.Hour(), parsed.Minute(), location, nil
}

func getenvOr(getenv func(string) string, key, fallback string) string {
	if value := strings.TrimSpace(getenv(key)); value != "" {
		return value
	}
	return fallback
}

func (c Config) EnsureDirectories() error {
	for _, path := range []string{c.InboxPath, c.OutboxPath, c.CardOutputDir} {
		if err := os.MkdirAll(filepath.Dir(path), 0o750); err != nil {
			return fmt.Errorf("create state directory: %w", err)
		}
	}
	return nil
}
