package rpc

import (
	"time"
)

type Block struct {
	Id            uint64        `json:"-"`
	Hash          string        `json:"hash"`
	Txsvalid      bool          `json:"txsvalid"`
	Confirmations uint32        `json:"confirmations"`
	Version       uint32        `json:"version"`
	Order         uint64        `json:"order"`
	Height        uint64        `json:"height"`
	TxRoot        string        `json:"txRoot"`
	Transactions  []Transaction `json:"transactions"`
	StateRoot     string        `json:"stateroot"`
	Bits          string        `json:"bits"`
	Difficulty    uint64        `json:"difficulty"`
	Nonce         uint64        `json:"nonce"`
	Timestamp     time.Time     `json:"timestamp"`
	ParentHash    []string      `json:"parents"`
	ChildrenHash  []string      `json:"children"`
}

type Transaction struct {
	Hex           string    `json:"hex"`
	Hexwit        string    `json:"hexwit"`
	Hexnowit      string    `json:"hexnowit"`
	Txid          string    `json:"txid"`
	Txhash        string    `json:"txhash"`
	Version       uint32    `json:"version"`
	Locktime      uint32    `json:"locktime"`
	Timestamp     time.Time `json:"timestamp"`
	Expire        uint32    `json:"expire"`
	Vin           []Vin     `json:"vin"`
	Vout          []Vout    `json:"vout"`
	Blockhash     string    `json:"blockhash"`
	BlockHeight   uint64    `json:"-"`
	BlockOrder    uint64    `json:"-"`
	Duplicate     bool      `json:"duplicate"`
	Confirmations uint32    `json:"confirmations"`
	IsCoinBase    bool      `json:"-"`
}

type Vin struct {
	Txid        string    `json:"txid"`
	Vout        uint64    `json:"vout"`
	Amountin    uint64    `json:"amountin"`
	Blockheight uint64    `json:"blockheight"`
	Blockindex  uint64    `json:"blockindex"`
	Coinbase    string    `json:"coinbase"`
	Txindex     uint64    `json:"txindex"`
	Sequence    uint64    `json:"sequence"`
	ScriptSig   ScriptSig `json:"scriptSig"`
}

type ScriptSig struct {
	Asm string `json:"asm"`
	Hex string `json:"hex"`
}

type Vout struct {
	Coin         string       `json:"coin"`
	CoinId       uint16       `json:"coinid"`
	Amount       uint64       `json:"amount"`
	ScriptPubKey ScriptPubKey `json:"scriptpubkey"`
}

type ScriptPubKey struct {
	Asm       string   `json:"asm"`
	Hex       string   `json:"hex"`
	ReqSigs   uint64   `json:"reqSigs"`
	Type      string   `json:"type"`
	Addresses []string `json:"addresses"`
}

type GraphState struct {
	Tips       []string `json:"tips"`
	Mainorder  uint64   `json:"mainorder"`
	Layer      uint64   `json:"layer"`
	MainHeight uint64   `json:"mainheight"`
}

type NodeInfo struct {
	Confirmations    uint32 `json:"confirmations"`
	Coinbasematurity uint32 `json:"coinbasematurity"`
	GraphState       `json:"graphstate"`
}
