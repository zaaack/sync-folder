package main

import (
	"fmt"
	"os"

	"github.com/energye/systray"
	"github.com/energye/systray/icon"
)

func onReady() {
	systray.SetIcon(icon.Data)
	systray.SetTitle("Sync folders")
	// systray.SetTooltip("")
	// systray.SetOnClick(func(menu systray.IMenu) {
	// 	fmt.Println("SetOnClick")
	// })
	// systray.SetOnDClick(func(menu systray.IMenu) {
	// 	fmt.Println("SetOnDClick")
	// })
	systray.SetOnRClick(func(menu systray.IMenu) {
		menu.ShowMenu()
		fmt.Println("SetOnRClick")
	})
	mShow := systray.AddMenuItem("Show", "Show the window")
	mShow.Click(func() {
		// runtime.Show(app.ctx)
		go openWindow()
	})
	mHide := systray.AddMenuItem("Hide", "Hide the window")
	mHide.Click(func() {
		closeWindow()
	})
	mQuit := systray.AddMenuItem("Quit", "Quit the whole app")
	mQuit.Click(func() {
		systray.Quit()
		os.Exit(0)
	})
}

func onExit() {
	// clean up here
	os.Exit(0)
}
