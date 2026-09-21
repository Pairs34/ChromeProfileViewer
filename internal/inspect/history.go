package inspect

import (
	"database/sql"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"time"

	_ "modernc.org/sqlite"

	"github.com/pairs/chrome-profile-viewer/internal/model"
)

const historyLimit = 300

// chromeEpochOffset is the number of seconds between 1601-01-01 (WebKit time) and the Unix epoch.
const chromeEpochOffset = 11644473600

type historyRow struct {
	url    string
	visits int
	last   time.Time
}

// readHistory aggregates visited pages per host, newest first. The database is
// copied first because the browser keeps it locked while running.
func readHistory(profile model.Profile) ([]model.HistorySite, error) {
	var name, query string
	var toTime func(int64) time.Time
	switch profile.Family {
	case model.FamilyChromium:
		name = "History"
		query = "SELECT url, visit_count, last_visit_time FROM urls WHERE last_visit_time > 0"
		toTime = func(v int64) time.Time { return time.Unix(v/1_000_000-chromeEpochOffset, 0) }
	default:
		name = "places.sqlite"
		query = "SELECT url, visit_count, last_visit_date FROM moz_places WHERE last_visit_date IS NOT NULL"
		toTime = func(v int64) time.Time { return time.UnixMicro(v) }
	}
	source := filepath.Join(profile.Path, name)
	if _, err := os.Stat(source); err != nil {
		return nil, err
	}

	temp, err := os.MkdirTemp("", "history-")
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(temp)
	copyPath := filepath.Join(temp, name)
	for _, suffix := range []string{"", "-wal", "-journal"} {
		if err := copyIfExists(source+suffix, copyPath+suffix); err != nil && suffix == "" {
			return nil, err
		}
	}

	db, err := sql.Open("sqlite", copyPath)
	if err != nil {
		return nil, err
	}
	defer db.Close()
	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	type stat struct {
		visits int
		last   time.Time
	}
	hosts := map[string]*stat{}
	for rows.Next() {
		var pageURL string
		var visits, raw int64
		if err := rows.Scan(&pageURL, &visits, &raw); err != nil {
			return nil, err
		}
		parsed, err := url.Parse(pageURL)
		if err != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") || parsed.Hostname() == "" {
			continue
		}
		entry := hosts[parsed.Hostname()]
		if entry == nil {
			entry = &stat{}
			hosts[parsed.Hostname()] = entry
		}
		entry.visits += int(visits)
		if t := toTime(raw); t.After(entry.last) {
			entry.last = t
		}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	out := make([]model.HistorySite, 0, len(hosts))
	for host, entry := range hosts {
		out = append(out, model.HistorySite{Host: host, Visits: entry.visits, LastVisit: formatTime(entry.last)})
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].LastVisit != out[j].LastVisit {
			return out[i].LastVisit > out[j].LastVisit
		}
		return out[i].Host < out[j].Host
	})
	if len(out) > historyLimit {
		out = out[:historyLimit]
	}
	return out, nil
}

func copyIfExists(source, destination string) error {
	input, err := os.Open(source)
	if err != nil {
		return err
	}
	defer input.Close()
	output, err := os.Create(destination)
	if err != nil {
		return err
	}
	_, copyErr := io.Copy(output, input)
	if closeErr := output.Close(); copyErr == nil {
		copyErr = closeErr
	}
	return copyErr
}
