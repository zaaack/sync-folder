package main

import (
	"embed"
	"os"

	"github.com/sirupsen/logrus"
	writer "github.com/sirupsen/logrus/hooks/writer"
	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

//go:embed all:frontend/dist
var assets embed.FS

var app *App

func openWindow() {
	if app != nil {
		return
	}
	// Create an instance of the app structure
	app = NewApp()

	// Create application with options
	err := wails.Run(&options.App{
		Title:  "wails",
		Width:  1024,
		Height: 768,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 27, G: 38, B: 54, A: 1},
		OnStartup:        app.startup,
		// Menu:             app.applicationMenu(),
		Bind: []interface{}{
			app,
		},
	})

	if err != nil {
		println("Error:", err.Error())
	}
}

func closeWindow() {
	runtime.Quit(app.ctx)
	app = nil
}

func main() {
	logFile, _ := os.OpenFile(getLogPath(), os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	logrus.AddHook(&writer.Hook{
		Writer:    logFile,
		LogLevels: []logrus.Level{logrus.InfoLevel},
	})
	syncConfigFolders(config)
	openWindow()
	// systray.Run(onReady, onExit)
}
