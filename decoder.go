package riblt

type Decoder[T Symbol[T]] struct {
	// coded symbols received so far
	cs []CodedSymbol[T]
	// set of source symbols that are exclusive to the decoder
	local codingWindow[T]
	// set of source symbols that the decoder initially has
	window codingWindow[T]
	// set of source symbols that are exclusive to the encoder
	remote codingWindow[T]
	// indices of coded symbols that can be decoded, i.e., degree equal to -1
	// or 1 and sum of hash equal to hash of sum, or degree equal to 0 and sum
	// of hash equal to 0
	decodable []int
	// number of coded symbols that are decoded
	decoded int
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
	// check if the coded symbol is decodable, and insert into decodable list if so
	if (c.Count == 1 || c.Count == -1) && (c.Hash == c.Symbol.Hash()) {
		d.decodable = append(d.decodable, len(d.cs)-1)
	} else if c.Count == 0 && c.Hash == 0 {
		d.decodable = append(d.decodable, len(d.cs)-1)
	}
	return
}

func (d *Decoder[T]) applyNewSymbol(t HashedSymbol[T], direction int64) randomMapping {
	m := randomMapping{t.Hash, 0}
	for int(m.lastIdx) < len(d.cs) {
		cidx := int(m.lastIdx)
		d.cs[cidx] = d.cs[cidx].apply(t, direction)
		// Check if the coded symbol is now decodable. We do not want to insert
		// a decodable symbol into the list if we already did, otherwise we
		// will visit the same coded symbol twice. To see how we achieve that,
		// notice the following invariant: if a coded symbol becomes decodable
		// with degree D (obviously -1 <= D <=1), it will stay that way, except
		// for that it's degree may become 0. For example, a decodable symbol
		// of degree -1 may not later become undecodable, or become decodable
		// but of degree 1. This is because each peeling removes a source
		// symbol from the coded symbol. So, if a coded symbol already contains
		// only 1 or 0 source symbol (the definition of decodable), the most we
		// can do is to peel off the only remaining source symbol.
		//
		// Meanwhile, notice that if a decodable symbol is of degree 0, then
		// there must be a point in the past when it was of degree 1 or -1 and
		// decodable, at which time we would have inserted it into the
		// decodable list. So, we do not insert degree-0 symbols to avoid
		// duplicates. On the other hand, it is fine that we insert all
		// degree-1 or -1 decodable symbols, because we only see them in such
		// state once.
		if (d.cs[cidx].Count == -1 || d.cs[cidx].Count == 1) && d.cs[cidx].Hash == d.cs[cidx].Symbol.Hash() {
			d.decodable = append(d.decodable, cidx)
		}
		m.nextIndex()
	}
	return m
}

func (d *Decoder[T]) TryDecode() {
	for didx := 0; didx < len(d.decodable); didx += 1 {
		cidx := d.decodable[didx]
		c := d.cs[cidx]
		// We do not need to compare Hash and Symbol.Hash() below, because we
		// have checked it before inserting into the decodable list. Per the
		// invariant mentioned in the comments in applyNewSymbol, a decodable
		// symbol does not turn undecodable, so there is no worry that
		// additional source symbols have been peeled off a coded symbol after
		// it was inserted into the decodable list and before we visit them
		// here.
		switch c.Count {
		case 1:
			// allocate a symbol and then XOR with the sum, so that we are
			// guaranted to copy the sum whether or not the symbol
			// interface is implemented as a pointer
			ns := HashedSymbol[T]{}
			ns.Symbol = ns.Symbol.XOR(c.Symbol)
			ns.Hash = c.Hash
			m := d.applyNewSymbol(ns, remove)
			d.remote.addHashedSymbolWithMapping(ns, m)
			d.decoded += 1
		case -1:
			ns := HashedSymbol[T]{}
			ns.Symbol = ns.Symbol.XOR(c.Symbol)
			ns.Hash = c.Hash
			m := d.applyNewSymbol(ns, add)
			d.local.addHashedSymbolWithMapping(ns, m)
			d.decoded += 1
		case 0:
			d.decoded += 1
		default:
			// a decodable symbol does not turn undecodable, so its degree must
			// be -1, 0, or 1
			panic("invalid degree for decodable coded symbol")
		}
	}
	d.decodable = d.decodable[:0]
}

func (d *Decoder[T]) Reset() {
	if len(d.cs) != 0 {
		d.cs = d.cs[:0]
	}
	if len(d.decodable) != 0 {
		d.decodable = d.decodable[:0]
	}
	d.local.reset()
	d.remote.reset()
	d.window.reset()
	d.decoded = 0
}
