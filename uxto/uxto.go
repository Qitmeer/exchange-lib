package uxto

import "github.com/Qitmeer/exchange-lib/rpc"

type Utxo struct {
	TxId    string
	TxIndex uint32
	Amount  uint64
	Address string
}

func GetUxtos(tx *rpc.Transaction) []*Utxo {
	utxos := make([]*Utxo, len(tx.Vout))
	for i, out := range tx.Vout {
		utxos[i] = &Utxo{
			TxId:    tx.Txid,
			TxIndex: uint32(i),
			Amount:  out.Amount,
			Address: out.ScriptPubKey.Addresses[0],
		}
	}
	return utxos
}
