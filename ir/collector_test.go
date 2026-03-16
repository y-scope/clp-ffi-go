package ir

import (
	"testing"

	"github.com/y-scope/clp-ffi-go/ffi"
)

func TestUniqueCountsObserve(t *testing.T) {
	t.Parallel()
	uc := NewUniqueCounts()

	uc.Observe("INFO")
	uc.Observe("WARN")
	uc.Observe("INFO")
	uc.Observe("ERROR")
	uc.Observe("INFO")

	if 3 != uc.UniqueCount() {
		t.Errorf("UniqueCount: got %d, want 3", uc.UniqueCount())
	}
	if 5 != uc.Total() {
		t.Errorf("Total: got %d, want 5", uc.Total())
	}

	counts := uc.ValueCounts()
	if 3 != counts["INFO"] {
		t.Errorf("INFO count: got %d, want 3", counts["INFO"])
	}
	if 1 != counts["WARN"] {
		t.Errorf("WARN count: got %d, want 1", counts["WARN"])
	}
	if 1 != counts["ERROR"] {
		t.Errorf("ERROR count: got %d, want 1", counts["ERROR"])
	}
}

func TestUniqueCountsNilIgnored(t *testing.T) {
	t.Parallel()
	uc := NewUniqueCounts()

	uc.Observe(nil)
	uc.Observe("INFO")
	uc.Observe(nil)

	if 1 != uc.UniqueCount() {
		t.Errorf("UniqueCount: got %d, want 1", uc.UniqueCount())
	}
	if 1 != uc.Total() {
		t.Errorf("Total: got %d, want 1", uc.Total())
	}
}

func TestUniqueCountsEmpty(t *testing.T) {
	t.Parallel()
	uc := NewUniqueCounts()

	if 0 != uc.UniqueCount() {
		t.Errorf("UniqueCount: got %d, want 0", uc.UniqueCount())
	}
	if 0 != uc.Total() {
		t.Errorf("Total: got %d, want 0", uc.Total())
	}
	counts := uc.ValueCounts()
	if 0 != len(counts) {
		t.Errorf("ValueCounts length: got %d, want 0", len(counts))
	}
}

func TestUniqueCountsValueCountsIsCopy(t *testing.T) {
	t.Parallel()
	uc := NewUniqueCounts()
	uc.Observe("INFO")

	counts := uc.ValueCounts()
	counts["INFO"] = 999

	original := uc.ValueCounts()
	if 1 != original["INFO"] {
		t.Errorf("ValueCounts should return a copy, but mutation leaked")
	}
}

func TestUniqueCountsUnhashableValue(t *testing.T) {
	t.Parallel()
	uc := NewUniqueCounts()

	// Slices are not hashable — should not panic.
	uc.Observe([]string{"a", "b"})
	uc.Observe(map[string]int{"x": 1})
	uc.Observe("normal")

	if 3 != uc.UniqueCount() {
		t.Errorf("UniqueCount: got %d, want 3", uc.UniqueCount())
	}
	if 3 != uc.Total() {
		t.Errorf("Total: got %d, want 3", uc.Total())
	}
}

func TestToHashablePreservesType(t *testing.T) {
	t.Parallel()

	// Hashable values should be returned as-is, preserving their original type.
	strVal := toHashable("foo")
	if _, ok := strVal.(string); !ok {
		t.Errorf("expected string type, got %T", strVal)
	}

	intVal := toHashable(42)
	if _, ok := intVal.(int); !ok {
		t.Errorf("expected int type, got %T", intVal)
	}

	// int(1) and string("1") should remain distinct.
	uc := NewUniqueCounts()
	uc.Observe(1)
	uc.Observe("1")
	if 2 != uc.UniqueCount() {
		t.Errorf("int(1) and string(\"1\") should be distinct, got UniqueCount=%d", uc.UniqueCount())
	}
}

func TestTrackedFieldUserField(t *testing.T) {
	t.Parallel()
	uc := NewUniqueCounts()
	tf := trackedField{"level", UserField, uc}

	event := ffi.LogEvent{
		UserKvPairs: map[string]any{"level": "INFO"},
		AutoKvPairs: map[string]any{"level": "auto-value"},
	}
	tf.observeEvent(event)

	if 1 != uc.UniqueCount() {
		t.Fatalf("UniqueCount: got %d, want 1", uc.UniqueCount())
	}
	counts := uc.ValueCounts()
	if 1 != counts["INFO"] {
		t.Errorf("expected UserKvPairs value 'INFO', got counts: %v", counts)
	}
}

func TestTrackedFieldAutoField(t *testing.T) {
	t.Parallel()
	uc := NewUniqueCounts()
	tf := trackedField{"timestamp", AutoField, uc}

	event := ffi.LogEvent{
		UserKvPairs: map[string]any{"timestamp": "user-value"},
		AutoKvPairs: map[string]any{"timestamp": "auto-value"},
	}
	tf.observeEvent(event)

	counts := uc.ValueCounts()
	if 1 != counts["auto-value"] {
		t.Errorf("expected AutoKvPairs value 'auto-value', got counts: %v", counts)
	}
}

func TestTrackedFieldMissingField(t *testing.T) {
	t.Parallel()
	uc := NewUniqueCounts()
	tf := trackedField{"missing", UserField, uc}

	event := ffi.LogEvent{
		UserKvPairs: map[string]any{"level": "INFO"},
		AutoKvPairs: map[string]any{},
	}
	tf.observeEvent(event)

	if 0 != uc.UniqueCount() {
		t.Errorf("UniqueCount: got %d, want 0 (nil should be ignored)", uc.UniqueCount())
	}
}

func TestFieldTrackerTrackAndCollector(t *testing.T) {
	t.Parallel()
	var ft fieldTracker

	uc := NewUniqueCounts()
	ft.Track("level", uc)

	got := ft.Collector("level")
	if got != uc {
		t.Error("Collector should return the registered collector")
	}

	if nil != ft.Collector("nonexistent") {
		t.Error("Collector should return nil for unregistered field")
	}
}

func TestFieldTrackerDefaultSource(t *testing.T) {
	t.Parallel()
	var ft fieldTracker

	uc := NewUniqueCounts()
	ft.Track("level", uc)

	if nil != ft.Collector("level", AutoField) {
		t.Error("Collector with AutoField should not match UserField registration")
	}
	if nil == ft.Collector("level") {
		t.Error("Collector with default source should match UserField registration")
	}
}

func TestFieldTrackerDuplicateReplaces(t *testing.T) {
	t.Parallel()
	var ft fieldTracker

	ucOld := NewUniqueCounts()
	ucNew := NewUniqueCounts()
	ft.Track("level", ucOld)
	ft.Track("level", ucNew)

	got := ft.Collector("level")
	if got != ucNew {
		t.Error("duplicate Track should replace the old collector")
	}
}

func TestFieldTrackerObserveAll(t *testing.T) {
	t.Parallel()
	var ft fieldTracker

	ucLevel := NewUniqueCounts()
	ucHost := NewUniqueCounts()
	ft.Track("level", ucLevel)
	ft.Track("host", ucHost, AutoField)

	event := ffi.LogEvent{
		UserKvPairs: map[string]any{"level": "WARN"},
		AutoKvPairs: map[string]any{"host": "server-1"},
	}
	ft.observeAll(event)

	if 1 != ucLevel.Total() {
		t.Errorf("level Total: got %d, want 1", ucLevel.Total())
	}
	if 1 != ucHost.Total() {
		t.Errorf("host Total: got %d, want 1", ucHost.Total())
	}
	levelCounts := ucLevel.ValueCounts()
	if 1 != levelCounts["WARN"] {
		t.Errorf("expected WARN=1, got: %v", levelCounts)
	}
	hostCounts := ucHost.ValueCounts()
	if 1 != hostCounts["server-1"] {
		t.Errorf("expected server-1=1, got: %v", hostCounts)
	}
}
