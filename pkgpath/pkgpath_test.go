package pkgpath

import (
	"encoding/base64"
	"testing"

	"github.com/goburrow/gol"
)

func TestOfTypeStandardPackage(t *testing.T) {
	//	p := NameOf((*base64.Encoding)(nil))
	p := OfType(base64.Encoding{})
	expected := "encoding/base64/Encoding"
	if expected != p {
		t.Fatalf("unexpected name: %v", p)
	}
}

func TestOfTypeExternalPackage(t *testing.T) {
	p := OfType((*gol.DefaultLogger)(nil))
	expected := "goburrow/gol/DefaultLogger"
	if expected != p {
		t.Fatalf("unexpected name: %v", p)
	}
}
