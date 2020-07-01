package sync

import (
	"fmt"
	"github.com/Qitmeer/exchange-lib/uxto"
	"testing"
	"time"
)

func TestSynchronizer_Start(t *testing.T) {
	opt := &Options{
		RpcAddr: "127.0.0.1:1234",
		RpcUser: "admin",
		RpcPwd:  "123",
		Https:   true,
		TxChLen: 100,
	}
	synchronizer := NewSynchronizer(opt)
	txChan, err := synchronizer.Start(&HistoryOrder{0, 0})
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
				fmt.Printf("%v\n", utxos)
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
	time.Sleep(1000 * time.Second)
}
