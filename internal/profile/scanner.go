package profile

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/pairs/chrome-profile-viewer/internal/model"
)

const maxDepth = 8

var ignoredDirectories = map[string]bool{
	"cache": true, "code cache": true, "gpucache": true, "shadercache": true,
	"crashpad": true, "dawncache": true, "grshadercache": true, "node_modules": true,
}

func Scan(root string, browsers []model.Browser) (model.ScanResult, error) {
	root = strings.TrimSpace(root)
	if root == "" {
		return model.ScanResult{}, errors.New("klasör yolu boş")
	}
	absRoot, err := filepath.Abs(root)
	if err != nil {
		return model.ScanResult{}, err
	}
	info, err := os.Stat(absRoot)
	if err != nil {
		return model.ScanResult{}, err
	}
	if !info.IsDir() {
		return model.ScanResult{}, errors.New("seçilen yol bir klasör değil")
	}

	result := model.ScanResult{Root: absRoot, Profiles: []model.Profile{}, Browsers: browsers, Warnings: []string{}}
	seen := make(map[string]bool)
	directories := 0
	err = filepath.WalkDir(absRoot, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			result.Warnings = append(result.Warnings, path+": "+walkErr.Error())
			return nil
		}
		if !entry.IsDir() {
			return nil
		}
		directories++
		if directories > 20000 {
			return errors.New("tarama güvenlik sınırını aştı (20.000 klasör)")
		}
		rel, _ := filepath.Rel(absRoot, path)
		depth := 0
		if rel != "." {
			depth = strings.Count(filepath.ToSlash(rel), "/") + 1
		}
		if depth > maxDepth || (path != absRoot && ignoredDirectories[strings.ToLower(entry.Name())]) {
			return filepath.SkipDir
		}

		if detected, ok := inspect(path, browsers); ok && !seen[detected.ID] {
			seen[detected.ID] = true
			result.Profiles = append(result.Profiles, detected)
		}
		return nil
	})
	if err != nil {
		return result, err
	}
	if len(result.Profiles) == 0 {
		result.Warnings = append(result.Warnings, "Bu klasörde desteklenen bir tarayıcı profili bulunamadı.")
	}
	if len(browsers) == 0 {
		result.Warnings = append(result.Warnings, "Sistemde desteklenen bir tarayıcı bulunamadı.")
	}
	sort.Slice(result.Profiles, func(i, j int) bool {
		if result.Profiles[i].Family != result.Profiles[j].Family {
			return result.Profiles[i].Family < result.Profiles[j].Family
		}
		return strings.ToLower(result.Profiles[i].Name) < strings.ToLower(result.Profiles[j].Name)
	})
	return result, nil
}

func inspect(path string, browsers []model.Browser) (model.Profile, bool) {
	if has(path, "Preferences") && hasAny(path, "Secure Preferences", "History", "Login Data", "Web Data", "Cookies", filepath.Join("Network", "Cookies")) {
		userDataDir := filepath.Dir(path)
		directoryName := filepath.Base(path)
		product := productHint(path, model.FamilyChromium)
		return model.Profile{
			ID: profileID(path), Name: chromiumName(path), Path: path, UserDataDir: userDataDir,
			DirectoryName: directoryName, Family: model.FamilyChromium, ProductHint: product,
			SuggestedBrowserID: suggestBrowser(product, model.FamilyChromium, browsers),
			Locked:             hasAny(userDataDir, "SingletonLock", "SingletonCookie", "SingletonSocket"),
			Evidence:           []string{"Preferences", "Chromium profil veritabanı"},
		}, true
	}
	if has(path, "prefs.js") && hasAny(path, "places.sqlite", "cookies.sqlite", "logins.json", "key4.db", "handlers.json") {
		product := productHint(path, model.FamilyFirefox)
		return model.Profile{
			ID: profileID(path), Name: filepath.Base(path), Path: path, UserDataDir: path,
			Family: model.FamilyFirefox, ProductHint: product,
			SuggestedBrowserID: suggestBrowser(product, model.FamilyFirefox, browsers),
			Locked:             hasAny(path, "parent.lock", ".parentlock", "lock"),
			Evidence:           []string{"prefs.js", "Firefox profil veritabanı"},
		}, true
	}
	return model.Profile{}, false
}

func chromiumName(path string) string {
	var preferences struct {
		Profile struct {
			Name string `json:"name"`
		} `json:"profile"`
	}
	data, err := os.ReadFile(filepath.Join(path, "Preferences"))
	if err == nil && json.Unmarshal(data, &preferences) == nil && strings.TrimSpace(preferences.Profile.Name) != "" {
		return preferences.Profile.Name
	}
	return filepath.Base(path)
}

func profileID(path string) string {
	sum := sha256.Sum256([]byte(strings.ToLower(filepath.Clean(path))))
	return hex.EncodeToString(sum[:8])
}

func has(dir, name string) bool {
	_, err := os.Lstat(filepath.Join(dir, name))
	return err == nil
}

func hasAny(dir string, names ...string) bool {
	for _, name := range names {
		if has(dir, name) {
			return true
		}
	}
	return false
}

func productHint(path string, family model.Family) string {
	normalized := strings.ToLower(filepath.ToSlash(path))
	for _, pair := range []struct{ token, id string }{
		{"bravesoftware/brave-browser", "brave"}, {"microsoft/edge", "edge"},
		{"google/chrome", "chrome"}, {"/chromium/", "chromium"}, {"/vivaldi/", "vivaldi"},
		{"opera software", "opera"}, {"/librewolf/", "librewolf"}, {"mozilla/firefox", "firefox"},
	} {
		if strings.Contains(normalized, pair.token) {
			return pair.id
		}
	}
	if family == model.FamilyFirefox {
		return "firefox"
	}
	return "chromium"
}

func suggestBrowser(product string, family model.Family, browsers []model.Browser) string {
	for _, candidate := range browsers {
		if candidate.ID == product {
			return candidate.ID
		}
	}
	for _, candidate := range browsers {
		if candidate.Family == family {
			return candidate.ID
		}
	}
	return ""
}
