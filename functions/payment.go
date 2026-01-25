package functions

import (
	"context"

	"github.com/gotd/td/tg"
)

// ExportInvoice exports an invoice for use with payment providers.
//
// Parameters:
//   - ctx: Context for the API call
//   - raw: The raw Telegram client
//   - inputMedia: The invoice media to export
//
// Returns exported invoice or an error.
func ExportInvoice(ctx context.Context, raw *tg.Client, inputMedia tg.InputMediaClass) (*tg.PaymentsExportedInvoice, error) {
	return raw.PaymentsExportInvoice(ctx, inputMedia)
}

// SetPreCheckoutResults sets pre-checkout query results for a bot.
//
// Parameters:
//   - ctx: Context for the API call
//   - raw: The raw Telegram client
//   - success: Whether to checkout succeeded
//   - queryID: The pre-checkout query ID
//   - errMsg: Optional error message
//
// Returns true if successful, or an error.
func SetPreCheckoutResults(ctx context.Context, raw *tg.Client, success bool, queryID int64, errMsg string) (bool, error) {
	return raw.MessagesSetBotPrecheckoutResults(ctx, &tg.MessagesSetBotPrecheckoutResultsRequest{
		Success: success,
		QueryID: queryID,
		Error:   errMsg,
	})
}
