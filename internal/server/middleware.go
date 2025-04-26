package server

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type flushWrapper struct{ gin.ResponseWriter }

func (w flushWrapper) Flush() {
	if f, ok := w.ResponseWriter.(http.Flusher); ok {
		f.Flush()
	}
}

func EnsureFlusher() gin.HandlerFunc {
	return func(c *gin.Context) {
		if _, ok := c.Writer.(http.Flusher); !ok {
			c.Writer = flushWrapper{c.Writer} // wrap once
		}
		c.Next()
	}
}