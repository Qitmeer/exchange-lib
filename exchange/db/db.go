package db

import (
	"encoding/json"
	"fmt"
	"github.com/Qitmeer/exchange-lib/exchange/db/base"
	"github.com/Qitmeer/exchange-lib/exchange/encode"
	sync2 "github.com/Qitmeer/exchange-lib/sync"
	"os"
	"sync"
)

const (
	block_bucket          = "block_bucket"
	coinbase_block_bucket = "coinbase_block_bucket"
	tx_bucket             = "tx_bucket"
	utxo_bucket           = "utxo_bucket"
	spent_bucket          = "spent_bucket"
	result_bucket         = "result_bucket"
	address_bucket        = "address_bucket"
	height_bucket         = "height_bucket"
)

type UTXODB struct {
	base         *base.Base
	mutex        sync.RWMutex
	addressCache map[string]bool
}

func NewUTXODB(path string) (*UTXODB, error) {
	base, err := base.Open(path)
	if err != nil {
		return nil, err
	}
	db := &UTXODB{base: base, addressCache: make(map[string]bool)}
	db.loadAddress()
	return db, nil
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

func (c *UTXODB) loadAddress() {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	addrs := c.getAllAddress()
	for _, addr := range addrs {
		c.addressCache[addr] = true
	}
}

func (c *UTXODB) GetAddresses() []string {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	return c.getAllAddress()
}

func (c *UTXODB) InsertAddress(address string) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.addressCache[address] = true

	return c.base.PutInBucket(address_bucket, []byte(address), []byte{})
}

func (c *UTXODB) AddressIsExist(address string) bool {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	_, ok := c.addressCache[address]

	return ok
}

func (c *UTXODB) getAllAddress() []string {
	addresses := []string{}
	iter := c.base.Iter(address_bucket)
	defer iter.Release()

	// Iter will affect RLP decoding and reallocate memory to value
	for iter.Next() {
		key := make([]byte, len(iter.Key()))
		copy(key, iter.Key())
		addresses = append(addresses, string(base.LeafKeyToKey(address_bucket, key)))
	}
	return addresses
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

func (c *UTXODB) UpdateAddressUTXOMandatory(address string, u *UTXO) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	err := c.saveAddressUTXO(address, u)
	if err != nil {
		return err
	}
	return nil
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

func (c *UTXODB) InsertSpentUTXO(spent *SpentUTXO) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	bytes, err := json.Marshal(spent)
	if err != nil {
		return err
	}
	return c.base.PutInBucket(spent_bucket, []byte(spent.SpentTxId), bytes)
}

func (c *UTXODB) DeleteSpentUTXO(txId string) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	return c.base.DeleteFromBucket(spent_bucket, []byte(txId))
}

func (c *UTXODB) GetSpents() []*SpentUTXO {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	rs := c.base.Foreach(spent_bucket)
	spents := []*SpentUTXO{}
	for _, value := range rs {
		spent := &SpentUTXO{}
		err := json.Unmarshal(value, spent)
		if err == nil {
			spents = append(spents, spent)
		}
	}
	return spents
}

func (c *UTXODB) GetAddressUTXO(address string, txId string, vout uint64) (*UTXO, error) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	return c.getAddressUTXO(address, txId, vout)
}

func (c *UTXODB) GetAddressUTXOs(address string, coin string, chainMainHeight uint64) ([]*UTXO, uint64, error) {
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
		if utxo != nil {
			if utxo.IsCoinBase && chainMainHeight-utxo.Height < sync2.DefaultCoinBaseThreshold {
				continue
			}
			if utxo.Spent == "" && utxo.Coin == coin && utxo.Lock < chainMainHeight {
				sum += utxo.Amount
				uxtos = append(uxtos, utxo)
			}
		}
	}
	return uxtos, sum, nil
}

func (c *UTXODB) GetAddressSpentUTXOs(address string, coin string) ([]*UTXO, uint64, error) {
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
		if utxo != nil && utxo.Spent != "" && utxo.Coin == coin {
			sum += utxo.Amount
			uxtos = append(uxtos, utxo)
		}
	}
	return uxtos, sum, nil
}

func (c *UTXODB) GetAddressLockUTXOs(address string, coin string, chainMainHeight uint64) ([]*UTXO, uint64, error) {
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
		if utxo.IsCoinBase && chainMainHeight-utxo.Height < sync2.DefaultCoinBaseThreshold {
			continue
		}
		if utxo.Spent == "" && utxo.Coin == coin && utxo.Lock >= chainMainHeight {
			sum += utxo.Amount
			uxtos = append(uxtos, utxo)
		}
	}
	return uxtos, sum, nil
}

func (c *UTXODB) SumUTXO(coin string) (uint64, error) {
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
		if utxo.Coin == coin && utxo.Spent == "" {
			sum += utxo.Amount
		}
	}
	return sum, nil
}

func (c *UTXODB) UpdateHeight(height uint64) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	bytes := encode.Uint64ToBytes(height)
	return c.base.PutInBucket(height_bucket, []byte(height_bucket), bytes)
}

func (c *UTXODB) GetHeight() (uint64, error) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	bytes, err := c.base.GetFromBucket(height_bucket, []byte(height_bucket))
	if err != nil {
		return 0, err
	}
	height := encode.BytesToUint64(bytes)
	return height, nil
}

type UTXO struct {
	TxId       string `json:"txid"`
	Vout       uint64 `json:"vout"`
	Address    string `json:"address"`
	Coin       string `json:"coin"`
	Amount     uint64 `json:"amount"`
	Spent      string `json:"spent"`
	Height     uint64 `json:"height"`
	Lock       uint64 `json:"lock"`
	IsCoinBase bool   `json:"iscoinbase"`
	PkHex      string `json:"pkhex"`
}

type SpentUTXO struct {
	SpentTxId string `json:"spenttxid"`
	UTXOList  []*UTXO
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
