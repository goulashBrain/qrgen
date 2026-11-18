// Package qrpng implements ports.ImageRenderer using github.com/skip2/go-qrcode.
package qrpng

import (
	qrcode "github.com/skip2/go-qrcode"

	"github.com/goulashBrain/qrgen/internal/core/ports"
)

type Renderer struct {
	size  int
	level qrcode.RecoveryLevel
}

var _ ports.ImageRenderer = (*Renderer)(nil)

func New(size int) *Renderer {
	return &Renderer{size: size, level: qrcode.Medium}
}

func (r *Renderer) Render(payload string) ([]byte, error) {
	return qrcode.Encode(payload, r.level, r.size)
}
