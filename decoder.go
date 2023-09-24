package riblt

type receivedSymbol[T Symbol[T]] struct {
	CodedSymbol[T]
	dirty bool
}

type Decoder[T Symbol[T]] struct {
	cs []receivedSymbol[T]	// coded symbols received so far
	local codingWindow[T]
	window codingWindow[T]	// set of the symbols that the decoder already has
	remote codingWindow[T]
	dirty []int
	pending int			// number of symbols that are not pure
}

func (d *Decoder[T]) Decoded() bool {
	return d.pending == 0
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
	if c.count == 0 && c.checksum == 0 {
		// still insert the codeword in case a symbol added later causes it to become dirty
		d.cs = append(d.cs, receivedSymbol[T]{c, false})
		return
	} else {
		if c.count >= -1 && c.count <= 1 {
			d.cs = append(d.cs, receivedSymbol[T]{c, true})
			d.dirty = append(d.dirty, len(d.cs)-1)
		} else {
			d.cs = append(d.cs, receivedSymbol[T]{c, false})
		}
		d.pending += 1
		return
	}
}

func (d *Decoder[T]) applyNewSymbol(t HashedSymbol[T], direction int64) randomMapping {
	m := randomMapping{t.Hash , 0}
	for int(m.lastIdx) < len(d.cs) {
		cidx := int(m.lastIdx)
		d.cs[cidx].CodedSymbol = d.cs[cidx].apply(t, direction)
		c := d.cs[cidx]
		if (!c.dirty) && c.count >= -1 && c.count <= 1 {
			d.cs[cidx].dirty = true
			d.dirty = append(d.dirty, cidx)
		}
		m.nextIndex()
	}
	return m
}

func (d *Decoder[T]) TryDecode() {
	for didx := 0; didx < len(d.dirty); didx += 1 {
		cidx := d.dirty[didx]
		c := d.cs[cidx]
		switch c.count {
		case 1:
			h := c.sum.Hash()
			if h == c.checksum {
				ns := HashedSymbol[T]{}
				ns.Symbol = ns.Symbol.XOR(c.sum)	// force duplicate the symbol data
				ns.Hash = h
				m := d.applyNewSymbol(ns, remove)
				d.remote.addHashedSymbolWithMapping(ns, m)
				d.pending -= 1
			}
		case -1:
			h := c.sum.Hash()
			if h == c.checksum {
				ns := HashedSymbol[T]{}
				ns.Symbol = ns.Symbol.XOR(c.sum)	// force duplicate the symbol data
				ns.Hash = h
				m := d.applyNewSymbol(ns, add)
				d.local.addHashedSymbolWithMapping(ns, m)
				d.pending -= 1
			}
		case 0:
			if c.checksum == 0 {
				d.pending -= 1
			}
		// one may want to add a panic here when coded symbol is not of
		// degree 1, -1 or 0, but this may be violated when a dirty coded symbol
		// is peeled before its turn
		}
		d.cs[cidx].dirty = false
	}
	d.dirty = d.dirty[:0]
}

func (d *Decoder[T]) Reset() {
	if len(d.cs) != 0 {
		d.cs= d.cs[:0]
	}
	if len(d.dirty) != 0 {
		d.dirty = d.dirty[:0]
	}
	d.local.reset()
	d.remote.reset()
	d.window.reset()
	d.pending = 0
}

