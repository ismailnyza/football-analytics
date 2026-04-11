package transfer

import (
	"testing"

	"github.com/ismael/football-analytics/internal/domain"
)

func makePlayer(id int64, ovr, potential int) domain.Player {
	return domain.Player{
		ID: id,
		Attributes: domain.PlayerAttributes{Overall: ovr, Potential: potential},
	}
}

func TestMarketValue_higherOvrMoreValue(t *testing.T) {
	elite := makePlayer(1, 90, 92)
	average := makePlayer(2, 72, 75)
	if MarketValue(elite, 25) <= MarketValue(average, 25) {
		t.Fatal("elite player should be worth more than average")
	}
}

func TestMarketValue_peakAgePremium(t *testing.T) {
	p := makePlayer(1, 82, 85)
	youngValue := MarketValue(p, 24)
	oldValue := MarketValue(p, 34)
	if youngValue <= oldValue {
		t.Fatal("peak-age player should be worth more than veteran")
	}
}

func TestMarketValue_potentialBonus(t *testing.T) {
	highPot := makePlayer(1, 75, 90)
	lowPot := makePlayer(2, 75, 76)
	if MarketValue(highPot, 20) <= MarketValue(lowPot, 20) {
		t.Fatal("high potential player should be worth more")
	}
}

func TestEvaluateTransfer_buyWhenAffordableAndUpgrade(t *testing.T) {
	p := makePlayer(1, 62, 65)
	val := MarketValue(p, 26)
	decision := EvaluateTransfer(p, 26, val+1, 55)
	if decision.DecisionType != "buy" {
		t.Fatalf("expected buy, got %s (value=%d)", decision.DecisionType, val)
	}
}

func TestEvaluateTransfer_passWhenTooExpensive(t *testing.T) {
	p := makePlayer(1, 95, 97)
	decision := EvaluateTransfer(p, 26, 1_000, 75)
	if decision.DecisionType != "pass" {
		t.Fatalf("expected pass when budget too low, got %s", decision.DecisionType)
	}
}

func TestEvaluateTransfer_passWhenNoUpgrade(t *testing.T) {
	p := makePlayer(1, 72, 75)
	decision := EvaluateTransfer(p, 26, 200_000_000, 80)
	if decision.DecisionType != "pass" {
		t.Fatalf("expected pass when player below squad avg, got %s", decision.DecisionType)
	}
}

func TestContractStatus(t *testing.T) {
	cases := []struct {
		years  int
		status ContractStatus
	}{
		{3, StatusSecure},
		{1, StatusMonitor},
		{0, StatusExpiring},
	}
	for _, tc := range cases {
		c := Contract{YearsRemaining: tc.years}
		if c.Status() != tc.status {
			t.Fatalf("years=%d: status=%s, want %s", tc.years, c.Status(), tc.status)
		}
	}
}

func TestRenewalOffer_wageIncrease(t *testing.T) {
	c := Contract{PlayerID: 1, WeeklyWage: 100_000, YearsRemaining: 1, ExpiryYear: 2025}
	offer := RenewalOffer(c, 0.10, 3)
	if offer.WeeklyWage != 110_000 {
		t.Fatalf("wage = %d, want 110000", offer.WeeklyWage)
	}
	if offer.YearsRemaining != 3 {
		t.Fatalf("years = %d, want 3", offer.YearsRemaining)
	}
}

func TestAnnualWageBill(t *testing.T) {
	contracts := []Contract{
		{WeeklyWage: 100_000},
		{WeeklyWage: 50_000},
	}
	bill := AnnualWageBill(contracts)
	expected := int64(150_000 * 52)
	if bill != expected {
		t.Fatalf("wage bill = %d, want %d", bill, expected)
	}
}

func TestContractFor_derivesWage(t *testing.T) {
	p := makePlayer(1, 85, 88)
	c := ContractFor(p, 10, 2024, 3)
	if c.WeeklyWage <= 0 {
		t.Fatal("weekly wage must be positive")
	}
	if c.ExpiryYear != 2027 {
		t.Fatalf("expiry = %d, want 2027", c.ExpiryYear)
	}
}
