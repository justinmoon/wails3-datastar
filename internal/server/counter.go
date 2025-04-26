package server

import (
	"embed"
	"fmt"
	"net/http"
	"path"
	"demo/internal/views"
	"strings"
	"sync"

	"github.com/gin-gonic/gin"
	datastar "github.com/starfederation/datastar/sdk/go"
)

// embedFS is set from main.go and passed here
var embedFS embed.FS

// SetEmbedFS sets the embed filesystem to be used by the server
func SetEmbedFS(fs embed.FS) {
	embedFS = fs
}

// Counter maintains the counter state with thread safety
type Counter struct {
	mu  sync.Mutex
	val int
}

// NewRouter creates and configures a gin router with counter endpoints
func NewRouter() *gin.Engine {
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(EnsureFlusher()) // Add middleware to ensure Flush() is available

	c := &Counter{}

	// main page with templ rendering
	r.GET("/", func(ctx *gin.Context) {
		component := views.Counter(c.val)
		component.Render(ctx.Request.Context(), ctx.Writer)
	})

	// We don't need to serve assets from a special location anymore

	// Serve static files directly from public directory
	r.GET("/:file", func(ctx *gin.Context) {
		filename := ctx.Param("file")
		if filename == "wails.png" || filename == "javascript.svg" {
			data, err := embedFS.ReadFile(path.Join("frontend/public", filename))
			if err != nil {
				ctx.String(http.StatusNotFound, "File not found: %s", filename)
				return
			}

			// Set content type
			contentType := "text/plain"
			if path.Ext(filename) == ".png" || path.Ext(filename) == ".svg" {
				contentType = "image/" + path.Ext(filename)[1:]
			}

			ctx.Data(http.StatusOK, contentType, data)
		}
	})

	// POST /inc — returns ONE Datastar event with merge-fragments
	r.POST("/inc", func(ctx *gin.Context) {
		c.mu.Lock()
		c.val++
		v := c.val
		c.mu.Unlock()

		fmt.Printf("Counter value: %d\n", v)

		sse := datastar.NewSSE(ctx.Writer, ctx.Request)

		// Use templ component for the fragment
		component := views.CountFragment(v)

		// Render to string for datastar
		var sb strings.Builder
		component.Render(ctx.Request.Context(), &sb)

		// Use the rendered fragment
		_ = sse.MergeFragments(
			sb.String(),
			datastar.WithSelectorID("count"), // target the span
			datastar.WithMergeMorph(),        // default, explicit for clarity
		)
	})

	// Debug endpoint to check if service is running
	r.GET("/api/status", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{
			"status": "ok",
			"count":  c.val,
		})
	})

	return r
}

