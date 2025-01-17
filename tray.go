package main

import (
	"embed"
	"fmt"
	"os"

	"github.com/energye/systray"
)

//go:embed build/trayicon.ico
var trayicon embed.FS

func toggleWindow() {
	defer recover()
	if app == nil {
		openWindow()
	} else {
		closeWindow()
	}
}

func onReady() {
	icon, err := trayicon.ReadFile("build/trayicon.ico")
	if err != nil {
		panic(err)
	}

	systray.SetIcon(icon)
	systray.SetTitle("Sync folders")
	// systray.SetTooltip("")
	systray.SetOnClick(func(menu systray.IMenu) {
		menu.ShowMenu()
	})
	systray.SetOnRClick(func(menu systray.IMenu) {
		menu.ShowMenu()
		fmt.Println("SetOnRClick")
	})
	mShow := systray.AddMenuItem("Toggle", "Toggle the window")
	mShow.Click(func() {
		toggleWindow()
	})
	mQuit := systray.AddMenuItem("Quit", "Quit the whole app")
	mQuit.Click(func() {
		systray.Quit()
		os.Exit(0)
	})
}

func onExit() {
	// clean up here
	systray.Quit()
	os.Exit(0)
}
