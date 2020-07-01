package sign

import (
	"github.com/Qitmeer/qitmeer/qx"
	"time"
)

func TxEncode(version uint32, lockTime uint32, timestamp *time.Time, inputs map[string]uint32, outputs map[string]uint64) (string, error) {
	return qx.TxEncode(version, lockTime, timestamp, inputs, outputs)
}

func TxSign(encode string, key string, network string) (string, bool) {
	rs, err := qx.TxSign(key, encode, network)
	if err != nil {
		return err.Error(), false
	}
	return rs, true
}
