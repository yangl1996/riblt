package riblt

// symbolMapping is a mapping from a source symbol to a coded symbol. The
// symbols are identified by their indices in codingWindow.
type symbolMapping struct {
	sourceIdx int
	codedIdx  int
}

// mappingHeap implements a priority queue of symbolMappings. The priority is
// the codedIdx of a symbolMapping. A smaller value means higher priority.  The
// first item of the queue is always the item with the highest priority.  The
// fixHead and fixTail methods should be called after the first or the last
// item is modified (or inserted, in the case of the tail), respectively. The
// implementation is a partial copy of container/heap in Go 1.21.
type mappingHeap []symbolMapping

// fixHead reestablishes the heap invariant when the first item is modified.
func (m mappingHeap) fixHead() {
	curr := 0
	for {
		child := curr*2 + 1
		if child >= len(m) {
			// no left child
			break
		}
		if rc := child + 1; rc < len(m) && m[rc].codedIdx < m[child].codedIdx {
			child = rc
		}
		if m[curr].codedIdx <= m[child].codedIdx {
			break
		}
		m[curr], m[child] = m[child], m[curr]
		curr = child
	}
}

// fixTail reestablishes the heap invariant when the last item is modified or
// just inserted.
func (m mappingHeap) fixTail() {
	curr := len(m) - 1
	for {
		parent := (curr - 1) / 2
		if curr == parent || m[parent].codedIdx <= m[curr].codedIdx {
			break
		}
		m[parent], m[curr] = m[curr], m[parent]
		curr = parent
	}
}

// codingWindow is a collection of source symbols and their mappings to coded symbols.
type codingWindow[T Symbol[T]] struct {
	symbols  []HashedSymbol[T] // source symbols
	mappings []randomMapping   // mapping generators of the source symbols
	queue    mappingHeap       // priority queue of source symbols by the next coded symbols they are mapped to
	nextIdx  int               // index of the next coded symbol to be generated
}

// addSymbol inserts a symbol to the codingWindow.
func (e *codingWindow[T]) addSymbol(t T) {
	th := HashedSymbol[T]{t, t.Hash()}
	e.addHashedSymbol(th)
}

// addHashedSymbol inserts a HashedSymbol to the codingWindow.
func (e *codingWindow[T]) addHashedSymbol(t HashedSymbol[T]) {
	e.addHashedSymbolWithMapping(t, randomMapping{t.Hash, 0})
}

// addHashedSymbolWithMapping inserts a HashedSymbol and the current state of its mapping generator to the codingWindow.
func (e *codingWindow[T]) addHashedSymbolWithMapping(t HashedSymbol[T], m randomMapping) {
	e.symbols = append(e.symbols, t)
	e.mappings = append(e.mappings, m)
	e.queue = append(e.queue, symbolMapping{len(e.symbols) - 1, int(m.lastIdx)})
	e.queue.fixTail()
}

// applyWindow maps the source symbols to the next coded symbol they should be
// mapped to, given as cw. The parameter direction controls how the counter
// of cw should be modified.
func (e *codingWindow[T]) applyWindow(cw CodedSymbol[T], direction int64) CodedSymbol[T] {
	if len(e.queue) == 0 {
		e.nextIdx += 1
		return cw
	}
	for e.queue[0].codedIdx == e.nextIdx {
		cw = cw.apply(e.symbols[e.queue[0].sourceIdx], direction)
		// generate the next mapping
		nextMap := e.mappings[e.queue[0].sourceIdx].nextIndex()
		e.queue[0].codedIdx = int(nextMap)
		e.queue.fixHead()
	}
	e.nextIdx += 1
	return cw
}

// reset clears a codingWindow.
func (e *codingWindow[T]) reset() {
	if len(e.symbols) != 0 {
		e.symbols = e.symbols[:0]
	}
	if len(e.mappings) != 0 {
		e.mappings = e.mappings[:0]
	}
	if len(e.queue) != 0 {
		e.queue = e.queue[:0]
	}
	e.nextIdx = 0
}

// Encoder is an incremental encoder of Rateless IBLTs. Once initialized with a
// set of source symbols by calling AddSymbol or AddHashedSymbol, a Encoder can
// incrementally generate coded symbols in the infinite sequence defined for
// the set. The set must not change after one or multiple coded symbols have
// been generated by calling ProduceNextCodedSymbol.
type Encoder[T Symbol[T]] codingWindow[T]

// AddSymbol adds source symbol s to e. It is undefined behavior to call AddSymbol
// after calling ProduceNextCodedSymbol.
func (e *Encoder[T]) AddSymbol(s T) {
	(*codingWindow[T])(e).addSymbol(s)
}

// AddHashedSymbol adds source symbol s to e. It is undefined behavior to call
// AddHashedSymbol after calling ProduceNextCodedSymbol.
func (e *Encoder[T]) AddHashedSymbol(s HashedSymbol[T]) {
	(*codingWindow[T])(e).addHashedSymbol(s)
}

// ProduceNextCodedSymbol returns the next coded symbol in the sequence.
func (e *Encoder[T]) ProduceNextCodedSymbol() CodedSymbol[T] {
	return (*codingWindow[T])(e).applyWindow(CodedSymbol[T]{}, add)
}

// Reset clears e. It is more efficient to call Reset to reuse an existing
// Encoder than creating a new one.
func (e *Encoder[T]) Reset() {
	(*codingWindow[T])(e).reset()
}
