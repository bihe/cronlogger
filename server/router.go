package server

import (
	"embed"
	"fmt"
	"io/fs"
	"net/http"
	"strings"
)

//go:embed assets/*
var content embed.FS

func SetupRoutes(mux *http.ServeMux, handler *CronLogHandler) {
	mux.HandleFunc("/", handler.StartPage())
	serveStaticDir(mux, "assets")
}

func serveStaticDir(mux *http.ServeMux, path string) {
	path, _ = strings.CutPrefix(path, "/")
	path, _ = strings.CutSuffix(path, "/")

	// Serve static files from the embedded filesystem.
	staticFS, err := fs.Sub(content, path)
	if err != nil {
		panic(fmt.Sprintf("Failed to create sub-filesystem: %v.", err))
	}
	mux.Handle("/"+path+"/", http.StripPrefix("/"+path+"/", http.FileServer(http.FS(staticFS))))
}
