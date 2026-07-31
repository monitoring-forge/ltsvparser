package ltsvparser

import (
	"bytes"
	"errors"
)

var byteTAB = []byte("\t")
var byteCOL = []byte(":")
var byteNULL = []byte("")

// Canceler is used to stop parser without errors
type Canceler struct{} //nolint:errname

type CallBackFunc func(int, []byte) error

func (e *Canceler) Error() string {
	return ""
}

// Canceler is used to stop parser without errors
var Cancel = &Canceler{} //nolint:errname

func seekField(d []byte, start int) (field []byte, next int) {
	p2 := bytes.Index(d[start:], byteTAB)
	if p2 == 0 {
		return nil, start + 1
	}
	if p2 < 0 {
		return d[start:], len(d)
	}
	return d[start : start+p2], start + p2 + 1
}

func splitField(field []byte) (key []byte, value []byte) {
	p3 := bytes.Index(field, byteCOL)
	if p3 < 0 {
		return field, byteNULL
	}
	if p3+1 >= len(field) {
		return field[:p3], byteNULL
	}
	return field[:p3], field[p3+1:]
}

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
func Each(d []byte, callback CallBackFunc, keys ...[]byte) error {
	p1 := 0
	dlen := len(d)
	for dlen > p1 {
		field, next := seekField(d, p1)
		p1 = next
		if field == nil {
			continue
		}
		key, value := splitField(field)
		err := matchAndCallback(key, value, callback, keys)
		if err != nil {
			var c *Canceler
			if errors.As(err, &c) {
				return nil
			}
			return err
		}
	}
	return nil
}
