package main

import (
	"embed"

	"companion/internal/window"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/windows"
)

//go:embed all:frontend/dist
var assets embed.FS

// Version is stamped at build time: -ldflags "-X main.Version=0.1.0".
// Dev builds report "dev".
var Version = "dev"

func main() {
	// Create an instance of the app structure
	app := NewApp()
	app.version = Version

	// Create application with options
	err := wails.Run(&options.App{
		Title:       "Wortel AI",
		Width:       window.PetSize.W,
		Height:      window.PetSize.H,
		MinWidth:    120,
		MinHeight:   120,
		MaxWidth:    460,
		MaxHeight:   680,
		Frameless:   true,
		AlwaysOnTop: true,
		AssetServer: &assetserver.Options{
			Assets:  assets,
			Handler: app.assetHandler(),
		},
		// Files dropped on an element carrying `--wails-drop-target: drop`
		// reach runtime.OnFileDrop as absolute paths; elsewhere they are
		// ignored (main.tsx also cancels the WebView's own file navigation).
		DragAndDrop:      &options.DragAndDrop{EnableFileDrop: true},
		BackgroundColour: &options.RGBA{R: 0, G: 0, B: 0, A: 0},
		Windows: &windows.Options{
			WebviewIsTransparent: true,
			// A transparent webview alone shows the window's own black
			// class brush. Translucent drops the window's redirection
			// surface, and backdrop None skips Mica/acrylic, so the desktop
			// shows through around the pet (ADR-041).
			WindowIsTranslucent:               true,
			BackdropType:                      windows.None,
			DisableFramelessWindowDecorations: true,
			WindowClassName:                   window.ClassName,
		},
		OnStartup:  app.startup,
		OnShutdown: app.shutdown,
		Bind: []interface{}{
			app,
		},
	})

	if err != nil {
		fatalDialog(err)
	}
}
