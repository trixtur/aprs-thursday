package rf

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestRFAddresses(t *testing.T) {
	for _, path := range []string{"WIDE1-16", "WIDE1-1,", "TOOLONG", "WIDE1-1*"} {
		if _, err := New("N7RIX-7", path, "default", time.Second); err == nil {
			t.Errorf("accepted invalid path %q", path)
		}
	}
}

func TestPublishAudio(t *testing.T) {
	decoder, _ := exec.LookPath("multimon-ng")
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "aplay"), []byte("#!/bin/sh\nprintf '%s\\n' \"$@\" > \"$CAPTURE.args\"\ncat > \"$CAPTURE\"\n"), 0700); err != nil {
		t.Fatal(err)
	}
	capture := filepath.Join(dir, "tx.pcm")
	t.Setenv("CAPTURE", capture)
	t.Setenv("PATH", dir+":"+os.Getenv("PATH"))
	p, err := New("N7RIX-7", "WIDE1-1,WIDE2-1", "plughw:0,0", 500*time.Millisecond)
	if err != nil {
		t.Fatal(err)
	}
	if err = p.Publish("CQ HOTG RF check"); err != nil {
		t.Fatal(err)
	}
	args, err := os.ReadFile(capture + ".args")
	if err != nil || strings.Join(strings.Fields(string(args)), " ") != "-q -D plughw:0,0 -t raw -f S16_LE -c 1 -r 22050" {
		t.Fatalf("aplay args %q, %v", args, err)
	}
	pcm, err := os.ReadFile(capture)
	if err != nil || len(pcm) <= sampleRate {
		t.Fatalf("PCM length %d, %v", len(pcm), err)
	}
	if decoder == "" {
		t.Log("multimon-ng unavailable; skipping independent decode")
		return
	}
	cmd := exec.Command(decoder, "-q", "-A", "-a", "AFSK1200", "-t", "raw", "-")
	cmd.Stdin = bytes.NewReader(pcm)
	out, err := cmd.CombinedOutput()
	if err != nil || !strings.Contains(string(out), "N7RIX-7>APRS,WIDE1-1,WIDE2-1::ANSRVR   :CQ HOTG RF check") {
		t.Fatalf("decode: %v %s", err, out)
	}
}
