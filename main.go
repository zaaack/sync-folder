package main

import (
	"context"
	"embed"
	"io"
	"os"
	"time"

	"github.com/energye/systray"
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

func runWindow() {
	defer recover()

	err := wails.Run(&options.App{
		Title:  "Sync Folder",
		Width:  1024,
		Height: 768,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 27, G: 38, B: 54, A: 1},
		OnStartup:        app.startup,
		OnShutdown: func(ctx context.Context) {
			app = nil
		},
		// Menu:             app.applicationMenu(),
		Bind: []interface{}{
			app,
		},
	})

	if err != nil {
		println("Error:", err.Error())
	}
}

func openWindow() {
	if app != nil {
		closeWindow()
	}
	// Create an instance of the app structure
	app = NewApp()
	// Create application with options
	go runWindow()
}

func closeWindow() {
	defer recover()
	if app == nil || app.ctx == nil {
		app = nil
		return
	}
	runtime.Quit(app.ctx)
	app = nil
}

var logs []string = make([]string, 0)

type LogWriter struct {
	io.Writer
}

var lastCheckLogTime int64 = 0

const DAY_IN_SECONDS = 24 * 60 * 60

func (w *LogWriter) Write(p []byte) (n int, err error) {
	logs = append(logs, string(p))
	if len(logs) > 100 {
		logs = logs[1:]
	}

	now := time.Now().Unix()

	if now-lastCheckLogTime > DAY_IN_SECONDS*3 {
		_, err = os.Stat(getLogPath() + ".old")
		if err == nil {
			os.Remove(getLogPath() + ".old")
		}
		os.Rename(getLogPath(), getLogPath()+".old")
		lastCheckLogTime = now
	}

	return len(p), nil
}

func main() {
	logFile, _ := os.OpenFile(getLogPath(), os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	logrus.AddHook(&writer.Hook{
		Writer:    logFile,
		LogLevels: []logrus.Level{logrus.InfoLevel},
	})

	logrus.AddHook(&writer.Hook{
		Writer:    &LogWriter{},
		LogLevels: []logrus.Level{logrus.InfoLevel},
	})
	syncConfigFolders(readConfig())

	// Create an instance of the app structure
	app = NewApp()
	// Create application with options
	runWindow()
	systray.Run(onReady, onExit)
}
