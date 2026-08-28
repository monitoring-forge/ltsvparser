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
                    // Also can use jsonparser.ParseFloat
					rt, _ := strconv.ParseFloat(string(v)), 64)
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

## Benchmarking

Benchmark codes https://gist.github.com/kazeburo/204efec4fab4a781f887ffa3d08a69c1

Parse 100k lines of LTSV

```
% make bench
go test -bench . -benchmem -run=^./...
goos: darwin
goarch: arm64
pkg: github.com/kazeburo/go-ltsvparser-bench
cpu: Apple M3
BenchmarkLtsv-8                        1        1596622375 ns/op        616017072 B/op  19000042 allocs/op
BenchmarkGoLtsv-8                      3         397009570 ns/op        672006845 B/op   7000022 allocs/op
BenchmarkLtsvParser-8                 10         110954279 ns/op            4216 B/op          4 allocs/op
PASS
ok      github.com/kazeburo/go-ltsvparser-bench 4.186s
```

With GOEXPERIMENT=simd
```
% GOAMD64=v3 GOEXPERIMENT=simd make bench
go test -bench . -benchmem -run=^./...
goos: darwin
goarch: arm64
pkg: github.com/kazeburo/go-ltsvparser-bench
cpu: Apple M3
BenchmarkLtsv-8                        1        1577866375 ns/op        616016528 B/op  19000037 allocs/op
BenchmarkGoLtsv-8                      3         413510764 ns/op        672007074 B/op   7000024 allocs/op
BenchmarkLtsvParser-8                 12          97328417 ns/op            4216 B/op          4 allocs/op
PASS
ok      github.com/kazeburo/go-ltsvparser-bench 4.193s
```

## Link

http://ltsv.org/

https://github.com/najeira/ltsv LTSV (Labeled Tab-separated Values) reader/writer for Go language.

https://github.com/Songmu/go-ltsv LTSV parser and encoder for Go with reflection
