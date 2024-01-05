package riblt_test

import (
	"encoding/binary"
	"fmt"
	"github.com/dchest/siphash"
	"github.com/yangl1996/riblt"
)

// item is the type of set elements we will reconcile. It implements
// riblt.Symbol.
type item uint64

// XOR implements the group operation. It is simply the bitwise exclusive-or of
// the operands.
func (t item) XOR(t2 item) item {
	return t ^ t2
}

// Hash hashes t using SipHash.
func (t item) Hash() uint64 {
	buf := [8]byte{}
	binary.LittleEndian.PutUint64(buf[0:8], uint64(t))
	return siphash.Hash(123, 456, buf[:])
}

func Example() {
	// Alice and Bob each holds a set of items. Bob wishes to know the items
	// that Alice has but he does not, as well as items that he has but Alice
	// does not. Their sets are mostly the same.
	alice := []item{1, 2, 3, 4, 5, 6, 7, 8, 9, 10} // only Alice has 2
	bob := []item{1, 3, 4, 5, 6, 7, 8, 9, 10, 11}  // only Bob has 11

	// Alice creates an encoder and gives it her set.
	enc := riblt.Encoder[item]{}
	for _, v := range alice {
		enc.AddSymbol(v)
	}
	// Bob creates a decoder and gives it his set.
	dec := riblt.Decoder[item]{}
	for _, v := range bob {
		dec.AddSymbol(v)
	}

	cost := 0
	for {
		// Alice generates the next coded symbol and sends to Bob.
		s := enc.ProduceNextCodedSymbol()
		cost += 1
		// Bob receives the next coded symbol from Alice, and tries to decode.
		dec.AddCodedSymbol(s)
		dec.TryDecode()
		if dec.Decoded() {
			break
		}
	}

	fmt.Println(dec.Remote()[0].Symbol, "is exclusive to Alice")
	fmt.Println(dec.Local()[0].Symbol, "is exclusive to Bob")
	fmt.Println(cost, "coded symbols sent")
	// Output:
	// 2 is exclusive to Alice
	// 11 is exclusive to Bob
	// 2 coded symbols sent
}
