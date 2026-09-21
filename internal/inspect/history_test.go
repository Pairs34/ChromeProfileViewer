package inspect

import (
	"database/sql"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/pairs/chrome-profile-viewer/internal/model"
)

func makeDB(t *testing.T, path string, statements ...string) {
	t.Helper()
	db, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	for _, statement := range statements {
		if _, err := db.Exec(statement); err != nil {
			t.Fatal(err)
		}
	}
}

func TestReadHistoryChromium(t *testing.T) {
	profile := t.TempDir()
	// 2024-01-02 03:04:05 UTC in WebKit microseconds.
	newer := (int64(1704164645) + chromeEpochOffset) * 1_000_000
	older := (int64(1704164645-86400) + chromeEpochOffset) * 1_000_000
	makeDB(t, filepath.Join(profile, "History"),
		"CREATE TABLE urls (url TEXT, visit_count INTEGER, last_visit_time INTEGER)",
		"INSERT INTO urls VALUES ('https://web.whatsapp.com/', 5, "+itoa(older)+")",
		"INSERT INTO urls VALUES ('https://web.whatsapp.com/x', 2, "+itoa(newer)+")",
		"INSERT INTO urls VALUES ('https://example.com/a', 1, "+itoa(older)+")",
		"INSERT INTO urls VALUES ('chrome://settings', 9, "+itoa(newer)+")",
	)
	history, err := readHistory(model.Profile{Family: model.FamilyChromium, Path: profile})
	if err != nil {
		t.Fatal(err)
	}
	if len(history) != 2 || history[0].Host != "web.whatsapp.com" || history[0].Visits != 7 || history[1].Host != "example.com" {
		t.Fatalf("unexpected history: %+v", history)
	}
	if history[0].LastVisit <= history[1].LastVisit {
		t.Fatalf("expected newest first: %+v", history)
	}
}

func TestReadHistoryFirefox(t *testing.T) {
	profile := t.TempDir()
	makeDB(t, filepath.Join(profile, "places.sqlite"),
		"CREATE TABLE moz_places (url TEXT, visit_count INTEGER, last_visit_date INTEGER)",
		"INSERT INTO moz_places VALUES ('https://web.whatsapp.com/', 3, 1704164645000000)",
		"INSERT INTO moz_places VALUES ('about:home', 1, 1704164645000000)",
		"INSERT INTO moz_places VALUES ('https://never.example/', 0, NULL)",
	)
	history, err := readHistory(model.Profile{Family: model.FamilyFirefox, Path: profile})
	if err != nil {
		t.Fatal(err)
	}
	if len(history) != 1 || history[0].Host != "web.whatsapp.com" || history[0].Visits != 3 {
		t.Fatalf("unexpected history: %+v", history)
	}
}

func itoa(n int64) string { return strconv.FormatInt(n, 10) }
