// Package finance provides the club finance engine — revenue, wage bill,
// transfer budget, and year-end balance calculation.
package finance

// ClubFinances holds the financial state of a club in a simulation year.
type ClubFinances struct {
	ClubID       int64
	// Revenue streams
	TicketRevenue  int64
	TVRevenue      int64
	SponsorRevenue int64
	// Costs
	WageBill       int64
	TransferSpend  int64
	TransferIncome int64
	// Balance
	OpeningBalance int64
	ClosingBalance int64
}

// FinanceParams configures the annual finance model.
type FinanceParams struct {
	// LeaguePosition (1=top) scales TV and ticket revenue.
	LeaguePosition int
	LeagueSize     int
	// BaseTicketRevenue and BaseTVRevenue are starting values before scaling.
	BaseTicketRevenue  int64
	BaseTVRevenue      int64
	BaseSponsorRevenue int64
}

// DefaultFinanceParams returns sensible defaults for a mid-table team.
func DefaultFinanceParams(leagueSize int) FinanceParams {
	return FinanceParams{
		LeaguePosition:     leagueSize / 2,
		LeagueSize:         leagueSize,
		BaseTicketRevenue:  20_000_000,
		BaseTVRevenue:      80_000_000,
		BaseSponsorRevenue: 30_000_000,
	}
}

// ComputeAnnualFinances calculates the full year-end financial position for a
// club given their performance and expenditure parameters.
func ComputeAnnualFinances(opening int64, wageBill, transferSpend, transferIncome int64, params FinanceParams) ClubFinances {
	// TV revenue scales inversely with position (top teams get more).
	positionMultiplier := 1.0
	if params.LeagueSize > 0 {
		positionMultiplier = 1.0 + float64(params.LeagueSize-params.LeaguePosition)/float64(params.LeagueSize)
	}

	ticketRevenue := int64(float64(params.BaseTicketRevenue) * positionMultiplier)
	tvRevenue := int64(float64(params.BaseTVRevenue) * positionMultiplier)
	sponsorRevenue := params.BaseSponsorRevenue

	totalRevenue := ticketRevenue + tvRevenue + sponsorRevenue + transferIncome
	totalCosts := wageBill + transferSpend

	closing := opening + totalRevenue - totalCosts

	return ClubFinances{
		TicketRevenue:  ticketRevenue,
		TVRevenue:      tvRevenue,
		SponsorRevenue: sponsorRevenue,
		WageBill:       wageBill,
		TransferSpend:  transferSpend,
		TransferIncome: transferIncome,
		OpeningBalance: opening,
		ClosingBalance: closing,
	}
}

// TransferBudget returns the recommended transfer spend ceiling based on
// the current financial position.
func TransferBudget(finances ClubFinances) int64 {
	surplus := finances.ClosingBalance
	if surplus <= 0 {
		return 0
	}
	// Spend up to 40% of closing balance on transfers.
	budget := surplus * 40 / 100
	return budget
}
