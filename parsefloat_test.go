package ltsvparser

import (
	"math"
	"strconv"
	"testing"
	"unsafe"
)

func BenchmarkByteToFloat_parsefloat_string(b *testing.B) {
	f := []byte("123.456")
	var n float64
	for b.Loop() {
		n, _ = strconv.ParseFloat(string(f), 64)
	}
	_ = n
}

func BenchmarkByteToFloat_parsefloat_bytes(b *testing.B) {
	f := []byte("123.456")
	var n float64
	for b.Loop() {
		n, _ = strconv.ParseFloat(unsafe.String(unsafe.SliceData(f), len(f)), 64)
	}
	_ = n
}

func BenchmarkByteToFloat_parsefloat_bytes2(b *testing.B) {
	f := []byte("123.456")
	var n float64
	for b.Loop() {
		n, _ = strconv.ParseFloat(*(*string)(unsafe.Pointer(&f)), 64)
	}
	_ = n
}

func BenchmarkByteToFloat_parsefloat_fast(b *testing.B) {
	f := []byte("123.456")
	var n float64
	for b.Loop() {
		n, _ = ParseFloat(f)
	}
	_ = n
}

func TestParseFloat64(t *testing.T) {
	testCases := []struct {
		input string
		want  float64
	}{
		{"0", 0},
		{"123.456", 123.456},
		{"-123.456", -123.456},
		{".5", 0.5},
		{"1e3", 1_000},
		{"1234567890123456", 1_234_567_890_123_456},
	}

	for _, testCase := range testCases {
		got, err := ParseFloat([]byte(testCase.input))
		if err != nil {
			t.Fatalf("ParseFloat(%q) returned an error: %v", testCase.input, err)
		}
		if got != testCase.want {
			t.Errorf("ParseFloat(%q) = %v, want %v", testCase.input, got, testCase.want)
		}
	}
}

func TestParseFloat64MatchesStrconvForSimpleDecimals(t *testing.T) {
	testCases := []string{
		"-0",
		"0.1",
		".5",
		"1.",
		"123.456",
		"999999999999999",
		"1.2345678901234",
	}

	for _, input := range testCases {
		got, err := ParseFloat([]byte(input))
		if err != nil {
			t.Fatalf("ParseFloat(%q) returned an error: %v", input, err)
		}
		want, err := strconv.ParseFloat(input, 64)
		if err != nil {
			t.Fatalf("strconv.ParseFloat(%q) returned an error: %v", input, err)
		}
		if math.Float64bits(got) != math.Float64bits(want) {
			t.Errorf("ParseFloat(%q) = %v, want %v", input, got, want)
		}
	}
}
