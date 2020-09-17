package main

import (
	"fmt"
	"github.com/Qitmeer/exchange-lib/sync"
	"github.com/Qitmeer/exchange-lib/uxto"
	"time"
)

func main() {
	opt := &sync.Options{
		RpcAddr: "127.0.0.1:1234",
		RpcUser: "admin",
		RpcPwd:  "123",
		Https:   true,
		TxChLen: 100,
	}
	synchronizer := sync.NewSynchronizer(opt)
	txChan, err := synchronizer.Start(&sync.HistoryOrder{13000, 0})
	if err != nil {
		fmt.Printf(err.Error())
		return
	}
	go func() {
		for {
			txs := <-txChan
			for _, tx := range txs {
				// save tx or uxto
				utxos := uxto.GetUxtos(&tx)
				for _, uxto := range utxos {
					fmt.Println(uxto)
				}
			}
		}
	}()
	go func() {
		for {
			historyId := synchronizer.GetHistoryOrder()
			if historyId.LastTxBlockOrder != 0 {
				// save historyID as the start id of the next synchronization
			}
			// update historyid every 10s
			time.Sleep(time.Second * 10)
		}
	}()
	time.Sleep(1000000 * time.Second)
}
