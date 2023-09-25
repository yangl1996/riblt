package riblt

type Decoder[T Symbol[T]] struct {
	cs      []CodedSymbol[T] // coded symbols received so far
	isDirty []bool           // if a coded symbol is in the dirty list
	local   codingWindow[T]  // set of source symbols that are exclusive to the decoder
	window  codingWindow[T]  // set of source symbols that the decoder initially has
	remote  codingWindow[T]  // set of source symbols that are exclusive to the encoder
	dirty   []int            // indices of coded symbols that have degrees -1, 0, or 1, of which we should examine the hashes
	decoded int              // number of coded symbols that are decoded
}

func (d *Decoder[T]) Decoded() bool {
	return d.decoded == len(d.cs)
}

func (d *Decoder[T]) Local() []HashedSymbol[T] {
	return d.local.symbols
}

func (d *Decoder[T]) Remote() []HashedSymbol[T] {
	return d.remote.symbols
}

func (d *Decoder[T]) AddSymbol(s T) {
	th := HashedSymbol[T]{s, s.Hash()}
	d.AddHashedSymbol(th)
}

func (d *Decoder[T]) AddHashedSymbol(s HashedSymbol[T]) {
	d.window.addHashedSymbol(s)
}

func (d *Decoder[T]) AddCodedSymbol(c CodedSymbol[T]) {
	// scan through decoded symbols to peel off matching ones
	c = d.window.applyWindow(c, remove)
	c = d.remote.applyWindow(c, remove)
	c = d.local.applyWindow(c, add)
	// insert the new coded symbol
	d.cs = append(d.cs, c)
	if c.Count <= 1 && c.Count >= -1 {
		d.dirty = append(d.dirty, len(d.cs)-1)
		d.isDirty = append(d.isDirty, true)
	} else {
		d.isDirty = append(d.isDirty, false)
	}
	return
}

func (d *Decoder[T]) applyNewSymbol(t HashedSymbol[T], direction int64) randomMapping {
	m := randomMapping{t.Hash, 0}
	for int(m.lastIdx) < len(d.cs) {
		cidx := int(m.lastIdx)
		d.cs[cidx] = d.cs[cidx].apply(t, direction)
		cnt := d.cs[cidx].Count
		if (!d.isDirty[cidx]) && cnt <= 1 && cnt >= -1 {
			d.dirty = append(d.dirty, cidx)
			d.isDirty[cidx] = true
		}
		m.nextIndex()
	}
	return m
}

func (d *Decoder[T]) TryDecode() {
	for didx := 0; didx < len(d.dirty); didx += 1 {
		cidx := d.dirty[didx]
		c := d.cs[cidx]
		switch c.Count {
		case 1:
			h := c.Symbol.Hash()
			if h == c.Hash {
				// allocate a symbol and then XOR with the sum, so that we are
				// guaranted to copy the sum whether or not the symbol
				// interface is implemented as a pointer
				ns := HashedSymbol[T]{}
				ns.Symbol = ns.Symbol.XOR(c.Symbol)
				ns.Hash = h
				m := d.applyNewSymbol(ns, remove)
				d.remote.addHashedSymbolWithMapping(ns, m)
				d.decoded += 1
			}
		case -1:
			h := c.Symbol.Hash()
			if h == c.Hash {
				ns := HashedSymbol[T]{}
				ns.Symbol = ns.Symbol.XOR(c.Symbol)
				ns.Hash = h
				m := d.applyNewSymbol(ns, add)
				d.local.addHashedSymbolWithMapping(ns, m)
				d.decoded += 1
			}
		case 0:
			if c.Hash == 0 {
				d.decoded += 1
			}
			// One may want to add a panic here when coded symbol is not of degree
			// 1, -1 or 0, but this may happen when a dirty coded symbol is
			// operated on before its turn.
		}
		d.isDirty[cidx] = false
	}
	d.dirty = d.dirty[:0]
}

func (d *Decoder[T]) Reset() {
	if len(d.cs) != 0 {
		d.cs = d.cs[:0]
	}
	if len(d.dirty) != 0 {
		d.dirty = d.dirty[:0]
	}
	if len(d.isDirty) != 0 {
		d.isDirty = d.isDirty[:0]
	}
	d.local.reset()
	d.remote.reset()
	d.window.reset()
	d.decoded = 0
}
