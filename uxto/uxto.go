package uxto

import "github.com/Qitmeer/exchange-lib/rpc"

type Utxo struct {
	Coin    string
	CoinId  uint16
	TxId    string
	TxIndex uint32
	Amount  uint64
	Address string
	Height  uint64
}

type Spent struct {
	TxId string
	Vout uint64
}

func GetUxtos(tx *rpc.Transaction) []*Utxo {
	utxos := make([]*Utxo, 0)
	for i, out := range tx.Vout {

		switch out.ScriptPubKey.Type {
		case "pubkeyhash":
		case "cltvpubkeyhash":
			utxos = append(utxos, &Utxo{
				TxId:    tx.Txid,
				TxIndex: uint32(i),
				Coin:    out.Coin,
				CoinId:  out.CoinId,
				Amount:  out.Amount,
				Address: out.ScriptPubKey.Addresses[0],
				Height:  tx.BlockHeight,
			})
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
				Coin:    out.Coin,
				CoinId:  out.CoinId,
				Amount:  out.Amount,
				Address: out.ScriptPubKey.Addresses[0],
				Height:  tx.BlockHeight,
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
