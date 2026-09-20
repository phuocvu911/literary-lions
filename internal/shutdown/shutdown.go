package shutdown

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"
)

// Graceful shutdown the HTTP server when the context is canceled (CTRL+C for example in this case).
func Graceful(ctx context.Context, server *http.Server, timeout time.Duration) error {
	<-ctx.Done()
	log.Println("Shutting down server...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("Server forced shutdown: %v", err)
	}
	log.Println("Process exited successfully.")
	return nil
}
