package watch

import (
	"fmt"
	"github.com/pmker/onegw/oneplus/backend/sdk"
	"github.com/pmker/onegw/oneplus/watch/plugin"
	"github.com/pmker/onegw/oneplus/watch/structs"
	"github.com/sirupsen/logrus"
)

//
//type AbstractWatcher1 struct {
//	Rpc rpc.IBlockChainRPC
//
//	//Ctx  context.Context
//	lock sync.RWMutex
//
//	NewBlockChan        chan *structs.RemovableBlock
//	NewTxAndReceiptChan chan *structs.RemovableTxAndReceipt
//	NewReceiptLogChan   chan *structs.RemovableReceiptLog
//
//	SyncedBlocks         *list.List
//	SyncedTxAndReceipts  *list.List
//	MaxSyncedBlockToKeep int
//
//	BlockPlugins      []plugin.IBlockPlugin
//	TxPlugins         []plugin.ITxPlugin
//	TxReceiptPlugins  []plugin.ITxReceiptPlugin
//	ReceiptLogPlugins []plugin.IReceiptLogPlugin
//
//	ReceiptCatchUpFromBlock uint64
//}
//
//func NewHttpBasedEthWatcher( api string) *Watcher {
//	rpc := rpc.NewEthRPCWithRetry(api, 5)
//
//	return &Watcher{
//		//Ctx:                  ctx,
//		Rpc:                  rpc,
//		NewBlockChan:         make(chan *structs.RemovableBlock, 32),
//		NewTxAndReceiptChan:  make(chan *structs.RemovableTxAndReceipt, 518),
//		NewReceiptLogChan:    make(chan *structs.RemovableReceiptLog, 518),
//		SyncedBlocks:         list.New(),
//		SyncedTxAndReceipts:  list.New(),
//		MaxSyncedBlockToKeep: 64,
//	}
//}

func (watcher *Watcher) RegisterBlockPlugin(plugin plugin.IBlockPlugin) {
	watcher.BlockPlugins = append(watcher.BlockPlugins, plugin)
}

func (watcher *Watcher) RegisterTxPlugin(plugin plugin.ITxPlugin) {
	watcher.TxPlugins = append(watcher.TxPlugins, plugin)
}

func (watcher *Watcher) RegisterTxReceiptPlugin(plugin plugin.ITxReceiptPlugin) {
	watcher.TxReceiptPlugins = append(watcher.TxReceiptPlugins, plugin)
}

func (watcher *Watcher) RegisterReceiptLogPlugin(plugin plugin.IReceiptLogPlugin) {
	watcher.ReceiptLogPlugins = append(watcher.ReceiptLogPlugins, plugin)
}


func (watcher *Watcher) RegisterEventPlugin(plugin plugin.IEventPlugin) {
	watcher.EventPlugins = append(watcher.EventPlugins, plugin)
}
// start sync from latest block
//func (watcher *Watcher) RunTillExit() error {
//	return watcher.RunTillExitFromBlock(0)
//}

// start sync from given block
// 0 means start from latest block
func (watcher *Watcher) Run() error {
	wg := &watcher.PluginWG //.WaitGroup{}

	wg.Add(1)
	go func() {
		for block := range watcher.NewBlockChan {
			// run thru block plugins
			for i := 0; i < len(watcher.BlockPlugins); i++ {
				blockPlugin := watcher.BlockPlugins[i]

				blockPlugin.AcceptBlock(block)
			}

			// run thru tx plugins
			txPlugins := watcher.TxPlugins
			for i := 0; i < len(txPlugins); i++ {
				txPlugin := txPlugins[i]

				for j := 0; j < len(block.GetTransactions()); j++ {
					tx := structs.NewRemovableTx(block.GetTransactions()[j], false)
					txPlugin.AcceptTx(tx)
				}
			}


		}

		wg.Done()
	}()

	wg.Add(1)
	go func() {
		for removableTxAndReceipt := range watcher.NewTxAndReceiptChan {

			txReceiptPlugins := watcher.TxReceiptPlugins
			for i := 0; i < len(txReceiptPlugins); i++ {
				txReceiptPlugin := txReceiptPlugins[i]

				if p, ok := txReceiptPlugin.(*plugin.TxReceiptPluginWithFilter); ok {
					// for filter plugin, only feed receipt it wants
					if p.NeedReceipt(removableTxAndReceipt.Tx) {
						txReceiptPlugin.Accept(removableTxAndReceipt)
					}
				} else {
					txReceiptPlugin.Accept(removableTxAndReceipt)
				}
			}
		}

		wg.Done()
	}()

	wg.Add(1)
	go func() {
		for removableReceiptLog := range watcher.NewReceiptLogChan {
			logrus.Debugf("get receipt log from chan: %+v, txHash: %s", removableReceiptLog, removableReceiptLog.IReceiptLog.GetTransactionHash())

			receiptLogsPlugins := watcher.ReceiptLogPlugins
			for i := 0; i < len(receiptLogsPlugins); i++ {
				p := receiptLogsPlugins[i]

				if p.NeedReceiptLog(removableReceiptLog) {
					logrus.Debugln("receipt log accepted")
					p.Accept(removableReceiptLog)
				} else {
					logrus.Debugln("receipt log not accepted")
				}
			}
		}

		wg.Done()
	}()
	//
	//wg.Add(1)
	//go func() {
	//	for event := range watcher.NewEventLog {
	//		 logrus.Info("get receipt log from chan: %+v", event.BlockHeader)
	//
	//		eventPlugins := watcher.EventPlugins
	//		for i := 0; i < len(eventPlugins); i++ {
	//			p := eventPlugins[i]
	//			p.AcceptEvent(event)
	//			logrus.Debugln("receipt log accepted")
	//			//
	//			//if p.NeedReceiptLog(removableReceiptLog) {
	//			//	logrus.Debugln("receipt log accepted")
	//			//	p.Accept(removableReceiptLog)
	//			//} else {
	//			//	logrus.Debugln("receipt log not accepted")
	//			//}
	//		}
	//	}
	//
	//	wg.Done()
	//}()

	return nil
}

func (watcher *Watcher) LatestSyncedBlockNum() uint64 {
	watcher.lock.RLock()
	defer watcher.lock.RUnlock()

	if watcher.SyncedBlocks.Len() <= 0 {
		return 0
	}

	b := watcher.SyncedBlocks.Back().Value.(sdk.Block)

	return b.Number()
}

// go thru plugins to check if this watcher need fetch receipt for tx
// network load for fetching receipts per tx is heavy,
// we use this method to make sure we only do the work we need
func (watcher *Watcher) needReceipt(tx sdk.Transaction) bool {
	plugins := watcher.TxReceiptPlugins

	for _, p := range plugins {
		if filterPlugin, ok := p.(plugin.TxReceiptPluginWithFilter); ok {
			if filterPlugin.NeedReceipt(tx) {
				return true
			}
		} else {
			// exist global tx-receipt listener
			return true
		}
	}

	return false
}

// return query map: contractAddress -> interested 1stTopics
func (watcher *Watcher) getReceiptLogQueryMap() (queryMap map[string][]string) {
	queryMap = make(map[string][]string, 16)

	for _, p := range watcher.ReceiptLogPlugins {
		key := p.FromContract()

		if v, exist := queryMap[key]; exist {
			queryMap[key] = append(v, p.InterestedTopics()...)
		} else {
			queryMap[key] = p.InterestedTopics()
		}
	}

	return
}

func (watcher *Watcher) AddNewBlock(block *structs.RemovableBlock, curHighestBlockNum uint64, event *structs.Event) error {
	watcher.lock.Lock()
	defer watcher.lock.Unlock()

	// get tx receipts in block, which is time consuming
	signals := make([]*SyncSignal, 0, len(block.GetTransactions()))
	for i := 0; i < len(block.GetTransactions()); i++ {
		tx := block.GetTransactions()[i]

		if !watcher.needReceipt(tx) {
			logrus.Debugf("ignored tx: %s", tx.GetHash())
			continue
		} else {
			logrus.Debugf("needReceipt of tx: %s", tx.GetHash())
		}

		syncSigName := fmt.Sprintf("B:%d T:%s", block.Number(), tx.GetHash())

		sig := newSyncSignal(syncSigName)
		signals = append(signals, sig)

		go func() {
			txReceipt, err := watcher.Rpc.GetTransactionReceipt(tx.GetHash())

			if err != nil {
				fmt.Printf("GetTransactionReceipt fail, err: %s", err)
				sig.err = err

				// one fails all
				return
			}

			sig.WaitPermission()

			sig.rst = structs.NewRemovableTxAndReceipt(tx, txReceipt, false, block.Timestamp())

			sig.Done()
		}()
	}

	for i := 0; i < len(signals); i++ {
		sig := signals[i]
		sig.Permit()
		sig.WaitDone()

		if sig.err != nil {
			return sig.err
		}
	}

	for i := 0; i < len(signals); i++ {
		watcher.SyncedTxAndReceipts.PushBack(signals[i].rst.TxAndReceipt)
		watcher.NewTxAndReceiptChan <- signals[i].rst
	}

	queryMap := watcher.getReceiptLogQueryMap()
	logrus.Debugln("getReceiptLogQueryMap:", queryMap)

	bigStep := uint64(50)
	if curHighestBlockNum-block.Number() > bigStep {
		// only do request with bigStep
		if watcher.ReceiptCatchUpFromBlock == 0 {
			// init
			logrus.Debugf("bigStep, init to %d", block.Number())
			watcher.ReceiptCatchUpFromBlock = block.Number()
		} else {
			// check if we need do requests
			if (block.Number() - watcher.ReceiptCatchUpFromBlock + 1) == bigStep {
				fromBlock := watcher.ReceiptCatchUpFromBlock
				toBlock := block.Number()

				logrus.Debugf("bigStep, doing request, range: %d -> %d (minus: %d)", fromBlock, toBlock, block.Number()-watcher.ReceiptCatchUpFromBlock)

				for k, v := range queryMap {
					err := watcher.fetchReceiptLogs(false, block, fromBlock, toBlock, k, v)
					if err != nil {
						return err
					}
				}

				// update catch up block
				watcher.ReceiptCatchUpFromBlock = block.Number() + 1
			} else {
				logrus.Debugf("bigStep, holding %d blocks: %d -> %d", block.Number()-watcher.ReceiptCatchUpFromBlock+1, watcher.ReceiptCatchUpFromBlock, block.Number())
			}
		}
	} else {
		// reset
		if watcher.ReceiptCatchUpFromBlock != 0 {
			logrus.Debugf("exit bigStep mode, ReceiptCatchUpFromBlock: %d, curBlock: %d, gap: %d", watcher.ReceiptCatchUpFromBlock, block.Number(), curHighestBlockNum-block.Number())
			watcher.ReceiptCatchUpFromBlock = 0
		}

		for k, v := range queryMap {
			err := watcher.fetchReceiptLogs(block.IsRemoved, block, block.Number(), block.Number(), k, v)
			if err != nil {
				return err
			}
		}
	}

	// clean synced data
	for watcher.SyncedBlocks.Len() >= watcher.MaxSyncedBlockToKeep {
		// clean block
		b := watcher.SyncedBlocks.Remove(watcher.SyncedBlocks.Front()).(sdk.Block)

		// clean txAndReceipt
		for watcher.SyncedTxAndReceipts.Front() != nil {
			head := watcher.SyncedTxAndReceipts.Front()

			if head.Value.(*structs.TxAndReceipt).Tx.GetBlockNumber() <= b.Number() {
				watcher.SyncedTxAndReceipts.Remove(head)
			} else {
				break
			}
		}
	}

	// block
	watcher.SyncedBlocks.PushBack(block.Block)
	watcher.NewBlockChan <- block

	fmt.Printf("%v",event)
	 watcher.NewEventLog <- event
	return nil
}

func (watcher *Watcher) fetchReceiptLogs(isRemoved bool, block sdk.Block, from, to uint64, address string, topics []string) error {

	receiptLogs, err := watcher.Rpc.GetLogs(from, to, address, topics)
	if err != nil {
		return err
	}

	for i := 0; i < len(receiptLogs); i++ {
		log := receiptLogs[i]
		logrus.Debugln("insert into chan: ", log.GetTransactionHash())

		watcher.NewReceiptLogChan <- &structs.RemovableReceiptLog{
			IReceiptLog: log,
			IsRemoved:   isRemoved,
		}
	}

	return nil
}

type SyncSignal struct {
	name       string
	permission chan bool
	jobDone    chan bool
	rst        *structs.RemovableTxAndReceipt
	err        error
}

func newSyncSignal(name string) *SyncSignal {
	return &SyncSignal{
		name:       name,
		permission: make(chan bool, 1),
		jobDone:    make(chan bool, 1),
	}
}

func (s *SyncSignal) Permit() {
	s.permission <- true
}

func (s *SyncSignal) WaitPermission() {
	<-s.permission
}

func (s *SyncSignal) Done() {
	s.jobDone <- true
}

func (s *SyncSignal) WaitDone() {
	<-s.jobDone
}

func (watcher *Watcher) PopBlocksUntilReachMainChain() error {
	watcher.lock.Lock()
	defer watcher.lock.Unlock()

	for {
		if watcher.SyncedBlocks.Back() == nil {
			return nil
		}

		// NOTE: instead of watcher.LatestSyncedBlockNum() cuz it has lock
		lastSyncedBlock := watcher.SyncedBlocks.Back().Value.(sdk.Block)
		block, err := watcher.Rpc.GetBlockByNum(lastSyncedBlock.Number())
		if err != nil {
			return err
		}

		if block.Hash() != lastSyncedBlock.Hash() {
			fmt.Println("removing tail block:", watcher.SyncedBlocks.Back())
			removedBlock := watcher.SyncedBlocks.Remove(watcher.SyncedBlocks.Back()).(sdk.Block)

			for watcher.SyncedTxAndReceipts.Back() != nil {

				tail := watcher.SyncedTxAndReceipts.Back()

				if tail.Value.(*structs.TxAndReceipt).Tx.GetBlockNumber() >= removedBlock.Number() {
					fmt.Printf("removing tail txAndReceipt: %+v", tail.Value)
					tuple := watcher.SyncedTxAndReceipts.Remove(tail).(*structs.TxAndReceipt)

					watcher.NewTxAndReceiptChan <- structs.NewRemovableTxAndReceipt(tuple.Tx, tuple.Receipt, true, block.Timestamp())
				} else {
					fmt.Printf("all txAndReceipts removed for block: %+v", removedBlock)
					break
				}
			}

			watcher.NewBlockChan <- structs.NewRemovableBlock(removedBlock, true)
		} else {
			return nil
		}
	}
}

func (watcher *Watcher) FoundFork(newBlock sdk.Block) bool {
	for e := watcher.SyncedBlocks.Back(); e != nil; e = e.Prev() {
		syncedBlock := e.Value.(sdk.Block)

		//if syncedBlock == nil {
		//	logrus.Warnln("error, syncedBlock is nil")
		//}
		//logrus.Debugf("syncedBlock: %+v", syncedBlock)

		//if newBlock == nil {
		//	logrus.Warnln("error, newBlock is nil")
		//}
		//logrus.Debugf("newBlock: %+v", newBlock)

		if syncedBlock.Number()+1 == newBlock.Number() {
			notMatch := (syncedBlock).Hash() != newBlock.ParentHash()

			if notMatch {
				fmt.Printf("found fork, new block(%d): %s, new block's parent: %s, parent we synced: %s",
					newBlock.Number(), newBlock.Hash(), newBlock.ParentHash(), syncedBlock.Hash())

				return true
			}
		}
	}

	return false
}
