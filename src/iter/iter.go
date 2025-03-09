//go:build !go1.23

// Delete this file when TinyGo drops support for Go 1.22.

package iter

// Seq is an iterator over sequences of individual values.
// When called as seq(yield), seq calls yield(v) for each value v in the sequence,
// stopping early if yield returns false.
// See the [iter] package documentation for more details.
type Seq[V any] func(yield func(V) bool)

// Seq2 is an iterator over sequences of pairs of values, most commonly key-value pairs.
// When called as seq(yield), seq calls yield(k, v) for each pair (k, v) in the sequence,
// stopping early if yield returns false.
// See the [iter] package documentation for more details.
type Seq2[K, V any] func(yield func(K, V) bool)
