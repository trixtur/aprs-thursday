package receive

import (
	"fmt"

	"github.com/trixtur/aprs-thursday/internal/aprs"
	"github.com/trixtur/aprs-thursday/internal/card"
	"github.com/trixtur/aprs-thursday/internal/inbox"
	"github.com/trixtur/aprs-thursday/internal/outbox"
)

// Pipeline turns accepted direct messages into cards and sends them through a
// delivery implementation. Group traffic is rejected before persistence.
type Pipeline struct {
	Operator string
	Inbox    *inbox.Store
	Cards    card.Config
	Delivery card.Delivery
	Outbox   *outbox.Store
	Notify   Notifier
}

type Notifier interface {
	CardPending(path string, message aprs.Message) error
}

// Handle processes one parsed packet. The bool reports whether a new card was
// created; duplicates and group traffic return false without an error.
func (p *Pipeline) Handle(message aprs.Message) (bool, error) {
	if p == nil || p.Inbox == nil {
		return false, fmt.Errorf("receive pipeline inbox is required")
	}
	if p.Outbox == nil {
		return false, fmt.Errorf("receive pipeline outbox is required")
	}
	if !DirectForOperator(message, p.Operator) {
		return false, nil
	}
	added, err := p.Inbox.Add(message)
	if err != nil {
		return false, err
	}
	record, exists := p.Inbox.Get(message.ID)
	if !exists {
		return false, fmt.Errorf("message disappeared from inbox")
	}
	if !added && record.Delivered {
		return false, nil
	}
	path, err := card.Render(p.Cards, record.Message)
	if err != nil {
		return false, err
	}
	entry, added, err := p.Outbox.Add(record.Message, path)
	if err != nil {
		return false, err
	}
	if !added && entry.Status == outbox.Sent {
		return false, nil
	}
	if added && p.Notify != nil {
		if err := p.Notify.CardPending(path, record.Message); err != nil {
			return false, err
		}
	}
	// With no delivery implementation configured, keep the rendered card in
	// the pending outbox for the operator. A local archive is not a delivery.
	if p.Delivery == nil {
		return added, nil
	}
	if err := p.Delivery.Deliver(path, record.Message); err != nil {
		_ = p.Outbox.MarkFailed(record.Message.ID, err.Error())
		return false, fmt.Errorf("deliver QSL card: %w", err)
	}
	if err := p.Outbox.MarkSent(record.Message.ID); err != nil {
		return false, err
	}
	if err := p.Inbox.MarkDelivered(message.ID); err != nil {
		return false, err
	}
	return true, nil
}
