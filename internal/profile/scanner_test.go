package profile

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/pairs/chrome-profile-viewer/internal/model"
)

func TestScanFindsChromiumAndFirefoxProfiles(t *testing.T) {
	root := t.TempDir()
	chrome := filepath.Join(root, "Google", "Chrome", "User Data", "Profile 1")
	firefox := filepath.Join(root, "Mozilla", "Firefox", "Profiles", "abc.default-release")
	mustWrite(t, filepath.Join(chrome, "Preferences"), `{"profile":{"name":"İş Profili"}}`)
	mustWrite(t, filepath.Join(chrome, "History"), "")
	mustWrite(t, filepath.Join(firefox, "prefs.js"), "")
	mustWrite(t, filepath.Join(firefox, "places.sqlite"), "")

	browsers := []model.Browser{{ID: "chrome", Family: model.FamilyChromium}, {ID: "firefox", Family: model.FamilyFirefox}}
	result, err := Scan(root, browsers)
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Profiles) != 2 {
		t.Fatalf("expected 2 profiles, got %d", len(result.Profiles))
	}
	if result.Profiles[0].Name != "İş Profili" || result.Profiles[0].SuggestedBrowserID != "chrome" {
		t.Fatalf("unexpected chromium profile: %+v", result.Profiles[0])
	}
	if result.Profiles[1].Family != model.FamilyFirefox {
		t.Fatalf("unexpected firefox profile: %+v", result.Profiles[1])
	}
}

func mustWrite(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
}
