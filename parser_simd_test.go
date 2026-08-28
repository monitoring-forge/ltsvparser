//go:build goexperiment.simd && (amd64 || arm64)

package ltsvparser

import (
	"bytes"
	"math/rand/v2"
	"simd/archsimd"
	"testing"
)

func TestScanFieldDelimitersSIMD(t *testing.T) {
	tabVector := archsimd.BroadcastUint8x16('\t')
	colonVector := archsimd.BroadcastUint8x16(':')
	for range 1_000 {
		d := make([]byte, rand.IntN(1024))
		for i := range d {
			d[i] = byte(rand.IntN(256))
		}
		wantTab := bytes.IndexByte(d, '\t')
		field := d
		if wantTab >= 0 {
			field = d[:wantTab]
		}
		wantColon := bytes.IndexByte(field, ':')
		gotTab, gotColon := scanFieldDelimitersSIMD(d, tabVector, colonVector)
		if gotTab != wantTab || gotColon != wantColon {
			t.Fatalf("len=%d: got (%d, %d), want (%d, %d)", len(d), gotTab, gotColon, wantTab, wantColon)
		}
	}
}

func BenchmarkStandardEach_WithCancel(b *testing.B) {
	for b.Loop() {
		_ = eachStandard(input, func(i int, value []byte) error {
			if i > 0 {
				return Cancel
			}
			return nil
		}, pk, sk)
	}
}

func BenchmarkStandardEach_WithoutCancel(b *testing.B) {
	for b.Loop() {
		_ = eachStandard(input, func(int, []byte) error { return nil }, pk, sk)
	}
}

func scanFieldDelimitersScalar(d []byte) (tab int, colon int) {
	colon = -1
	for i, c := range d {
		switch c {
		case '\t':
			return i, colon
		case ':':
			if colon < 0 {
				colon = i
			}
		}
	}
	return -1, colon
}

func scanFieldDelimitersStdlib(d []byte) (tab int, colon int) {
	tab = bytes.IndexByte(d, '\t')
	field := d
	if tab >= 0 {
		field = d[:tab]
	}
	return tab, bytes.IndexByte(field, ':')
}

func BenchmarkFieldDelimiterScan(b *testing.B) {
	tabVector := archsimd.BroadcastUint8x16('\t')
	colonVector := archsimd.BroadcastUint8x16(':')
	cases := map[string][]byte{
		"first-field":  input,
		"middle-field": input[bytes.IndexByte(input, '\t')+1:],
		"last-field":   input[bytes.LastIndexByte(input, '\t')+1:],
	}
	for name, d := range cases {
		b.Run(name+"/stdlib-two-pass", func(b *testing.B) {
			for b.Loop() {
				_, _ = scanFieldDelimitersStdlib(d)
			}
		})
		b.Run(name+"/scalar-one-pass", func(b *testing.B) {
			for b.Loop() {
				_, _ = scanFieldDelimitersScalar(d)
			}
		})
		b.Run(name+"/simd-one-pass", func(b *testing.B) {
			for b.Loop() {
				_, _ = scanFieldDelimitersSIMD(d, tabVector, colonVector)
			}
		})
	}
}
