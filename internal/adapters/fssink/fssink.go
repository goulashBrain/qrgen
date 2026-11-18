// Package fssink implements ports.CodeSink by writing PNG files to a folder.
package fssink

import (
	"context"
	"os"
	"path/filepath"

	"github.com/goulashBrain/qrgen/internal/core/domain"
	"github.com/goulashBrain/qrgen/internal/core/ports"
)

type Sink struct{ dir string }

var _ ports.CodeSink = (*Sink)(nil)

func New(dir string) (*Sink, error) {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, err
	}
	return &Sink{dir: dir}, nil
}

func (s *Sink) Save(_ context.Context, code domain.QRCode) error {
	return os.WriteFile(filepath.Join(s.dir, filepath.Base(code.Name)), code.PNG, 0o644)
}
