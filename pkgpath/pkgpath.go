package pkgpath

import (
	"reflect"
	"strings"
)

func OfType(v interface{}) string {
	t := reflect.TypeOf(v)
	if t == nil {
		return ""
	}
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	p := t.PkgPath() + "/" + t.Name()
	return stripHostname(p)
}

func stripHostname(p string) string {
	idx := strings.IndexRune(p, '/')
	if idx > 0 && strings.ContainsRune(p, '.') {
		p = p[idx+1:]
	}
	return p
}
