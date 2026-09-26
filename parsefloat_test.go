package ltsvparser

import (
	"math"
	"strconv"
	"testing"
	"unsafe"
)

var benchmarkCases = [][]byte{
	[]byte("0.001"),
	[]byte("0.321"),
	[]byte("0.000001"),
	[]byte("1.234567"),
	[]byte("-0.001"),
	[]byte("-0.000001"),
	[]byte("0"),
	[]byte("0.1"),
	[]byte("0.32"),
	[]byte("12.345"),
	[]byte("123.456"),
	[]byte("12.34567"),
	[]byte("12345678"),
	[]byte("1234.5678"),
	[]byte("12345678.123456"),
}

func BenchmarkByteToFloat_parsefloat_string(b *testing.B) {
	var n float64
	for i := range b.N {
		f := benchmarkCases[i%len(benchmarkCases)]
		n, _ = strconv.ParseFloat(string(f), 64)
	}
	_ = n
}

func BenchmarkByteToFloat_parsefloat_bytes(b *testing.B) {
	var n float64
	for i := range b.N {
		f := benchmarkCases[i%len(benchmarkCases)]
		n, _ = strconv.ParseFloat(unsafe.String(unsafe.SliceData(f), len(f)), 64)
	}
	_ = n
}

func BenchmarkByteToFloat_parsefloat_bytes2(b *testing.B) {
	var n float64
	for i := range b.N {
		f := benchmarkCases[i%len(benchmarkCases)]
		n, _ = strconv.ParseFloat(*(*string)(unsafe.Pointer(&f)), 64)
	}
	_ = n
}

func BenchmarkByteToFloat_parsefloat_fast(b *testing.B) {
	var n float64
	for i := range b.N {
		f := benchmarkCases[i%len(benchmarkCases)]
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
		"0.001",
		"-0.001",
		"0.000",
		"-0.000",
		"0.000001",
		"-0.000001",
		"-0.000000",
		"9.999",
		"9.999999",
		"1.234567",
		"+0.001",
		"+0.000001",
		"0.1e3",
		"0.123e-3",
		"1.234_56",
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

func TestParseFloatFixedMicroseconds(t *testing.T) {
	// All subsecond microsecond values, including negative zero.
	v := []byte("-0.000000")
	for n := range 1000000 {
		x := n
		for i := 8; i >= 3; i-- {
			v[i] = byte(x%10) + '0'
			x /= 10
		}
		checkFixedFloat(t, v[1:])
		checkFixedFloat(t, v)
	}
	for _, base := range []string{"1.234", "1.234567", "-1.234", "-1.234567"} {
		for pos := range len(base) {
			for c := range 256 {
				input := []byte(base)
				input[pos] = byte(c)
				checkFixedFloat(t, input)
			}
		}
	}
}

func checkFixedFloat(t *testing.T, input []byte) {
	t.Helper()
	want, wantErr := strconv.ParseFloat(string(input), 64)
	got, err := ParseFloat(input)
	if (err == nil) != (wantErr == nil) || math.Float64bits(got) != math.Float64bits(want) {
		t.Fatalf("%q: got %v (%v), want %v (%v)", input, got, err, want, wantErr)
	}
}
