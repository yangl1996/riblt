package riblt

// TODO: encoder should send conflicting transactions as-is when detected; since it happens very rarely and each pair of peers can use a secret hash key, an adversary cannot forge too many conflicts.
// TODO: replace siphash with xxhash (or whatever that supports native 4-byte output)

type symbolMapping struct {
	sourceIdx int
	codedIdx int
}

// TODO: remove the heap?
type mappingHeap []symbolMapping

func (m mappingHeap) fixHead() {
	curr := 0
	for {
		child := curr * 2 + 1
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

func (m mappingHeap) fixTail() {
	curr := len(m)-1
	for {
		parent := (curr - 1) / 2
		if curr == parent || m[parent].codedIdx <= m[curr].codedIdx {
			break
		}
		m[parent], m[curr] = m[curr], m[parent]
		curr = parent
	}
}

type codingWindow[T Symbol[T]] struct {
	symbols []HashedSymbol[T]
	mappings []randomMapping
	queue mappingHeap
	nextIdx int
}

func (e *codingWindow[T]) addSymbol(t T) {
	th := HashedSymbol[T]{t, t.Hash()}
	e.addHashedSymbol(th)
}

func (e *codingWindow[T]) addHashedSymbol(t HashedSymbol[T]) {
	e.addHashedSymbolWithMapping(t, randomMapping{t.Hash , 0})
}

func (e *codingWindow[T]) addHashedSymbolWithMapping(t HashedSymbol[T], m randomMapping) {
	e.symbols = append(e.symbols, t)
	e.mappings = append(e.mappings, m)
	e.queue = append(e.queue, symbolMapping{len(e.symbols)-1, int(m.lastIdx)})
	e.queue.fixTail()
}

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

type Encoder[T Symbol[T]] codingWindow[T]

func (e *Encoder[T]) AddSymbol(s T) {
	(*codingWindow[T])(e).addSymbol(s)
}

func (e *Encoder[T]) AddHashedSymbol(s HashedSymbol[T]) {
	(*codingWindow[T])(e).addHashedSymbol(s)
}

func (e *Encoder[T]) ProduceNextCodedSymbol() CodedSymbol[T] {
	return (*codingWindow[T])(e).applyWindow(CodedSymbol[T]{}, add)
}

func (e *Encoder[T]) Reset() {
	(*codingWindow[T])(e).reset()
}

