package api

import (
	"encoding/json"
	"github.com/Qitmeer/exchange-lib/exchange/db"
	"github.com/Qitmeer/exchange-lib/sync"
	"strconv"
)

type Api struct {
	rest         *RestApi
	storage      *db.UTXODB
	synchronizer *sync.Synchronizer
}

func NewApi(listen string, db *db.UTXODB, synchronizer *sync.Synchronizer) (*Api, error) {
	return &Api{
		rest:         NewRestApi(listen),
		storage:      db,
		synchronizer: synchronizer,
	}, nil
}

func (a *Api) Run() error {
	a.addApi()
	return a.rest.Start()
}

func (a *Api) Stop() {
	a.rest.Stop()
}

func (a *Api) addApi() {
	a.rest.AuthRouteSet("api/v1/utxo").Get(a.getUTXO)
	a.rest.AuthRouteSet("api/v1/utxo/spent").Get(a.getSpentUTXO)
	a.rest.AuthRouteSet("api/v1/utxo").Post(a.updateUTXO)
	a.rest.AuthRouteSet("api/v1/transaction").Post(a.sendTransaction)
	a.rest.AuthRouteSet("api/v1/address").Post(a.addAddress)
	a.rest.AuthRouteSet("api/v1/address").Get(a.getAddress)
	a.rest.AuthRouteSet("api/v1/address/utxo").Get(a.getAddressUTXO)
}

func (a *Api) getUTXO(ct *Context) (interface{}, *Error) {
	addr, ok := ct.Query["address"]
	if !ok {
		return nil, &Error{ERROR_UNKNOWN, "address is required"}
	}
	utxos, balance, _ := a.storage.GetAddressUTXOs(addr)
	rs := map[string]interface{}{
		"utxo":    utxos,
		"balance": balance,
	}
	return rs, nil
}

func (a *Api) getSpentUTXO(ct *Context) (interface{}, *Error) {
	addr, ok := ct.Query["address"]
	if !ok {
		return nil, &Error{ERROR_UNKNOWN, "address is required"}
	}
	spent, amount, _ := a.storage.GetAddressSpentUTXOs(addr)
	rs := map[string]interface{}{
		"spent":  spent,
		"amount": amount,
	}
	return rs, nil
}

func (a *Api) updateUTXO(ct *Context) (interface{}, *Error) {
	txid, _ := ct.Form["txid"]
	if len(txid) == 0 {
		return nil, &Error{ERROR_UNKNOWN, "txid is required"}
	}
	vout, _ := ct.Form["vout"]
	if len(vout) == 0 {
		return nil, &Error{ERROR_UNKNOWN, "vout is required"}
	}
	amount, _ := ct.Form["amount"]
	if len(amount) == 0 {
		return nil, &Error{ERROR_UNKNOWN, "amount is required"}
	}
	address, _ := ct.Form["address"]
	if len(address) == 0 {
		return nil, &Error{ERROR_UNKNOWN, "address is required"}
	}
	spent, _ := ct.Form["spent"]
	iVout, err := strconv.ParseUint(vout, 10, 64)
	if err != nil {
		return nil, &Error{ERROR_UNKNOWN, "wrong vout"}
	}
	iAmount, err := strconv.ParseUint(amount, 10, 64)
	if err != nil {
		return nil, &Error{ERROR_UNKNOWN, "wrong amount"}
	}
	err = a.storage.UpdateAddressUTXOMandatory(address, &db.UTXO{
		TxId:    txid,
		Vout:    iVout,
		Address: address,
		Amount:  iAmount,
		Spent:   spent,
	})
	if err != nil {
		return nil, &Error{ERROR_UNKNOWN, err.Error()}
	}
	return true, nil
}

type Utxo struct {
}

func (a *Api) sendTransaction(ct *Context) (interface{}, *Error) {
	raw, ok := ct.Form["raw"]
	if !ok {
		return nil, &Error{ERROR_UNKNOWN, "raw is required"}
	}
	utxos, ok := ct.Form["spent"]
	if !ok {
		return nil, &Error{ERROR_UNKNOWN, "spent is required"}
	}
	utxoList := []*db.UTXO{}
	err := json.Unmarshal([]byte(utxos), &utxoList)
	if err != nil {
		return nil, &Error{ERROR_UNKNOWN, err.Error()}
	}
	txId, err := a.synchronizer.SendTx(raw)
	if err == nil {

		for _, utxo := range utxoList {
			utxo.Spent = txId
			a.storage.UpdateAddressUTXO(utxo.Address, utxo)
		}
		spentUtxo := &db.SpentUTXO{
			SpentTxId: txId,
			UTXOList:  utxoList,
		}
		a.storage.InsertSpentUTXO(spentUtxo)
	} else {
		return nil, &Error{ERROR_UNKNOWN, err.Error()}
	}
	return txId, nil
}

func (a *Api) addAddress(ct *Context) (interface{}, *Error) {
	addr, ok := ct.Form["address"]
	if !ok {
		return nil, &Error{ERROR_UNKNOWN, "address is required"}
	}
	err := a.storage.InsertAddress(addr)
	if err != nil {
		return nil, &Error{ERROR_UNKNOWN, err.Error()}
	}
	return addr, nil
}

func (a *Api) getAddress(ct *Context) (interface{}, *Error) {
	addresses := a.storage.GetAddresses()
	return addresses, nil
}

func (a *Api) getAddressUTXO(ct *Context) (interface{}, *Error) {
	address := ct.Query["address"]
	if len(address) == 0 {
		return nil, &Error{ERROR_UNKNOWN, "address is required"}
	}
	txid := ct.Query["txid"]
	if len(txid) == 0 {
		return nil, &Error{ERROR_UNKNOWN, "txid is required"}
	}
	vout := ct.Query["vout"]
	if len(vout) == 0 {
		return nil, &Error{ERROR_UNKNOWN, "vout is required"}
	}
	iVout, err := strconv.ParseUint(vout, 10, 64)
	if err != nil {
		return nil, &Error{ERROR_UNKNOWN, "wrong vout"}
	}
	utxo, err := a.storage.GetAddressUTXO(address, txid, iVout)
	if err != nil {
		return nil, &Error{ERROR_UNKNOWN, err.Error()}
	}
	return utxo, nil
}
