package model

// Family describes browsers that share the same profile layout and launch flags.
type Family string

const (
	FamilyChromium Family = "chromium"
	FamilyFirefox  Family = "firefox"
)

type Browser struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	Family     Family `json:"family"`
	Executable string `json:"executable"`
}

type Profile struct {
	ID                 string   `json:"id"`
	Name               string   `json:"name"`
	Path               string   `json:"path"`
	UserDataDir        string   `json:"userDataDir"`
	DirectoryName      string   `json:"directoryName"`
	Family             Family   `json:"family"`
	ProductHint        string   `json:"productHint"`
	SuggestedBrowserID string   `json:"suggestedBrowserId"`
	Locked             bool     `json:"locked"`
	Evidence           []string `json:"evidence"`
}

type ScanResult struct {
	Root     string    `json:"root"`
	Profiles []Profile `json:"profiles"`
	Browsers []Browser `json:"browsers"`
	Warnings []string  `json:"warnings"`
}

type LaunchResult struct {
	PID          int      `json:"pid"`
	Browser      string   `json:"browser"`
	Profile      string   `json:"profile"`
	EffectiveDir string   `json:"effectiveDir"`
	Isolated     bool     `json:"isolated"`
	Arguments    []string `json:"arguments"`
}
