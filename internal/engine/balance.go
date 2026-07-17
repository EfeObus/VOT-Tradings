// Package engine aggregates positions and balances across brokers into a
// single unified view, and enforces cross-broker risk checks such as the
// Pattern Day Trader rule.
package engine

import "vot-tradings/internal/models"

// UnifiedBalance is the cross-broker rollup of buying power and equity,
// expressed in both native per-broker currency and a single USD total.
type UnifiedBalance struct {
	TotalEquityUSD      float64          `json:"total_equity_usd"`
	TotalBuyingPowerUSD float64          `json:"total_buying_power_usd"`
	TotalCashUSD        float64          `json:"total_cash_usd"`
	ByAccount           []models.Account `json:"by_account"`
}

// USDRate returns the conversion multiplier for an account's native
// currency into USD (e.g. 0.73 means 1 CAD = 0.73 USD for a "CAD" account);
// 1.0 for any currency other than "CAD". Exported so callers that need a
// single account's USD-equivalent value (e.g. a per-broker allocation
// breakdown) use the same conversion AggregateBalances does internally.
func USDRate(currency string, usdCadRate float64) float64 {
	if currency == "CAD" {
		return usdCadRate
	}
	return 1.0
}

// AggregateBalances rolls up per-broker accounts into a unified view. CAD
// balances are converted to USD using usdCadRate (e.g. 0.73 means 1 CAD =
// 0.73 USD); pass 1.0 if an account's Currency is already "USD".
func AggregateBalances(accounts []models.Account, usdCadRate float64) UnifiedBalance {
	var unified UnifiedBalance
	unified.ByAccount = accounts

	for _, acct := range accounts {
		rate := USDRate(acct.Currency, usdCadRate)
		unified.TotalEquityUSD += acct.Equity * rate
		unified.TotalBuyingPowerUSD += acct.BuyingPower * rate
		unified.TotalCashUSD += acct.Cash * rate
	}

	return unified
}
