package ingestion

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestReadFbrefURLFile(t *testing.T) {
	dir := t.TempDir()
	if _, err := ReadFbrefURLFile(dir); err == nil {
		t.Fatal("expected error")
	}
	if err := os.WriteFile(filepath.Join(dir, FbrefURLFile), []byte("# c\n\n  https://example.com/x  \n"), 0o644); err != nil {
		t.Fatal(err)
	}
	u, err := ReadFbrefURLFile(dir)
	if err != nil || u != "https://example.com/x" {
		t.Fatalf("%q %v", u, err)
	}
}

func TestRunFbrefScrape_endToEnd(t *testing.T) {
	html, err := os.ReadFile("fbref/testdata/minimal_table.html")
	if err != nil {
		t.Fatal(err)
	}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write(html)
	}))
	defer srv.Close()

	dir := t.TempDir()
	if err := RunFbrefScrape(context.Background(), dir, srv.URL, time.Now()); err != nil {
		t.Fatal(err)
	}
	reg, err := LoadOrCreateRegistry(dir)
	if err != nil {
		t.Fatal(err)
	}
	var fb *SourceSnapshot
	for _, s := range reg.Snapshots() {
		if s.Name == "fbref" {
			fb = &s
			break
		}
	}
	if fb == nil || fb.Records != 2 {
		t.Fatalf("%+v", fb)
	}
}
