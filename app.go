package main

import (
	"context"
	"errors"

	"github.com/pairs/chrome-profile-viewer/internal/browser"
	"github.com/pairs/chrome-profile-viewer/internal/inspect"
	"github.com/pairs/chrome-profile-viewer/internal/launcher"
	"github.com/pairs/chrome-profile-viewer/internal/model"
	"github.com/pairs/chrome-profile-viewer/internal/profile"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type App struct {
	ctx context.Context
}

func NewApp() *App { return &App{} }

func (a *App) startup(ctx context.Context) { a.ctx = ctx }

func (a *App) SelectDirectory() (string, error) {
	if a.ctx == nil {
		return "", errors.New("uygulama henüz hazır değil")
	}
	return runtime.OpenDirectoryDialog(a.ctx, runtime.OpenDialogOptions{
		Title:           "Tarayıcı profillerinin bulunduğu klasörü seçin",
		ShowHiddenFiles: true,
		ResolvesAliases: true,
	})
}

func (a *App) DetectBrowsers() []model.Browser { return browser.Detect() }

func (a *App) ScanDirectory(path string) (model.ScanResult, error) {
	return profile.Scan(path, browser.Detect())
}

func (a *App) InspectProfile(selected model.Profile) (model.ProfileDetails, error) {
	return inspect.Inspect(selected)
}

func (a *App) LaunchProfile(selected model.Profile, browserID string, isolated bool) (model.LaunchResult, error) {
	return launcher.New(browser.Detect()).Launch(selected, browserID, isolated)
}
