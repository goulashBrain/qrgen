// Command qrgen is the composition root: it reads config, wires adapters
// into the core service, and starts either the CLI or the HTTP server.
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/goulashBrain/qrgen/internal/adapters/cli"
	"github.com/goulashBrain/qrgen/internal/adapters/csvrepo"
	"github.com/goulashBrain/qrgen/internal/adapters/httpapi"
	"github.com/goulashBrain/qrgen/internal/adapters/qrpng"
	"github.com/goulashBrain/qrgen/internal/adapters/vcard"
	"github.com/goulashBrain/qrgen/internal/core/service"
)

const maxBatch = 1000

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	showVersion := flag.Bool("version", false, "print version and exit")
	addr := flag.String("addr", envOr("QRGEN_ADDR", ":8080"), "listen address (server mode)")
	in := flag.String("in", "", "CLI mode: input folder with CSV files (empty = server mode)")
	out := flag.String("out", "./qrcodes", "CLI mode: output folder")
	sepStr := flag.String("sep", envOr("QRGEN_SEP", ";"), "CSV separator")
	size := flag.Int("size", 512, "QR image size in px")
	flag.Parse()

	sep, err := csvrepo.ParseSeparator(*sepStr)
	if err != nil {
		log.Fatal(err)
	}

	if *showVersion {
		fmt.Printf("qrgen %s (commit %s, built %s)\n", version, commit, date)
		return
	}
	log.Printf("qrgen %s starting", version)

	// Wiring: driven adapters -> core service.
	svc := service.NewQrService(vcard.New(), qrpng.New(*size), maxBatch)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// Driving adapter #1: CLI
	if *in != "" {
		if err := cli.Run(ctx, svc, cli.Config{InputDir: *in, OutputDir: *out, Separator: sep}); err != nil {
			log.Fatal(err)
		}
		return
	}

	// Driving adapter #2: HTTP
	if err := serve(ctx, *addr, httpapi.New(svc, sep)); err != nil {
		log.Fatal(err)
	}
}

func serve(ctx context.Context, addr string, h http.Handler) error {
	srv := &http.Server{
		Addr:              addr,
		Handler:           h,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      30 * time.Second,
	}

	errCh := make(chan error, 1)
	go func() {
		log.Printf("listening on %s", addr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
		close(errCh)
	}()

	select {
	case err := <-errCh:
		return err
	case <-ctx.Done(): // SIGTERM from k8s/systemd or Ctrl+C
	}

	log.Print("shutting down")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	return srv.Shutdown(shutdownCtx)
}

func envOr(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
