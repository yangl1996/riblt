// Package riblt implements Rateless Invertible Bloom Lookup Tables (Rateless
// IBLTs), a family of rateless codes that reconcile set differences.
//
// Rateless IBLTs solve the "set reconciliation" problem. Imagine that two
// computers, Alice and Bob, each holds a set of fixed-length bit strings. We
// call the bit strings "source symbols". Alice and Bob wish to distributedly
// compute the symmetric difference of their sets, i.e., source symbols that
// are exclusively present on either Alice or Bob but not both computers.
// 
// Rateless IBLTs operate by defining an infinite sequence of "coded symbols"
// for any set. The coded symbol sequences have two strong properties. First,
// for any two sets, prefixes of their coded symbol sequences alone are
// sufficient for computing their symmetric difference. Second, the number of
// coded symbols, i.e., the length of the prefixes, required is linear to the
// size of the symmetric difference, where the coefficient converges to 1.35.
//
// In practice, Alice and Bob locally generate the coded symbol sequences for
// their respective sets. Alice streams an increasingly long prefix of her
// sequence to Bob, who keeps trying to use the received prefix along with his
// locally generated sequence to recover the symmetric difference, which will
// happen after Bob receives a sufficiently long prefix.
//
// To use this library, the user needs to define the source symbol being
// reconciled. See type Symbol. Then, the user instantiates an Encoder for
// Alice, and a Decoder for Bob. Alice and Bob's sets should be imported into
// the Encoder and the Decoder, respectively. The user should program Alice to
// stream coded symbols over a reliable transport to Bob, and program Bob to
// decode the symbols and signal Alice to stop when successful. See the
// example.
package riblt

// Symbol is the interface that source symbols (set elements being reconciled)
// should implement. It specifies a Boolean group, where type T (or its subset)
// is the underlying set, and $ is the group operation. It should satisfy the
// following properties:
//  1. For all a, b, c in the group, (a $ b) $ c = a $ (b $ c).
//  2. Let e be the default value of T. For every a in the group, e $ a = a
//     and a $ e = a.
//  3. For every a in the group, a $ a = e.
// As an example, when source symbols are plain byte strings of length 32, T is
// [32]byte. $ can be the bitwise exclusive-or (XOR) operation. e is the byte
// string where every byte is zero.
type Symbol[T any] interface {
	// XOR returns t $ t2, where t is the method receiver. XOR is allowed to
	// modify the method receiver in-place (when T is a pointer) and return the
	// modified t. Although the method is called XOR (because the bitwise
	// exclusive-or operation is a valid group operation for groups of
	// fixed-length byte strings), it can implement any operation that satisfy
	// the aforementioned properties.
	XOR(t2 T) T
	// Hash returns the hash of the method receiver. It must not modify the
	// method receiver. It must not be homomorphic over the group operation.
	// That is, the probability that
	//   (a $ b).Hash() == a.Hash() ^ b.Hash()
	// must be negligible. Here, ^ on the right-hand side is the bitwise
	// exclusive-or operation.
	Hash() uint64
}

// HashedSymbol is the bundle of a symbol and its hash computed using its Hash
// method.
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
