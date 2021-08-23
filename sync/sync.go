package sync

import (
	"fmt"
	"github.com/Qitmeer/exchange-lib/rpc"
	"github.com/bCoder778/log"
	"time"
)

const (
	defaultHost                 = "127.0.0.1:1234"
	defaultTxChLen              = 100
	defaultRepeatCount          = 5
	defaultCoinBaseThreshold    = 720
	defaultTransactionThreshold = 10
)

type Synchronizer struct {
	rpcClient             *rpc.Client
	opt                   *Options
	threshold             *threshold
	txChannel             chan []rpc.Transaction
	stopSyncTxCh          chan bool
	stopSyncCoinBaseCh    chan bool
	curTxBlockOrder       uint64
	curCoinBaseBlockOrder uint64
}

type Options struct {
	// Rpc option
	RpcAddr string
	RpcUser string
	RpcPwd  string
	Https   bool
	// tx channel length
	TxChLen uint
}

type HistoryOrder struct {
	LastTxBlockOrder       uint64
	LastCoinBaseBlockOrder uint64
	Confirmations          uint64
}

func NewSynchronizer(opt *Options) *Synchronizer {
	if opt.RpcAddr == "" {
		opt.RpcAddr = defaultHost
	}
	if opt.TxChLen == 0 {
		opt.TxChLen = defaultTxChLen
	}

	client := rpc.NewClient(&rpc.RpcConfig{
		Address: opt.RpcAddr,
		User:    opt.RpcUser,
		Pwd:     opt.RpcPwd,
		Https:   opt.Https,
	})
	return &Synchronizer{
		rpcClient:          client,
		opt:                opt,
		txChannel:          make(chan []rpc.Transaction, opt.TxChLen),
		stopSyncTxCh:       make(chan bool),
		stopSyncCoinBaseCh: make(chan bool),
		threshold: &threshold{
			coinBaseThreshold:    defaultCoinBaseThreshold,
			transactionThreshold: defaultTransactionThreshold,
		},
	}
}

// start syncing at 0
// or start syncing at last stop return id
func (s *Synchronizer) Start(info *HistoryOrder) (<-chan []rpc.Transaction, error) {
	if err := s.setThreshold(info.Confirmations); err != nil {
		return nil, fmt.Errorf("failed to set threshold %s", err.Error())
	}

	go s.startSync(info)

	return s.txChannel, nil
}

// use the return value as the parameter for the next startup
func (s *Synchronizer) Stop() {
	s.stopSyncTxCh <- true
	//s.stopSyncCoinBaseCh <- true
}

func (s *Synchronizer) GetHistoryOrder() *HistoryOrder {
	return &HistoryOrder{
		LastTxBlockOrder:       s.curTxBlockOrder,
		LastCoinBaseBlockOrder: s.curCoinBaseBlockOrder,
	}
}

func (s *Synchronizer) startSync(hisOrder *HistoryOrder) {
	s.curTxBlockOrder = hisOrder.LastTxBlockOrder
	if s.curTxBlockOrder >= defaultRepeatCount {
		s.curTxBlockOrder -= defaultRepeatCount
	} else {
		s.curTxBlockOrder = 0
	}

	s.curCoinBaseBlockOrder = hisOrder.LastCoinBaseBlockOrder
	if s.curCoinBaseBlockOrder >= defaultRepeatCount {
		s.curCoinBaseBlockOrder -= defaultRepeatCount
	} else {
		s.curCoinBaseBlockOrder = 0
	}

	go s.SyncTxs()
	//go s.SyncCoinBaseTx()
}

func (s *Synchronizer) SyncTxs() {
	s.requestTxs()
}

func (s *Synchronizer) SyncCoinBaseTx() {
	for {
		select {
		case _ = <-s.stopSyncCoinBaseCh:
			log.Infof("stop sync coinbase tx")
			return
		default:
			block, err := s.rpcClient.GetBlockByOrder(s.curCoinBaseBlockOrder)
			if err != nil {
				time.Sleep(time.Second * 30)
				break
			}
			if !s.isBlockConfirmed(block) {
				time.Sleep(time.Second * 1)
				break
			}
			if usable, err := s.IsCoinBaseUsable(block); err != nil {
				time.Sleep(time.Second * 5)
				break
			} else {
				if usable {
					txs := getConfirmedCoinBase(block)
					if len(txs) != 0 {
						s.txChannel <- txs
					}
				}
				s.curCoinBaseBlockOrder++
			}
		}
	}
}

func (s *Synchronizer) requestTxs() {
	for {
		select {
		case _ = <-s.stopSyncTxCh:
			log.Infof("stop sync tx")
			return
		default:
			block, err := s.rpcClient.GetBlockByOrder(s.curTxBlockOrder)
			if err != nil {
				time.Sleep(time.Second * 30)
				break
			}
			if s.isTxConfirmed(block) {
				if block.Txsvalid {
					txs := s.getConfirmedTx(block)
					if len(txs) != 0 {
						s.txChannel <- txs
					}
				}
				s.curTxBlockOrder++
			} else {
				time.Sleep(time.Second * 30)
			}
		}
	}
}

func (s *Synchronizer) isBlockConfirmed(block *rpc.Block) bool {
	return block.Confirmations > s.threshold.coinBaseThreshold
}

func (s *Synchronizer) isTxConfirmed(block *rpc.Block) bool {
	return block.Confirmations > s.threshold.transactionThreshold
}

func (s *Synchronizer) IsCoinBaseUsable(block *rpc.Block) (bool, error) {
	color, err := s.rpcClient.IsBlue(block.Hash)
	if err != nil {
		return false, err
	}
	switch color {
	case 0:
		return false, nil
	case 1:
		return true, nil
	}
	return false, nil
}

func (s *Synchronizer) SendTx(raw string) (string, error) {
	return s.rpcClient.SendTransaction(raw)
}

func (s *Synchronizer) GetTx(txId string) (*rpc.Transaction, error) {
	return s.rpcClient.GetTransaction(txId)
}

type threshold struct {
	coinBaseThreshold    uint32
	transactionThreshold uint32
}

func (s *Synchronizer) setThreshold(confirmations uint64) error {
	nodeInfo, err := s.rpcClient.GetNodeInfo()
	if err != nil {
		return err
	}
	if confirmations != 0 {
		s.threshold.transactionThreshold = uint32(confirmations)
	} else {
		s.threshold.transactionThreshold = nodeInfo.Confirmations
	}
	s.threshold.coinBaseThreshold = nodeInfo.Coinbasematurity

	return nil
}

func (s *Synchronizer) getConfirmedTx(block *rpc.Block) []rpc.Transaction {
	txs := []rpc.Transaction{}
	for _, tx := range block.Transactions {
		if tx.Duplicate {
			continue
		}
		if isCoinBase(&tx) {
			ok, _ := s.IsCoinBaseUsable(block)
			if !ok {
				continue
			}
		}
		tx.BlockOrder = block.Order
		txs = append(txs, tx)
	}
	return txs
}

func getConfirmedCoinBase(block *rpc.Block) []rpc.Transaction {
	for _, tx := range block.Transactions {
		if isCoinBase(&tx) {
			tx.IsCoinBase = true
			if tx.Duplicate {
				return []rpc.Transaction{}
			} else {
				tx.BlockOrder = block.Order
				return []rpc.Transaction{tx}
			}
		}
	}
	return []rpc.Transaction{}
}

func isCoinBase(tx *rpc.Transaction) bool {
	if tx != nil && len(tx.Vin) > 0 && tx.Vin[0].Coinbase != "" {
		return true
	}
	return false
}
