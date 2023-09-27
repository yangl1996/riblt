package riblt

import (
	"testing"
	"crypto/sha256"
)

func BenchmarkSketchAddSymbol(b *testing.B) {
	benches := []struct{
		name string
		size int
	}{
		{"1000", 1000},
		{"100000", 100000},
		{"10000000", 10000000},
	}
	for _, bench := range benches {
		s := make(Sketch[testSymbol], bench.size)
		b.Run(bench.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				s.AddSymbol(newTestSymbol(uint64(i)))
			}
		})
	}
}

func BenchmarkSHA256(b *testing.B) {
	for i := 0; i < b.N; i++ {
		t := newTestSymbol(uint64(i))
		sha256.Sum256(t[:])
	}
}
