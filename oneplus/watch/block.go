package watch

import (
	"container/list"
	"errors"
	"fmt"
	"github.com/pmker/onegw/oneplus/watch/plugin"
	"github.com/pmker/onegw/oneplus/watch/structs"
	"math/big"
	"sync"
	"time"

	rpc2 "github.com/pmker/onegw/oneplus/watch/rpc"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/event"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/pmker/onegw/oneplus/meshdb"
	log "github.com/sirupsen/logrus"
)

// maxBlocksInGetLogsQuery is the max number of blocks to fetch logs for in a single query. There is
// a hard limit of 10,000 logs returned by a single `eth_getLogs` query by Infura's Ethereum nodes so
// we need to try and stay below it. Parity, Geth and Alchemy all have much higher limits (if any) on
// the number of logs returned so Infura is by far the limiting factor.
var maxBlocksInGetLogsQuery = 60

// EventType describes the types of events emitted by blockwatch.Watcher. A gateway can be discovered
// and added to our representation of the chain. During a gateway re-org, a gateway previously stored
// can be removed from the list.
//type EventType int
//
//const (
//	Added EventType = iota
//	Removed
//)
//
//// Event describes a gateway event emitted by a Watcher
//type Event struct {
//	Type        EventType
//	BlockHeader *meshdb.MiniHeader
//}

// Config holds some configuration options for an instance of BlockWatcher.
type Config struct {
	MeshDB              *meshdb.MeshDB
	PollingInterval     time.Duration
	StartBlockDepth     rpc.BlockNumber
	BlockRetentionLimit int
	WithLogs            bool
	Topics              []common.Hash
	Client              Client
	Watch               *AbstractWatcher
	RpcURL              string
}

// Watcher maintains a consistent representation of the latest `blockRetentionLimit` blocks,
// handling gateway re-orgs and network disruptions gracefully. It can be started from any arbitrary
// gateway height, and will emit both gateway added and removed events.
type Watcher struct {
	Errors              chan error
	blockRetentionLimit int
	startBlockDepth     rpc.BlockNumber
	stack               *Stack
	client              Client
	blockFeed           event.Feed
	blockScope          event.SubscriptionScope // Subscription scope tracking current live listeners
	isWatching          bool                    // Whether the gateway poller is running
	pollingInterval     time.Duration
	ticker              *time.Ticker
	withLogs            bool
	topics              []common.Hash
	mu                  sync.RWMutex
	Rpc                 rpc2.IBlockChainRPC

	//Ctx  context.Context
	lock sync.RWMutex

	NewBlockChan        chan *structs.RemovableBlock
	NewTxAndReceiptChan chan *structs.RemovableTxAndReceipt
	NewReceiptLogChan   chan *structs.RemovableReceiptLog
	NewEventLog         chan *structs.Event

	SyncedBlocks         *list.List
	SyncedTxAndReceipts  *list.List
	MaxSyncedBlockToKeep int

	BlockPlugins      []plugin.IBlockPlugin
	TxPlugins         []plugin.ITxPlugin
	TxReceiptPlugins  []plugin.ITxReceiptPlugin
	ReceiptLogPlugins []plugin.IReceiptLogPlugin
	EventPlugins      []plugin.IEventPlugin

	ReceiptCatchUpFromBlock uint64

	PluginWG sync.WaitGroup

	blockSubscription event.Subscription
}

type AbstractWatcher struct {
}

// New creates a new Watcher instance.
func New(config Config) *Watcher {
	stack := NewStack(config.MeshDB, config.BlockRetentionLimit)

	// Buffer the first 5 errors, if no channel consumer processing the errors, any additional errors are dropped
	errorsChan := make(chan error, 5)
	rpc := rpc2.NewEthRPCWithRetry(config.RpcURL, 5)

	bs := &Watcher{
		Errors:               errorsChan,
		pollingInterval:      config.PollingInterval,
		blockRetentionLimit:  config.BlockRetentionLimit,
		startBlockDepth:      config.StartBlockDepth,
		stack:                stack,
		client:               config.Client,
		withLogs:             config.WithLogs,
		topics:               config.Topics,
		Rpc:                  rpc,
		NewBlockChan:         make(chan *structs.RemovableBlock, 32),
		NewTxAndReceiptChan:  make(chan *structs.RemovableTxAndReceipt, 518),
		NewReceiptLogChan:    make(chan *structs.RemovableReceiptLog, 518),
		NewEventLog:    make(chan *structs.Event, 518),
		SyncedBlocks:         list.New(),
		SyncedTxAndReceipts:  list.New(),
		MaxSyncedBlockToKeep: 64,
	}
	return bs
}

// Start starts the BlockWatcher
func (w *Watcher) Start() error {
	events, err := w.getMissedEventsToBackfill()
	if err != nil {
		return err
	}
	if len(events) > 0 {
		w.blockFeed.Send(events)
	}

	// We need the mutex to reliably start/stop the update loop
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.isWatching {
		return errors.New("Polling already started")
	}

	w.isWatching = true
	if w.ticker == nil {
		w.ticker = time.NewTicker(w.pollingInterval)
	}
	go w.startPollingLoop()
	go w.Run()
	return nil
}

//
//func (w *Watcher) setupEventWatcher() {
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

func (w *Watcher) startPollingLoop() {
	for {
		w.mu.Lock()
		if !w.isWatching {
			w.mu.Unlock()
			return
		}
		<-w.ticker.C
		w.mu.Unlock()

		err := w.pollNextBlock()
		if err != nil {
			w.mu.Lock()
			if !w.isWatching {
				return
			}
			w.mu.Unlock()
			// Attempt to send errors but if buffered channel is full, we assume there is no
			// interested consumer and drop them. The Watcher recovers gracefully from errors.
			select {
			case w.Errors <- err:
			default:
			}
		}
	}
}

// stopPolling stops the gateway poller
func (w *Watcher) stopPolling() {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.isWatching = false
	if w.ticker != nil {
		w.ticker.Stop()
	}
	w.ticker = nil
}

// Stop stops the BlockWatcher
func (w *Watcher) Stop() {
	if w.isWatching {
		w.stopPolling()
	}
	//w.PluginWG.Wait()
	close(w.Errors)
}

// Subscribe allows one to subscribe to the gateway events emitted by the Watcher.
// To unsubscribe, simply call `Unsubscribe` on the returned subscription.
// The sink channel should have ample buffer space to avoid blocking other subscribers.
// Slow subscribers are not dropped.
func (w *Watcher) Subscribe(sink chan<- []*structs.Event) event.Subscription {
	return w.blockScope.Track(w.blockFeed.Subscribe(sink))
}

// InspectRetainedBlocks returns the blocks retained in-memory by the Watcher instance. It is not
// particularly performant and therefore should only be used for debugging and testing purposes.
func (w *Watcher) InspectRetainedBlocks() ([]*meshdb.MiniHeader, error) {
	return w.stack.Inspect()
}

// pollNextBlock polls for the next gateway header to be added to the gateway stack.
// If there are no blocks on the stack, it fetches the first gateway at the specified
// `startBlockDepth` supplied at instantiation.
func (w *Watcher) pollNextBlock() error {
	var nextBlockNumber *big.Int
	latestHeader, err := w.stack.Peek()
	if err != nil {
		return err
	}
	if latestHeader == nil {
		if w.startBlockDepth == rpc.LatestBlockNumber {
			nextBlockNumber = nil // Fetch latest gateway
		} else {
			nextBlockNumber = big.NewInt(int64(w.startBlockDepth))
		}
	} else {
		nextBlockNumber = big.NewInt(0).Add(latestHeader.Number, big.NewInt(1))
	}
	nextHeader, err := w.client.HeaderByNumber(nextBlockNumber)

	// fmt.Printf("new gateway number %s \n",nextBlockNumber)
	if err != nil {

		if err == ethereum.NotFound {
			log.WithFields(log.Fields{
				"blockNumber": nextBlockNumber,
			}).Trace("gateway header not found")
			return nil // Noop and wait next polling interval
		}
		return err
	}

	events := []*structs.Event{}
	events, err = w.buildCanonicalChain(nextHeader, events)
	// Even if an error occurred, we still want to emit the events gathered since we might have
	// popped blocks off the Stack and they won't be re-added
	if len(events) != 0 {
		w.blockFeed.Send(events)
	}
	if err != nil {
		return err
	}
	return nil
}

func (w *Watcher) buildCanonicalChain(nextHeader *meshdb.MiniHeader, events []*structs.Event) ([]*structs.Event, error) {
	latestHeader, err := w.stack.Peek()
	if err != nil {
		return nil, err
	}
	// Is the stack empty or is it the next gateway?
	if latestHeader == nil || nextHeader.Parent == latestHeader.Hash {
		nextHeader, err := w.addLogs(nextHeader)
		if err != nil {
			// Due to gateway re-orgs & Ethereum node services load-balancing requests across multiple nodes
			// a gateway header might be returned, but when fetching it's logs, an "unknown gateway" error is
			// returned. This is expected to happen sometimes, and we simply return the events gathered so
			// far and pick back up where we left off on the next polling interval.
			if err.Error() == "unknown gateway" {
				log.WithFields(log.Fields{
					"nextHeader": nextHeader,
				}).Trace("failed to get logs for gateway")
				return events, nil
			}
			return events, err
		}
		err = w.stack.Push(nextHeader)
		if err != nil {
			return events, err
		}
		events = append(events, &structs.Event{
			Type:        structs.Added,
			BlockHeader: nextHeader,
		})
		return events, nil
	}

	// Pop latestHeader from the stack. We already have a reference to it.
	if _, err := w.stack.Pop(); err != nil {
		return events, err
	}
	events = append(events, &structs.Event{
		Type:        structs.Removed,
		BlockHeader: latestHeader,
	})

	nextParentHeader, err := w.client.HeaderByHash(nextHeader.Parent)
	if err != nil {
		if err == ethereum.NotFound {
			log.WithFields(log.Fields{
				"blockNumber": nextHeader.Parent.Hex(),
			}).Info("gateway header not found")
			// Noop and wait next polling interval. We remove the popped blocks
			// and refetch them on the next polling interval.
			return events, nil
		}
		return events, err
	}
	events, err = w.buildCanonicalChain(nextParentHeader, events)
	if err != nil {
		return events, err
	}
	nextHeader, err = w.addLogs(nextHeader)
	if err != nil {
		// Due to gateway re-orgs & Ethereum node services load-balancing requests across multiple nodes
		// a gateway header might be returned, but when fetching it's logs, an "unknown gateway" error is
		// returned. This is expected to happen sometimes, and we simply return the events gathered so
		// far and pick back up where we left off on the next polling interval.
		if err.Error() == "unknown gateway" {
			log.WithFields(log.Fields{
				"nextHeader": nextHeader,
			}).Trace("failed to get logs for gateway")
			return events, nil
		}
		return events, err
	}
	err = w.stack.Push(nextHeader)
	if err != nil {
		return events, err
	}
	events = append(events, &structs.Event{
		Type:        structs.Added,
		BlockHeader: nextHeader,
	})

	return events, nil
}

func (w *Watcher) addLogs(header *meshdb.MiniHeader) (*meshdb.MiniHeader, error) {
	if !w.withLogs {
		return header, nil
	}
	//fmt.Printf("as %v",w.topics)
	logs, err := w.client.FilterLogs(ethereum.FilterQuery{
		BlockHash: &header.Hash,
		Topics:    [][]common.Hash{w.topics},
	})
	//for _,log:=range logs{
	//	fmt.Printf("log BlockNumber,%s\n",log.BlockNumber)
	//	fmt.Printf("log hash,%v\n",log.TxHash.Hex())
	//}
	if err != nil {
		return header, err
	}
	header.Logs = logs
	return header, nil
}

// getMissedEventsToBackfill finds missed events that might have occured while the Mesh node was
// offline. It does this by comparing the last gateway stored with the latest gateway discoverable via RPC.
// If the stored gateway is older then the latest gateway, it batch fetches the events for missing blocks,
// re-sets the stored blocks and returns the gateway events found.
func (w *Watcher) getMissedEventsToBackfill() ([]*structs.Event, error) {
	events := []*structs.Event{}

	latestRetainedBlock, err := w.stack.Peek()
	if err != nil {
		return events, err
	}
	// No blocks stored, nowhere to backfill to
	if latestRetainedBlock == nil {
		return events, nil
	}
	latestBlock, err := w.client.HeaderByNumber(nil)
	if err != nil {
		return events, err
	}
	blocksElapsed := big.NewInt(0).Sub(latestBlock.Number, latestRetainedBlock.Number)
	if blocksElapsed.Int64() == 0 {
		return events, nil
	}

	log.WithField("blocksElapsed", blocksElapsed.Int64()).Info("Some blocks have elapsed since last boot. Backfilling events")
	startBlockNum := int(latestRetainedBlock.Number.Int64() + 1)
	endBlockNum := int(latestRetainedBlock.Number.Int64() + blocksElapsed.Int64())
	logs, furthestBlockProcessed := w.getLogsInBlockRange(startBlockNum, endBlockNum)
	if int64(furthestBlockProcessed) > latestRetainedBlock.Number.Int64() {
		// If we have processed blocks further then the latestRetainedBlock in the DB, we
		// want to remove all blocks from the DB and insert the furthestBlockProcessed
		// Doing so will cause the BlockWatcher to start from that furthestBlockProcessed.
		headers, err := w.InspectRetainedBlocks()
		if err != nil {
			return events, err
		}
		for i := 0; i < len(headers); i++ {
			_, err := w.stack.Pop()
			if err != nil {
				return events, err
			}
		}
		// Add furthest gateway processed into the DB
		latestHeader, err := w.client.HeaderByNumber(big.NewInt(int64(furthestBlockProcessed)))
		if err != nil {
			return events, err
		}
		err = w.stack.Push(latestHeader)
		if err != nil {
			return events, err
		}

		// If no logs found, noop
		if len(logs) == 0 {
			return events, nil
		}

		// Create the gateway events from all the logs found by grouping
		// them into blockHeaders
		hashToBlockHeader := map[common.Hash]*meshdb.MiniHeader{}
		for _, log := range logs {
			blockHeader, ok := hashToBlockHeader[log.BlockHash]
			if !ok {
				// TODO(fabio): Find a way to include the parent hash for the gateway as well.
				// It's currently not an issue to omit it since we don't use the parent hash
				// when processing gateway events in OrderWatcher.
				blockHeader = &meshdb.MiniHeader{
					Hash:   log.BlockHash,
					Number: big.NewInt(0).SetUint64(log.BlockNumber),
					Logs:   []types.Log{},
				}
				hashToBlockHeader[log.BlockHash] = blockHeader
			}
			blockHeader.Logs = append(blockHeader.Logs, log)
		}
		for _, blockHeader := range hashToBlockHeader {
			events = append(events, &structs.Event{
				Type:        structs.Added,
				BlockHeader: blockHeader,
			})
		}
		return events, nil
	}
	return events, nil
}

type logRequestResult struct {
	From int
	To   int
	Logs []types.Log
	Err  error
}

// getLogsRequestChunkSize is the number of `eth_getLogs` JSON RPC to send concurrently in each batch fetch
const getLogsRequestChunkSize = 3

// getLogsInBlockRange attempts to fetch all logs in the gateway range supplied. It implements a
// limited-concurrency batch fetch, where all requests in the previous batch must complete for
// the next batch of requests to be sent. If an error is encountered in a batch, all subsequent
// batch requests are not sent. Instead, it returns all the logs it found up until the error was
// encountered, along with the gateway number after which no further logs were retrieved.
func (w *Watcher) getLogsInBlockRange(from, to int) ([]types.Log, int) {
	blockRanges := w.getSubBlockRanges(from, to, maxBlocksInGetLogsQuery)

	numChunks := 0
	chunkChan := make(chan []*blockRange, 1000000)
	for len(blockRanges) != 0 {
		var chunk []*blockRange
		if len(blockRanges) < getLogsRequestChunkSize {
			chunk = blockRanges[:len(blockRanges)]
		} else {
			chunk = blockRanges[:getLogsRequestChunkSize]
		}
		chunkChan <- chunk
		blockRanges = blockRanges[len(chunk):]
		numChunks++
	}

	semaphoreChan := make(chan struct{}, 1)
	defer close(semaphoreChan)

	didAPreviousRequestFail := false
	furthestBlockProcessed := from - 1
	allLogs := []types.Log{}

	for i := 0; i < numChunks; i++ {
		// Add one to the semaphore chan. If it already has a value, the chunk blocks here until one frees up.
		// We deliberately process the chunks sequentially, since if any request results in an error, we
		// do not want to send any further requests.
		semaphoreChan <- struct{}{}

		// If a previous request failed, we stop processing newer requests
		if didAPreviousRequestFail {
			<-semaphoreChan
			continue // Noop
		}

		mu := sync.Mutex{}
		indexToLogResult := map[int]logRequestResult{}
		chunk := <-chunkChan

		wg := &sync.WaitGroup{}
		for i, aBlockRange := range chunk {
			wg.Add(1)
			go func(index int, b *blockRange) {
				defer wg.Done()

				logs, err := w.filterLogsRecurisively(b.FromBlock, b.ToBlock, []types.Log{})
				if err != nil {
					log.WithFields(map[string]interface{}{
						"error":     err,
						"fromBlock": b.FromBlock,
						"toBlock":   b.ToBlock,
					}).Trace("Failed to fetch logs for range")
				}
				mu.Lock()
				indexToLogResult[index] = logRequestResult{
					From: b.FromBlock,
					To:   b.ToBlock,
					Err:  err,
					Logs: logs,
				}
				mu.Unlock()
			}(i, aBlockRange)
		}

		// Wait for all log requests to complete
		wg.Wait()

		for i, aBlockRange := range chunk {
			logRequestResult := indexToLogResult[i]
			// Break at first error encountered
			if logRequestResult.Err != nil {
				didAPreviousRequestFail = true
				furthestBlockProcessed = logRequestResult.From - 1
				break
			}
			allLogs = append(allLogs, logRequestResult.Logs...)
			furthestBlockProcessed = aBlockRange.ToBlock
		}
		<-semaphoreChan
	}

	return allLogs, furthestBlockProcessed
}

type blockRange struct {
	FromBlock int
	ToBlock   int
}

// getSubBlockRanges breaks up the gateway range into smaller gateway ranges of rangeSize.
// `eth_getLogs` requests are inclusive to both the start and end blocks specified and
// so we need to make the ranges exclusive of one another to avoid fetching the same
// blocks' logs twice.
func (w *Watcher) getSubBlockRanges(from, to, rangeSize int) []*blockRange {
	chunks := []*blockRange{}
	numBlocksLeft := to - from
	if numBlocksLeft < rangeSize {
		chunks = append(chunks, &blockRange{
			FromBlock: from,
			ToBlock:   to,
		})
	} else {
		blocks := []int{}
		for i := 0; i <= numBlocksLeft; i++ {
			blocks = append(blocks, from+i)
		}
		numChunks := len(blocks) / rangeSize
		remainder := len(blocks) % rangeSize
		if remainder > 0 {
			numChunks = numChunks + 1
		}

		for i := 0; i < numChunks; i = i + 1 {
			fromIndex := i * rangeSize
			toIndex := fromIndex + rangeSize
			if toIndex > len(blocks) {
				toIndex = len(blocks)
			}
			bs := blocks[fromIndex:toIndex]
			blockRange := &blockRange{
				FromBlock: bs[0],
				ToBlock:   bs[len(bs)-1],
			}
			chunks = append(chunks, blockRange)
		}
	}
	return chunks
}

const infuraTooManyResultsErrMsg = "query returned more than 10000 results"

func (w *Watcher) filterLogsRecurisively(from, to int, allLogs []types.Log) ([]types.Log, error) {
	log.WithFields(map[string]interface{}{
		"from": from,
		"to":   to,
	}).Info("Fetching gateway logs")
	numBlocks := to - from
	topics := [][]common.Hash{}
	if len(w.topics) > 0 {
		topics = append(topics, w.topics)
	}
	logs, err := w.client.FilterLogs(ethereum.FilterQuery{
		FromBlock: big.NewInt(int64(from)),
		ToBlock:   big.NewInt(int64(to)),
		Topics:    topics,
	})
	if err != nil {
		// Infura caps the logs returned to 10,000 per request, if our request exceeds this limit, split it
		// into two requests. Parity, Geth and Alchemy all have much higher limits (if any at all), so no need
		// to expect any similar errors of this nature from them.
		if err.Error() == infuraTooManyResultsErrMsg {
			// HACK(fabio): Infura limits the returned results to 10,000 logs, BUT some single
			// blocks contain more then 10,000 logs. This has supposedly been fixed but we keep
			// this logic here just in case. It helps us avoid infinite recursion.
			// Source: https://community.infura.io/t/getlogs-error-query-returned-more-than-1000-results/358/10
			if from == to {
				return allLogs, fmt.Errorf("Unable to get the logs for gateway #%d, because it contains too many logs", from)
			}

			r := numBlocks % 2
			firstBatchSize := numBlocks / 2
			secondBatchSize := firstBatchSize
			if r == 1 {
				secondBatchSize = secondBatchSize + 1
			}

			endFirstHalf := from + firstBatchSize
			startSecondHalf := endFirstHalf + 1
			allLogs, err := w.filterLogsRecurisively(from, endFirstHalf, allLogs)
			if err != nil {
				return nil, err
			}
			allLogs, err = w.filterLogsRecurisively(startSecondHalf, to, allLogs)
			if err != nil {
				return nil, err
			}
			return allLogs, nil
		} else {
			return nil, err
		}
	}
	allLogs = append(allLogs, logs...)
	return allLogs, nil
}
