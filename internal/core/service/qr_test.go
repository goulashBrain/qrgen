package service_test

import (
	"context"
	"errors"
	"testing"

	"github.com/goulashBrain/qrgen/internal/core/domain"
	"github.com/goulashBrain/qrgen/internal/core/service"
)

// Fakes: the core is testable without CSV files, QR libs or disk.
type fakeRepo []domain.Contact

func (f fakeRepo) List(context.Context) ([]domain.Contact, error) { return f, nil }

type fakePayload struct{}

func (fakePayload) Encode(c domain.Contact) (string, error) { return c.FullName(), nil }

type fakeRenderer struct{}

func (fakeRenderer) Render(p string) ([]byte, error) { return []byte(p), nil }

type memSink struct{ codes []domain.QRCode }

func (m *memSink) Save(_ context.Context, c domain.QRCode) error {
	m.codes = append(m.codes, c)
	return nil
}

func TestGenerateAll(t *testing.T) {
	svc := service.NewQrService(fakePayload{}, fakeRenderer{}, 10)
	sink := &memSink{}
	repo := fakeRepo{{FirstName: "Anna", LastName: "Bianchi"}, {FirstName: "Mario", LastName: "Rossi"}}

	n, err := svc.GenerateAll(context.Background(), repo, sink)
	if err != nil {
		t.Fatal(err)
	}
	if n != 2 || len(sink.codes) != 2 {
		t.Fatalf("want 2 codes, got n=%d saved=%d", n, len(sink.codes))
	}
	if got := sink.codes[0].Name; got != "001_bianchi_anna.png" {
		t.Errorf("unexpected name %q", got)
	}
}

func TestGenerateRejectsEmptyContact(t *testing.T) {
	svc := service.NewQrService(fakePayload{}, fakeRenderer{}, 0)
	_, err := svc.Generate(context.Background(), domain.Contact{Email: "x@y.z"})
	if !errors.Is(err, domain.ErrInvalidInput) {
		t.Fatalf("want ErrInvalidInput, got %v", err)
	}
}

func TestGenerateAllTooMany(t *testing.T) {
	svc := service.NewQrService(fakePayload{}, fakeRenderer{}, 1)
	repo := fakeRepo{{FirstName: "a"}, {FirstName: "b"}}
	if _, err := svc.GenerateAll(context.Background(), repo, &memSink{}); !errors.Is(err, domain.ErrTooManyContacts) {
		t.Fatalf("want ErrTooManyContacts, got %v", err)
	}
}
