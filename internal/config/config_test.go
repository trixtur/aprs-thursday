package config_test

import (
	"testing"

	"github.com/trixtur/aprs-thursday/internal/config"
)

func TestFromEnvLoadsRequiredAndDefaultPaths(t *testing.T) {
	values := map[string]string{"OPERATOR_CALLSIGN": "n0call", "APRS_CALLSIGN": "n0call-1", "APRS_PASSCODE": "12345", "APRS_IS_SERVER": "rotate.aprs2.net:14580"}
	c, err := config.FromEnv(func(key string) string { return values[key] })
	if err != nil {
		t.Fatal(err)
	}
	if c.OperatorCallsign != "N0CALL" || c.APRSCallsign != "N0CALL-1" || c.OutboxPath == "" {
		t.Fatalf("config = %#v", c)
	}
}

func TestFromEnvRejectsMissingCredentials(t *testing.T) {
	if _, err := config.FromEnv(func(string) string { return "" }); err == nil {
		t.Fatal("missing configuration was accepted")
	}
}
