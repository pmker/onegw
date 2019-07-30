package watch

import (
	"fmt"
	ethrpc "github.com/ethereum/go-ethereum/rpc"
	"github.com/pmker/onegw/oneplus/meshdb"

	"github.com/pmker/onegw/oneplus/watch/structs"
	"github.com/sirupsen/logrus"
	"path/filepath"
	"time"
)

const (
	blockWatcherRetentionLimit = 20
	ethereumRPCRequestTimeout  = 30 * time.Second
	ethWatcherPollingInterval  = 1 * time.Minute
	peerConnectTimeout         = 60 * time.Second
	checkNewAddrInterval       = 20 * time.Second
	expirationPollingInterval  = 50 * time.Millisecond
	BlockPollingInterval       = 5 * time.Second
	EthereumRPCURL             = "https://rinkeby.infura.io/v3/19d753b2600445e292d54b1ef58d4df4"
)

//type BlockerWatch struct {
//	Watcher           *Watcher
//	blockSubscription event.Subscription
//}

func NewWatch() (*Watcher, error) {

	databasePath := filepath.Join("0vedb", "db")
	meshDB, err := meshdb.NewMeshDB(databasePath)
	if err != nil {
		return nil, err
	}

	// Initialize gateway watcher (but don't start it yet).
	blockWatcherClient, err := NewRpcClient(EthereumRPCURL, ethereumRPCRequestTimeout)
	if err != nil {
		return nil, err
	}

	//topics := orderwatch.GetRelevantTopics()
	blockWatcherConfig := Config{
		MeshDB:              meshDB,
		PollingInterval:     BlockPollingInterval,
		StartBlockDepth:     ethrpc.LatestBlockNumber,
		BlockRetentionLimit: blockWatcherRetentionLimit,
		WithLogs:            true,
		//Topics:              topics,
		Client: blockWatcherClient,
		RpcURL: EthereumRPCURL,
	}
	blockWatcher := New(blockWatcherConfig)
	go func() {
		for {
			err, isOpen := <-blockWatcher.Errors
			if isOpen {
				logrus.WithField("error", err).Error("BlockWatcher error encountered")
			} else {
				return // Exit when the error channel is closed
			}
		}
	}()


	return blockWatcher, nil
}

func (w *Watcher) Listen() error {

	w.Start()
	blockEvents := make(chan []*structs.Event, 1)
	w.blockSubscription = w.Subscribe(blockEvents)

	go func() {
		for {
			select {
			case err, isOpen := <-w.blockSubscription.Err():
				close(blockEvents)
				if !isOpen {
					// event.Subscription closes the Error channel on unsubscribe.
					// We therefore cleanup this goroutine on channel closure.
					return
				}
				logrus.WithFields(logrus.Fields{
					"error": err.Error(),
				}).Error("subscription error encountered")
				return

			case events := <-blockEvents:
				latestBlockNum, err := w.Rpc.GetCurrentBlockNum()
				if err != nil {
				}

				//hashToOrderWithTxHashes := map[tool.Hash]*OrderWithTxHashes{}
				for _, event := range events {

					newBlock, err := w.Rpc.GetBlockByNum(uint64(event.BlockHeader.Number.Int64()))
					if err != nil {
						 //return err
					}

					if newBlock == nil {
						// msg := fmt.Sprintf("GetBlockByNum(%d) returns nil block", newBlock)
						 //return errors.New(msg)
					}

					if w.FoundFork(newBlock) {
						logrus.Infoln("found fork, popping")
						err = w.PopBlocksUntilReachMainChain()
					} else {
						logrus.Infoln("adding new block:", newBlock.Number())
						err = w.AddNewBlock(structs.NewRemovableBlock(newBlock, false), latestBlockNum,event)
					}
					fmt.Printf("current block number,%s \n", event.BlockHeader.Number)
				}

				//fmt.Printf("orderr hash %s\n", hashToOrderWithTxHashes)
				//w.generateOrderEventsIfChanged(hashToOrderWithTxHashes)
			}
		}
	}()
	return nil
}
