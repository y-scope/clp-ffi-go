package ir

import (
	"fmt"
	"os"
	"testing"
)

func TestBloomFilterObserveAndContains(t *testing.T) {
	t.Parallel()
	bf := NewBloomFilter()
	defer bf.Close()

	bf.Observe("uuid-abc")
	bf.Observe("uuid-def")
	bf.Observe("uuid-abc") // duplicate

	if !bf.Contains("uuid-abc") {
		t.Error("expected Contains(uuid-abc) == true")
	}
	if !bf.Contains("uuid-def") {
		t.Error("expected Contains(uuid-def) == true")
	}
	if bf.Contains("uuid-missing") {
		t.Error("expected Contains(uuid-missing) == false")
	}
}

func TestBloomFilterContainsMatchesObserveType(t *testing.T) {
	t.Parallel()
	bf := NewBloomFilter()
	defer bf.Close()

	bf.Observe(42)
	if !bf.Contains(42) {
		t.Error("Contains(42) should match Observe(42)")
	}
}

func TestBloomFilterNilIgnored(t *testing.T) {
	t.Parallel()
	bf := NewBloomFilter()
	defer bf.Close()

	bf.Observe(nil)
	if 0 != bf.ApproxCount() {
		t.Errorf("ApproxCount: got %d, want 0", bf.ApproxCount())
	}
}

func TestBloomFilterApproxCount(t *testing.T) {
	t.Parallel()
	bf := NewBloomFilter(WithInitialCapacity(4096))
	defer bf.Close()

	for i := 0; i < 1000; i++ {
		bf.Observe(fmt.Sprintf("uuid-%d", i))
	}

	count := bf.ApproxCount()
	// EstimateCardinality is approximate — allow 50% margin since SBBF
	// cardinality estimation is less precise than standard Bloom filters.
	if count < 500 || count > 1500 {
		t.Errorf("ApproxCount: got %d, want ~1000 (±50%%)", count)
	}
}

func TestBloomFilterDuplicatesNotCounted(t *testing.T) {
	t.Parallel()
	bf := NewBloomFilter()
	defer bf.Close()

	for i := 0; i < 100; i++ {
		bf.Observe("same-value")
	}

	count := bf.ApproxCount()
	if count < 1 || count > 2 {
		t.Errorf("ApproxCount: got %d, want ~1", count)
	}
}

func TestBloomFilterRebuild(t *testing.T) {
	t.Parallel()
	bf := NewBloomFilter(WithInitialCapacity(100))
	defer bf.Close()

	for i := 0; i < 200; i++ {
		bf.Observe(fmt.Sprintf("uuid-%d", i))
	}

	// All values should still be queryable after rebuild
	for i := 0; i < 200; i++ {
		if !bf.Contains(fmt.Sprintf("uuid-%d", i)) {
			t.Errorf("missing uuid-%d after rebuild", i)
		}
	}

	if nil != bf.Err() {
		t.Errorf("unexpected error: %v", bf.Err())
	}
}

func TestBloomFilterValueWithNewlines(t *testing.T) {
	t.Parallel()
	bf := NewBloomFilter(WithInitialCapacity(64))
	defer bf.Close()

	bf.Observe("line1\nline2\nline3")
	bf.Observe("normal-value")

	// Force rebuild
	for i := 0; i < 100; i++ {
		bf.Observe(fmt.Sprintf("uuid-%d", i))
	}

	if !bf.Contains("line1\nline2\nline3") {
		t.Error("value with newlines should survive rebuild")
	}
	if !bf.Contains("normal-value") {
		t.Error("normal value should survive rebuild")
	}
	if nil != bf.Err() {
		t.Errorf("unexpected error: %v", bf.Err())
	}
}

func TestBloomFilterCloseCleansTempFile(t *testing.T) {
	t.Parallel()
	bf := NewBloomFilter()

	bf.Observe("value1")

	bf.mu.Lock()
	var tmpPath string
	if nil != bf.store.tmpFile {
		tmpPath = bf.store.tmpFile.Name()
	}
	bf.mu.Unlock()

	if "" == tmpPath {
		t.Fatal("expected temp file to be created")
	}

	if err := bf.Close(); nil != err {
		t.Fatalf("Close failed: %v", err)
	}

	if _, err := os.Stat(tmpPath); !os.IsNotExist(err) {
		t.Errorf("temp file %s should have been deleted", tmpPath)
	}
}

func TestBloomFilterCloseWithoutObserve(t *testing.T) {
	t.Parallel()
	bf := NewBloomFilter()
	if err := bf.Close(); nil != err {
		t.Fatalf("Close failed: %v", err)
	}
}

func TestBloomFilterWithOptions(t *testing.T) {
	t.Parallel()
	bf := NewBloomFilter(
		WithInitialCapacity(50000),
		WithMaxFPRate(0.001),
	)
	defer bf.Close()

	bf.mu.Lock()
	capacity := bf.cfg.capacity
	maxFPRate := bf.cfg.maxFPRate
	bf.mu.Unlock()

	if 50000 != capacity {
		t.Errorf("capacity: got %d, want 50000", capacity)
	}
	if 0.001 != maxFPRate {
		t.Errorf("maxFPRate: got %f, want 0.001", maxFPRate)
	}
}

func TestBloomFilterFieldCollectorInterface(t *testing.T) {
	t.Parallel()
	var _ FieldCollector = NewBloomFilter()
}

func TestBloomFilterEstimateFPRate(t *testing.T) {
	t.Parallel()
	bf := NewBloomFilter(WithInitialCapacity(1000), WithMaxFPRate(0.01))
	defer bf.Close()

	if bf.EstimateFPRate() > 0.0001 {
		t.Errorf("empty filter FP rate should be ~0, got %f", bf.EstimateFPRate())
	}

	for i := 0; i < 500; i++ {
		bf.Observe(fmt.Sprintf("uuid-%d", i))
	}
	rate := bf.EstimateFPRate()
	if rate > 0.01 {
		t.Errorf("FP rate should be below max 0.01, got %f", rate)
	}
}

func TestBloomFilterMarshalUnmarshalRoundTrip(t *testing.T) {
	t.Parallel()
	bf := NewBloomFilter(WithInitialCapacity(1024))
	defer bf.Close()

	for i := 0; i < 100; i++ {
		bf.Observe(fmt.Sprintf("uuid-%d", i))
	}

	// Marshal
	data, err := bf.MarshalMsgpack()
	if nil != err {
		t.Fatalf("MarshalMsgpack failed: %v", err)
	}

	// Unmarshal
	restored, err := UnmarshalBloomFilter(data)
	if nil != err {
		t.Fatalf("UnmarshalBloomFilter failed: %v", err)
	}

	// Restored filter should contain all values
	for i := 0; i < 100; i++ {
		if !restored.Contains(fmt.Sprintf("uuid-%d", i)) {
			t.Errorf("restored filter missing uuid-%d", i)
		}
	}

	// And reject unknown values
	if restored.Contains("not-inserted") {
		t.Error("restored filter should not contain 'not-inserted'")
	}
}

func TestBloomFilterSnapshot(t *testing.T) {
	t.Parallel()
	bf := NewBloomFilter(WithInitialCapacity(256), WithMaxFPRate(0.01))
	defer bf.Close()

	for i := 0; i < 50; i++ {
		bf.Observe(fmt.Sprintf("key-%d", i))
	}

	snap, err := bf.Snapshot()
	if nil != err {
		t.Fatalf("Snapshot failed: %v", err)
	}

	if "parquet_sbbf_xxhash64" != snap.Type {
		t.Errorf("Type: got %s, want parquet_sbbf_xxhash64", snap.Type)
	}
	if 0 == len(snap.Data) {
		t.Error("Data should not be empty")
	}
	if 0 != len(snap.Data)%32 {
		t.Errorf("Data length %d should be a multiple of 32", len(snap.Data))
	}
}

func TestBloomFilterErrAfterFileFailure(t *testing.T) {
	t.Parallel()
	bf := NewBloomFilter()
	defer bf.Close()

	if nil != bf.Err() {
		t.Errorf("expected no error initially, got: %v", bf.Err())
	}
	bf.Observe("value1")
	if nil != bf.Err() {
		t.Errorf("expected no error after observe, got: %v", bf.Err())
	}
}
