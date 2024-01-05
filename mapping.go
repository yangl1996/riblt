package riblt

import (
	"math"
)

// randomMapping generates a sequence of indices indicating the coded symbols
// that a source symbol should be mapped to. The generator is deterministic,
// dependent only on its initial PRNG state. When seeded with a uniformly
// random initial PRNG state, index i will be present in the generated sequence
// with probability 1/(1+i/2), for any non-negative i.
type randomMapping struct {
	prng    uint64 // PRNG state
	lastIdx uint64 // the last index the symbol was mapped to
}

// nextIndex returns the next index in the sequence.
func (s *randomMapping) nextIndex() uint64 {
	// Update the PRNG. TODO: prove that the following update rule gives us
	// high quality randomness, assuming the multiplier is coprime to 2^64.
	r := s.prng * 0xda942042e4dd58b5
	s.prng = r
	// Calculate the difference from the current index (s.lastIdx) to the next
	// index. See the paper for details. We use the approximated form
	//   diff = (1.5+i)((1-u)^(-1/2)-1)
	// where i is the current index, i.e., lastIdx; u is a number uniformly
	// sampled from [0, 1). We apply the following optimization. Notice that
	// our u actually comes from sampling a random uint64 r, and then dividing
	// it by maxUint64, i.e., 1<<64. So we can replace (1-u)^(-1/2) with
	//   1<<32 / sqrt(r).
	s.lastIdx += uint64(math.Ceil((float64(s.lastIdx) + 1.5) * ((1<<32)/math.Sqrt(float64(r)+1) - 1)))
	return s.lastIdx
}
