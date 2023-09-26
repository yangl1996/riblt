package riblt

import (
	"testing"
)

func BenchmarkMapping(b *testing.B) {
	m := randomMapping{123456789, 0}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.nextIndex()
	}
}
