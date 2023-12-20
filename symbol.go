package riblt

// Symbol is the interface that source symbols should implement. It specifies a
// Boolean group, where T (or its subset) is the underlying set, and ^ is the
// group operation. It should satisfy the following properties:
//   1. For all a, b, c in the group, (a ^ b) ^ c = a ^ (b ^ c).
//   2. Let e be the default value of T. For every a in the group, e ^ a = a
//      and a ^ e = a.
//   3. For every a in the group, a ^ a = e.
type Symbol[T any] interface {
	// XOR returns t ^ t2, where t is the method receiver. XOR is allowed to
	// modify the method receiver. Although the method is called XOR (because
	// the bitwise exclusive-or operation is a valid group operation for groups
	// of fixed-length bit strings), it can implement any operation that
	// satisfy the aforementioned properties.
	XOR(t2 T) T
	// Hash returns the hash of the method receiver. It must not modify the
	// method receiver. It must not be homomorphic over the group operation.
	// That is, the probability that
	//   (a ^ b).Hash() == a.Hash() ^ b.Hash()
	// must be negligible. Here, ^ is the group operation on the left-hand
	// side, and bitwise exclusive-or on the right side.
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

// apply maps s to c and modifies the counter of c according to direction. add
// increments the counter, and remove decrements the counter.
func (c CodedSymbol[T]) apply(s HashedSymbol[T], direction int64) CodedSymbol[T] {
	c.Symbol = c.Symbol.XOR(s.Symbol)
	c.Hash ^= s.Hash
	c.Count += direction
	return c
}
