package jdextract

import (
	"context"
	"fmt"
	"net/http"
	"slices"
)

// Serve starts the web UI on the given port and blocks until ctx is cancelled,
// at which point it performs a graceful shutdown via http.Server.Shutdown.
// port is a bare number (e.g. "8080"); the colon prefix is added internally.
// All routes pass through csrfMiddleware before reaching their handlers.
func (a *App) Serve(ctx context.Context, port string) error {
	if err := a.Setup(); err != nil {
		return fmt.Errorf("setup: %w", err)
	}
	mux := http.NewServeMux()
	a.registerRoutes(mux)

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: csrfMiddleware(port, mux),
	}

	errCh := make(chan error, 1)
	go func() {
		fmt.Printf("Serving on http://localhost:%s\n", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errCh <- err
		}
		close(errCh)
	}()

	select {
	case err := <-errCh:
		return err
	case <-ctx.Done():
		return srv.Shutdown(context.Background())
	}
}

// csrfMiddleware rejects requests where Origin is present but doesn't match
// the expected localhost origin, and requires Content-Type: application/json
// on POST/PATCH requests. curl and other non-browser clients (no Origin) pass through.
func csrfMiddleware(port string, next http.Handler) http.Handler {
	allowed := []string{
		"http://localhost:" + port,
		"http://127.0.0.1:" + port,
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if origin := r.Header.Get("Origin"); origin != "" {
			if !slices.Contains(allowed, origin) {
				http.Error(w, "forbidden: origin not allowed", http.StatusForbidden)
				return
			}
		}
		if r.Method == http.MethodPost || r.Method == http.MethodPatch {
			if r.Header.Get("Content-Type") != "application/json" {
				http.Error(w, "Content-Type must be application/json", http.StatusUnsupportedMediaType)
				return
			}
		}
		next.ServeHTTP(w, r)
	})
}
