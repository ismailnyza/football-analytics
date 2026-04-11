package ingestion

import (
	"testing"
)

func TestNormalizeFbrefPlayerRecord_ok(t *testing.T) {
	raw := SourceRecord{
		SourceName: "fbref",
		EntityType: "player",
		ExternalID: "d70fd3fd",
		RawJSON:    `{"player":"Erling Haaland","nation":"no NOR","pos":"FW","age":"24"}`,
	}
	norm, err := NormalizeFbrefPlayerRecord(raw)
	if err != nil {
		t.Fatal(err)
	}
	if norm.Name != "Erling Haaland" || norm.Attributes["position"] != "ST" {
		t.Fatalf("%+v", norm)
	}
	if norm.Attributes["nationality"] != "NOR" {
		t.Fatalf("nat %q", norm.Attributes["nationality"])
	}
	v := ValidatePlayerRecord(norm.Attributes)
	if !v.Valid {
		t.Fatalf("%v", v.Errors)
	}
}

func TestNormalizeFbrefPlayerRecord_pipeline(t *testing.T) {
	raw := SourceRecord{
		SourceName: "fbref",
		EntityType: "player",
		ExternalID: "x",
		RawJSON:    `{"player":"A B","nation":"x X","pos":"CM"}`,
	}
	out, errs := Normalize([]SourceRecord{raw}, NormalizeFbrefPlayerRecord)
	if len(errs) != 0 || len(out) != 1 || out[0].Attributes["position"] != "CM" {
		t.Fatalf("out=%v errs=%v", out, errs)
	}
}

func TestMapFBrefPosition_codes(t *testing.T) {
	for in, want := range map[string]string{
		"GK": "GK", "FW": "ST", "LW": "LW", "DM": "DM", "CB": "CB",
	} {
		if got := mapFBrefPosition(in); got != want {
			t.Fatalf("%q -> %q want %q", in, got, want)
		}
	}
}
