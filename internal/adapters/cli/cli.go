// Package cli is a driving adapter: batch-generates QR codes from a CSV folder.
package cli

import (
	"context"
	"log"

	"github.com/goulashBrain/qrgen/internal/adapters/csvrepo"
	"github.com/goulashBrain/qrgen/internal/adapters/fssink"
	"github.com/goulashBrain/qrgen/internal/core/ports"
)

type Config struct {
	InputDir  string
	OutputDir string
	Separator rune
}

func Run(ctx context.Context, svc ports.QRService, cfg Config) error {
	sink, err := fssink.New(cfg.OutputDir)
	if err != nil {
		return err
	}
	repo := csvrepo.NewDir(cfg.InputDir, cfg.Separator)

	n, err := svc.GenerateAll(ctx, repo, sink)
	if err != nil {
		return err
	}
	log.Printf("generated %d QR codes in %s", n, cfg.OutputDir)
	return nil
}
