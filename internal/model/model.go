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

type Field struct {
	Label string `json:"label"`
	Value string `json:"value"`
}

type FileEntry struct {
	Name     string `json:"name"`
	Size     int64  `json:"size"`
	Modified string `json:"modified"`
	IsDir    bool   `json:"isDir"`
}

// Site is an origin that stores data in the profile; Modified is the newest
// change to that data (a proxy for last use, since access times are unreliable).
type Site struct {
	Origin   string `json:"origin"`
	Modified string `json:"modified"`
}

// HistorySite aggregates browsing history per host.
type HistorySite struct {
	Host      string `json:"host"`
	Visits    int    `json:"visits"`
	LastVisit string `json:"lastVisit"`
}

// ProfileDetails is a read-only summary of a profile; passwords and cookie values are never read.
type ProfileDetails struct {
	History    []HistorySite `json:"history"`
	Fields     []Field     `json:"fields"`
	Extensions []string    `json:"extensions"`
	Sites      []Site      `json:"sites"`
	Files      []FileEntry `json:"files"`
	Notes      []string    `json:"notes"`
}
