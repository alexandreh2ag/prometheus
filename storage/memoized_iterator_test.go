// Copyright 2021 The Prometheus Authors
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package storage

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/prometheus/prometheus/tsdb/chunkenc"
)

func TestMemoizedSeriesIterator(t *testing.T) {
	// TODO(beorn7): Include histograms in testing.
	var it *MemoizedSeriesIterator

	sampleEq := func(ets int64, ev float64) {
		ts, v := it.At()
		require.Equal(t, ets, ts, "timestamp mismatch")
		require.Equal(t, ev, v, "value mismatch")
	}
	prevSampleEq := func(ets int64, ev float64, eok bool) {
		ts, v, _, _, ok := it.PeekPrev()
		require.Equal(t, eok, ok, "exist mismatch")
		require.Equal(t, ets, ts, "timestamp mismatch")
		require.Equal(t, ev, v, "value mismatch")
	}

	it = NewMemoizedIterator(NewListSeriesIterator(samples{
		fSample{t: 1, f: 2},
		fSample{t: 2, f: 3},
		fSample{t: 3, f: 4},
		fSample{t: 4, f: 5},
		fSample{t: 5, f: 6},
		fSample{t: 99, f: 8},
		fSample{t: 100, f: 9},
		fSample{t: 101, f: 10},
	}), 2)

	require.Equal(t, it.Seek(-123), chunkenc.ValFloat, "seek failed")
	sampleEq(1, 2)
	prevSampleEq(0, 0, false)

	require.Equal(t, it.Next(), chunkenc.ValFloat, "next failed")
	sampleEq(2, 3)
	prevSampleEq(1, 2, true)

	require.Equal(t, it.Next(), chunkenc.ValFloat, "next failed")
	require.Equal(t, it.Next(), chunkenc.ValFloat, "next failed")
	require.Equal(t, it.Next(), chunkenc.ValFloat, "next failed")
	sampleEq(5, 6)
	prevSampleEq(4, 5, true)

	require.Equal(t, it.Seek(5), chunkenc.ValFloat, "seek failed")
	sampleEq(5, 6)
	prevSampleEq(4, 5, true)

	require.Equal(t, it.Seek(101), chunkenc.ValFloat, "seek failed")
	sampleEq(101, 10)
	prevSampleEq(100, 9, true)

	require.Equal(t, it.Next(), chunkenc.ValNone, "next succeeded unexpectedly")
	require.Equal(t, it.Seek(1024), chunkenc.ValNone, "seek succeeded unexpectedly")
}

func BenchmarkMemoizedSeriesIterator(b *testing.B) {
	// Simulate a 5 minute rate.
	it := NewMemoizedIterator(newFakeSeriesIterator(int64(b.N), 30), 5*60)

	b.SetBytes(16)
	b.ReportAllocs()
	b.ResetTimer()

	for it.Next() != chunkenc.ValNone {
		// scan everything
	}
	require.NoError(b, it.Err())
}
