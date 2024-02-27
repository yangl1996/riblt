package riblt

import (
	"testing"
)

func BenchmarkMapping(b *testing.B) {
	m := randomMapping{prng: 123456789}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.nextIndex()
	}
}
