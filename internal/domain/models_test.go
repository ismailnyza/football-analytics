package domain

import "testing"

func TestParsePosition(t *testing.T) {
	got, err := ParsePosition(" cm ")
	if err != nil {
		t.Fatalf("ParsePosition() error = %v", err)
	}
	if got != PositionCM {
		t.Fatalf("ParsePosition() = %q, want %q", got, PositionCM)
	}
}

func TestParsePositionRejectsUnknownCode(t *testing.T) {
	if _, err := ParsePosition("xyz"); err == nil {
		t.Fatal("ParsePosition() error = nil, want error")
	}
}

func TestFixtureValidate(t *testing.T) {
	fixture := Fixture{
		SeasonID:   1,
		Matchday:   3,
		HomeClubID: 10,
		AwayClubID: 20,
	}
	if err := fixture.Validate(); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
}

func TestFixtureValidateRejectsSameClub(t *testing.T) {
	fixture := Fixture{
		SeasonID:   1,
		Matchday:   3,
		HomeClubID: 10,
		AwayClubID: 10,
	}
	if err := fixture.Validate(); err == nil {
		t.Fatal("Validate() error = nil, want error")
	}
}

func TestPlayerDisplayNamePrefersPreferredName(t *testing.T) {
	player := Player{
		FirstName:     "Martin",
		LastName:      "Odegaard",
		PreferredName: "Martin Odegaard",
	}
	if got, want := player.DisplayName(), "Martin Odegaard"; got != want {
		t.Fatalf("DisplayName() = %q, want %q", got, want)
	}
}
