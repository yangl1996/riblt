package riblt

import (
	"testing"
)

func BenchmarkSketchAddSymbol(b *testing.B) {
	benches := []struct {
		name string
		size int
	}{
		{"m=1000", 1000},
		{"m=10000", 10000},
		{"m=100000", 100000},
		{"m=1000000", 1000000},
		{"m=10000000", 10000000},
	}
	for _, bench := range benches {
		s := make(Sketch[testSymbol], bench.size)
		b.Run(bench.name, func(b *testing.B) {
			b.SetBytes(testSymbolSize)
			for i := 0; i < b.N; i++ {
				s.AddSymbol(newTestSymbol(uint64(i)))
			}
		})
	}
}

