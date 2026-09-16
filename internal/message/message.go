// Package message resolves and validates the body for a scheduled APRS post.
package message

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const maxBodyBytes = 67

// Resolve returns the special message for the seven-day period anchored by the
// designated Thursday when one exists, otherwise the regular message. An
// override is not deleted, so retries during that period are stable.
func Resolve(regular, overridesDir string, scheduledAt time.Time) (string, error) {
	if overridesDir == "" {
		return "", fmt.Errorf("message override directory is required")
	}

	periodStart := scheduledAt
	daysSinceThursday := (int(scheduledAt.Weekday()) - int(time.Thursday) + 7) % 7
	periodStart = periodStart.AddDate(0, 0, -daysSinceThursday)
	path := filepath.Join(overridesDir, periodStart.Format("2006-01-02")+".txt")
	contents, err := os.ReadFile(path)
	switch {
	case err == nil:
		return validate(strings.TrimSuffix(strings.TrimSuffix(string(contents), "\n"), "\r"))
	case !os.IsNotExist(err):
		return "", fmt.Errorf("read dated message override: %w", err)
	default:
		return validate(regular)
	}
}

func validate(body string) (string, error) {
	if !strings.HasPrefix(strings.ToUpper(body), "CQ HOTG ") {
		return "", fmt.Errorf("message must begin with %q", "CQ HOTG ")
	}
	if len([]byte(body)) > maxBodyBytes {
		return "", fmt.Errorf("message is %d bytes; APRS message bodies are limited to %d bytes", len([]byte(body)), maxBodyBytes)
	}
	return body, nil
}
