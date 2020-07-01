package sync

import (
	"fmt"
	"github.com/Qitmeer/exchange-lib/sync"
	"os"
	"testing"
	"time"
)

func TestSynchronizer_Start(t *testing.T) {
	opt := &sync.Options{
		RpcAddr: "127.0.0.1:1234",
		RpcUser: "admin",
		RpcPwd:  "123",
		Https:   false,
		TxChLen: 100,
	}
	synchronizer := sync.NewSynchronizer(opt)
	txChan, err := synchronizer.Start(&sync.HistoryOrder{0, 0})
	if err != nil {
		fmt.Printf(err.Error())
		return
	}
	go func() {
		for {
			txs := <-txChan
			for _, tx := range txs {
				// save tx or uxto
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
}
