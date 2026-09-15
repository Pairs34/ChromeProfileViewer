package launcher

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/pairs/chrome-profile-viewer/internal/model"
)

func TestBuildArguments(t *testing.T) {
	tests := []struct {
		name    string
		profile model.Profile
		want    []string
	}{
		{"chromium", model.Profile{Family: model.FamilyChromium, UserDataDir: "/data", DirectoryName: "Profile 1"}, []string{"--user-data-dir=/data", "--profile-directory=Profile 1", "--new-window"}},
		{"firefox", model.Profile{Family: model.FamilyFirefox, Path: "/profile"}, []string{"-no-remote", "-profile", "/profile"}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := BuildArguments(test.profile)
			if err != nil || !reflect.DeepEqual(got, test.want) {
				t.Fatalf("got %v, %v; want %v", got, err, test.want)
			}
		})
	}
}

func TestIsolatedCopyCopiesProfileAndSkipsCaches(t *testing.T) {
	root := t.TempDir()
	profileDir := filepath.Join(root, "Profile 1")
	mustWrite(t, filepath.Join(root, "Local State"), "state")
	mustWrite(t, filepath.Join(profileDir, "Preferences"), "preferences")
	mustWrite(t, filepath.Join(profileDir, "Cache", "cached"), "cache")

	service := &Service{tempRoot: t.TempDir()}
	copied, err := service.isolatedCopy(model.Profile{
		Family: model.FamilyChromium, Path: profileDir, UserDataDir: root, DirectoryName: "Profile 1", Locked: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if copied.Locked || copied.UserDataDir == root {
		t.Fatalf("profile was not isolated: %+v", copied)
	}
	assertContent(t, filepath.Join(copied.UserDataDir, "Local State"), "state")
	assertContent(t, filepath.Join(copied.Path, "Preferences"), "preferences")
	if _, err := os.Stat(filepath.Join(copied.Path, "Cache", "cached")); !os.IsNotExist(err) {
		t.Fatalf("cache should not be copied, got error %v", err)
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

func assertContent(t *testing.T, path, want string) {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil || string(data) != want {
		t.Fatalf("%s: got %q, %v; want %q", path, data, err, want)
	}
}
