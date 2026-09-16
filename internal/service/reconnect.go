package service

import (
	"context"
	"fmt"
	"time"

	"github.com/trixtur/aprs-thursday/internal/observe"
	"github.com/trixtur/aprs-thursday/internal/receive"
)

type SourceFactory func() (LineSource, error)

// ReceiveWithReconnect keeps the receive side alive across connection drops.
// Outbound scheduling is intentionally outside this loop, so reconnecting
// cannot trigger another weekly APRS transmission.
func ReceiveWithReconnect(ctx context.Context, factory SourceFactory, operator string, pipeline *receive.Pipeline, logger *observe.Logger, backoff time.Duration) error {
	if factory == nil {
		return fmt.Errorf("source factory is required")
	}
	if backoff <= 0 {
		backoff = time.Second
	}
	for {
		if err := ctx.Err(); err != nil {
			return err
		}
		source, err := factory()
		if err != nil {
			if logger != nil {
				_ = logger.Event("aprs.connect_failed", map[string]any{"error": err.Error()})
			}
			if err := wait(ctx, backoff); err != nil {
				return err
			}
			continue
		}
		err = ReceiveOnly(ctx, source, operator, pipeline, logger)
		if err == nil || err == context.Canceled || err == context.DeadlineExceeded || ctx.Err() != nil {
			return err
		}
		if logger != nil {
			_ = logger.Event("aprs.reconnect", map[string]any{"error": err.Error()})
		}
		if err := wait(ctx, backoff); err != nil {
			return err
		}
	}
}

func wait(ctx context.Context, duration time.Duration) error {
	timer := time.NewTimer(duration)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}
