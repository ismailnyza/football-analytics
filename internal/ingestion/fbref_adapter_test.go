package ingestion

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestFbrefPlayerTableAdapter_Fetch(t *testing.T) {
	html, err := os.ReadFile("fbref/testdata/minimal_table.html")
	if err != nil {
		t.Fatal(err)
	}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write(html)
	}))
	defer srv.Close()

	store := NewInMemoryStagingStore()
	ad := NewFbrefPlayerTableAdapter(srv.URL)
	n, err := ad.Fetch(context.Background(), store)
	if err != nil {
		t.Fatal(err)
	}
	if n != 2 {
		t.Fatalf("n=%d", n)
	}
	recs, err := store.ListStaged(context.Background(), "fbref", "player")
	if err != nil || len(recs) != 2 {
		t.Fatalf("list %v err %v", len(recs), err)
	}
}
