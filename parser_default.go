//go:build !goexperiment.simd || (!amd64 && !arm64)

package ltsvparser

// Each parses an LTSV record and calls callback for every requested key found.
func Each(d []byte, callback CallBackFunc, keys ...[]byte) error {
	return eachStandard(d, callback, keys...)
}
