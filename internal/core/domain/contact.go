// Package domain holds the core entities and business rules.
// It has no dependencies on infrastructure (no CSV, HTTP, filesystem, QR libs).
package domain

import (
	"regexp"
	"strings"
)

// Contact is the core entity: a person to be encoded in a QR code.
type Contact struct {
	FirstName string
	LastName  string
	Org       string
	Title     string
	Phone     string
	Email     string
}

// Normalize returns a copy with all fields trimmed.
func (c Contact) Normalize() Contact {
	t := strings.TrimSpace
	return Contact{
		FirstName: t(c.FirstName),
		LastName:  t(c.LastName),
		Org:       t(c.Org),
		Title:     t(c.Title),
		Phone:     t(c.Phone),
		Email:     t(c.Email),
	}
}

// IsBlank reports whether every field is empty (e.g. an empty CSV row).
func (c Contact) IsBlank() bool {
	n := c.Normalize()
	return n == Contact{}
}

// Validate enforces the business rule: a contact needs at least a name.
func (c Contact) Validate() error {
	if strings.TrimSpace(c.FirstName) == "" && strings.TrimSpace(c.LastName) == "" {
		return ErrEmptyContact
	}
	return nil
}

// FullName returns "First Last", trimmed.
func (c Contact) FullName() string {
	return strings.TrimSpace(strings.TrimSpace(c.FirstName) + " " + strings.TrimSpace(c.LastName))
}

var unsafeChars = regexp.MustCompile(`[^a-z0-9_-]+`)

// Slug returns a filesystem-safe identifier like "simoncelli_leonardo".
func (c Contact) Slug() string {
	s := strings.ToLower(c.LastName + "_" + c.FirstName)
	s = strings.Trim(unsafeChars.ReplaceAllString(s, "_"), "_")
	if s == "" {
		return "contact"
	}
	return s
}
