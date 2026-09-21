// Package inspect reads non-sensitive, human-readable facts from a browser profile.
// It never opens password, cookie or history databases.
package inspect

import (
	"bufio"
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/pairs/chrome-profile-viewer/internal/model"
)

func Inspect(profile model.Profile) (model.ProfileDetails, error) {
	details := model.ProfileDetails{}
	if _, err := os.Stat(profile.Path); err != nil {
		return details, err
	}
	switch profile.Family {
	case model.FamilyChromium:
		inspectChromium(profile, &details)
	case model.FamilyFirefox:
		inspectFirefox(profile, &details)
	default:
		return details, errors.New("desteklenmeyen profil türü")
	}
	if history, err := readHistory(profile); err != nil {
		if !errors.Is(err, fs.ErrNotExist) {
			details.Notes = append(details.Notes, "Geçmiş okunamadı: "+err.Error())
		}
	} else {
		details.History = history
	}
	details.Sites = mergeSites(details.Sites)
	sort.Strings(details.Extensions)

	newest := ""
	for _, file := range details.Files {
		if file.Modified > newest {
			newest = file.Modified
		}
	}
	if newest != "" {
		details.Fields = append(details.Fields, model.Field{Label: "Son değişiklik (tahmini erişim)", Value: newest})
	}
	return details, nil
}

// mergeSites keeps one entry per origin (the newest), most recently used first.
func mergeSites(sites []model.Site) []model.Site {
	newest := make(map[string]string, len(sites))
	for _, site := range sites {
		if site.Modified >= newest[site.Origin] {
			newest[site.Origin] = site.Modified
		}
	}
	out := make([]model.Site, 0, len(newest))
	for origin, modified := range newest {
		out = append(out, model.Site{Origin: origin, Modified: modified})
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Modified != out[j].Modified {
			return out[i].Modified > out[j].Modified
		}
		return out[i].Origin < out[j].Origin
	})
	return out
}

func addField(d *model.ProfileDetails, label, value string) {
	if value != "" {
		d.Fields = append(d.Fields, model.Field{Label: label, Value: value})
	}
}

// addSites records one site per directory of dir, using origin() to name it.
func addSites(d *model.ProfileDetails, dir string, origin func(string) string) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return
	}
	for _, entry := range entries {
		name := origin(entry.Name())
		if name == "" {
			continue
		}
		_, newest := dirStats(filepath.Join(dir, entry.Name()))
		d.Sites = append(d.Sites, model.Site{Origin: name, Modified: formatTime(newest)})
	}
}

func readJSON(path string) (map[string]any, bool) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, false
	}
	var out map[string]any
	if json.Unmarshal(data, &out) != nil {
		return nil, false
	}
	return out, true
}

func dig(value any, keys ...string) any {
	for _, key := range keys {
		m, ok := value.(map[string]any)
		if !ok {
			return nil
		}
		value = m[key]
	}
	return value
}

func str(value any) string {
	s, _ := value.(string)
	return s
}

func formatTime(t time.Time) string { return t.Local().Format("2006-01-02 15:04") }

func addFiles(d *model.ProfileDetails, root string, names ...string) {
	for _, name := range names {
		path := filepath.Join(root, filepath.FromSlash(name))
		info, err := os.Stat(path)
		if err != nil {
			continue
		}
		entry := model.FileEntry{Name: name, Size: info.Size(), Modified: formatTime(info.ModTime()), IsDir: info.IsDir()}
		if info.IsDir() {
			size, newest := dirStats(path)
			entry.Size = size
			if !newest.IsZero() {
				entry.Modified = formatTime(newest)
			}
		}
		d.Files = append(d.Files, entry)
	}
}

// dirStats returns the total file size and the newest modification time under path.
func dirStats(path string) (int64, time.Time) {
	var total int64
	var newest time.Time
	_ = filepath.WalkDir(path, func(_ string, entry fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		info, err := entry.Info()
		if err != nil {
			return nil
		}
		if info.ModTime().After(newest) {
			newest = info.ModTime()
		}
		if !entry.IsDir() {
			total += info.Size()
		}
		return nil
	})
	return total, newest
}

// ---- Chromium ----

func inspectChromium(profile model.Profile, d *model.ProfileDetails) {
	prefs, ok := readJSON(filepath.Join(profile.Path, "Preferences"))
	if !ok {
		d.Notes = append(d.Notes, "Preferences dosyası okunamadı (tarayıcı açık olabilir).")
	}
	addField(d, "Profil adı", str(dig(prefs, "profile", "name")))

	account := ""
	if accounts, _ := prefs["account_info"].([]any); len(accounts) > 0 {
		email, name := str(dig(accounts[0], "email")), str(dig(accounts[0], "full_name"))
		account = email
		if name != "" && email != "" {
			account = name + " <" + email + ">"
		} else if email == "" {
			account = name
		}
	}
	if localState, ok := readJSON(filepath.Join(profile.UserDataDir, "Local State")); ok {
		info := dig(localState, "profile", "info_cache", profile.DirectoryName)
		if account == "" {
			account = str(dig(info, "user_name"))
		}
		if active, ok := dig(info, "active_time").(float64); ok && active > 0 {
			addField(d, "Son kullanım", formatTime(time.Unix(int64(active), 0)))
		}
	}
	addField(d, "Hesap", account)

	// 4 = "keep cookies for the session only".
	if dig(prefs, "profile", "default_content_setting_values", "cookies") == float64(4) {
		d.Notes = append(d.Notes, "Çerezler tarayıcı kapanınca siliniyor; siteler oturumu unutabilir.")
	}

	if settings, ok := dig(prefs, "extensions", "settings").(map[string]any); ok {
		for id, raw := range settings {
			location, _ := dig(raw, "location").(float64)
			if location == 5 || location == 10 || dig(raw, "manifest") == nil { // component extensions
				continue
			}
			name := str(dig(raw, "manifest", "name"))
			if name == "" || strings.HasPrefix(name, "__MSG_") {
				name = id
			}
			d.Extensions = append(d.Extensions, name)
		}
	}

	if bookmarks, ok := readJSON(filepath.Join(profile.Path, "Bookmarks")); ok {
		count := 0
		if roots, ok := bookmarks["roots"].(map[string]any); ok {
			for _, root := range roots {
				count += countBookmarks(root)
			}
		}
		addField(d, "Yer imi", strconv.Itoa(count))
	}

	addSites(d, filepath.Join(profile.Path, "IndexedDB"), chromiumOrigin)

	addFiles(d, profile.Path, "IndexedDB", "Local Storage", "Service Worker", "Extensions",
		"Network/Cookies", "Cookies", "Login Data", "History", "Bookmarks", "Web Data", "Preferences")
}

func countBookmarks(node any) int {
	if str(dig(node, "type")) == "url" {
		return 1
	}
	total := 0
	if children, ok := dig(node, "children").([]any); ok {
		for _, child := range children {
			total += countBookmarks(child)
		}
	}
	return total
}

// "https_web.whatsapp.com_0.indexeddb.leveldb" -> "https://web.whatsapp.com"
func chromiumOrigin(dirName string) string {
	name := strings.TrimSuffix(strings.TrimSuffix(dirName, ".indexeddb.leveldb"), ".indexeddb.blob")
	if name == dirName {
		return ""
	}
	if i := strings.LastIndex(name, "_"); i > 0 {
		name = name[:i]
	}
	scheme, host, found := strings.Cut(name, "_")
	if !found || (scheme != "https" && scheme != "http") {
		return ""
	}
	return scheme + "://" + host
}

// ---- Firefox ----

var prefLine = regexp.MustCompile(`^user_pref\("([^"]+)",\s*(.+)\);\s*$`)

func readPrefs(path string) map[string]string {
	file, err := os.Open(path)
	if err != nil {
		return nil
	}
	defer file.Close()
	prefs := map[string]string{}
	scanner := bufio.NewScanner(file)
	scanner.Buffer(make([]byte, 0, 64*1024), 4*1024*1024)
	for scanner.Scan() {
		if m := prefLine.FindStringSubmatch(scanner.Text()); m != nil {
			prefs[m[1]] = strings.Trim(m[2], `"`)
		}
	}
	return prefs
}

func inspectFirefox(profile model.Profile, d *model.ProfileDetails) {
	prefs := readPrefs(filepath.Join(profile.Path, "prefs.js"))
	addField(d, "Profil klasörü", filepath.Base(profile.Path))
	addField(d, "Sync hesabı", prefs["services.sync.username"])
	addField(d, "Ana sayfa", prefs["browser.startup.homepage"])
	addField(d, "Dil", prefs["general.useragent.locale"])

	if data, err := os.ReadFile(filepath.Join(profile.Path, "compatibility.ini")); err == nil {
		for _, line := range strings.Split(string(data), "\n") {
			if v, ok := strings.CutPrefix(strings.TrimSpace(line), "LastVersion="); ok {
				addField(d, "Son çalışan sürüm", v)
			}
		}
	}
	if times, ok := readJSON(filepath.Join(profile.Path, "times.json")); ok {
		if created, ok := times["created"].(float64); ok {
			addField(d, "Oluşturulma", formatTime(time.UnixMilli(int64(created))))
		}
	}
	if logins, ok := readJSON(filepath.Join(profile.Path, "logins.json")); ok {
		if list, ok := logins["logins"].([]any); ok {
			addField(d, "Kayıtlı giriş sayısı", strconv.Itoa(len(list)))
		}
	}

	if prefs["privacy.sanitize.sanitizeOnShutdown"] == "true" || prefs["network.cookie.lifetimePolicy"] == "2" {
		d.Notes = append(d.Notes, "Firefox kapanırken çerez/site verilerini temizliyor; siteler oturumu unutabilir.")
	}

	if ext, ok := readJSON(filepath.Join(profile.Path, "extensions.json")); ok {
		addons, _ := ext["addons"].([]any)
		for _, addon := range addons {
			if str(dig(addon, "type")) == "extension" && str(dig(addon, "location")) == "app-profile" {
				name := str(dig(addon, "defaultLocale", "name"))
				if name == "" {
					name = str(dig(addon, "id"))
				}
				d.Extensions = append(d.Extensions, name)
			}
		}
	}

	addSites(d, filepath.Join(profile.Path, "storage", "default"), firefoxOrigin)

	addFiles(d, profile.Path, "storage", "cookies.sqlite", "places.sqlite", "logins.json", "key4.db",
		"extensions", "sessionstore-backups", "prefs.js")
}

// "https+++web.whatsapp.com^userContextId=2" -> "https://web.whatsapp.com"
func firefoxOrigin(dirName string) string {
	name, _, _ := strings.Cut(dirName, "^")
	scheme, rest, found := strings.Cut(name, "+++")
	if !found || (scheme != "https" && scheme != "http") {
		return ""
	}
	return scheme + "://" + rest
}
