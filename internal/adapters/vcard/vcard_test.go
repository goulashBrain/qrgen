package vcard

import (
	"strings"
	"testing"

	"github.com/goulashBrain/qrgen/internal/core/domain"
)

func TestEncodeEscapes(t *testing.T) {
	out, _ := New().Encode(domain.Contact{FirstName: "Leonardo", LastName: "Simoncelli", Org: "Dolomiti Energia; ICT"})
	if !strings.Contains(out, `ORG:Dolomiti Energia\; ICT`+"\r\n") {
		t.Fatalf("ORG not escaped:\n%s", out)
	}
	if !strings.HasPrefix(out, "BEGIN:VCARD\r\nVERSION:3.0\r\n") || !strings.HasSuffix(out, "END:VCARD\r\n") {
		t.Fatalf("bad envelope:\n%s", out)
	}
}
