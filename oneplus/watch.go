package main

import (
	"fmt"
	"github.com/pmker/onegw/oneplus/watch"
	"github.com/pmker/onegw/oneplus/watch/plugin"
	"github.com/sirupsen/logrus"
)

//
//func block() {
//	api := "https://mainnet.infura.io/v3/19d753b2600445e292d54b1ef58d4df4"
//	w := blockwatch.NewHttpBasedEthWatcher(api)
//
//	w.RegisterBlockPlugin(plugin.NewBlockNumPlugin(func(i uint64, b bool) {
//		fmt.Println(">>", i, b)
//	}))
//
//}
//func receipt() {
//	logrus.SetLevel(logrus.DebugLevel)
//
//	ctx := context.Background()
//	api := "https://kovan.infura.io/v3/19d753b2600445e292d54b1ef58d4df4"
//	startBlock := 12220000
//	contract := "0xAc34923B2b8De9d441570376e6c811D5aA5ed72f"
//	interestedTopics := []string{
//		"0x23b872dd7302113369cda2901243429419bec145408fa8b352b3dd92b66c680b",
//		"0x6bf96fcc2cec9e08b082506ebbc10114578a497ff1ea436628ba8996b750677c",
//		"0x5a746ce5ce37fc996a6e682f4f84b6f90d3be79fd8ac9a8a11264345f3d29edd",
//		"0x9c4e90320be51bb93d854d0ab9ba8aa249dabc21192529efcd76ae7c22c6bc0b",
//		"0x0ce31a5f70780bb6770b52a6793403d856441ccb475715e8382a0525d35b0558",
//	}
//
//	handler := func(log structs.RemovableReceiptLog) {
//		logrus.Infof("log from tx: %s", log.GetTransactionHash())
//	}
//
//	stepsForBigLag := 100
//
//	highestProcessed := watch.ListenForReceiptLogTillExit(ctx, api, startBlock, contract, interestedTopics, handler, stepsForBigLag)
//	logrus.Infof("highestProcessed: %d", highestProcessed)
//}
//
//const (
//	blockWatcherRetentionLimit = 20
//	ethereumRPCRequestTimeout  = 30 * time.Second
//	ethWatcherPollingInterval  = 1 * time.Minute
//	peerConnectTimeout         = 60 * time.Second
//	checkNewAddrInterval       = 20 * time.Second
//	expirationPollingInterval  = 50 * time.Millisecond
//	BlockPollingInterval       = 5 * time.Second
//	EthereumRPCURL             = "https://rinkeby.infura.io/v3/19d753b2600445e292d54b1ef58d4df4"
//)
//
//type Watch struct {
//	blockWatcher      *blockwatch.Watcher
//	blockSubscription event.Subscription
//}
//
//func BlockWatch() (*Watch, error) {
//
//	databasePath := filepath.Join("0vedb", "db")
//	meshDB, err := meshdb.NewMeshDB(databasePath)
//	if err != nil {
//		return nil, err
//	}
//
//	// Initialize gateway watcher (but don't start it yet).
//	blockWatcherClient, err := blockwatch.NewRpcClient(EthereumRPCURL, ethereumRPCRequestTimeout)
//	if err != nil {
//		return nil, err
//	}
//
//	watch := blockwatch.NewHttpBasedEthWatcher(EthereumRPCURL)
//	//topics := orderwatch.GetRelevantTopics()
//	blockWatcherConfig := blockwatch.Config{
//		MeshDB:              meshDB,
//		PollingInterval:     BlockPollingInterval,
//		StartBlockDepth:     ethrpc.LatestBlockNumber,
//		BlockRetentionLimit: blockWatcherRetentionLimit,
//		WithLogs:            true,
//		//Topics:              topics,
//		Client: blockWatcherClient,
//		Watch:  watch,
//	}
//	blockWatcher := blockwatch.New(blockWatcherConfig)
//	go func() {
//		for {
//			err, isOpen := <-blockWatcher.Errors
//			if isOpen {
//				logrus.WithField("error", err).Error("BlockWatcher error encountered")
//			} else {
//				return // Exit when the error channel is closed
//			}
//		}
//	}()
//
//	watchblock := &Watch{
//		blockWatcher: blockWatcher,
//	}
//	return watchblock, nil
//}
//
//func (w *Watch) setupEventWatcher() {
//	blockEvents := make(chan []*blockwatch.Event, 1)
//	w.blockSubscription = w.blockWatcher.Subscribe(blockEvents)
//
//	go func() {
//		for {
//			select {
//			case err, isOpen := <-w.blockSubscription.Err():
//				close(blockEvents)
//				if !isOpen {
//					// event.Subscription closes the Error channel on unsubscribe.
//					// We therefore cleanup this goroutine on channel closure.
//					return
//				}
//				logrus.WithFields(logrus.Fields{
//					"error": err.Error(),
//				}).Error("subscription error encountered")
//				return
//
//			case events := <-blockEvents:
//				latestBlockNum, err := w.blockWatcher.Watch.Rpc.GetCurrentBlockNum()
//				if err!=nil{}
//
//				//hashToOrderWithTxHashes := map[tool.Hash]*OrderWithTxHashes{}
//				for _, event := range events {
//
//
//					newBlock, err := w.blockWatcher.Watch.Rpc.GetBlockByNum(uint64(event.BlockHeader.Number.Int64()))
//					if err != nil {
//						//return err
//					}
//
//					if newBlock == nil {
//						//msg := fmt.Sprintf("GetBlockByNum(%d) returns nil block", newBlockNumToSync)
//						//return errors.New(msg)
//					}
//
//					if w.blockWatcher.Watch.FoundFork(newBlock) {
//						logrus.Infoln("found fork, popping")
//						err = w.blockWatcher.Watch.PopBlocksUntilReachMainChain()
//					} else {
//						logrus.Infoln("adding new block:", newBlock.Number())
//						err = w.blockWatcher.Watch.AddNewBlock(structs.NewRemovableBlock(newBlock, false), latestBlockNum)
//					}
//					fmt.Printf("current block number,%s \n", event.BlockHeader.Number)
//				}
//
//				//fmt.Printf("orderr hash %s\n", hashToOrderWithTxHashes)
//				//w.generateOrderEventsIfChanged(hashToOrderWithTxHashes)
//			}
//		}
//	}()
//}
//
//func main1() {
//	//watch, err1 := BlockWatch()
//	//	//
//	//	//
//	//	//watch.blockWatcher.Watch.RegisterBlockPlugin(plugin.NewBlockNumPlugin(func(i uint64, b bool) {
//	//	//	fmt.Println(">>", i, b)
//	//	//}))
//	//	//if err1 != nil {
//	//	//	fmt.Printf(err1.Error())
//	//	//}
//	//	//if err := watch.blockWatcher.Start(); err != nil {
//	//	//	logrus.WithField("error", err).Error("BlockWatcher start")
//	//	//
//	//	//}
//	//	//go watch.blockWatcher.Watch.Start()
//	//	//
//	//	//go watch.setupEventWatcher()
//	//	//select {}
//
//
//
//}

func main() {
	Watcher, err1 := watch.NewWatch()

	Watcher.RegisterBlockPlugin(plugin.NewBlockNumPlugin(func(i uint64, b bool) {
		fmt.Println(">>", i, b)
	}))
	//
	//Watcher.RegisterTxReceiptPlugin(plugin.NewTxReceiptPlugin(func(txAndReceipt *structs.RemovableTxAndReceipt) {
	//	if txAndReceipt.IsRemoved {
	//		fmt.Println("Removed >>", txAndReceipt.Tx.GetHash(), txAndReceipt.Receipt.GetTxIndex())
	//	} else {
	//		fmt.Println("Adding >>", txAndReceipt.Tx.GetHash(), txAndReceipt.Receipt.GetTxIndex())
	//	}
	//}))
	//

	//callback:=plugin.NewEventHashPlugin(func(event *structs.Event) {
	//	fmt.Printf("%v",event)
	//})
	//
	//Watcher.RegisterEventPlugin(callback)
	if err1 != nil {
		fmt.Printf(err1.Error())
	}
	if err := Watcher.Listen(); err != nil {
		logrus.WithField("error", err).Error("BlockWatcher start")

	}

	select {}
}
