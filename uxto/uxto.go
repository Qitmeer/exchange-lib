package uxto

import "github.com/Qitmeer/exchange-lib/rpc"

type Utxo struct {
	TxId    string
	TxIndex uint32
	Amount  uint64
	Address string
}

type Spent struct {
	TxId string
	Vout uint64
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

func GetAddressUxtos(tx *rpc.Transaction, address map[string]bool) []*Utxo {
	utxos := []*Utxo{}
	for i, out := range tx.Vout {
		addr := out.ScriptPubKey.Addresses[0]
		_, ok := address[addr]
		if ok {
			utxos = append(utxos, &Utxo{
				TxId:    tx.Txid,
				TxIndex: uint32(i),
				Amount:  out.Amount,
				Address: addr,
			})
		}
	}
	return utxos
}

func GetSpentTxs(tx *rpc.Transaction) []*Spent {
	spents := []*Spent{}
	for _, vin := range tx.Vin {
		spents = append(spents, &Spent{
			TxId: vin.Txid,
			Vout: vin.Vout,
		})
	}
	return spents
}
