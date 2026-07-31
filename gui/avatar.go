package gui

import (
	"context"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/windows"
)

// LaunchAvatarApp creates and runs the NEXA transparent floating avatar mascot window.
// Window properties: 320x320, Frameless: true, AlwaysOnTop: true, WebviewIsTransparent: true.
func LaunchAvatarApp(app *App) error {
	return wails.Run(&options.App{
		Title:            "NEXA Avatar",
		Width:            320,
		Height:           320,
		MinWidth:         200,
		MinHeight:        200,
		MaxWidth:         500,
		MaxHeight:        500,
		DisableResize:    true,
		Frameless:        true,
		AlwaysOnTop:      true,
		StartHidden:      false,
		BackgroundColour: &options.RGBA{R: 0, G: 0, B: 0, A: 0}, // Fully transparent
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
			WebviewIsTransparent:              true,
			WindowIsTranslucent:               true,
			DisableWindowIcon:                 false,
			DisableFramelessWindowDecorations: true,
			Theme:                             windows.Dark,
		},
	})
}
