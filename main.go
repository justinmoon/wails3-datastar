package main

import (
	"embed"
	"log"
	"net/http"
	"os"

	"github.com/wailsapp/wails/v3/pkg/application"
	"demo/internal/server"
)

//go:embed all:frontend
var embedFS embed.FS

func main() {
	// Make embedFS available to the server package
	server.SetEmbedFS(embedFS)

	ginEngine := server.NewRouter()

	// middleware that always wraps responses with flushWriter
	// This ensures all responses can be flushed (for SSE and templ)

	// Important: Setting empty dev server URL to disable dev server check
	os.Setenv("FRONTEND_DEVSERVER_URL", "")

	app := application.New(application.Options{
		Name:        "Counter-DS",
		Description: "Counter demo with Datastar",
		Assets: application.AssetOptions{
			Handler: ginEngine, // serve bundle, handle API
			// No middleware - Wails AssetServer already implements http.Flusher
		},
		Mac: application.MacOptions{
			ApplicationShouldTerminateAfterLastWindowClosed: true,
		},
	})

	// Create a new window with the necessary options
	window := app.NewWebviewWindowWithOptions(application.WebviewWindowOptions{
		Title:  "Counter",
		URL:    "/", // Gin root
		Width:  500,
		Height: 400,
		Mac: application.MacWindow{
			InvisibleTitleBarHeight: 50,
			Backdrop:                application.MacBackdropTranslucent,
			TitleBar:                application.MacTitleBarHiddenInset,
		},
		BackgroundColour: application.NewRGB(27, 38, 54),
	})

	log.Println("Window created with ID:", window.ID())

	if err := app.Run(); err != nil {
		log.Fatal(err)
	}
}

// flushWriter is a wrapper that implements http.Flusher
type flushWriter struct {
	w http.ResponseWriter
}

func (fw *flushWriter) Header() http.Header {
	return fw.w.Header()
}

func (fw *flushWriter) Write(data []byte) (int, error) {
	return fw.w.Write(data)
}

func (fw *flushWriter) WriteHeader(statusCode int) {
	fw.w.WriteHeader(statusCode)
}

// Flush implements http.Flusher
func (fw *flushWriter) Flush() {
	if f, ok := fw.w.(http.Flusher); ok {
		f.Flush()
	}
	// If not a Flusher, we simply ignore the Flush request
}
