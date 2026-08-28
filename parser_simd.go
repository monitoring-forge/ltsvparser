//go:build goexperiment.simd && (amd64 || arm64)

package ltsvparser

import (
	"bytes"
	"errors"
	"math/bits"
	"simd/archsimd"
)

func firstMatchingLane(matches archsimd.Uint8x16) int {
	words := matches.ReshapeToUint64s()
	if lo := words.GetElem(0); lo != 0 {
		return bits.TrailingZeros64(lo) / 8
	}
	if hi := words.GetElem(1); hi != 0 {
		return 8 + bits.TrailingZeros64(hi)/8
	}
	return -1
}

// scanFieldDelimitersSIMD finds the next tab and the first colon before it.
// A single vector load is shared by both comparisons. Once the first colon is
// found, the remaining single-delimiter search is delegated to bytes.IndexByte.
//nolint:gocognit // Keeping delimiter branches together avoids calls in the hot loop.
func scanFieldDelimitersSIMD(
	d []byte,
	tabVector archsimd.Uint8x16,
	colonVector archsimd.Uint8x16,
) (tab int, colon int) {
	const vectorSize = 16

	colon = -1
	i := 0
	for ; i+vectorSize <= len(d); i += vectorSize {
		chunk := archsimd.LoadUint8x16(d[i:])
		tabMatches := chunk.Equal(tabVector).ToInt8x16().ToBits()
		colonMatches := chunk.Equal(colonVector).ToInt8x16().ToBits()
		firstDelimiter := firstMatchingLane(tabMatches.Or(colonMatches))
		if firstDelimiter < 0 {
			continue
		}
		if d[i+firstDelimiter] != ':' {
			return i + firstDelimiter, colon
		}

		colon = i + firstDelimiter
		if tabLane := firstMatchingLane(tabMatches); tabLane >= 0 {
			return i + tabLane, colon
		}

		// IndexByte's architecture-specific assembly is faster for a long
		// value. The colon is already known, so delegate the remaining
		// single-delimiter search to it.
		next := i + vectorSize
		if tabTail := bytes.IndexByte(d[next:], '\t'); tabTail >= 0 {
			return next + tabTail, colon
		}
		return -1, colon
	}

	// A scalar tail avoids loading beyond the end of d and still searches for
	// both delimiters in one pass.
	for ; i < len(d); i++ {
		switch d[i] {
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

// Each parses an LTSV record and calls callback for every requested key found.
// This implementation is selected when GOEXPERIMENT=simd is enabled on amd64
// or arm64.
//nolint:gocognit // Parsing and callback error handling intentionally share one hot loop.
func Each(d []byte, callback CallBackFunc, keys ...[]byte) error {
	tabVector := archsimd.BroadcastUint8x16('\t')
	colonVector := archsimd.BroadcastUint8x16(':')

	for p1 := 0; p1 < len(d); {
		p2, p3 := scanFieldDelimitersSIMD(d[p1:], tabVector, colonVector)
		if p2 == 0 {
			p1++
			continue
		}

		end := len(d)
		if p2 >= 0 {
			end = p1 + p2
		}

		var key, value []byte
		if p3 < 0 {
			key = d[p1:end]
			value = byteNULL
		} else {
			colon := p1 + p3
			key = d[p1:colon]
			if colon+1 >= end {
				value = byteNULL
			} else {
				value = d[colon+1 : end]
			}
		}

		err := matchAndCallback(key, value, callback, keys)
		if err != nil {
			if _, ok := err.(*Canceler); ok { //nolint:errorlint
				return nil
			}
			if errors.As(err, &cancel) {
				return nil
			}
			return err
		}

		if p2 < 0 {
			break
		}
		p1 = end + 1
	}
	return nil
}
