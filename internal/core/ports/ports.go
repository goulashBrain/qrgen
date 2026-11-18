// Package ports defines the hexagon's boundaries.
//
// Driving (primary) port:  QRService        – what the app offers (used by CLI, HTTP).
// Driven (secondary) ports: ContactRepository, PayloadEncoder, ImageRenderer, CodeSink
//
//	– what the app needs (implemented by adapters).
package ports

import (
	"context"

	"github.com/goulashBrain/qrgen/internal/core/domain"
)

// ---------- driving port ----------

// QRService is the application's use-case API.
type QRService interface {
	// Generate builds a single QR code for one contact.
	Generate(ctx context.Context, c domain.Contact) (domain.QRCode, error)
	// GenerateAll reads every contact from repo and stores a QR code for each in sink.
	GenerateAll(ctx context.Context, repo ContactRepository, sink CodeSink) (int, error)
}

// ---------- driven ports ----------

// ContactRepository provides contacts (CSV folder, uploaded CSV, DB, ...).
type ContactRepository interface {
	List(ctx context.Context) ([]domain.Contact, error)
}

// PayloadEncoder turns a contact into the text stored in the QR (vCard, MeCard, ...).
type PayloadEncoder interface {
	Encode(c domain.Contact) (string, error)
}

// ImageRenderer turns a payload into an image (PNG via go-qrcode, SVG, ...).
type ImageRenderer interface {
	Render(payload string) ([]byte, error)
}

// CodeSink stores generated codes (folder, ZIP stream, S3, ...).
type CodeSink interface {
	Save(ctx context.Context, code domain.QRCode) error
}
