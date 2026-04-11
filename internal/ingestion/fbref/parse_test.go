package fbref

import (
	"os"
	"testing"

	"github.com/PuerkitoBio/goquery"
)

func TestParsePlayerStatRows_minimalFixture(t *testing.T) {
	f, err := os.Open("testdata/minimal_table.html")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	doc, err := goquery.NewDocumentFromReader(f)
	if err != nil {
		t.Fatal(err)
	}
	rows, err := ParsePlayerStatRows(doc)
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 2 {
		t.Fatalf("rows=%d", len(rows))
	}
	if rows[0].ExternalID != "d70fd3fd" || rows[0].Stats["player"] != "Erling Haaland" {
		t.Fatalf("%+v", rows[0])
	}
	if rows[0].Stats["pos"] != "FW" {
		t.Fatalf("pos %q", rows[0].Stats["pos"])
	}
	if rows[1].ExternalID != "abcdef12" || rows[1].Stats["pos"] != "GK" {
		t.Fatalf("%+v", rows[1])
	}
}

func TestPlayerIDFromHref(t *testing.T) {
	id := playerIDFromHref("/en/players/d70fd3fd/Erling-Haaland")
	if id != "d70fd3fd" {
		t.Fatalf("got %q", id)
	}
}
