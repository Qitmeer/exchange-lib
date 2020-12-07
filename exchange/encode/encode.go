package encode

import "encoding/binary"

func Uint64ToBytes(val uint64) []byte {
	var buf = make([]byte, 8)
	binary.BigEndian.PutUint64(buf, val)
	return buf
}

func BytesToUint64(val []byte) uint64 {
	return binary.BigEndian.Uint64(val)
}
