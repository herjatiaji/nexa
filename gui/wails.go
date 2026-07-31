package gui

import (
	"context"
	"embed"

	"github.com/heraji/jarvis/config"
	"github.com/heraji/jarvis/core"
	"github.com/heraji/jarvis/memory"
	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/windows"
)

//go:embed all:frontend/dist
var assets embed.FS

// LaunchDesktopApp creates and runs the NEXA Wails native desktop window.
func LaunchDesktopApp(agent *core.Agent, cfg *config.Config, memStore *memory.MemoryStore) error {
	app := NewApp(agent, cfg, memStore)

	return wails.Run(&options.App{
		Title:            "NEXA — Personal AI Assistant",
		Width:            1280,
		Height:           820,
		MinWidth:         900,
		MinHeight:        600,
		DisableResize:    false,
		Frameless:        false,
		StartHidden:      false,
		BackgroundColour: &options.RGBA{R: 8, G: 12, B: 20, A: 255},
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		OnStartup: func(ctx context.Context) {
			app.Startup(ctx)
			app.AddTrayIcon()
			app.RegisterGlobalHotkey()
		},
		OnShutdown: func(ctx context.Context) {
			app.RemoveTrayIcon()
		},
		Bind: []interface{}{
			app,
		},
		Windows: &windows.Options{
			WebviewIsTransparent:              false,
			WindowIsTranslucent:               false,
			DisableWindowIcon:                 false,
			DisableFramelessWindowDecorations: false,
			WebviewUserDataPath:               "",
			Theme:                             windows.Dark,
		},
	})
}
