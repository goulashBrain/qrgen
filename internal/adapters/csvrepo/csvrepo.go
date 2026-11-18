// Package csvrepo implements ports.ContactRepository on top of CSV data.
package csvrepo

import (
	"context"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/goulashBrain/qrgen/internal/core/domain"
	"github.com/goulashBrain/qrgen/internal/core/ports"
)

var (
	_ ports.ContactRepository = (*StreamRepository)(nil)
	_ ports.ContactRepository = (*DirRepository)(nil)
)

// ---------- single stream (e.g. HTTP upload) ----------

// StreamRepository reads contacts from one CSV stream. It is single-use.
type StreamRepository struct {
	r   io.Reader
	sep rune
}

func NewStream(r io.Reader, sep rune) *StreamRepository {
	return &StreamRepository{r: r, sep: sep}
}

func (s *StreamRepository) List(context.Context) ([]domain.Contact, error) {
	return parse(s.r, s.sep)
}

// ---------- folder of *.csv files (CLI) ----------

type DirRepository struct {
	dir string
	sep rune
}

func NewDir(dir string, sep rune) *DirRepository {
	return &DirRepository{dir: dir, sep: sep}
}

func (d *DirRepository) List(ctx context.Context) ([]domain.Contact, error) {
	files, err := filepath.Glob(filepath.Join(d.dir, "*.csv"))
	if err != nil {
		return nil, err
	}
	if len(files) == 0 {
		return nil, fmt.Errorf("%w: no CSV files in %s", domain.ErrInvalidInput, d.dir)
	}

	var all []domain.Contact
	for _, path := range files {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		contacts, err := parseFile(path, d.sep)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", filepath.Base(path), err)
		}
		all = append(all, contacts...)
	}
	return all, nil
}

func parseFile(path string, sep rune) ([]domain.Contact, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return parse(f, sep)
}

// ---------- parsing ----------

// Column names expected in the CSV header (case-insensitive, any order).
const (
	colFirst = "first_name"
	colLast  = "last_name"
	colOrg   = "org"
	colTitle = "title"
	colPhone = "phone"
	colEmail = "email"
)

func parse(r io.Reader, sep rune) ([]domain.Contact, error) {
	cr := csv.NewReader(r)
	cr.Comma = sep
	cr.TrimLeadingSpace = true
	cr.FieldsPerRecord = -1 // tolerate ragged rows

	header, err := cr.Read()
	if err != nil {
		return nil, invalid("reading header", err)
	}
	idx := make(map[string]int, len(header))
	for i, h := range header {
		h = strings.TrimPrefix(h, "\ufeff") // Excel BOM
		idx[strings.ToLower(strings.TrimSpace(h))] = i
	}
	if _, okF := idx[colFirst]; !okF {
		if _, okL := idx[colLast]; !okL {
			return nil, fmt.Errorf("%w: header must contain %q or %q", domain.ErrInvalidInput, colFirst, colLast)
		}
	}

	get := func(rec []string, key string) string {
		if i, ok := idx[key]; ok && i < len(rec) {
			return rec[i]
		}
		return ""
	}

	var out []domain.Contact
	for {
		rec, err := cr.Read()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return nil, invalid("reading row", err)
		}
		c := domain.Contact{
			FirstName: get(rec, colFirst),
			LastName:  get(rec, colLast),
			Org:       get(rec, colOrg),
			Title:     get(rec, colTitle),
			Phone:     get(rec, colPhone),
			Email:     get(rec, colEmail),
		}
		if c.IsBlank() {
			continue
		}
		out = append(out, c)
	}
	return out, nil
}

func invalid(what string, err error) error {
	return fmt.Errorf("%w: %s: %v", domain.ErrInvalidInput, what, err)
}

// ParseSeparator converts a flag/form value into a CSV separator rune.
func ParseSeparator(s string) (rune, error) {
	switch s {
	case "":
		return ';', nil
	case `\t`, "tab":
		return '\t', nil
	}
	r := []rune(s)
	if len(r) != 1 {
		return 0, fmt.Errorf("%w: separator must be a single character, got %q", domain.ErrInvalidInput, s)
	}
	return r[0], nil
}
