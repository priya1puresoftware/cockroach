// Copyright 2014 The Cockroach Authors.
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package storage

import (
	"bytes"
	"context"
	"fmt"
	"reflect"
	"strconv"
	"sync/atomic"
	"testing"

	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/storage/enginepb"
	"github.com/cockroachdb/cockroach/pkg/testutils"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
	"github.com/cockroachdb/cockroach/pkg/util/iterutil"
	"github.com/cockroachdb/cockroach/pkg/util/leaktest"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/protoutil"
	"github.com/cockroachdb/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func mvccKey(k interface{}) MVCCKey {
	switch k := k.(type) {
	case string:
		return MakeMVCCMetadataKey(roachpb.Key(k))
	case []byte:
		return MakeMVCCMetadataKey(roachpb.Key(k))
	case roachpb.Key:
		return MakeMVCCMetadataKey(k)
	case roachpb.RKey:
		return MakeMVCCMetadataKey(roachpb.Key(k))
	default:
		panic(fmt.Sprintf("unsupported type: %T", k))
	}
}

func mustMarshal(m protoutil.Message) []byte {
	b, err := protoutil.Marshal(m)
	if err != nil {
		panic(err)
	}
	return b
}

func appender(s string) []byte {
	val := roachpb.MakeValueFromString(s)
	v := &enginepb.MVCCMetadataSubsetForMergeSerialization{RawBytes: val.RawBytes}
	return mustMarshal(v)
}

func testBatchBasics(t *testing.T, writeOnly bool, commit func(e Engine, b Batch) error) {
	for _, engineImpl := range mvccEngineImpls {
		t.Run(engineImpl.name, func(t *testing.T) {
			e := engineImpl.create()
			defer e.Close()

			var b Batch
			if writeOnly {
				b = e.NewUnindexedBatch(true /* writeOnly */)
			} else {
				b = e.NewBatch()
			}
			defer b.Close()

			require.NoError(t, b.PutUnversioned(mvccKey("a").Key, []byte("value")))

			// Write an engine value to be deleted.
			require.NoError(t, e.PutUnversioned(mvccKey("b").Key, []byte("value")))
			require.NoError(t, b.ClearUnversioned(mvccKey("b").Key))

			// Write an engine value to be merged.
			require.NoError(t, e.PutUnversioned(mvccKey("c").Key, appender("foo")))
			require.NoError(t, b.Merge(mvccKey("c"), appender("bar")))

			// Write a key with an empty value.
			require.NoError(t, b.PutUnversioned(mvccKey("e").Key, nil))

			// Write an engine value to be single deleted.
			require.NoError(t, e.PutUnversioned(mvccKey("d").Key, []byte("before")))
			require.NoError(t, b.SingleClearEngineKey(EngineKey{Key: mvccKey("d").Key}))

			// Check all keys are in initial state (nothing from batch has gone
			// through to engine until commit).
			expValues := []MVCCKeyValue{
				{Key: mvccKey("b"), Value: []byte("value")},
				{Key: mvccKey("c"), Value: appender("foo")},
				{Key: mvccKey("d"), Value: []byte("before")},
			}
			kvs, err := Scan(e, localMax, roachpb.KeyMax, 0)
			require.NoError(t, err)
			require.Equal(t, expValues, kvs)

			// Now, merged values should be:
			expValues = []MVCCKeyValue{
				{Key: mvccKey("a"), Value: []byte("value")},
				{Key: mvccKey("c"), Value: appender("foobar")},
				{Key: mvccKey("e"), Value: []byte{}},
			}
			if !writeOnly {
				// Scan values from batch directly.
				kvs, err = Scan(b, localMax, roachpb.KeyMax, 0)
				require.NoError(t, err)
				require.Equal(t, expValues, kvs)
			}

			// Commit batch and verify direct engine scan yields correct values.
			require.NoError(t, commit(e, b))
			kvs, err = Scan(e, localMax, roachpb.KeyMax, 0)
			require.NoError(t, err)
			require.Equal(t, expValues, kvs)
		})
	}
}

// TestBatchBasics verifies that all commands work in a batch, aren't
// visible until commit, and then are all visible after commit.
func TestBatchBasics(t *testing.T) {
	defer leaktest.AfterTest(t)()
	defer log.Scope(t).Close(t)
	testBatchBasics(t, false /* writeOnly */, func(e Engine, b Batch) error {
		return b.Commit(false /* sync */)
	})
}

func shouldPanic(t *testing.T, f func(), funcName string, expectedPanicStr string) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatalf("%v: test did not panic", funcName)
		} else if r != expectedPanicStr {
			t.Fatalf("%v: unexpected panic: %v", funcName, r)
		}
	}()
	f()
}

func shouldNotPanic(t *testing.T, f func(), funcName string) {
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("%v: unexpected panic: %v", funcName, r)
		}
	}()
	f()
}

// TestReadOnlyBasics verifies that for a read-only ReadWriter (obtained via
// engine.NewReadOnly()) all Reader methods work, and all Writer methods panic
// as "not implemented". Also basic iterating functionality is verified.
func TestReadOnlyBasics(t *testing.T) {
	defer leaktest.AfterTest(t)()
	defer log.Scope(t).Close(t)

	for _, engineImpl := range mvccEngineImpls {
		t.Run(engineImpl.name, func(t *testing.T) {
			e := engineImpl.create()
			defer e.Close()

			ro := e.NewReadOnly(StandardDurability)
			if ro.Closed() {
				t.Fatal("read-only is expectedly found to be closed")
			}
			a := mvccKey("a")
			successTestCases := []func(){
				func() {
					_ = ro.MVCCIterate(a.Key, a.Key, MVCCKeyIterKind, IterKeyTypePointsOnly,
						func(MVCCKeyValue, MVCCRangeKeyStack) error { return iterutil.StopIteration() })
				},
				func() { ro.NewMVCCIterator(MVCCKeyIterKind, IterOptions{UpperBound: roachpb.KeyMax}).Close() },
				func() {
					ro.NewMVCCIterator(MVCCKeyIterKind, IterOptions{
						MinTimestampHint: hlc.MinTimestamp,
						MaxTimestampHint: hlc.MaxTimestamp,
						UpperBound:       roachpb.KeyMax,
					}).Close()
				},
			}
			defer func(engineName string) {
				ro.Close()
				if !ro.Closed() {
					t.Fatal("even after calling Close, a read-only should not be closed")
				}
				name := "rocksDBReadOnly"
				if engineName == "pebble" {
					name = "pebbleReadOnly"
				}
				shouldPanic(t, func() { ro.Close() }, "Close", "closing an already-closed "+name)
				for i, f := range successTestCases {
					shouldPanic(t, f, strconv.Itoa(i), "using a closed "+name)
				}
			}(engineImpl.name)

			for i, f := range successTestCases {
				shouldNotPanic(t, f, strconv.Itoa(i))
			}

			// For a read-only ReadWriter, all Writer methods should panic.
			failureTestCases := []func(){
				func() { _ = ro.ApplyBatchRepr(nil, false) },
				func() { _ = ro.ClearUnversioned(a.Key) },
				func() { _ = ro.SingleClearEngineKey(EngineKey{Key: a.Key}) },
				func() { _ = ro.ClearRawRange(a.Key, a.Key, true, true) },
				func() { _ = ro.Merge(a, nil) },
				func() { _ = ro.PutUnversioned(a.Key, nil) },
			}
			for i, f := range failureTestCases {
				shouldPanic(t, f, strconv.Itoa(i), "not implemented")
			}

			if err := e.PutUnversioned(mvccKey("a").Key, []byte("value")); err != nil {
				t.Fatal(err)
			}
			if err := e.PutUnversioned(mvccKey("b").Key, []byte("value")); err != nil {
				t.Fatal(err)
			}
			if err := e.ClearUnversioned(mvccKey("b").Key); err != nil {
				t.Fatal(err)
			}
			if err := e.PutUnversioned(mvccKey("c").Key, appender("foo")); err != nil {
				t.Fatal(err)
			}
			if err := e.Merge(mvccKey("c"), appender("bar")); err != nil {
				t.Fatal(err)
			}
			if err := e.PutUnversioned(mvccKey("d").Key, []byte("value")); err != nil {
				t.Fatal(err)
			}
			if err := e.SingleClearEngineKey(EngineKey{Key: mvccKey("d").Key}); err != nil {
				t.Fatal(err)
			}

			// Now, merged values should be:
			expValues := []MVCCKeyValue{
				{Key: mvccKey("a"), Value: []byte("value")},
				{Key: mvccKey("c"), Value: appender("foobar")},
			}

			kvs, err := Scan(e, localMax, roachpb.KeyMax, 0)
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(expValues, kvs) {
				t.Errorf("%v != %v", kvs, expValues)
			}
		})
	}
}

func TestBatchRepr(t *testing.T) {
	defer leaktest.AfterTest(t)()
	defer log.Scope(t).Close(t)
	testBatchBasics(t, false /* writeOnly */, func(e Engine, b Batch) error {
		repr := b.Repr()

		r, err := NewPebbleBatchReader(repr)
		require.NoError(t, err)

		const expectedCount = 5
		require.Equal(t, expectedCount, r.Count())
		count, err := PebbleBatchCount(repr)
		require.NoError(t, err)
		require.Equal(t, expectedCount, count)

		var ops []string
		for i := 0; i < r.Count(); i++ {
			require.True(t, r.Next())
			switch r.BatchType() {
			case BatchTypeDeletion:
				ops = append(ops, fmt.Sprintf("delete(%s)", string(r.Key())))
			case BatchTypeValue:
				ops = append(ops, fmt.Sprintf("put(%s,%s)", string(r.Key()), string(r.Value())))
			case BatchTypeMerge:
				// The merge value is a protobuf and not easily displayable.
				ops = append(ops, fmt.Sprintf("merge(%s)", string(r.Key())))
			case BatchTypeSingleDeletion:
				ops = append(ops, fmt.Sprintf("single_delete(%s)", string(r.Key())))
			}
		}
		require.NoError(t, r.Error())
		require.False(t, r.Next())

		// The keys in the batch have the internal MVCC encoding applied which for
		// this test implies an appended 0 byte.
		expOps := []string{
			"put(a\x00,value)",
			"delete(b\x00)",
			"merge(c\x00)",
			"put(e\x00,)",
			"single_delete(d\x00)",
		}
		require.Equal(t, expOps, ops)

		return e.ApplyBatchRepr(repr, false /* sync */)
	})
}

func TestWriteBatchBasics(t *testing.T) {
	defer leaktest.AfterTest(t)()
	defer log.Scope(t).Close(t)
	testBatchBasics(t, true /* writeOnly */, func(e Engine, b Batch) error {
		return b.Commit(false /* sync */)
	})
}

// Regression test for flush issue which caused
// b2.ApplyBatchRepr(b1.Repr()).Repr() to not equal a noop.
func TestApplyBatchRepr(t *testing.T) {
	defer leaktest.AfterTest(t)()
	defer log.Scope(t).Close(t)

	for _, engineImpl := range mvccEngineImpls {
		t.Run(engineImpl.name, func(t *testing.T) {
			e := engineImpl.create()
			defer e.Close()

			// Failure to represent the absorbed Batch again.
			{
				b1 := e.NewBatch()
				defer b1.Close()

				if err := b1.PutUnversioned(mvccKey("lost").Key, []byte("update")); err != nil {
					t.Fatal(err)
				}

				repr1 := b1.Repr()

				b2 := e.NewBatch()
				defer b2.Close()
				if err := b2.ApplyBatchRepr(repr1, false /* sync */); err != nil {
					t.Fatal(err)
				}
				repr2 := b2.Repr()

				if !reflect.DeepEqual(repr1, repr2) {
					t.Fatalf("old batch represents to:\n%q\nrestored batch to:\n%q", repr1, repr2)
				}
			}

			// Failure to commit what was absorbed.
			{
				b3 := e.NewBatch()
				defer b3.Close()

				key := mvccKey("phantom")
				val := []byte("phantom")

				if err := b3.PutUnversioned(key.Key, val); err != nil {
					t.Fatal(err)
				}

				repr := b3.Repr()

				b4 := e.NewBatch()
				defer b4.Close()
				if err := b4.ApplyBatchRepr(repr, false /* sync */); err != nil {
					t.Fatal(err)
				}
				// Intentionally don't call Repr() because the expected user wouldn't.
				if err := b4.Commit(false /* sync */); err != nil {
					t.Fatal(err)
				}

				require.Equal(t, val, mvccGetRaw(t, e, key))
			}
		})
	}
}

func TestBatchGet(t *testing.T) {
	defer leaktest.AfterTest(t)()
	defer log.Scope(t).Close(t)

	for _, engineImpl := range mvccEngineImpls {
		t.Run(engineImpl.name, func(t *testing.T) {
			e := engineImpl.create()
			defer e.Close()

			b := e.NewBatch()
			defer b.Close()

			// Write initial values, then write to batch.
			if err := e.PutUnversioned(mvccKey("b").Key, []byte("value")); err != nil {
				t.Fatal(err)
			}
			if err := e.PutUnversioned(mvccKey("c").Key, appender("foo")); err != nil {
				t.Fatal(err)
			}
			// Write batch values.
			if err := b.PutUnversioned(mvccKey("a").Key, []byte("value")); err != nil {
				t.Fatal(err)
			}
			if err := b.ClearUnversioned(mvccKey("b").Key); err != nil {
				t.Fatal(err)
			}
			if err := b.Merge(mvccKey("c"), appender("bar")); err != nil {
				t.Fatal(err)
			}
			if err := b.PutUnversioned(mvccKey("d").Key, []byte("before")); err != nil {
				t.Fatal(err)
			}
			if err := b.SingleClearEngineKey(EngineKey{Key: mvccKey("d").Key}); err != nil {
				t.Fatal(err)
			}
			if err := b.PutUnversioned(mvccKey("d").Key, []byte("after")); err != nil {
				t.Fatal(err)
			}

			expValues := []MVCCKeyValue{
				{Key: mvccKey("a"), Value: []byte("value")},
				{Key: mvccKey("b"), Value: nil},
				{Key: mvccKey("c"), Value: appender("foobar")},
				{Key: mvccKey("d"), Value: []byte("after")},
			}
			for _, expKV := range expValues {
				require.Equal(t, expKV.Value, mvccGetRaw(t, b, expKV.Key))
			}
		})
	}
}

func compareMergedValues(t *testing.T, result, expected []byte) bool {
	var resultV, expectedV enginepb.MVCCMetadata
	if err := protoutil.Unmarshal(result, &resultV); err != nil {
		t.Fatal(err)
	}
	if err := protoutil.Unmarshal(expected, &expectedV); err != nil {
		t.Fatal(err)
	}
	return reflect.DeepEqual(resultV, expectedV)
}

func TestBatchMerge(t *testing.T) {
	defer leaktest.AfterTest(t)()
	defer log.Scope(t).Close(t)

	for _, engineImpl := range mvccEngineImpls {
		t.Run(engineImpl.name, func(t *testing.T) {
			e := engineImpl.create()
			defer e.Close()

			b := e.NewBatch()
			defer b.Close()

			// Write batch put, delete & merge.
			if err := b.PutUnversioned(mvccKey("a").Key, appender("a-value")); err != nil {
				t.Fatal(err)
			}
			if err := b.ClearUnversioned(mvccKey("b").Key); err != nil {
				t.Fatal(err)
			}
			if err := b.Merge(mvccKey("c"), appender("c-value")); err != nil {
				t.Fatal(err)
			}

			// Now, merge to all three keys.
			if err := b.Merge(mvccKey("a"), appender("append")); err != nil {
				t.Fatal(err)
			}
			if err := b.Merge(mvccKey("b"), appender("append")); err != nil {
				t.Fatal(err)
			}
			if err := b.Merge(mvccKey("c"), appender("append")); err != nil {
				t.Fatal(err)
			}

			// Verify values.
			val := mvccGetRaw(t, b, mvccKey("a"))
			if !compareMergedValues(t, val, appender("a-valueappend")) {
				t.Error("mismatch of \"a\"")
			}

			val = mvccGetRaw(t, b, mvccKey("b"))
			if !compareMergedValues(t, val, appender("append")) {
				t.Error("mismatch of \"b\"")
			}

			val = mvccGetRaw(t, b, mvccKey("c"))
			if !compareMergedValues(t, val, appender("c-valueappend")) {
				t.Error("mismatch of \"c\"")
			}
		})
	}
}

func TestBatchScan(t *testing.T) {
	defer leaktest.AfterTest(t)()
	defer log.Scope(t).Close(t)

	for _, engineImpl := range mvccEngineImpls {
		t.Run(engineImpl.name, func(t *testing.T) {
			e := engineImpl.create()
			defer e.Close()

			b := e.NewBatch()
			defer b.Close()

			existingVals := []MVCCKeyValue{
				{Key: mvccKey("a"), Value: []byte("1")},
				{Key: mvccKey("b"), Value: []byte("2")},
				{Key: mvccKey("c"), Value: []byte("3")},
				{Key: mvccKey("d"), Value: []byte("4")},
				{Key: mvccKey("e"), Value: []byte("5")},
				{Key: mvccKey("f"), Value: []byte("6")},
				{Key: mvccKey("g"), Value: []byte("7")},
				{Key: mvccKey("h"), Value: []byte("8")},
				{Key: mvccKey("i"), Value: []byte("9")},
				{Key: mvccKey("j"), Value: []byte("10")},
				{Key: mvccKey("k"), Value: []byte("11")},
				{Key: mvccKey("l"), Value: []byte("12")},
				{Key: mvccKey("m"), Value: []byte("13")},
			}
			for _, kv := range existingVals {
				if err := e.PutUnversioned(kv.Key.Key, kv.Value); err != nil {
					t.Fatal(err)
				}
			}

			batchVals := []MVCCKeyValue{
				{Key: mvccKey("a"), Value: []byte("b1")},
				{Key: mvccKey("bb"), Value: []byte("b2")},
				{Key: mvccKey("c"), Value: []byte("b3")},
				{Key: mvccKey("dd"), Value: []byte("b4")},
				{Key: mvccKey("e"), Value: []byte("b5")},
				{Key: mvccKey("ff"), Value: []byte("b6")},
				{Key: mvccKey("g"), Value: []byte("b7")},
				{Key: mvccKey("hh"), Value: []byte("b8")},
				{Key: mvccKey("i"), Value: []byte("b9")},
				{Key: mvccKey("jj"), Value: []byte("b10")},
			}
			for _, kv := range batchVals {
				if err := b.PutUnversioned(kv.Key.Key, kv.Value); err != nil {
					t.Fatal(err)
				}
			}

			scans := []struct {
				start, end roachpb.Key
				max        int64
			}{
				// Full monty.
				{start: roachpb.Key("a"), end: roachpb.Key("z"), max: 0},
				// Select ~half.
				{start: roachpb.Key("a"), end: roachpb.Key("z"), max: 9},
				// Select one.
				{start: roachpb.Key("a"), end: roachpb.Key("z"), max: 1},
				// Select half by end key.
				{start: roachpb.Key("a"), end: roachpb.Key("f0"), max: 0},
				// Start at half and select rest.
				{start: roachpb.Key("f"), end: roachpb.Key("z"), max: 0},
				// Start at last and select max=10.
				{start: roachpb.Key("m"), end: roachpb.Key("z"), max: 10},
			}

			// Scan each case using the batch and store the results.
			results := map[int][]MVCCKeyValue{}
			for i, scan := range scans {
				kvs, err := Scan(b, scan.start, scan.end, scan.max)
				if err != nil {
					t.Fatal(err)
				}
				results[i] = kvs
			}

			// Now, commit batch and re-scan using engine direct to compare results.
			if err := b.Commit(false /* sync */); err != nil {
				t.Fatal(err)
			}
			for i, scan := range scans {
				kvs, err := Scan(e, scan.start, scan.end, scan.max)
				if err != nil {
					t.Fatal(err)
				}
				if !reflect.DeepEqual(kvs, results[i]) {
					t.Errorf("%d: expected %v; got %v", i, results[i], kvs)
				}
			}
		})
	}
}

// TestBatchScanWithDelete verifies that a scan containing
// a single deleted value returns nothing.
func TestBatchScanWithDelete(t *testing.T) {
	defer leaktest.AfterTest(t)()
	defer log.Scope(t).Close(t)

	for _, engineImpl := range mvccEngineImpls {
		t.Run(engineImpl.name, func(t *testing.T) {
			e := engineImpl.create()
			defer e.Close()

			b := e.NewBatch()
			defer b.Close()

			// Write initial value, then delete via batch.
			if err := e.PutUnversioned(mvccKey("a").Key, []byte("value")); err != nil {
				t.Fatal(err)
			}
			if err := b.ClearUnversioned(mvccKey("a").Key); err != nil {
				t.Fatal(err)
			}
			kvs, err := Scan(b, localMax, roachpb.KeyMax, 0)
			if err != nil {
				t.Fatal(err)
			}
			if len(kvs) != 0 {
				t.Errorf("expected empty scan with batch-deleted value; got %v", kvs)
			}
		})
	}
}

// TestBatchScanMaxWithDeleted verifies that if a deletion
// in the updates map shadows an entry from the engine, the
// max on a scan is still reached.
func TestBatchScanMaxWithDeleted(t *testing.T) {
	defer leaktest.AfterTest(t)()
	defer log.Scope(t).Close(t)

	for _, engineImpl := range mvccEngineImpls {
		t.Run(engineImpl.name, func(t *testing.T) {
			e := engineImpl.create()
			defer e.Close()

			b := e.NewBatch()
			defer b.Close()

			// Write two values.
			if err := e.PutUnversioned(mvccKey("a").Key, []byte("value1")); err != nil {
				t.Fatal(err)
			}
			if err := e.PutUnversioned(mvccKey("b").Key, []byte("value2")); err != nil {
				t.Fatal(err)
			}
			// Now, delete "a" in batch.
			if err := b.ClearUnversioned(mvccKey("a").Key); err != nil {
				t.Fatal(err)
			}
			// A scan with max=1 should scan "b".
			kvs, err := Scan(b, localMax, roachpb.KeyMax, 1)
			if err != nil {
				t.Fatal(err)
			}
			if len(kvs) != 1 || !bytes.Equal(kvs[0].Key.Key, []byte("b")) {
				t.Errorf("expected scan of \"b\"; got %v", kvs)
			}
		})
	}
}

func TestBatchVisibleAfterApplyBatchRepr(t *testing.T) {
	defer leaktest.AfterTest(t)()
	defer log.Scope(t).Close(t)

	for _, engineImpl := range mvccEngineImpls {
		t.Run(engineImpl.name, func(t *testing.T) {
			e := engineImpl.create()
			defer e.Close()

			wb := func() []byte {
				batch := e.NewBatch()
				defer batch.Close()

				if err := batch.PutUnversioned(mvccKey("batchkey").Key, []byte("b")); err != nil {
					t.Fatal(err)
				}

				return batch.Repr()
			}()

			batch := e.NewBatch()
			defer batch.Close()

			assert.NoError(t, batch.ApplyBatchRepr(wb, false /* sync */))

			// The batch can see the earlier write.
			require.Equal(t, []byte("b"), mvccGetRaw(t, batch, mvccKey("batchkey")))
		})
	}
}

func TestUnindexedBatchThatSupportsReader(t *testing.T) {
	defer leaktest.AfterTest(t)()
	defer log.Scope(t).Close(t)

	for _, engineImpl := range mvccEngineImpls {
		t.Run(engineImpl.name, func(t *testing.T) {
			e := engineImpl.create()
			defer e.Close()

			if err := e.PutUnversioned(mvccKey("b").Key, []byte("b")); err != nil {
				t.Fatal(err)
			}

			b := e.NewUnindexedBatch(false /* writeOnly */)
			defer b.Close()
			if err := b.PutUnversioned(mvccKey("b").Key, []byte("c")); err != nil {
				t.Fatal(err)
			}

			// Verify that reads on the distinct batch go to the underlying engine, not
			// to the unindexed batch.
			iter := b.NewMVCCIterator(MVCCKeyIterKind, IterOptions{UpperBound: roachpb.KeyMax})
			iter.SeekGE(mvccKey("a"))
			if ok, err := iter.Valid(); !ok {
				t.Fatalf("expected iterator to be valid, err=%v", err)
			}
			if string(iter.Key().Key) != "b" {
				t.Fatalf("expected b, but got %s", iter.Key())
			}
			iter.Close()

			require.Equal(t, []byte("b"), mvccGetRaw(t, b, mvccKey("b")))
			require.NoError(t, b.Commit(true))
			require.Equal(t, []byte("c"), mvccGetRaw(t, e, mvccKey("b")))
		})
	}
}

func TestUnindexedBatchThatDoesNotSupportReaderPanics(t *testing.T) {
	defer leaktest.AfterTest(t)()
	defer log.Scope(t).Close(t)

	for _, engineImpl := range mvccEngineImpls {
		t.Run(engineImpl.name, func(t *testing.T) {
			e := engineImpl.create()
			defer e.Close()

			batch := e.NewUnindexedBatch(true /* writeOnly */)
			defer batch.Close()

			// The various Reader methods on the batch should panic.
			a := mvccKey("a")
			b := mvccKey("b")
			testCases := []func(){
				func() { _ = batch.MVCCIterate(a.Key, b.Key, MVCCKeyIterKind, IterKeyTypePointsOnly, nil) },
				func() { _ = batch.NewMVCCIterator(MVCCKeyIterKind, IterOptions{UpperBound: roachpb.KeyMax}) },
			}
			for i, f := range testCases {
				func() {
					defer func(i int) {
						if r := recover(); r == nil {
							t.Fatalf("%d: test did not panic", i)
						} else if r != "write-only batch" {
							t.Fatalf("%d: unexpected panic: %v", i, r)
						}
					}(i)
					f()
				}()
			}
		})
	}
}

func TestBatchIteration(t *testing.T) {
	defer leaktest.AfterTest(t)()
	defer log.Scope(t).Close(t)

	for _, engineImpl := range mvccEngineImpls {
		t.Run(engineImpl.name, func(t *testing.T) {
			e := engineImpl.create()
			defer e.Close()

			b := e.NewBatch()
			defer b.Close()

			k1 := MakeMVCCMetadataKey(roachpb.Key("c"))
			k2 := MakeMVCCMetadataKey(roachpb.Key("d"))
			k3 := MakeMVCCMetadataKey(roachpb.Key("e"))
			v1 := []byte("value1")
			v2 := []byte("value2")

			if err := b.PutUnversioned(k1.Key, v1); err != nil {
				t.Fatal(err)
			}
			if err := b.PutUnversioned(k2.Key, v2); err != nil {
				t.Fatal(err)
			}
			if err := b.PutUnversioned(k3.Key, []byte("doesn't matter")); err != nil {
				t.Fatal(err)
			}

			iterOpts := IterOptions{UpperBound: k3.Key}
			iter := b.NewMVCCIterator(MVCCKeyIterKind, iterOpts)
			defer iter.Close()

			// Forward iteration,
			iter.SeekGE(k1)
			if ok, err := iter.Valid(); !ok {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(iter.Key(), k1) {
				t.Fatalf("expected %s, got %s", k1, iter.Key())
			}
			if !reflect.DeepEqual(iter.Value(), v1) {
				t.Fatalf("expected %s, got %s", v1, iter.Value())
			}
			iter.Next()
			if ok, err := iter.Valid(); !ok {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(iter.Key(), k2) {
				t.Fatalf("expected %s, got %s", k2, iter.Key())
			}
			if !reflect.DeepEqual(iter.Value(), v2) {
				t.Fatalf("expected %s, got %s", v2, iter.Value())
			}
			iter.Next()
			if ok, err := iter.Valid(); err != nil {
				t.Fatal(err)
			} else if ok {
				t.Fatalf("expected invalid, got valid at key %s", iter.Key())
			}

			// Reverse iteration.
			switch engineImpl.name {
			case "pebble":
				// Reverse iteration in batches works on Pebble.
				iter.SeekLT(k3)
				if ok, err := iter.Valid(); !ok {
					t.Fatal(err)
				}
				if !reflect.DeepEqual(iter.Key(), k2) {
					t.Fatalf("expected %s, got %s", k2, iter.Key())
				}
				if !reflect.DeepEqual(iter.Value(), v2) {
					t.Fatalf("expected %s, got %s", v2, iter.Value())
				}

				iter.Prev()
				if ok, err := iter.Valid(); !ok || err != nil {
					t.Fatalf("expected success, but got invalid: %v", err)
				}
				if !reflect.DeepEqual(iter.Key(), k1) {
					t.Fatalf("expected %s, got %s", k1, iter.Key())
				}
				if !reflect.DeepEqual(iter.Value(), v1) {
					t.Fatalf("expected %s, got %s", v1, iter.Value())
				}
			default:
				// Reverse iteration in batches is not supported with RocksDB.
				iter.SeekLT(k3)
				if ok, err := iter.Valid(); ok {
					t.Fatalf("expected invalid, got valid at key %s", iter.Key())
				} else if !testutils.IsError(err, "SeekForPrev\\(\\) not supported") {
					t.Fatalf("expected 'SeekForPrev() not supported', got %s", err)
				}

				iter.Prev()
				if ok, err := iter.Valid(); ok {
					t.Fatalf("expected invalid, got valid at key %s", iter.Key())
				} else if !testutils.IsError(err, "Prev\\(\\) not supported") {
					t.Fatalf("expected 'Prev() not supported', got %s", err)
				}
			}
		})
	}
}

// Test combining of concurrent commits of write-only batches, verifying that
// all of the keys written by the individual batches are subsequently readable.
func TestBatchCombine(t *testing.T) {
	defer leaktest.AfterTest(t)()
	defer log.Scope(t).Close(t)

	for _, engineImpl := range mvccEngineImpls {
		t.Run(engineImpl.name, func(t *testing.T) {
			e := engineImpl.create()
			defer e.Close()

			var n uint32
			const count = 10000

			errs := make(chan error, 10)
			for i := 0; i < cap(errs); i++ {
				go func() {
					for {
						v := atomic.AddUint32(&n, 1) - 1
						if v >= count {
							break
						}
						k := fmt.Sprint(v)

						b := e.NewUnindexedBatch(true /* writeOnly */)
						if err := b.PutUnversioned(mvccKey(k).Key, []byte(k)); err != nil {
							errs <- errors.Wrap(err, "put failed")
							return
						}
						if err := b.Commit(false); err != nil {
							errs <- errors.Wrap(err, "commit failed")
							return
						}

						// Verify we can read the key we just wrote immediately.
						if v, err := mvccGetRawWithError(t, e, mvccKey(k)); err != nil {
							errs <- errors.Wrap(err, "get failed")
							return
						} else if string(v) != k {
							errs <- errors.Errorf("read %q from engine, expected %q", v, k)
							return
						}
					}
					errs <- nil
				}()
			}

			for i := 0; i < cap(errs); i++ {
				if err := <-errs; err != nil {
					t.Error(err)
				}
			}
		})
	}
}

func TestDecodeKey(t *testing.T) {
	defer leaktest.AfterTest(t)()
	defer log.Scope(t).Close(t)

	e, err := Open(context.Background(), InMemory(), CacheSize(1<<20 /* 1 MiB */))
	assert.NoError(t, err)
	defer e.Close()

	tests := []MVCCKey{
		{Key: []byte("foo")},
		{Key: []byte("foo"), Timestamp: hlc.Timestamp{WallTime: 1}},
		{Key: []byte("foo"), Timestamp: hlc.Timestamp{WallTime: 1, Logical: 1}},
		{Key: []byte("foo"), Timestamp: hlc.Timestamp{WallTime: 1, Logical: 1, Synthetic: true}},
	}
	for _, test := range tests {
		t.Run(test.String(), func(t *testing.T) {
			b := e.NewBatch()
			defer b.Close()
			if test.Timestamp.IsEmpty() {
				if err := b.PutUnversioned(test.Key, nil); err != nil {
					t.Fatalf("%+v", err)
				}
			} else {
				if err := b.PutMVCC(test, MVCCValue{}); err != nil {
					t.Fatalf("%+v", err)
				}
			}
			repr := b.Repr()

			r, err := NewPebbleBatchReader(repr)
			if err != nil {
				t.Fatalf("%+v", err)
			}
			if !r.Next() {
				t.Fatalf("could not get the first entry: %+v", r.Error())
			}
			decodedKey, err := DecodeMVCCKey(r.Key())
			if err != nil {
				t.Fatalf("unexpected err: %+v", err)
			}
			if !reflect.DeepEqual(test, decodedKey) {
				t.Errorf("expected %+v got %+v", test, decodedKey)
			}
		})
	}
}

func TestPebbleBatchReader(t *testing.T) {
	defer leaktest.AfterTest(t)()
	defer log.Scope(t).Close(t)

	DisableMetamorphicSimpleValueEncoding(t)

	eng := NewDefaultInMemForTesting()
	defer eng.Close()

	b := eng.NewBatch()
	defer b.Close()

	// Write some basic data.
	require.NoError(t, b.PutEngineKey(engineKey("engineKey", 0), []byte("engineValue")))
	require.NoError(t, b.PutMVCC(pointKey("mvccKey", 1), stringValue("mvccValue")))
	require.NoError(t, b.PutMVCC(pointKey("mvccTombstone", 1), MVCCValue{}))
	require.NoError(t, b.PutEngineRangeKey(roachpb.Key("rangeFrom"), roachpb.Key("rangeTo"), []byte{7}, []byte("engineRangeKey")))

	// Clear some already empty keys.
	require.NoError(t, b.ClearMVCC(pointKey("mvccKey", 9)))
	require.NoError(t, b.ClearMVCCRangeKey(rangeKey("rangeFrom", "rangeTo", 9)))
	require.NoError(t, b.ClearRawRange(roachpb.Key("clearFrom"), roachpb.Key("clearTo"), true, true))

	// Read it back.
	expect := []struct {
		batchType BatchType
		key       string
		ts        int
		endKey    string
		value     []byte
		rangeKeys []EngineRangeKeyValue
	}{
		{BatchTypeValue, "engineKey", 0, "", []byte("engineValue"), nil},
		{BatchTypeValue, "mvccKey", 1, "", stringValueRaw("mvccValue"), nil},
		{BatchTypeValue, "mvccTombstone", 1, "", []byte{}, nil},
		{BatchTypeRangeKeySet, "rangeFrom", 0, "rangeTo", nil, []EngineRangeKeyValue{
			{Version: []byte{7}, Value: []byte("engineRangeKey")},
		}},

		{BatchTypeDeletion, "mvccKey", 9, "", nil, nil},
		{BatchTypeRangeKeyUnset, "rangeFrom", 0, "rangeTo", nil, []EngineRangeKeyValue{
			{Version: EncodeMVCCTimestampSuffix(wallTS(9)), Value: nil},
		}},
		{BatchTypeRangeDeletion, "clearFrom", 0, "clearTo", []byte("clearTo\000"), nil},
		{BatchTypeRangeKeyDelete, "clearFrom", 0, "clearTo", []byte("clearTo\000"), nil},
	}

	r, err := NewPebbleBatchReader(b.Repr())
	require.NoError(t, err)

	require.Equal(t, len(expect), r.Count())

	for _, e := range expect {
		t.Logf("batchType=%v key=%s endKey=%s", e.batchType, e.key, e.endKey)
		require.True(t, r.Next())
		require.Equal(t, e.batchType, r.BatchType())

		key, err := r.MVCCKey()
		require.NoError(t, err)
		require.Equal(t, pointKey(e.key, e.ts), key)

		eKey, err := r.EngineKey()
		require.NoError(t, err)
		require.Equal(t, engineKey(e.key, e.ts), eKey)

		if len(e.endKey) > 0 {
			eKey, err = r.EngineEndKey()
			require.NoError(t, err)
			require.Equal(t, engineKey(e.endKey, e.ts), eKey)
		} else {
			_, err = r.EngineEndKey()
			require.Error(t, err)
		}

		switch e.batchType {
		case BatchTypeRangeKeySet, BatchTypeRangeKeyUnset:
			rkvs, err := r.EngineRangeKeys()
			require.NoError(t, err)
			require.Equal(t, e.rangeKeys, rkvs)
		case BatchTypeDeletion:
			require.Nil(t, e.value)
		default:
			require.Equal(t, e.value, r.Value())
		}
	}

	require.False(t, r.Next())
	require.NoError(t, r.Error())
}
