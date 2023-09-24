package riblt

type Symbol[T any] interface {
	// XOR returns the XOR result of the method receiver and t2. It is allowed
	// to modify the method receiver during the operation. When the method
	// receiver is the default value of T, the result is equal to t2.
	XOR(t2 T) T
	// Hash returns the cryptographic hash of the method receiver. It is guaranteed not to modify the method receiver.
	Hash() uint64
}

type HashedSymbol[T Symbol[T]] struct {
	Symbol T
	Hash uint64
}

type CodedSymbol[T Symbol[T]] struct {
    sum T
    count int64
    checksum uint64
}

const (
	add = 1
	remove = -1
)

func (c CodedSymbol[T]) apply(s HashedSymbol[T], direction int64) CodedSymbol[T] {
	c.sum = c.sum.XOR(s.Symbol)
	c.count += direction
	c.checksum ^= s.Hash
	return c
}
