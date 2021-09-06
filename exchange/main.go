package main

import (
	"flag"
	"fmt"
	"github.com/Qitmeer/exchange-lib/exchange/api"
	"github.com/Qitmeer/exchange-lib/exchange/conf"
	"github.com/Qitmeer/exchange-lib/exchange/db"
	"github.com/Qitmeer/exchange-lib/exchange/version"
	"github.com/Qitmeer/exchange-lib/sync"
	"github.com/Qitmeer/exchange-lib/uxto"
	"github.com/bCoder778/log"
	"os"
	"os/signal"
	"strings"
	sync2 "sync"
	"time"
)

var interrupt chan struct{}

func main() {
	dealCommand()

	db, err := openDB("data")
	if err != nil {
		fmt.Println("failed to open db, ", err)
		os.Exit(1)
	}
	log.SetOption(&log.Option{
		LogLevel: conf.Setting.Log.Level,
		Mode:     conf.Setting.Log.Mode,
		Path:     conf.Setting.Log.Path,
	})
	opt := &sync.Options{
		RpcAddr: conf.Setting.Rpc.Host,
		RpcUser: conf.Setting.Rpc.Admin,
		RpcPwd:  conf.Setting.Rpc.Password,
		Https:   conf.Setting.Rpc.Tls,
		TxChLen: 100,
	}
	synchronizer := sync.NewSynchronizer(opt)
	listenInterrupt()

	wg := sync2.WaitGroup{}
	wg.Add(1)
	go startSync(db, synchronizer, &wg)
	wg.Add(1)
	go startApi(db, synchronizer, &wg)

	wg.Wait()

	db.Close()
}

func dealCommand() {
	v := flag.Bool("v", false, "show bin info")
	c := flag.Bool("c", false, "clear data")
	flag.Parse()

	if *v {
		_, _ = fmt.Fprint(os.Stderr, version.StringifyMultiLine())
		os.Exit(1)
	}
	if *c {
		clearDB()
		os.Exit(1)
	}
}

func startSync(storage *db.UTXODB, synchronizer *sync.Synchronizer, wg *sync2.WaitGroup) {
	defer wg.Done()

	for _, addr := range conf.Setting.Sync.Address {
		storage.InsertAddress(addr)
	}

	start := conf.Setting.Sync.Start
	lastOrder := storage.LastBlockOrder()
	if lastOrder != 0 {
		start = lastOrder
	}


	txChan, err := synchronizer.Start(&sync.HistoryOrder{
		LastTxBlockOrder:       start,
		Confirmations:          conf.Setting.Sync.Confirmations,
	})
	if err != nil {
		log.Errorf("Failed to start sync block, %s", err.Error())
		return
	}

	go dealSpent(storage, synchronizer)

	var preOrder uint64
	go func() {
		for {
			select {
			case <-interrupt:
				synchronizer.Stop()
				log.Infof("Stop sync block")
				return
			default:

				txs := <-txChan
				for _, tx := range txs {
					storage.UpdateHeight(tx.BlockHeight)
					utxoFlag := false
					// save tx or uxto
					utxos := uxto.GetUxtos(&tx)
					for _, u := range utxos {
						if storage.AddressIsExist(u.Address) {
							utxoFlag = true
							dbUtxo := &db.UTXO{
								TxId:       u.TxId,
								Vout:       uint64(u.TxIndex),
								Address:    u.Address,
								Amount:     u.Amount,
								Coin:       u.Coin,
								Height:     u.Height,
								IsCoinBase: tx.IsCoinBase,
							}
							storage.UpdateAddressUTXO(u.Address, dbUtxo)
							storage.SaveUTXO(dbUtxo)
						}
					}
					if utxoFlag {
						spentTxs := uxto.GetSpentTxs(&tx)
						for _, spentTx := range spentTxs {
							u, err := storage.GetUTXO(spentTx.TxId, spentTx.Vout)
							if err != nil {
								continue
							}
							// 标记这些utxo已经被花费掉
							storage.UpdateAddressUTXO(u.Address, &db.UTXO{
								TxId:   u.TxId,
								Coin:   u.Coin,
								Vout:   u.Vout,
								Amount: u.Amount,
								Spent:  tx.Txid,
							})
						}
					}

					if preOrder != tx.BlockOrder {
						preOrder = tx.BlockOrder
						storage.UpdateLastOrder(preOrder)
						log.Infof("Sync tx block order %d", preOrder)
					}
				}
			}

		}
	}()
}

func dealSpent(storage *db.UTXODB, synchronizer *sync.Synchronizer) {
	t := time.NewTicker(time.Second * 3 * 60 * 60)
	defer t.Stop()

	for {
		select {
		case <-interrupt:
			log.Infof("Stop deal spent")
			return
		case <-t.C:
			spents := storage.GetSpents()
			for _, spent := range spents {
				_, err := synchronizer.GetTx(spent.SpentTxId)
				if err != nil && isNoTx(err)  {
					log.Debugf("could not found tx %s", spent.SpentTxId)
					for _, utxo := range spent.UTXOList {
						utxo.Spent = ""
						err = storage.UpdateAddressUTXOMandatory(utxo.Address, utxo)
						log.Debugf("update utxo %s %d unspent", utxo.TxId, utxo.Vout)
						if err == nil {
							storage.DeleteSpentUTXO(spent.SpentTxId)
						}
					}
				}
			}
		}
	}
}

func isNoTx(err error) bool {
	if strings.Contains(err.Error(), "No information available about transaction") {
		return true
	}
	return false
}

func startApi(db *db.UTXODB, synchronizer *sync.Synchronizer, wg *sync2.WaitGroup) {
	defer wg.Done()

	a, err := api.NewApi(conf.Setting.Api.Listen, db, synchronizer)
	if err != nil {
		log.Errorf("Failed to start up api, %s", err.Error())
		return
	}
	go a.Run()

	func() {
		<-interrupt
		a.Stop()
	}()
}

func openDB(path string) (*db.UTXODB, error) {
	db, err := db.NewUTXODB(path)
	if err != nil {
		return nil, err
	}
	return db, nil
}

func listenInterrupt() {
	interrupt = make(chan struct{}, 1)
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, os.Kill)

	go func() {
		s := <-c
		close(interrupt)
		fmt.Println("exit", s)
	}()
}

func clearDB() {
	fmt.Println("Are you sure you want to clear all data?(y/n)")
	readBytes := [10]byte{}
	_, err := os.Stdin.Read(readBytes[:])
	if err != nil {
		fmt.Println("Failed to read input, ", err.Error())
		os.Exit(1)
	}
	rs := string(readBytes[:1])
	switch rs {
	case "y":
		fallthrough
	case "Y":
		fmt.Println("Start to clear db data...")
		storage := &db.UTXODB{}
		if err := storage.Clear(); err != nil {
			fmt.Printf("Clear db failed! %s\n", err)
		} else {
			fmt.Println("Clear db success!")
		}
	}
}
