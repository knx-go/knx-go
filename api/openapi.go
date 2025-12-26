package api

import (
	"embed"
	"io/fs"
	"net/http"
)

//go:embed openapi.yaml swagger/*
var embedded embed.FS

func (a api) OpenAPIHandler() (http.Handler, error) {
	root, err := fs.Sub(embedded, ".")
	if err != nil {
		return nil, err
	}

	swagger, err := fs.Sub(embedded, "swagger")
	if err != nil {
		return nil, err
	}

	rootFS := http.FileServer(http.FS(root))
	swaggerFS := http.FileServer(http.FS(swagger))

	mux := http.NewServeMux()

	mux.Handle("/openapi/", http.StripPrefix("/openapi/",
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "openapi.yaml" {
				rootFS.ServeHTTP(w, r)
				return
			}
			swaggerFS.ServeHTTP(w, r)
		}),
	))

	return mux, nil
}
