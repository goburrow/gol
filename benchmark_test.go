package gol

import (
	"io/ioutil"
	"testing"
)

func BenchmarkGolWithoutFields(b *testing.B) {
	logger := NewFactory(ioutil.Discard).GetLogger("main")
	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			logger.Infof("go")
		}
	})
}

func BenchmarkGolWithFields(b *testing.B) {
	logger := NewFactory(ioutil.Discard).GetLogger("main")
	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			logger.Infof(
				`go {"int"=%q, "string"=%q, "float"=%q, "bool"=%q`,
				1, "two", 3.0, true)
		}
	})
}
