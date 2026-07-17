package engine

import (
	"time"

	"vot-tradings/internal/models"
)

// PDTEquityThreshold is FINRA's minimum equity (USD) below which the
// Pattern Day Trader day-trade limit applies to margin accounts.
const PDTEquityThreshold = 25000.0

// PDTDayTradeLimit is the maximum number of day trades a sub-threshold
// account may execute within the rolling window before being flagged.
const PDTDayTradeLimit = 3

// businessDaysBack walks backward from `from`, skipping Saturdays and
// Sundays, to find the start of a trailing N-business-day window.
//
// Note: this does not account for market holidays — a real production
// implementation should consult an exchange calendar.
func businessDaysBack(from time.Time, days int) time.Time {
	d := from
	counted := 0
	for counted < days {
		d = d.AddDate(0, 0, -1)
		if d.Weekday() != time.Saturday && d.Weekday() != time.Sunday {
			counted++
		}
	}
	return d
}

// CountDayTradesInWindow counts how many day trades an account has executed
// within the trailing 5-business-day window ending at asOf.
func CountDayTradesInWindow(trades []models.DayTrade, accountID string, asOf time.Time) int {
	windowStart := businessDaysBack(asOf, 5)

	count := 0
	for _, t := range trades {
		if t.AccountID != accountID {
			continue
		}
		if !t.TradeDate.Before(windowStart) && !t.TradeDate.After(asOf) {
			count++
		}
	}
	return count
}

// PDTCheckResult reports whether an account is clear to execute another day
// trade without triggering (or worsening) a Pattern Day Trader restriction.
type PDTCheckResult struct {
	Restricted        bool
	DayTradesInWindow int
	Reason            string
}

// CheckPDT evaluates whether an account under the $25,000 equity threshold
// is at or past the FINRA day-trade limit for the rolling 5-business-day
// window. Accounts at or above the threshold are never restricted by this
// rule.
func CheckPDT(account models.Account, trades []models.DayTrade, asOf time.Time) PDTCheckResult {
	if account.Equity >= PDTEquityThreshold {
		return PDTCheckResult{Restricted: false}
	}

	count := CountDayTradesInWindow(trades, account.ID, asOf)
	if count >= PDTDayTradeLimit {
		return PDTCheckResult{
			Restricted:        true,
			DayTradesInWindow: count,
			Reason:            "account equity is below $25,000 and has reached the FINRA day-trade limit for the rolling 5-business-day window; one more day trade will trigger a Pattern Day Trader restriction",
		}
	}

	return PDTCheckResult{DayTradesInWindow: count}
}
