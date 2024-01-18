package riblt

// Sketch is a prefix of the coded symbol sequence for a set of source symbols.
type Sketch[T Symbol[T]] []CodedSymbol[T]

// AddHashedSymbol efficiently updates s when t is added to the set.
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

// RemoveHashedSymbol efficiently updates s when t is removed from the set.
func (s Sketch[T]) RemoveHashedSymbol(t HashedSymbol[T]) {
	m := randomMapping{t.Hash, 0}
	for int(m.lastIdx) < len(s) {
		idx := m.lastIdx
		s[idx].Symbol = s[idx].Symbol.XOR(t.Symbol)
		s[idx].Count -= 1
		s[idx].Hash ^= t.Hash
		m.nextIndex()
	}
}

// AddSymbol efficiently updates s when t is added to the set.
func (s Sketch[T]) AddSymbol(t T) {
	hs := HashedSymbol[T]{t, t.Hash()}
	s.AddHashedSymbol(hs)
}

// RemoveSymbol efficiently updates s when t is removed from the set.
func (s Sketch[T]) RemoveSymbol(t T) {
	hs := HashedSymbol[T]{t, t.Hash()}
	s.RemoveHashedSymbol(hs)
}

// Subtract subtracts s2 from s, modifying s in place. s and s2 must be of
// equal length. If s is a sketch of set S and s2 is a sketch of set S2, then
// the result is a sketch of the symmetric difference between S and S2.
func (s Sketch[T]) Subtract(s2 Sketch[T]) {
	if len(s) != len(s2) {
		panic("subtracting sketches of different sizes")
	}

	for i := range s {
		s[i].Symbol = s[i].Symbol.XOR(s2[i].Symbol)
		s[i].Count = s[i].Count - s2[i].Count
		s[i].Hash ^= s2[i].Hash
	}
	return
}

// Decode tries to decode s, where s can be one of the following
//  1. A sketch of set S.
//  2. Content of s after calling s.Subtract(s2), where s is a sketch of set
//     S, and s2 is a sketch of set S2.
//
// When successful, fwd contains all source symbols in S in case 1, or S \ S2
// in case 2 (\ is the set subtraction operation). rev is empty in case 1, or
// S2 \ S in case 2. succ is true. When unsuccessful, fwd and rev are
// undefined, and succ is false.
func (s Sketch[T]) Decode() (fwd []HashedSymbol[T], rev []HashedSymbol[T], succ bool) {
	dec := Decoder[T]{}
	for _, c := range s {
		dec.AddCodedSymbol(c)
	}
	dec.TryDecode()
	return dec.Remote(), dec.Local(), dec.Decoded()
}
