package riblt

type Sketch[T Symbol[T]] []CodedSymbol[T]

func (s Sketch[T]) AddHashedSymbol(t HashedSymbol[T]) {
	m := randomMapping{t.Hash, 0}
	for int(m.lastIdx) < len(s) {
		idx := m.lastIdx
		s[idx].sum = s[idx].sum.XOR(t.Symbol)
		s[idx].count += 1
		s[idx].checksum ^= t.Hash
		m.nextIndex()
	}
}

func (s Sketch[T]) AddSymbol(t T) {
	hs := HashedSymbol[T]{t, t.Hash()}
	s.AddHashedSymbol(hs)
}

func (s Sketch[T]) Subtract(s2 Sketch[T]) Sketch[T] {
	if len(s) != len(s2) {
		panic("subtracting sketches of different sizes")
	}

	for i := range s {
		s[i].sum = s[i].sum.XOR(s2[i].sum)
		s[i].count = s[i].count - s2[i].count
		s[i].checksum = s[i].checksum ^ s2[i].checksum
	}
	return s
}

func (s Sketch[T]) Decode() ([]HashedSymbol[T], []HashedSymbol[T], bool) {
	dec := Decoder[T]{}
	for _, c := range s {
		dec.AddCodedSymbol(c)
	}
	dec.TryDecode()
	return dec.Remote(), dec.Local(), dec.Decoded()
}
