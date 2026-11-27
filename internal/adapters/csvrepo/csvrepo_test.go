package csvrepo

import (
	"context"
	"strings"
	"testing"
)

func TestStreamRepository(t *testing.T) {
	in := "\ufeffFirst_Name;last_name;email\nAnna;Bianchi;l@x.it\n;;\n"
	got, err := NewStream(strings.NewReader(in), ';').List(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0].LastName != "Bianchi" || got[0].Email != "l@x.it" {
		t.Fatalf("unexpected result: %+v", got)
	}
}
