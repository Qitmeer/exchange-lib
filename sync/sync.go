package sync

import (
	"fmt"
	"github.com/Qitmeer/exchange-lib/rpc"
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
func (s *Synchronizer) Start(order *HistoryOrder) (<-chan []rpc.Transaction, error) {
	if err := s.setThreshold(); err != nil {
		return nil, fmt.Errorf("failed to set threshold %s", err.Error())
	}

	go s.startSync(order)

	return s.txChannel, nil
}

// use the return value as the parameter for the next startup
func (s *Synchronizer) Stop() {
	s.stopSyncTxCh <- true
	s.stopSyncCoinBaseCh <- true
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
	go s.SyncCoinBaseTx()
}

func (s *Synchronizer) SyncTxs() {
	s.requestTxs()
}

func (s *Synchronizer) SyncCoinBaseTx() {
	for {
		select {
		case _ = <-s.stopSyncCoinBaseCh:
			return
		default:
			block, err := s.rpcClient.GetBlockByOrder(s.curCoinBaseBlockOrder)
			if err != nil {
				time.Sleep(time.Second * 5)
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
			return
		default:
			block, err := s.rpcClient.GetBlockByOrder(s.curTxBlockOrder)
			if err != nil {
				time.Sleep(time.Second * 5)
				break
			}
			if block.Txsvalid {
				if s.isTxConfirmed(block) {
					txs := getConfirmedTx(block)
					if len(txs) != 0 {
						s.txChannel <- txs
					}
					s.curTxBlockOrder++
				} else {
					time.Sleep(time.Second * 1)
				}
			} else {
				if s.isTxConfirmed(block) {
					s.curTxBlockOrder++
				} else {
					time.Sleep(time.Second * 1)
				}
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

type threshold struct {
	coinBaseThreshold    uint32
	transactionThreshold uint32
}

func (s *Synchronizer) setThreshold() error {
	nodeInfo, err := s.rpcClient.GetNodeInfo()
	if err != nil {
		return err
	}
	s.threshold.coinBaseThreshold = nodeInfo.Coinbasematurity
	s.threshold.transactionThreshold = nodeInfo.Confirmations
	return nil
}

func getConfirmedTx(block *rpc.Block) []rpc.Transaction {
	txs := []rpc.Transaction{}
	for _, tx := range block.Transactions {
		if isCoinBase(&tx) || tx.Duplicate {
			continue
		} else {
			txs = append(txs, tx)
		}
	}
	return []rpc.Transaction{}
}

func getConfirmedCoinBase(block *rpc.Block) []rpc.Transaction {
	for _, tx := range block.Transactions {
		if isCoinBase(&tx) {
			if tx.Duplicate {
				return []rpc.Transaction{}
			} else {
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
