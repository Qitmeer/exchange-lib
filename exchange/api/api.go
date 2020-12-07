package api

import (
	"encoding/json"
	"github.com/Qitmeer/exchange-lib/exchange/db"
	"github.com/Qitmeer/exchange-lib/sync"
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
	a.rest.AuthRouteSet("api/v1/transaction").Post(a.sendTransaction)
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
	} else {
		return nil, &Error{ERROR_UNKNOWN, err.Error()}
	}
	return txId, nil
}
