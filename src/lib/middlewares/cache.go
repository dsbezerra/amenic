package middlewares

import (
	"net/http"

	"github.com/gin-contrib/static"
	"github.com/gin-gonic/gin"
)

// StaticWithCache little modification of static.Serve that adds Cache-Control header
// to all static files.
func StaticWithCache(urlPrefix string, fs static.ServeFileSystem) gin.HandlerFunc {
	fileserver := http.FileServer(fs)
	if urlPrefix != "" {
		fileserver = http.StripPrefix(urlPrefix, fileserver)
	}
	return func(c *gin.Context) {
		if fs.Exists(urlPrefix, c.Request.URL.Path) {
			// Define cache values here.
			c.Writer.Header().Set("Cache-Control", "public, max-age=21600")

			fileserver.ServeHTTP(c.Writer, c.Request)
			c.Abort()
		}
	}
}
