package api

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"
)

func Run(ctx context.Context, opts Options) error {
	// API
	a := api{
		ctx:     ctx,
		store:   opts.Store,
		catalog: opts.Catalog,
		tunnel:  opts.Tunnel,
	}
	openapiHandler, err := a.OpenAPIHandler()
	if err != nil {
		log.Fatalf("failed to init openapi handler: %v", err)
	}
	healthzHandler, err := a.HealthzAPIHandler()
	if err != nil {
		log.Fatalf("failed to init openapi handler: %v", err)
	}
	v1Handler, err := a.V1APIHandler()
	if err != nil {
		log.Fatalf("failed to init openapi handler: %v", err)
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/openapi/", http.StatusMovedPermanently)
	})
	mux.Handle("/openapi/", openapiHandler)
	mux.Handle("/healthz", healthzHandler)
	mux.Handle("/v1/", v1Handler)

	// HTTP
	listenAddr := opts.ListenAddress
	if listenAddr == "" {
		listenAddr = ":8080"
	}

	server := &http.Server{
		Addr:    listenAddr,
		Handler: mux,
	}

	go a.listen(ctx)

	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := server.Shutdown(shutdownCtx); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Printf("HTTP server shutdown error: %v", err)
		}
	}()
	log.Println("Listening on :8080")

	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("HTTP server error: %w", err)
	}
	log.Println("HTTP server shutdown complete")
	return nil
}
