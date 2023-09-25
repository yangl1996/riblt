package riblt

// Symbol is the interfact that source symbols should implement.
type Symbol[T any] interface {
	// XOR returns the XOR result of the method receiver and t2. It is allowed
	// to modify the method receiver. When the method receiver is the default
	// value of T, the result should be equal to t2.
	XOR(t2 T) T
	// Hash returns the hash of the method receiver. It must not modify the
	// method receiver.
	Hash() uint64
}

// HashedSymbol is the bundle of a symbol and its hash.
type HashedSymbol[T Symbol[T]] struct {
	Symbol T
	Hash   uint64
}

// CodedSymbol is a coded symbol produced by a Rateless IBLT encoder.
type CodedSymbol[T Symbol[T]] struct {
	HashedSymbol[T]
	Count int64
}

const (
	add    = 1
	remove = -1
)

func (c CodedSymbol[T]) apply(s HashedSymbol[T], direction int64) CodedSymbol[T] {
	c.Symbol = c.Symbol.XOR(s.Symbol)
	c.Hash ^= s.Hash
	c.Count += direction
	return c
}
