// Package vcard implements ports.PayloadEncoder using vCard 3.0.
package vcard

import (
	"fmt"
	"strings"

	"github.com/goulashBrain/qrgen/internal/core/domain"
	"github.com/goulashBrain/qrgen/internal/core/ports"
)

type Encoder struct{}

var _ ports.PayloadEncoder = Encoder{}

func New() Encoder { return Encoder{} }

var escaper = strings.NewReplacer(`\`, `\\`, ";", `\;`, ",", `\,`, "\r\n", `\n`, "\n", `\n`)

func esc(s string) string { return escaper.Replace(strings.TrimSpace(s)) }

func (Encoder) Encode(c domain.Contact) (string, error) {
	var b strings.Builder
	line := func(s string) { b.WriteString(s); b.WriteString("\r\n") }

	line("BEGIN:VCARD")
	line("VERSION:3.0")
	line(fmt.Sprintf("N:%s;%s;;;", esc(c.LastName), esc(c.FirstName)))
	line("FN:" + esc(c.FullName()))
	if c.Org != "" {
		line("ORG:" + esc(c.Org))
	}
	if c.Title != "" {
		line("TITLE:" + esc(c.Title))
	}
	if c.Phone != "" {
		line("TEL;TYPE=WORK,VOICE:" + esc(c.Phone))
	}
	if c.Email != "" {
		line("EMAIL;TYPE=INTERNET:" + esc(c.Email))
	}
	line("END:VCARD")
	return b.String(), nil
}
