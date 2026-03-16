package ir

import (
	"fmt"
	"math"
	"sync"

	"github.com/axiomhq/splitblockbloom"
	"github.com/cespare/xxhash/v2"
	"github.com/vmihailenco/msgpack/v5"
)

const (
	// defaultInitialCapacity is the starting capacity for filters.
	defaultInitialCapacity = 128
	// defaultMaxFPRate is the maximum false positive rate for filters.
	defaultMaxFPRate = 0.01
)

// FilterOption configures a [BloomFilter].
type FilterOption func(*filterConfig)

type filterConfig struct {
	capacity  uint
	maxFPRate float64
}

// WithInitialCapacity sets the initial filter capacity.
func WithInitialCapacity(n uint) FilterOption {
	return func(cfg *filterConfig) { cfg.capacity = n }
}

// WithMaxFPRate sets the maximum allowed false positive rate. When the estimated
// false positive rate exceeds this threshold, the filter is automatically
// rebuilt with doubled capacity from the temp file. Defaults to 0.01 (1%).
func WithMaxFPRate(rate float64) FilterOption {
	return func(cfg *filterConfig) { cfg.maxFPRate = rate }
}

func newFilterConfig(opts []FilterOption) filterConfig {
	cfg := filterConfig{
		capacity:  defaultInitialCapacity,
		maxFPRate: defaultMaxFPRate,
	}
	for _, opt := range opts {
		opt(&cfg)
	}
	return cfg
}

// toBytes converts a value to its byte representation for filter operations.
func toBytes(v any) []byte {
	return []byte(fmt.Sprintf("%v", v))
}

// hashValue hashes a byte slice using xxHash64. This is the standard hash
// function used by all filter types, enabling cross-language compatibility —
// any language with an xxHash64 implementation can query the serialized filter.
func hashValue(data []byte) uint64 {
	return xxhash.Sum64(data)
}

func newSBBF(capacity uint, maxFPRate float64) splitblockbloom.Filter {
	bpv := splitblockbloom.RecommendedBitsPerValue(uint64(capacity), maxFPRate)
	return splitblockbloom.NewFilter(uint64(capacity), uint64(bpv))
}

// BloomFilter tracks unique values for a high-cardinality field using a
// Parquet-compatible Split Block Bloom Filter (SBBF) for deduplication and
// membership queries, and a temporary LZ4-compressed file for storing all
// observed values on disk. Values are hashed with xxHash64 for cross-language
// compatibility — the same algorithm and block layout used by Apache Parquet.
// The Bloom filter tracks approximate cardinality and provides queryable
// membership testing. All values are written to the temp file (including
// duplicates) so that no data is lost to false positives. When the
// estimated false positive rate exceeds the maximum, the filter is automatically
// rebuilt from the temp file with doubled capacity. The file is append-only; LZ4
// compression handles duplicate reduction at the byte level. The temp file is
// automatically deleted when [Close] is called.
type BloomFilter struct {
	filter splitblockbloom.Filter
	cfg    filterConfig

	mu    sync.Mutex
	store filterStore
	err   error
}

// NewBloomFilter creates a new [BloomFilter]. Use [WithInitialCapacity] and
// [WithMaxFPRate] to override defaults (128 capacity, 1% max FP rate).
func NewBloomFilter(opts ...FilterOption) *BloomFilter {
	cfg := newFilterConfig(opts)
	return &BloomFilter{
		cfg:    cfg,
		filter: newSBBF(cfg.capacity, cfg.maxFPRate),
	}
}

// Observe records a field value. Nil values are ignored. The value is always
// appended to the temp file. If the Bloom filter has not seen the value before,
// it is added to the filter. When the estimated false positive rate exceeds the
// maximum, the filter is automatically rebuilt with doubled capacity.
func (bf *BloomFilter) Observe(fieldValue any) {
	if nil == fieldValue {
		return
	}

	data := toBytes(fieldValue)
	h := hashValue(data)

	bf.mu.Lock()
	defer bf.mu.Unlock()

	if nil != bf.err {
		return
	}

	if err := bf.store.append(data); nil != err {
		bf.err = err
		return
	}

	if bf.filter.AddHashIfNotContains(h) {
		cardinality := uint64(bf.filter.EstimateCardinality())
		if cardinality > 0 {
			fpRate := estimateSBBFFPRate(bf.filter, cardinality)
			if fpRate > bf.cfg.maxFPRate {
				bf.rebuild()
			}
		}
	}
}

// Contains returns true if the value has likely been observed (subject to Bloom
// filter false positive rate). The value is converted to bytes and hashed with
// xxHash64, the same as [Observe], so Contains(x) after Observe(x) is
// consistent.
func (bf *BloomFilter) Contains(value any) bool {
	bf.mu.Lock()
	defer bf.mu.Unlock()
	return bf.filter.Contains(hashValue(toBytes(value)))
}

// ApproxCount returns the approximate number of unique values observed, derived
// from the Bloom filter's bit density.
func (bf *BloomFilter) ApproxCount() uint64 {
	bf.mu.Lock()
	defer bf.mu.Unlock()
	return uint64(bf.filter.EstimateCardinality())
}

// EstimateFPRate returns the current estimated false positive rate of the Bloom
// filter.
func (bf *BloomFilter) EstimateFPRate() float64 {
	bf.mu.Lock()
	defer bf.mu.Unlock()
	return estimateSBBFFPRate(bf.filter, uint64(bf.filter.EstimateCardinality()))
}

// Filter returns a copy of the underlying [splitblockbloom.Filter] for direct
// access or serialization. The filter uses the same 256-bit block layout and
// algorithm as Apache Parquet's SBBF, so its [splitblockbloom.Filter.WriteTo]
// output can be read by any Parquet-compatible SBBF implementation.
func (bf *BloomFilter) Filter() splitblockbloom.Filter {
	bf.mu.Lock()
	defer bf.mu.Unlock()
	cp := make(splitblockbloom.Filter, len(bf.filter))
	copy(cp, bf.filter)
	return cp
}

// BloomFilterSnapshot is the serializable representation of a [BloomFilter].
// It contains a type discriminator and the raw SBBF block data. The type string
// "parquet_sbbf_xxhash64" fully identifies the format: Parquet-compatible Split
// Block Bloom Filter with 256-bit blocks (8 × uint32 LE words), standard salt
// constants, and xxHash64 with seed 0. NumBlocks is derived as len(Data)/32.
// Cardinality can be re-estimated from the loaded filter via EstimateCardinality.
type BloomFilterSnapshot struct {
	Type string `msgpack:"type"` // "parquet_sbbf_xxhash64"
	Data []byte `msgpack:"data"` // raw 256-bit blocks (32 bytes each)
}

// Snapshot returns a [BloomFilterSnapshot] containing the filter metadata and
// raw block data, ready for serialization with msgpack. The block data is
// serialized using each block's native binary format (8 × uint32 LE words per
// 256-bit block), which matches the Apache Parquet SBBF layout.
func (bf *BloomFilter) Snapshot() (BloomFilterSnapshot, error) {
	bf.mu.Lock()
	defer bf.mu.Unlock()

	var data []byte
	for i := 0; i < bf.filter.NumBlocks(); i++ {
		data = bf.filter[i].AppendTo(data)
	}

	return BloomFilterSnapshot{
		Type: "parquet_sbbf_xxhash64",
		Data: data,
	}, nil
}

// MarshalMsgpack serializes the [BloomFilter] to msgpack bytes.
func (bf *BloomFilter) MarshalMsgpack() ([]byte, error) {
	snap, err := bf.Snapshot()
	if nil != err {
		return nil, err
	}
	return msgpack.Marshal(snap)
}

// LoadBloomFilterSnapshot restores a [BloomFilter] from a [BloomFilterSnapshot].
// The restored filter supports [Contains] and [ApproxCount] queries but not
// [Observe] (no temp file is created). NumBlocks is derived from len(Data)/32.
func LoadBloomFilterSnapshot(snap BloomFilterSnapshot) (*BloomFilter, error) {
	if snap.Type != "parquet_sbbf_xxhash64" {
		return nil, fmt.Errorf(
			"unsupported bloom filter type %q, want %q",
			snap.Type, "parquet_sbbf_xxhash64",
		)
	}
	if 0 != len(snap.Data)%32 {
		return nil, fmt.Errorf(
			"bloom filter data size %d is not a multiple of 32 bytes",
			len(snap.Data),
		)
	}
	numBlocks := len(snap.Data) / 32
	if 0 == numBlocks {
		return nil, fmt.Errorf("bloom filter data is empty")
	}
	filter := make(splitblockbloom.Filter, numBlocks)
	for i := range filter {
		offset := i * 32
		if err := filter[i].UnmarshalBinary(snap.Data[offset : offset+32]); nil != err {
			return nil, err
		}
	}
	return &BloomFilter{
		filter: filter,
		cfg: filterConfig{
			capacity:  uint(filter.EstimateCardinality()),
			maxFPRate: defaultMaxFPRate,
		},
		err: errSnapshotReadOnly,
	}, nil
}

// errSnapshotReadOnly is set on filters loaded from a snapshot to prevent
// Observe from silently creating a temp file. Contains and ApproxCount still work.
var errSnapshotReadOnly = fmt.Errorf("bloom filter loaded from snapshot: Observe is not supported")

// UnmarshalBloomFilter deserializes a [BloomFilter] from msgpack bytes produced
// by [BloomFilter.MarshalMsgpack].
func UnmarshalBloomFilter(data []byte) (*BloomFilter, error) {
	var snap BloomFilterSnapshot
	if err := msgpack.Unmarshal(data, &snap); nil != err {
		return nil, err
	}
	return LoadBloomFilterSnapshot(snap)
}

// Err returns the first error encountered during file operations. Once an error
// occurs, subsequent [Observe] calls are no-ops.
func (bf *BloomFilter) Err() error {
	bf.mu.Lock()
	defer bf.mu.Unlock()
	return bf.err
}

// Close deletes the temporary file. Must be called to avoid leaving temp files
// on disk.
func (bf *BloomFilter) Close() error {
	bf.mu.Lock()
	defer bf.mu.Unlock()
	return bf.store.close()
}

func (bf *BloomFilter) rebuild() {
	if err := bf.store.closeLz4Writer(); nil != err {
		bf.err = err
		return
	}

	bf.cfg.capacity *= 2
	bf.filter = newSBBF(bf.cfg.capacity, bf.cfg.maxFPRate)

	if err := bf.store.forEach(func(v []byte) {
		bf.filter.AddHash(hashValue(v))
	}); nil != err {
		bf.err = err
		return
	}

	if err := bf.store.reopenWriter(); nil != err {
		bf.err = err
	}
}

// estimateSBBFFPRate estimates the false positive rate for a SBBF given the
// number of items inserted. Uses the standard Bloom filter approximation
// adjusted for blocked layout.
func estimateSBBFFPRate(f splitblockbloom.Filter, nkeys uint64) float64 {
	if 0 == nkeys || 0 == f.NumBlocks() {
		return 0
	}
	// Bits per block = 256 (8 words × 32 bits). Number of hashes is fixed at 8
	// for SBBF (one per word in the block).
	const bitsPerBlock = 256
	const nhashes = 8
	numBlocks := float64(f.NumBlocks())
	n := float64(nkeys)
	// Average items per block
	itemsPerBlock := n / numBlocks
	// Standard Bloom FP rate within a single block
	bitNotSet := math.Pow(1.0-1.0/bitsPerBlock, itemsPerBlock*nhashes)
	return math.Pow(1.0-bitNotSet, nhashes)
}
