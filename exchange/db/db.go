package db

import (
	"encoding/json"
	"fmt"
	"github.com/Qitmeer/exchange-lib/exchange/db/base"
	"github.com/Qitmeer/exchange-lib/exchange/encode"
	"os"
	"sync"
)

const (
	block_bucket          = "block_bucket"
	coinbase_block_bucket = "coinbase_block_bucket"
	tx_bucket             = "tx_bucket"
	utxo_bucket           = "utxo_bucket"
	result_bucket         = "result_bucket"
)

type UTXODB struct {
	base  *base.Base
	mutex sync.RWMutex
}

func NewUTXODB(path string) (*UTXODB, error) {
	base, err := base.Open(path)
	if err != nil {
		return nil, err
	}
	return &UTXODB{base: base}, nil
}

func (c *UTXODB) Close() {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.base.Close()
}

func (c *UTXODB) Clear() error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	return os.RemoveAll("data")
}

func (c *UTXODB) LastBlockOrder() uint64 {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	bytes, err := c.base.GetFromBucket(block_bucket, []byte(block_bucket))
	if err != nil {
		return 0
	}
	return encode.BytesToUint64(bytes)
}

func (c *UTXODB) LastCoinBaseBlockOrder() uint64 {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	bytes, err := c.base.GetFromBucket(coinbase_block_bucket, []byte(coinbase_block_bucket))
	if err != nil {
		return 0
	}
	return encode.BytesToUint64(bytes)
}

func (c *UTXODB) UpdateLastOrder(order uint64) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.base.PutInBucket(block_bucket, []byte(block_bucket), encode.Uint64ToBytes(order))
}

func (c *UTXODB) UpdateCoinBaseLastOrder(order uint64) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.base.PutInBucket(coinbase_block_bucket, []byte(coinbase_block_bucket), encode.Uint64ToBytes(order))
}

func (c *UTXODB) AddWrong(w *Wrong) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	bytes, _ := w.Bytes()
	c.base.PutInBucket(result_bucket, []byte(w.Hash), bytes)
}

func (c *UTXODB) WrongList() []Wrong {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	rs := c.base.Foreach(result_bucket)
	wrongs := make([]Wrong, 0)
	for _, value := range rs {
		w, _ := BytesToWrong(value)
		wrongs = append(wrongs, *w)
	}
	return wrongs
}

func (c *UTXODB) getAddressUTXO(address, txId string, index uint64) (*UTXO, error) {
	bytes, err := c.base.GetFromBucket(getUTXOBucket(address), []byte(getOutKey(txId, index)))
	if err != nil {
		return nil, err
	}
	var utxo *UTXO
	err = json.Unmarshal(bytes, &utxo)
	if err != nil {
		return nil, err
	}
	return utxo, nil
}

func (c *UTXODB) GetUTXO(txId string, index uint64) (*UTXO, error) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	bytes, err := c.base.GetFromBucket(utxo_bucket, []byte(getOutKey(txId, index)))
	if err != nil {
		return nil, err
	}
	var utxo *UTXO
	err = json.Unmarshal(bytes, &utxo)
	if err != nil {
		return nil, err
	}
	return utxo, nil
}

func (c *UTXODB) saveAddressUTXO(address string, uxto *UTXO) error {
	bytes, err := json.Marshal(uxto)
	if err != nil {
		return err
	}
	return c.base.PutInBucket(getUTXOBucket(address), []byte(getOutKey(uxto.TxId, uxto.Vout)), bytes)
}

func (c *UTXODB) SaveUTXO(uxto *UTXO) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	bytes, err := json.Marshal(uxto)
	if err != nil {
		return err
	}
	return c.base.PutInBucket(utxo_bucket, []byte(getOutKey(uxto.TxId, uxto.Vout)), bytes)
}

func (c *UTXODB) UpdateAddressUTXO(address string, u *UTXO) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	utxo, _ := c.getAddressUTXO(address, u.TxId, u.Vout)
	if utxo != nil {
		if u.Spent == "" {
			u.Spent = utxo.Spent
		}
	}
	err := c.saveAddressUTXO(address, u)
	if err != nil {
		return err
	}
	return nil
}

func (c *UTXODB) GetAddressUTXOs(address string) ([]*UTXO, uint64, error) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	var sum uint64
	uxtos := []*UTXO{}
	iter := c.base.Iter(getUTXOBucket(address))
	defer iter.Release()

	// Iter will affect RLP decoding and reallocate memory to value
	for iter.Next() {
		value := make([]byte, len(iter.Value()))
		copy(value, iter.Value())
		var utxo *UTXO
		err := json.Unmarshal(value, &utxo)
		if err != nil {
			return nil, 0, err
		}
		if utxo != nil && utxo.Spent == "" {
			sum += utxo.Amount
			uxtos = append(uxtos, utxo)
		}
	}
	return uxtos, sum, nil
}

func (c *UTXODB) SumUTXO() (uint64, error) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	var sum uint64
	iter := c.base.Iter(utxo_bucket)
	defer iter.Release()

	// Iter will affect RLP decoding and reallocate memory to value
	for iter.Next() {
		value := make([]byte, len(iter.Value()))
		copy(value, iter.Value())
		var utxo *UTXO
		err := json.Unmarshal(value, &utxo)
		if err != nil {
			return 0, err
		}
		if utxo.Spent == "" {
			sum += utxo.Amount
		}
	}
	return sum, nil
}

type UTXO struct {
	TxId    string `json:"txid"`
	Vout    uint64 `json:"vout"`
	Address string `json:"address"`
	Amount  uint64 `json:"amount"`
	Spent   string `json:"spent"`
}

type Wrong struct {
	Order       uint64
	Hash        string
	Coinbase    uint64
	CalCoinbase uint64
}

func (w *Wrong) Bytes() ([]byte, error) {
	return json.Marshal(w)
}

func BytesToWrong(bytes []byte) (*Wrong, error) {
	var w *Wrong
	err := json.Unmarshal(bytes, &w)
	if err != nil {
		return nil, err
	}
	return w, nil
}

func getOutKey(txId string, idx interface{}) string {
	return fmt.Sprintf("%s-%d", txId, idx)
}

func getUTXOBucket(address string) string {
	return fmt.Sprintf("%s-%s", tx_bucket, address)
}
