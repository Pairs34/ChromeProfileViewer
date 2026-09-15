package main

import (
	"embed"
	"log"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	app := NewApp()
	err := wails.Run(&options.App{
		Title: "Browser Profile Viewer",
		Width: 1180, Height: 760, MinWidth: 880, MinHeight: 600,
		AssetServer:      &assetserver.Options{Assets: assets},
		OnStartup:        app.startup,
		Bind:             []interface{}{app},
		BackgroundColour: options.NewRGB(9, 16, 29),
		DragAndDrop:      &options.DragAndDrop{EnableFileDrop: true},
	})
	if err != nil {
		log.Fatal(err)
	}
}
