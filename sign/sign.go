package sign

import (
	"fmt"
	"github.com/Qitmeer/qitmeer/core/types"
	"github.com/Qitmeer/qitmeer/qx"
	"time"
)

func TxEncode(version uint32, lockTime uint32, timestamp *time.Time, inputs map[string]uint32, outputs map[string]uint64, coin string) (string, error) {
	qxInputs := []qx.Input{}
	qxOutput := map[string]qx.Amount{}
	coinId := types.MEERID
	for txId, vout:= range inputs{
		qxInputs = append(qxInputs, qx.Input{
			TxID:     txId,
			OutIndex: vout,
		})
	}
	switch coin {
	case "MEER":
		coinId = types.MEERID
	default:
		return "", fmt.Errorf("incorrect coin name %s", coin)
	}
	for address, amount := range outputs{
		qxOutput[address] = qx.Amount{
			TargetLockTime: 0,
			Value:         int64(amount),
			Id:            coinId,
		}
	}
	return qx.TxEncode(version, lockTime, timestamp, qxInputs, qxOutput)
}

func TxSign(encode string, keys []string, network string, pks []string) (string, bool) {
	rs, err := qx.TxSign(keys, encode, network, pks)
	if err != nil {
		return err.Error(), false
	}
	return rs, true
}
