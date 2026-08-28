# ltsvparser

LTSV (Labeled Tab-separated Values) parser for Go language.

ltsvparser provides bare API for performance. it's able to parse Millions LTSV record in a second.

This code is used in mackerel-plugin-axslog.

## Reference

### Each

```
func Each(d []byte, cb func(int, []byte) error, keys ...[]byte) error
```

This API refers EachKey of https://github.com/buger/jsonparser.

`Each` parses `[]byte` payload and calls callback function when key is found.

On Go 1.27 and later, building with `GOEXPERIMENT=simd` enables an
architecture-specific SIMD implementation of `Each` on amd64 and arm64. The
API and parsing behavior are unchanged. Builds without the experiment continue
to use the standard implementation.

```console
GOEXPERIMENT=simd go test ./...
```

On amd64, add `GOAMD64=v3` to enable the instruction set required for the
optimized SIMD path:

```console
GOEXPERIMENT=simd GOAMD64=v3 go test ./...
```

```
func main() {
	data := `
time:05/Feb/2013:15:34:47 +0000	host:192.168.50.1	req:GET / HTTP/1.1	status:200	reqtime:0.030
time:05/Feb/2013:15:35:15 +0000	host:192.168.50.1	req:GET /foo HTTP/1.1   status:200	reqtime:0.050
time:05/Feb/2013:15:35:54 +0000	host:192.168.50.1	req:GET /bar HTTP/1.1   status:404	reqtime:0.110
`
	b := bytes.NewBufferString(data)
	var statusOK = 0
	var statusNotOK = 0
	var totalReqTime = 0.0
	scanner := bufio.NewScanner(b)
	for scanner.Scan() {
		err := ltsvparser.Each(
			scanner.Bytes(),
			func(idx int, v []byte) error {
				switch idx {
				case 0:
					// status
					if bytes.Equal(v, []byte("200")) {
						statusOK++
					} else {
						statusNotOK++
					}
				case 1:
					// reqtime
					rt, _ := ltsvparser.ParseFloat(v)
					totalReqTime = totalReqTime + rt
				}
				return nil
			},
			[]byte("status"),
			[]byte("reqtime"),
		)
		if err != nil {
			panic(err)
		}
	}
	fmt.Printf("OK:%d NotOK:%d TotalReqTime:%f\n", statusOK, statusNotOK, totalReqTime)
}
```

### ParseFloat

```
func ParseFloat(value []byte) (float64, error)
```

`ParseFloat` converts an LTSV value to a `float64`. Common decimal values are
parsed directly for performance; other valid Go floating-point formats fall
back to `strconv.ParseFloat` with 64-bit precision.

```go
requestTime, err := ltsvparser.ParseFloat([]byte("0.030"))
if err != nil {
	panic(err)
}
fmt.Println(requestTime) // 0.03
```

## Benchmarking

Parse 100k lines of LTSV

```
%  make bench                       
go test -bench '^BenchmarkParser' -benchmem -run=^./...
goos: darwin
goarch: arm64
pkg: github.com/kazeburo/go-ltsvparser-bench
cpu: Apple M3
BenchmarkParser_Ltsv-8         	       1	1572635458 ns/op	616011272 B/op	19000032 allocs/op
BenchmarkParser_GoLtsv-8       	       3	 409148389 ns/op	672009458 B/op	 7000018 allocs/op
BenchmarkParser_LtsvParser-8   	      12	  95090236 ns/op	    4216 B/op	       4 allocs/op
PASS
ok  	github.com/kazeburo/go-ltsvparser-bench	4.101s
```

With GOEXPERIMENT=simd
```
% GOEXPERIMENT=simd make bench      
go test -bench '^BenchmarkParser' -benchmem -run=^./...
goos: darwin
goarch: arm64
pkg: github.com/kazeburo/go-ltsvparser-bench
cpu: Apple M3
BenchmarkParser_Ltsv-8         	       1	1556602333 ns/op	616011416 B/op	19000033 allocs/op
BenchmarkParser_GoLtsv-8       	       3	 410863583 ns/op	672010536 B/op	 7000027 allocs/op
BenchmarkParser_LtsvParser-8   	      13	  88663721 ns/op	    4216 B/op	       4 allocs/op
PASS
ok  	github.com/kazeburo/go-ltsvparser-bench	4.108s
```



## Link

- http://ltsv.org/
- https://github.com/najeira/ltsv LTSV (Labeled Tab-separated Values) reader/writer for Go language.
- https://github.com/Songmu/go-ltsv LTSV parser and encoder for Go with reflection
