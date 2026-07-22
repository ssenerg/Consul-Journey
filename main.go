// Command consul-journey runs one instance of a self-clustering service.
//
// Every instance registers itself with the local Consul agent, publishes both
// an HTTP and a TTL health check, discovers all of its peers through Consul's
// health API using blocking queries, and participates in leader election via a
// Consul session + KV lock. Run several instances (each on its own port) and
// they will find one another and elect a single leader — killing the leader
// causes a survivor to take over automatically.
package main

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	cfg, err := LoadConfig(os.Args[1:])
	if err != nil {
		fmt.Fprintln(os.Stderr, "config error:", err)
		os.Exit(2)
	}

	log := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})).With("node", cfg.NodeID)

	node, err := NewNode(cfg, log)
	if err != nil {
		log.Error("startup failed", "err", err)
		os.Exit(1)
	}

	// Cancel the root context on SIGINT/SIGTERM for graceful shutdown.
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := node.Run(ctx); err != nil {
		log.Error("run failed", "err", err)
		os.Exit(1)
	}
	log.Info("shutdown complete")
}

// --- small shared helpers ---------------------------------------------------

// pid returns the current process id.
func pid() int { return os.Getpid() }

// short truncates an id for readable logs.
func short(id string) string {
	if len(id) <= 8 {
		return id
	}
	return id[:8]
}

// sleep waits for d or until ctx is cancelled. It returns true if the context
// was cancelled (the caller should stop).
func sleep(ctx context.Context, d time.Duration) bool {
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-ctx.Done():
		return true
	case <-t.C:
		return false
	}
}

// nextBackoff doubles a backoff duration up to a 30s ceiling.
func nextBackoff(d time.Duration) time.Duration {
	d *= 2
	if d > 30*time.Second {
		return 30 * time.Second
	}
	return d
}

// writeHTMLf writes a formatted HTML fragment to the response, ignoring errors
// (the client disconnecting mid-render is not actionable).
func writeHTMLf(w io.Writer, format string, args ...any) {
	_, _ = fmt.Fprintf(w, format, args...)
}
