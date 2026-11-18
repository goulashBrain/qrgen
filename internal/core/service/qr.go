// Package service implements the use cases defined by ports.QRService.
// It orchestrates driven ports and contains no infrastructure code.
package service

import (
	"context"
	"fmt"

	"github.com/goulashBrain/qrgen/internal/core/domain"
	"github.com/goulashBrain/qrgen/internal/core/ports"
)

type QRService struct {
	payload  ports.PayloadEncoder
	renderer ports.ImageRenderer
	maxBatch int // 0 = unlimited
}

// Compile-time check that QRService satisfies the driving port.
var _ ports.QRService = (*QRService)(nil)

func NewQrService(payload ports.PayloadEncoder, renderer ports.ImageRenderer, maxBatch int) *QRService {
	return &QRService{payload: payload, renderer: renderer, maxBatch: maxBatch}
}

func (s *QRService) Generate(_ context.Context, c domain.Contact) (domain.QRCode, error) {
	return s.build(c, 1)
}

func (s *QRService) GenerateAll(ctx context.Context, repo ports.ContactRepository, sink ports.CodeSink) (int, error) {
	contacts, err := repo.List(ctx)
	if err != nil {
		return 0, err
	}
	if len(contacts) == 0 {
		return 0, domain.ErrNoContacts
	}
	if s.maxBatch > 0 && len(contacts) > s.maxBatch {
		return 0, fmt.Errorf("%w: %d (max %d)", domain.ErrTooManyContacts, len(contacts), s.maxBatch)
	}

	for i, c := range contacts {
		if err := ctx.Err(); err != nil {
			return i, err
		}
		code, err := s.build(c, i+1)
		if err != nil {
			return i, fmt.Errorf("contact %d: %w", i+1, err)
		}
		if err := sink.Save(ctx, code); err != nil {
			return i, fmt.Errorf("saving %s: %w", code.Name, err)
		}
	}
	return len(contacts), nil
}

func (s *QRService) build(c domain.Contact, n int) (domain.QRCode, error) {
	c = c.Normalize()
	if err := c.Validate(); err != nil {
		return domain.QRCode{}, err
	}
	payload, err := s.payload.Encode(c)
	if err != nil {
		return domain.QRCode{}, fmt.Errorf("encoding payload: %w", err)
	}
	png, err := s.renderer.Render(payload)
	if err != nil {
		return domain.QRCode{}, fmt.Errorf("rendering QR: %w", err)
	}
	return domain.QRCode{Name: fmt.Sprintf("%03d_%s.png", n, c.Slug()), PNG: png}, nil
}
