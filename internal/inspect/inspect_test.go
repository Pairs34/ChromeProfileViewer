package inspect

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/pairs/chrome-profile-viewer/internal/model"
)

func write(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func hasField(d model.ProfileDetails, label, value string) bool {
	for _, f := range d.Fields {
		if f.Label == label && f.Value == value {
			return true
		}
	}
	return false
}

func TestInspectChromium(t *testing.T) {
	root := t.TempDir()
	profile := filepath.Join(root, "Profile 1")
	write(t, filepath.Join(profile, "Preferences"), `{
		"profile": {"name": "İş", "default_content_setting_values": {"cookies": 4}},
		"account_info": [{"email": "a@b.c", "full_name": "Ada"}],
		"extensions": {"settings": {
			"abc": {"location": 1, "manifest": {"name": "uBlock"}},
			"def": {"location": 5, "manifest": {"name": "Component"}},
			"ghi": {"location": 1, "manifest": {"name": "__MSG_appName__"}}
		}}}`)
	write(t, filepath.Join(profile, "Bookmarks"), `{"roots":{"bookmark_bar":{"type":"folder","children":[{"type":"url"},{"type":"url"}]}}}`)
	write(t, filepath.Join(profile, "IndexedDB", "https_web.whatsapp.com_0.indexeddb.leveldb", "000003.log"), "x")
	write(t, filepath.Join(profile, "IndexedDB", "chrome-extension_abc_0.indexeddb.leveldb", "000003.log"), "x")

	d, err := Inspect(model.Profile{Family: model.FamilyChromium, Path: profile, UserDataDir: root, DirectoryName: "Profile 1"})
	if err != nil {
		t.Fatal(err)
	}
	if !hasField(d, "Profil adı", "İş") || !hasField(d, "Hesap", "Ada <a@b.c>") || !hasField(d, "Yer imi", "2") {
		t.Fatalf("unexpected fields: %+v", d.Fields)
	}
	if len(d.Sites) != 1 || d.Sites[0].Origin != "https://web.whatsapp.com" || d.Sites[0].Modified == "" {
		t.Fatalf("unexpected sites: %v", d.Sites)
	}
	if strings.Join(d.Extensions, ",") != "ghi,uBlock" {
		t.Fatalf("unexpected extensions: %v", d.Extensions)
	}
	if len(d.Notes) != 1 {
		t.Fatalf("expected cookie note, got %v", d.Notes)
	}
}

func TestInspectFirefox(t *testing.T) {
	profile := filepath.Join(t.TempDir(), "abc.default-release")
	write(t, filepath.Join(profile, "prefs.js"), "user_pref(\"privacy.sanitize.sanitizeOnShutdown\", true);\nuser_pref(\"services.sync.username\", \"me@x.y\");\n")
	write(t, filepath.Join(profile, "compatibility.ini"), "[Compatibility]\nLastVersion=130.0_2024/GRE\n")
	write(t, filepath.Join(profile, "logins.json"), `{"logins":[{},{}]}`)
	write(t, filepath.Join(profile, "extensions.json"), `{"addons":[{"type":"extension","location":"app-profile","defaultLocale":{"name":"Bitwarden"}},{"type":"extension","location":"app-builtin","id":"x"}]}`)
	write(t, filepath.Join(profile, "storage", "default", "https+++web.whatsapp.com", ".metadata"), "x")
	write(t, filepath.Join(profile, "storage", "default", "https+++web.whatsapp.com^userContextId=2", ".metadata"), "x")
	write(t, filepath.Join(profile, "storage", "default", "moz-extension+++uuid", ".metadata"), "x")

	d, err := Inspect(model.Profile{Family: model.FamilyFirefox, Path: profile})
	if err != nil {
		t.Fatal(err)
	}
	if !hasField(d, "Sync hesabı", "me@x.y") || !hasField(d, "Son çalışan sürüm", "130.0_2024/GRE") || !hasField(d, "Kayıtlı giriş sayısı", "2") {
		t.Fatalf("unexpected fields: %+v", d.Fields)
	}
	if len(d.Sites) != 1 || d.Sites[0].Origin != "https://web.whatsapp.com" {
		t.Fatalf("unexpected sites: %v", d.Sites)
	}
	if len(d.Extensions) != 1 || d.Extensions[0] != "Bitwarden" || len(d.Notes) != 1 {
		t.Fatalf("unexpected extensions/notes: %v %v", d.Extensions, d.Notes)
	}
}
