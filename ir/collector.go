package ir

import (
	"errors"
	"fmt"
	"io"

	"github.com/y-scope/clp-ffi-go/ffi"
)

// FieldSource specifies which key-value pairs in a [ffi.LogEvent] to extract a
// field from.
type FieldSource int

const (
	// UserField extracts from [ffi.LogEvent.UserKvPairs].
	UserField FieldSource = iota
	// AutoField extracts from [ffi.LogEvent.AutoKvPairs].
	AutoField
)

// FieldCollector observes field values from log events and computes statistics.
// Implementations include [UniqueCounts] for exact value counts on
// low-cardinality fields, and [BloomFilter] for membership queries on
// high-cardinality fields using a cross-language compatible Split Block Bloom
// Filter (SBBF) with xxHash64.
type FieldCollector interface {
	Observe(fieldValue any)
}

// fieldTracker provides shared [Track] and [Collector] methods for [Reader] and
// [Writer].
type fieldTracker struct {
	trackedFields []trackedField
}

// Track registers a [FieldCollector] to observe the field named fieldName on
// every log event processed. By default the field is looked up in
// [ffi.LogEvent.UserKvPairs]. Pass [AutoField] to look up in
// [ffi.LogEvent.AutoKvPairs] instead. If a collector is already registered for
// the same fieldName and source, it is replaced.
func (ft *fieldTracker) Track(
	fieldName string,
	collector FieldCollector,
	source ...FieldSource,
) {
	src := UserField
	if 0 < len(source) {
		src = source[0]
	}
	for i := range ft.trackedFields {
		if ft.trackedFields[i].name == fieldName &&
			ft.trackedFields[i].source == src {
			ft.trackedFields[i].collector = collector
			return
		}
	}
	ft.trackedFields = append(
		ft.trackedFields,
		trackedField{fieldName, src, collector},
	)
}

// Collector returns the [FieldCollector] registered for fieldName, or nil if
// none is registered. If a [FieldSource] is provided it must also match;
// otherwise [UserField] is assumed.
func (ft *fieldTracker) Collector(
	fieldName string,
	source ...FieldSource,
) FieldCollector {
	src := UserField
	if 0 < len(source) {
		src = source[0]
	}
	for i := range ft.trackedFields {
		if ft.trackedFields[i].name == fieldName &&
			ft.trackedFields[i].source == src {
			return ft.trackedFields[i].collector
		}
	}
	return nil
}

// observeAll calls [trackedField.observeEvent] for all registered collectors.
func (ft *fieldTracker) observeAll(event ffi.LogEvent) {
	for i := range ft.trackedFields {
		ft.trackedFields[i].observeEvent(event)
	}
}

// closeAll closes all registered collectors that implement [io.Closer]. Always
// attempts all collectors and returns a joined error if any fail.
func (ft *fieldTracker) closeAll() error {
	var errs []error
	for i := range ft.trackedFields {
		if c, ok := ft.trackedFields[i].collector.(io.Closer); ok {
			if err := c.Close(); nil != err {
				errs = append(errs, err)
			}
		}
	}
	return errors.Join(errs...)
}

// trackedField pairs a field name and source with its collector.
type trackedField struct {
	name      string
	source    FieldSource
	collector FieldCollector
}

// observeEvent extracts the named field from a [ffi.LogEvent] based on the
// [FieldSource] and calls [FieldCollector.Observe].
func (tf *trackedField) observeEvent(event ffi.LogEvent) {
	var pairs map[string]any
	if AutoField == tf.source {
		pairs = event.AutoKvPairs
	} else {
		pairs = event.UserKvPairs
	}
	if v, ok := pairs[tf.name]; ok {
		tf.collector.Observe(v)
	} else {
		tf.collector.Observe(nil)
	}
}

// UniqueCounts tracks value counts for a field. Suitable for low-cardinality
// fields like "level" where the number of distinct values is small. Counts are
// exact for comparable values (strings, numbers, bools). Unhashable values
// (slices, maps) are converted to their string representation, so distinct
// values that format identically will be counted as one.
type UniqueCounts struct {
	counts map[any]int
	total  int
}

// NewUniqueCounts creates a new [UniqueCounts].
func NewUniqueCounts() *UniqueCounts {
	return &UniqueCounts{counts: make(map[any]int)}
}

// Observe records a field value. Nil values (field not present) are ignored.
// Unhashable values (slices, maps) are converted to their string representation.
func (uc *UniqueCounts) Observe(fieldValue any) {
	if nil == fieldValue {
		return
	}
	key := toHashable(fieldValue)
	uc.counts[key]++
	uc.total++
}

// UniqueCount returns the number of distinct values observed.
func (uc *UniqueCounts) UniqueCount() int {
	return len(uc.counts)
}

// ValueCounts returns a copy of the value-to-count mapping.
func (uc *UniqueCounts) ValueCounts() map[any]int {
	result := make(map[any]int, len(uc.counts))
	for k, v := range uc.counts {
		result[k] = v
	}
	return result
}

// Total returns the total number of observations (excluding nil).
func (uc *UniqueCounts) Total() int {
	return uc.total
}

// toHashable returns v if it is comparable and usable as a map key, otherwise
// falls back to its fmt.Sprint string representation. Note that different types
// formatting to the same string (e.g., int(1) vs string("1")) will be treated
// as the same key in the fallback path.
func toHashable(v any) (key any) {
	defer func() {
		if nil != recover() {
			key = fmt.Sprintf("%v", v)
		}
	}()
	_ = map[any]struct{}{v: {}}
	return v
}
