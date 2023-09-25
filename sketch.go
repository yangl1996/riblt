package riblt

type Sketch[T Symbol[T]] []CodedSymbol[T]

func (s Sketch[T]) AddHashedSymbol(t HashedSymbol[T]) {
	m := randomMapping{t.Hash, 0}
	for int(m.lastIdx) < len(s) {
		idx := m.lastIdx
		s[idx].Symbol = s[idx].Symbol.XOR(t.Symbol)
		s[idx].Count += 1
		s[idx].Hash ^= t.Hash
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
		s[i].Symbol = s[i].Symbol.XOR(s2[i].Symbol)
		s[i].Count = s[i].Count - s2[i].Count
		s[i].Hash ^= s2[i].Hash
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
