// Package service contains the long-running APRS service loops.
package service

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/trixtur/aprs-thursday/internal/aprs"
	"github.com/trixtur/aprs-thursday/internal/observe"
	"github.com/trixtur/aprs-thursday/internal/receive"
)

type LineSource interface {
	Receive() (string, error)
	Close() error
}

// ReceiveOnly reads packets until cancellation or stream closure. The APRS
// client remains responsible for login and reconnect; this loop is deliberately
// independent of outbound weekly scheduling.
func ReceiveOnly(ctx context.Context, source LineSource, operator string, pipeline *receive.Pipeline, logger *observe.Logger) error {
	if source == nil || pipeline == nil {
		return fmt.Errorf("receive source and pipeline are required")
	}
	defer source.Close()
	if logger != nil {
		_ = logger.Event("aprs.receive_started", nil)
	}
	for {
		select {
		case <-ctx.Done():
			if logger != nil {
				_ = logger.Event("aprs.receive_stopped", map[string]any{"reason": "context canceled"})
			}
			return ctx.Err()
		default:
		}
		line, err := source.Receive()
		if err != nil {
			if err == io.EOF {
				return io.EOF
			}
			return fmt.Errorf("receive APRS-IS packet: %w", err)
		}
		message, err := aprs.ParseMessage(line, time.Now().UTC())
		if err != nil {
			if logger != nil {
				_ = logger.Event("aprs.packet_ignored", map[string]any{"reason": err.Error()})
			}
			continue
		}
		if !receive.DirectForOperator(message, operator) {
			if logger != nil {
				_ = logger.Event("aprs.message_ignored", map[string]any{"sender": message.From, "group": message.IsHOTGReply()})
			}
			continue
		}
		created, err := pipeline.Handle(message)
		if err != nil {
			return err
		}
		if logger != nil {
			kind := "aprs.message_duplicate"
			if created {
				kind = "aprs.card_delivered"
			}
			_ = logger.Event(kind, map[string]any{"sender": message.From, "message_id": message.ID})
		}
	}
}
