// Package httpapi is a driving adapter exposing the QR service over HTTP.
package httpapi

import (
	"bytes"
	_ "embed"
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/goulashBrain/qrgen/internal/adapters/csvrepo"
	"github.com/goulashBrain/qrgen/internal/adapters/zipsink"
	"github.com/goulashBrain/qrgen/internal/core/domain"
	"github.com/goulashBrain/qrgen/internal/core/ports"
)

//go:embed web/index.html
var indexHTML []byte

const (
	maxFormBytes   = 64 << 10 // single-contact form
	maxUploadBytes = 1 << 20  // CSV upload
)

type Handler struct {
	svc        ports.QRService
	defaultSep rune
}

// New returns the router with all routes registered.
func New(svc ports.QRService, defaultSep rune) http.Handler {
	h := &Handler{svc: svc, defaultSep: defaultSep}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /{$}", h.index)
	mux.HandleFunc("GET /healthz", h.health)
	mux.HandleFunc("POST /api/qr", h.single)
	mux.HandleFunc("POST /api/batch", h.batch)
	return mux
}

func (h *Handler) index(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write(indexHTML)
}

func (h *Handler) health(w http.ResponseWriter, _ *http.Request) {
	_, _ = w.Write([]byte("ok"))
}

func (h *Handler) single(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, maxFormBytes)
	if err := r.ParseForm(); err != nil {
		writeError(w, err)
		return
	}
	c := domain.Contact{
		FirstName: r.FormValue("first_name"),
		LastName:  r.FormValue("last_name"),
		Org:       r.FormValue("org"),
		Title:     r.FormValue("title"),
		Phone:     r.FormValue("phone"),
		Email:     r.FormValue("email"),
	}
	code, err := h.svc.Generate(r.Context(), c)
	if err != nil {
		writeError(w, err)
		return
	}
	w.Header().Set("Content-Type", "image/png")
	w.Header().Set("Content-Disposition", fmt.Sprintf(`inline; filename="%s"`, code.Name))
	_, _ = w.Write(code.PNG)
}

func (h *Handler) batch(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, maxUploadBytes)
	if err := r.ParseMultipartForm(maxUploadBytes); err != nil {
		writeError(w, err)
		return
	}
	file, _, err := r.FormFile("file")
	if err != nil {
		writeError(w, fmt.Errorf("%w: missing 'file' field", domain.ErrInvalidInput))
		return
	}
	defer file.Close()

	sep := h.defaultSep
	if s := r.FormValue("sep"); s != "" {
		if sep, err = csvrepo.ParseSeparator(s); err != nil {
			writeError(w, err)
			return
		}
	}

	// Build the ZIP in memory so errors can still produce a proper status code.
	var buf bytes.Buffer
	sink := zipsink.New(&buf)
	if _, err := h.svc.GenerateAll(r.Context(), csvrepo.NewStream(file, sep), sink); err != nil {
		writeError(w, err)
		return
	}
	if err := sink.Close(); err != nil {
		writeError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/zip")
	w.Header().Set("Content-Disposition", `attachment; filename="qrcodes.zip"`)
	_, _ = w.Write(buf.Bytes())
}

// writeError maps domain/transport errors to HTTP status codes.
func writeError(w http.ResponseWriter, err error) {
	var tooBig *http.MaxBytesError
	switch {
	case errors.As(err, &tooBig):
		http.Error(w, "request too large", http.StatusRequestEntityTooLarge)
	case errors.Is(err, domain.ErrInvalidInput):
		http.Error(w, err.Error(), http.StatusBadRequest)
	default:
		log.Printf("internal error: %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
	}
}
