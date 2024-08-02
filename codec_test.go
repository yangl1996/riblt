package riblt

import (
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

func BenchmarkEncodeAndDecode(bc *testing.B) {
	cases := []struct {
        name string
        size int
    }{
        {"d=10", 10},
        {"d=20", 20},
        {"d=40", 40},
        {"d=100", 100},
        {"d=1000", 1000},
        {"d=10000", 10000},
        {"d=50000", 50000},
        {"d=100000", 100000},
    }
	for _, tc := range cases {
		bc.Run(tc.name, func(b *testing.B) {
			b.SetBytes(testSymbolSize * int64(tc.size))
			nlocal := tc.size/2
			nremote := tc.size/2
			ncommon := tc.size
			ncw := 0
			var nextId uint64
			b.ResetTimer()
			b.StopTimer()
			for iter := 0; iter < b.N; iter++ {
				enc := Encoder[testSymbol]{}
				dec := Decoder[testSymbol]{}

				for i := 0; i < nlocal; i++ {
					s := newTestSymbol(nextId)
					nextId += 1
					dec.AddSymbol(s)
				}
				for i := 0; i < nremote; i++ {
					s := newTestSymbol(nextId)
					nextId += 1
					enc.AddSymbol(s)
				}
				for i := 0; i < ncommon; i++ {
					s := newTestSymbol(nextId)
					nextId += 1
					enc.AddSymbol(s)
					dec.AddSymbol(s)
				}
				b.StartTimer()
				for {
					dec.AddCodedSymbol(enc.ProduceNextCodedSymbol())
					dec.TryDecode()
					ncw += 1
					if dec.Decoded() {
						break
					}
				}
				b.StopTimer()
			}
			b.ReportMetric(float64(ncw)/float64(b.N * tc.size), "symbols/diff")
		})
	}
}

func TestEncodeAndDecode(t *testing.T) {
	enc := Encoder[testSymbol]{}
	dec := Decoder[testSymbol]{}
	local := make(map[uint64]struct{})
	remote := make(map[uint64]struct{})

	var nextId uint64
	nlocal := 200000/2
	nremote := 200000/2
	ncommon := 200000
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
}

