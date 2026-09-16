// Package receive selects direct inbound messages for card generation.
package receive

import "github.com/trixtur/aprs-thursday/internal/aprs"

// DirectForOperator accepts only a direct APRS message to operator. Group
// traffic, including ANSRVR HOTG deliveries, is intentionally excluded.
func DirectForOperator(message aprs.Message, operator string) bool {
	return !message.IsHOTGReply() && message.IsFor(operator)
}
