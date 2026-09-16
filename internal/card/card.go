// Package card creates a small, portable QSL card artifact.
package card

import (
	"fmt"
	"html"
	"os"
	"path/filepath"
	"strings"

	"github.com/trixtur/aprs-thursday/internal/aprs"
)

type Config struct {
	OperatorCallsign string
	Greeting         string
	Location         string
	PhotoURL         string
	OutputDir        string
}

type Delivery interface {
	Deliver(path string, message aprs.Message) error
}

func Render(cfg Config, message aprs.Message) (string, error) {
	if strings.TrimSpace(cfg.OutputDir) == "" {
		return "", fmt.Errorf("card output directory is required")
	}
	if strings.TrimSpace(cfg.OperatorCallsign) == "" {
		return "", fmt.Errorf("operator callsign is required")
	}
	if strings.TrimSpace(cfg.Greeting) == "" {
		return "", fmt.Errorf("card greeting is required")
	}
	if message.From == "" || message.Text == "" {
		return "", fmt.Errorf("card message sender and text are required")
	}
	if err := os.MkdirAll(cfg.OutputDir, 0o750); err != nil {
		return "", fmt.Errorf("create card output directory: %w", err)
	}
	name := filepath.Join(cfg.OutputDir, message.ID+".svg")
	photo := ""
	if cfg.PhotoURL != "" {
		photo = fmt.Sprintf(`<image href="%s" x="20" y="20" width="760" height="180" preserveAspectRatio="xMidYMid slice"/>`, html.EscapeString(cfg.PhotoURL))
	}
	contents := fmt.Sprintf(`<svg xmlns="http://www.w3.org/2000/svg" width="800" height="600" viewBox="0 0 800 600"><rect width="800" height="600" fill="#f5efe3"/>%s<rect x="20" y="220" width="760" height="350" rx="12" fill="#ffffff" fill-opacity="0.94"/><text x="40" y="262" font-family="sans-serif" font-size="28" font-weight="bold">APRS MESSAGE RECEIPT</text><text x="40" y="295" font-family="sans-serif" font-size="16">QSL-style record — receipt of an APRS message, not a confirmed QSO</text><line x1="40" y1="315" x2="760" y2="315" stroke="#777"/><text x="40" y="350" font-family="sans-serif" font-size="20">FROM: %s</text><text x="420" y="350" font-family="sans-serif" font-size="20">TO: %s</text><text x="40" y="382" font-family="sans-serif" font-size="18">DATE (UTC): %s</text><text x="420" y="382" font-family="sans-serif" font-size="18">TIME (UTC): %s</text><text x="40" y="414" font-family="sans-serif" font-size="18">MODE: APRS messaging</text><text x="420" y="414" font-family="sans-serif" font-size="18">BAND/FREQ: N/A</text><text x="40" y="446" font-family="sans-serif" font-size="18">RST: N/A</text><text x="420" y="446" font-family="sans-serif" font-size="18">QTH: %s</text><text x="40" y="485" font-family="sans-serif" font-size="18">MESSAGE:</text><text x="40" y="515" font-family="sans-serif" font-size="18">%s</text><text x="40" y="550" font-family="sans-serif" font-size="20">%s</text></svg>`, photo, html.EscapeString(message.From), html.EscapeString(cfg.OperatorCallsign), html.EscapeString(message.Received.UTC().Format("2006-01-02")), html.EscapeString(message.Received.UTC().Format("15:04:05")), html.EscapeString(cfg.Location), html.EscapeString(message.Text), html.EscapeString(cfg.Greeting))
	if err := os.WriteFile(name, []byte(contents), 0o640); err != nil {
		return "", fmt.Errorf("write card: %w", err)
	}
	return name, nil
}
