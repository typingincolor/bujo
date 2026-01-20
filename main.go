package main

import (
	"context"
	"embed"
	"os"
	"path/filepath"

	"github.com/typingincolor/bujo/internal/adapter/wails"
	"github.com/typingincolor/bujo/internal/app"
	wailsrt "github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/linux"
	"github.com/wailsapp/wails/v2/pkg/options/mac"
	"github.com/wailsapp/wails/v2/pkg/options/windows"
)

//go:embed all:frontend/dist
var assets embed.FS

func getDefaultDBPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return "bujo.db"
	}

	bujoDir := filepath.Join(home, ".bujo")
	if err := os.MkdirAll(bujoDir, 0755); err != nil {
		return "bujo.db"
	}

	return filepath.Join(bujoDir, "bujo.db")
}

func main() {
	factory := app.NewServiceFactory()
	services, cleanup, err := factory.Create(context.Background(), getDefaultDBPath())
	if err != nil {
		println("Error creating services:", err.Error())
		os.Exit(1)
	}
	defer cleanup()

	wailsApp := wails.NewApp(services)

	err = wailsrt.Run(&options.App{
		Title:  "Bujo",
		Width:  1024,
		Height: 768,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 27, G: 38, B: 54, A: 1},
		OnStartup:        wailsApp.Startup,
		DragAndDrop: &options.DragAndDrop{
			EnableFileDrop:     true,
			DisableWebViewDrop: false,
			CSSDropProperty:    "--wails-drop-target",
			CSSDropValue:       "drop",
		},
		Mac: &mac.Options{
			About: &mac.AboutInfo{
				Title:   "Bujo",
				Message: "A Bullet Journal for your terminal and desktop",
			},
		},
		Windows: &windows.Options{},
		Linux:   &linux.Options{},
		Bind: []interface{}{
			wailsApp,
		},
	})

	if err != nil {
		println("Error:", err.Error())
	}
}
