package riblt

import (
	"fmt"
	"encoding/binary"
	"github.com/dchest/siphash"
	"testing"
	"unsafe"
)

const testSymbolSize = 64

type testSymbol [testSymbolSize]byte

func (d testSymbol) XOR(t2 testSymbol) testSymbol {
	dw := (*[testSymbolSize / 8]uint64)(unsafe.Pointer(&d))
	t2w := (*[testSymbolSize / 8]uint64)(unsafe.Pointer(&t2))
	for i := 0; i < testSymbolSize/8; i++ {
		(*dw)[i] ^= (*t2w)[i]
	}
	return d
}

func (d testSymbol) Hash() uint64 {
	return siphash.Hash(567, 890, d[:])
}

func newTestSymbol(i uint64) testSymbol {
	data := testSymbol{}
	binary.LittleEndian.PutUint64(data[0:8], i)
	return data
}

func TestEncodeAndDecode(t *testing.T) {
	for _, scale := range []int{1, 2, 4, 10, 100, 1000, 10000, 20000} {
		t.Run(fmt.Sprintf("diff %d", scale * 10), func(t *testing.T) {
			enc := Encoder[testSymbol]{}
			dec := Decoder[testSymbol]{}
			local := make(map[uint64]struct{})
			remote := make(map[uint64]struct{})

			var nextId uint64
			nlocal := 5*scale
			nremote := 5*scale
			ncommon := 10*scale
			for i := 0; i < nlocal; i++ {
				s := newTestSymbol(nextId)
				nextId += 1
				dec.AddSymbol(s)
				local[s.Hash()] = struct{}{}
			}
			for i := 0; i < nremote; i++ {
				s := newTestSymbol(nextId)
				nextId += 1
				enc.AddSymbol(s)
				remote[s.Hash()] = struct{}{}
			}
			for i := 0; i < ncommon; i++ {
				s := newTestSymbol(nextId)
				nextId += 1
				enc.AddSymbol(s)
				dec.AddSymbol(s)
			}

			ncw := 0
			for {
				dec.AddCodedSymbol(enc.ProduceNextCodedSymbol())
				ncw += 1
				dec.TryDecode()
				if dec.Decoded() {
					break
				}
			}
			for _, v := range dec.Remote() {
				delete(remote, v.Hash)
			}
			for _, v := range dec.Local() {
				delete(local, v.Hash)
			}
			if len(remote) != 0 || len(local) != 0 {
				t.Errorf("missing symbols: %d remote and %d local", len(remote), len(local))
			}
			if !dec.Decoded() {
				t.Errorf("decoder not marked as decoded")
			}
			t.Logf("%d coded symbols used for decoding %d diffs overhead %.4f", ncw, nlocal+nremote, float64(ncw)/float64(nlocal+nremote))
		})
	}
}

func BenchmarkEncoding(b *testing.B) {
	n := 10000
	m := 15000
	enc := Encoder[testSymbol]{}
	data := []testSymbol{}
	for j := 0; j < n; j++ {
		s := newTestSymbol(uint64(j))
		data = append(data, s)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		enc.Reset()
		for j := 0; j < n; j++ {
			enc.AddSymbol(data[j])
		}
		for j := 0; j < m; j++ {
			enc.ProduceNextCodedSymbol()
		}
	}
}
