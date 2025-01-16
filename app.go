package main

import (
	"context"
	"fmt"

	"github.com/wailsapp/wails/v2/pkg/menu"
	"github.com/wailsapp/wails/v2/pkg/menu/keys"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// App struct
type App struct {
	ctx    context.Context
	config Config
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	a.config = config
}

func (a *App) applicationMenu() *menu.Menu {
	AppMenu := menu.NewMenu()
	FileMenu := AppMenu.AddSubmenu("File")
	FileMenu.AddText("&Open", keys.CmdOrCtrl("o"), func(_ *menu.CallbackData) {
	})
	FileMenu.AddSeparator()
	FileMenu.AddText("Quit", keys.CmdOrCtrl("q"), func(_ *menu.CallbackData) {
		// runtime.qu(a.ctx)
	})

	return AppMenu

}

// Greet returns a greeting for the given name
func (a *App) Greet(name string) string {
	return fmt.Sprintf("Hello %s, It's show time!", name)
}

func (a *App) LoadConfig() Config {
	return app.config
}

func (a *App) SaveConfig(config Config) error {
	a.config = config
	syncConfigFolders(config)
	err := writeConfig(config)
	return err
}

func (a *App) OpenDirectory() (string, error) {
	dir, err := runtime.OpenDirectoryDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "Select a directory",
	})
	return dir, err
}
