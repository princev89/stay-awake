package main

import (
	"embed"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	app := NewApp()

	err := wails.Run(&options.App{
		Title:             "Stay Awake",
		Width:             320,
		Height:            440,
		DisableResize:     true,
		Fullscreen:        false,
		Frameless:         false,
		MinWidth:          320,
		MinHeight:         440,
		MaxWidth:          320,
		MaxHeight:         440,
		BackgroundColour:  &options.RGBA{R: 12, G: 13, B: 18, A: 255},
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		OnStartup:     app.startup,
		OnDomReady:    app.domReady,
		OnBeforeClose: app.beforeClose,
		OnShutdown:    app.shutdown,
		Bind: []interface{}{
			app,
		},
	})

	if err != nil {
		println("Error:", err.Error())
	}
}
