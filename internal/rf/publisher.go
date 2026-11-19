// Package rf transmits ANSRVR messages as Bell 202 audio through ALSA.
package rf

import (
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"math"
	"os/exec"
	"path/filepath"

	"github.com/gofrs/flock"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const sampleRate = 22050

var address = regexp.MustCompile("^([A-Z0-9]{1,6})(-([0-9]{1,2}))?$")

type Publisher struct {
	header   []byte
	device   string
	aplay    string
	lockFile string
	lead     time.Duration
}

func New(call, path, device, lockFile string, lead time.Duration) (*Publisher, error) {
	if lead < 0 || lead > 5*time.Second {
		return nil, fmt.Errorf("RF_TX_DELAY must be between 0s and 5s")
	}
	calls := []string{"APRS", call}
	if path != "" {
		calls = append(calls, strings.Split(path, ",")...)
	}
	if len(calls) > 10 {
		return nil, fmt.Errorf("RF_PATH exceeds eight digipeaters")
	}
	var header []byte
	for i, c := range calls {
		parts := address.FindStringSubmatch(strings.ToUpper(strings.TrimSpace(c)))
		if parts == nil {
			return nil, fmt.Errorf("invalid RF address %q", c)
		}
		ssid, _ := strconv.Atoi(parts[3])
		if ssid > 15 {
			return nil, fmt.Errorf("invalid RF SSID in %q", c)
		}
		for _, ch := range fmt.Sprintf("%-6s", parts[1]) {
			header = append(header, byte(ch)<<1)
		}
		flags := byte(0x60 | ssid<<1)
		if i == 0 {
			flags |= 0x80
		} // UI command
		if i == len(calls)-1 {
			flags |= 1
		}
		header = append(header, flags)
	}
	if !filepath.IsAbs(lockFile) {
		return nil, fmt.Errorf("RF_LOCK_FILE requires an absolute path")
	}
	aplay, err := exec.LookPath("aplay")
	if err != nil {
		return nil, fmt.Errorf("RF mode requires aplay: %w", err)
	}
	return &Publisher{header: append(header, 0x03, 0xf0), device: device, lead: lead, aplay: aplay, lockFile: lockFile}, nil
}

func (p *Publisher) Publish(body string) error {
	if len(body) == 0 || len(body) > 67 || strings.ContainsAny(body, "\r\n") {
		return fmt.Errorf("invalid APRS message body")
	}
	frame := append([]byte(nil), p.header...)
	frame = append(frame, ":ANSRVR   :"...)
	frame = append(frame, body...)
	pcm := audio(frame, p.lead)
	// Use a separate descriptor for every send, including concurrent callers.
	lock := flock.New(p.lockFile)
	wait, stopWaiting := context.WithTimeout(context.Background(), 30*time.Second)
	defer stopWaiting()
	acquired, err := lock.TryLockContext(wait, 50*time.Millisecond)
	if err != nil {
		return fmt.Errorf("acquire RF playback lock: %w", err)
	}
	if !acquired {
		return fmt.Errorf("RF playback lock unavailable")
	}
	defer lock.Unlock()
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, p.aplay, "-q", "-D", p.device, "-t", "raw", "-f", "S16_LE", "-c", "1", "-r", "22050")
	cmd.Stdin = bytes.NewReader(pcm)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("RF playback: %w: %s", err, strings.TrimSpace(string(out)))
	}
	return nil
}

func audio(frame []byte, lead time.Duration) []byte {
	// AX.25 FCS (CRC-16/X-25), low byte first.
	crc := uint16(0xffff)
	for _, b := range frame {
		crc ^= uint16(b)
		for range 8 {
			if crc&1 != 0 {
				crc = crc>>1 ^ 0x8408
			} else {
				crc >>= 1
			}
		}
	}
	frame = binary.LittleEndian.AppendUint16(frame, ^crc)
	var pcm []byte
	phase := 0.0
	mark := true
	remainder := 0
	// NRZI changes tone on a zero. Carry the fractional samples and phase across bits.
	emit := func(one bool) {
		if !one {
			mark = !mark
		}
		hz := 2200.0
		if mark {
			hz = 1200
		}
		step := 2 * math.Pi * hz / sampleRate
		remainder += sampleRate
		for ; remainder >= 1200; remainder -= 1200 {
			pcm = binary.LittleEndian.AppendUint16(pcm, uint16(int16(math.Sin(phase)*16000)))
			phase = math.Mod(phase+step, 2*math.Pi)
		}
	}
	flag := func() {
		for i := 0; i < 8; i++ {
			emit(0x7e&(1<<i) != 0)
		}
	}
	for range 1 + int(math.Ceil(lead.Seconds()*1200/8)) {
		flag()
	}
	ones := 0
	for _, b := range frame {
		for i := 0; i < 8; i++ {
			one := b&(1<<i) != 0
			emit(one)
			if one {
				ones++
			} else {
				ones = 0
			}
			if ones == 5 {
				emit(false)
				ones = 0
			}
		}
	}
	for range 4 {
		flag()
	}
	return pcm
}
