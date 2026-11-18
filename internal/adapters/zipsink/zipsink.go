// Package zipsink implements ports.CodeSink by writing entries into a ZIP stream.
package zipsink

import (
	"archive/zip"
	"context"
	"io"

	"github.com/goulashBrain/qrgen/internal/core/domain"
	"github.com/goulashBrain/qrgen/internal/core/ports"
)

type Sink struct{ zw *zip.Writer }

var _ ports.CodeSink = (*Sink)(nil)

func New(w io.Writer) *Sink { return &Sink{zw: zip.NewWriter(w)} }

func (s *Sink) Save(_ context.Context, code domain.QRCode) error {
	f, err := s.zw.Create(code.Name)
	if err != nil {
		return err
	}
	_, err = f.Write(code.PNG)
	return err
}

// Close finalizes the archive; call it after the last Save.
func (s *Sink) Close() error { return s.zw.Close() }
