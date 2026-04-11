package finance

import (
	"testing"
)

func TestComputeAnnualFinances_positiveSurplus(t *testing.T) {
	params := DefaultFinanceParams(20)
	params.LeaguePosition = 5
	f := ComputeAnnualFinances(10_000_000, 50_000_000, 20_000_000, 15_000_000, params)
	if f.ClosingBalance <= 0 {
		t.Fatalf("expected positive closing balance, got %d", f.ClosingBalance)
	}
}

func TestComputeAnnualFinances_topTeamMoreRevenue(t *testing.T) {
	params := DefaultFinanceParams(20)
	top := params
	top.LeaguePosition = 1
	mid := params
	mid.LeaguePosition = 10

	fTop := ComputeAnnualFinances(0, 0, 0, 0, top)
	fMid := ComputeAnnualFinances(0, 0, 0, 0, mid)

	topRevenue := fTop.TicketRevenue + fTop.TVRevenue
	midRevenue := fMid.TicketRevenue + fMid.TVRevenue

	if topRevenue <= midRevenue {
		t.Fatalf("top team revenue %d should exceed mid table %d", topRevenue, midRevenue)
	}
}

func TestTransferBudget_zeroWhenNegativeBalance(t *testing.T) {
	f := ClubFinances{ClosingBalance: -5_000_000}
	if TransferBudget(f) != 0 {
		t.Fatal("budget must be 0 when closing balance is negative")
	}
}

func TestTransferBudget_fortyPercent(t *testing.T) {
	f := ClubFinances{ClosingBalance: 100_000_000}
	budget := TransferBudget(f)
	if budget != 40_000_000 {
		t.Fatalf("budget = %d, want 40_000_000", budget)
	}
}

func TestComputeAnnualFinances_openingCarriedForward(t *testing.T) {
	params := DefaultFinanceParams(20)
	f := ComputeAnnualFinances(50_000_000, 0, 0, 0, params)
	if f.OpeningBalance != 50_000_000 {
		t.Fatal("opening balance not preserved")
	}
}
