package ltsvparser

import (
	"bytes"
	"errors"
)

var byteNULL = []byte("")

// Canceler is used to stop parser without errors
type Canceler struct{} //nolint:errname

type CallBackFunc func(int, []byte) error

func (e *Canceler) Error() string {
	return ""
}

// Canceler is used to stop parser without errors
var Cancel = &Canceler{} //nolint:errname
// non public variable for errors.As check
var cancel = &Canceler{} //nolint:errname

func matchAndCallback(
	key []byte,
	value []byte,
	callback CallBackFunc,
	keys [][]byte,
) error {
	for i := range keys {
		if bytes.Equal(key, keys[i]) {
			return callback(i, value)
		}
	}
	return nil
}

// Extract multiple keys from LTSV
// BEGIN-NOSCAN
// nolint:gocognit
func Each(d []byte, callback CallBackFunc, keys ...[]byte) error {
	p1 := 0
	dlen := len(d)
	for dlen > p1 {
		p2 := bytes.IndexByte(d[p1:], '\t')
		if p2 == 0 {
			p1++
			continue
		}

		var field []byte
		if p2 < 0 {
			field = d[p1:]
			p1 = dlen
		} else {
			field = d[p1 : p1+p2]
			p1 = p1 + p2 + 1
		}

		p3 := bytes.IndexByte(field, ':')
		var key, value []byte
		if p3 < 0 {
			key = field
			value = byteNULL
		} else if p3+1 >= len(field) {
			key = field[:p3]
			value = byteNULL
		} else {
			key = field[:p3]
			value = field[p3+1:]
		}

		err := matchAndCallback(key, value, callback, keys)
		if err != nil {
			// Performance optimization: avoid errors.As check if err is not *Canceler
			if _, ok := err.(*Canceler); ok { //nolint:errorlint
				return nil
			}
			if errors.As(err, &cancel) {
				return nil
			}
			return err
		}
	}
	return nil
}
