package middleware

import (
	"net/http"
	"strings"
)

func uiMiddleware(uiDir, apiBasePath string, api http.Handler) http.Handler {
	ui := http.FileServer(http.Dir(uiDir))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, apiBasePath) || r.URL.Path == "/swagger.json" {
			api.ServeHTTP(w, r)
		} else {
			ui.ServeHTTP(w, r)
		}
	})
}

// MaybeServeUI adds UI middleware if options.UIPath is set.
func MaybeServeUI(apiBasePath string, uiDir string, handler http.Handler) http.Handler {
	if len(uiDir) == 0 {
		return handler
	}
	return uiMiddleware(uiDir, apiBasePath, handler)
}
