package browser

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"

	"github.com/pairs/chrome-profile-viewer/internal/model"
)

type candidate struct {
	id, name string
	family   model.Family
	paths    []string
	commands []string
	registry []string
}

// Detect returns locally installed supported browsers without invoking them.
func Detect() []model.Browser {
	seen := make(map[string]bool)
	var result []model.Browser
	for _, item := range candidates() {
		path := firstExecutable(item.paths, item.commands, item.registry)
		if path == "" {
			continue
		}
		key := strings.ToLower(filepath.Clean(path))
		if seen[key] {
			continue
		}
		seen[key] = true
		result = append(result, model.Browser{ID: item.id, Name: item.name, Family: item.family, Executable: path})
	}
	sort.Slice(result, func(i, j int) bool { return result[i].Name < result[j].Name })
	return result
}

func candidates() []candidate {
	chromium := model.FamilyChromium
	firefox := model.FamilyFirefox
	items := []candidate{
		{id: "chrome", name: "Google Chrome", family: chromium, commands: []string{"google-chrome", "google-chrome-stable", "chrome"}},
		{id: "edge", name: "Microsoft Edge", family: chromium, commands: []string{"microsoft-edge", "microsoft-edge-stable"}},
		{id: "brave", name: "Brave", family: chromium, commands: []string{"brave-browser", "brave"}},
		{id: "chromium", name: "Chromium", family: chromium, commands: []string{"chromium", "chromium-browser"}},
		{id: "vivaldi", name: "Vivaldi", family: chromium, commands: []string{"vivaldi", "vivaldi-stable"}},
		{id: "opera", name: "Opera", family: chromium, commands: []string{"opera"}},
		{id: "firefox", name: "Mozilla Firefox", family: firefox, commands: []string{"firefox"}},
		{id: "librewolf", name: "LibreWolf", family: firefox, commands: []string{"librewolf"}},
	}

	switch runtime.GOOS {
	case "darwin":
		home, _ := os.UserHomeDir()
		paths := map[string][]string{
			"chrome":    {"/Applications/Google Chrome.app/Contents/MacOS/Google Chrome"},
			"edge":      {"/Applications/Microsoft Edge.app/Contents/MacOS/Microsoft Edge"},
			"brave":     {"/Applications/Brave Browser.app/Contents/MacOS/Brave Browser"},
			"chromium":  {"/Applications/Chromium.app/Contents/MacOS/Chromium"},
			"vivaldi":   {"/Applications/Vivaldi.app/Contents/MacOS/Vivaldi"},
			"opera":     {"/Applications/Opera.app/Contents/MacOS/Opera"},
			"firefox":   {"/Applications/Firefox.app/Contents/MacOS/firefox", "/Applications/Firefox.app/Contents/MacOS/firefox-bin"},
			"librewolf": {"/Applications/LibreWolf.app/Contents/MacOS/librewolf"},
		}
		for i := range items {
			items[i].paths = append(items[i].paths, paths[items[i].id]...)
			for _, systemPath := range paths[items[i].id] {
				if home != "" && strings.HasPrefix(systemPath, "/Applications/") {
					items[i].paths = append(items[i].paths, filepath.Join(home, systemPath[1:]))
				}
			}
		}
	case "windows":
		local := os.Getenv("LOCALAPPDATA")
		programFiles := []string{os.Getenv("PROGRAMFILES"), os.Getenv("PROGRAMFILES(X86)"), local}
		relative := map[string][]string{
			"chrome": {"Google/Chrome/Application/chrome.exe"}, "edge": {"Microsoft/Edge/Application/msedge.exe"},
			"brave": {"BraveSoftware/Brave-Browser/Application/brave.exe"}, "vivaldi": {"Vivaldi/Application/vivaldi.exe"},
			"chromium": {"Chromium/Application/chrome.exe"}, "opera": {"Programs/Opera/opera.exe", "Opera/launcher.exe"}, "firefox": {"Mozilla Firefox/firefox.exe"},
			"librewolf": {"LibreWolf/librewolf.exe"},
		}
		exeNames := map[string]string{"chrome": "chrome.exe", "edge": "msedge.exe", "brave": "brave.exe", "vivaldi": "vivaldi.exe", "opera": "opera.exe", "firefox": "firefox.exe", "librewolf": "librewolf.exe"}
		for i := range items {
			for _, base := range programFiles {
				for _, rel := range relative[items[i].id] {
					if base != "" {
						items[i].paths = append(items[i].paths, filepath.Join(base, filepath.FromSlash(rel)))
					}
				}
			}
			if exe := exeNames[items[i].id]; exe != "" {
				items[i].registry = []string{`HKCU\SOFTWARE\Microsoft\Windows\CurrentVersion\App Paths\` + exe, `HKLM\SOFTWARE\Microsoft\Windows\CurrentVersion\App Paths\` + exe}
			}
		}
	}
	return items
}

func firstExecutable(paths, commands, registryKeys []string) string {
	for _, path := range paths {
		if executable(path) {
			return path
		}
	}
	for _, command := range commands {
		if path, err := exec.LookPath(command); err == nil {
			return path
		}
	}
	if runtime.GOOS == "windows" {
		for _, key := range registryKeys {
			if path := queryWindowsAppPath(key); executable(path) {
				return path
			}
		}
	}
	return ""
}

func executable(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

func queryWindowsAppPath(key string) string {
	out, err := exec.Command("reg", "query", key, "/ve").Output()
	if err != nil {
		return ""
	}
	for _, line := range strings.Split(string(out), "\n") {
		if index := strings.Index(line, "REG_SZ"); index >= 0 {
			return strings.TrimSpace(line[index+len("REG_SZ"):])
		}
	}
	return ""
}
