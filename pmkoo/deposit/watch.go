package deposit

import (
	"context"
	"errors"
	"github.com/nats-io/nats.go"
	"github.com/pmker/onegw/oneplus/watch/structs"

	"github.com/pmker/onegw/oneplus/common/bgdb"
	bw "github.com/pmker/onegw/oneplus/watch"
	"github.com/shopspring/decimal"
	"math/big"
	"strings"

	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/event"
	"github.com/pmker/onegw/ethereum"
	logger "github.com/sirupsen/logrus"
)

// minCleanupInterval specified the minimum amount of time between orderbook cleanup intervals. These
// cleanups are meant to catch any stale orders that somehow were not caught by the event watcher
// process.
var minCleanupInterval = 1 * time.Hour

// lastUpdatedBuffer specifies how long it must have been since an order was last updated in order to
// be re-validated by the cleanup worker
var lastUpdatedBuffer = 30 * time.Minute

// permanentlyDeleteAfter specifies how long after an order is marked as IsRemoved and not updated that
// it should be considered for permanent deletion. Blocks get mined on avg. every 12 sec, so 4 minutes
// corresponds to a gateway depth of ~20.
var permanentlyDeleteAfter = 4 * time.Minute

// expirationPollingInterval specifies the interval in which the order watcher should check for expired
// orders
var expirationPollingInterval = 50 * time.Millisecond

// Watcher watches all order-relevant state and handles the state transitions
type Watcher struct {

	bowdb                      *bgdb.Store
	blockWatcher               *bw.Watcher
	eventDecoder               *Decoder
	blockFeed                  event.Feed
	blockScope                 event.SubscriptionScope
	blockSubscription          event.Subscription
	contractAddresses          ethereum.ContractAddresses
	orderFeed                  event.Feed
	orderScope                 event.SubscriptionScope // Subscription scope tracking current live listeners
	cleanupCtx                 context.Context
	cleanupCancelFunc          context.CancelFunc
	contractAddressToSeenCount map[common.Address]uint
	isSetup                    bool
	setupMux                   sync.RWMutex

	nat *nats.EncodedConn
}
type ERC20AssetData struct {
	Address common.Address
}


type VgwOrder struct {
	From        string
	To          string
	Amount      *big.Int
	Token       string
	BlockNumber uint64 `json:"blockNumber"`
	// hash of the transaction
	TxHash            string
	Status            int `default:"0"`
	ViteWalletAddress string
	ViteAmout         *big.Int
	ViteToken         string
	ViteHash          string
	ViteStaus         int `default:"0"`
}

// New instantiates a new order watcher
func Deposit( blockWatcher *bw.Watcher, bowdb *bgdb.Store, networkID int, nat *nats.EncodedConn) (*Watcher, error) {
	decoder, err := NewDecoder(bowdb)
	if err != nil {
		return nil, err
	}
	contractAddresses, err := ethereum.GetContractAddressesForNetworkID(networkID)
	if err != nil {
		return nil, err
	}
	cleanupCtx, cleanupCancelFunc := context.WithCancel(context.Background())

	w := &Watcher{

		bowdb:        bowdb,
		blockWatcher: blockWatcher,
		//expirationWatcher:          expirationwatch.New(expirationBuffer),
		cleanupCtx:                 cleanupCtx,
		cleanupCancelFunc:          cleanupCancelFunc,
		contractAddressToSeenCount: map[common.Address]uint{},
		//orderValidator:             orderValidator,
		eventDecoder:      decoder,
		contractAddresses: contractAddresses,
		nat:               nat,
	}

	return w, nil
}

// Start sets up the event & expiration watchers as well as the cleanup worker. Event
// watching will require the blockwatch.Watcher to be started however.
func (w *Watcher) Start() error {
	w.setupMux.Lock()
	defer w.setupMux.Unlock()
	if w.isSetup {
		return errors.New("Setup can only be called once")
	}

	//if err := w.nat.Publish("vgw.transfer", &stock{Symbol: "GOOG", Price: 331200}); err != nil {
	//	fmt.Printf("asdfasdf err %s", err.Error())
	//}

	w.AddPlug()

	logger.Printf("start depoist")

	w.isSetup = true
	return nil
}

// Stop closes the gateway subscription, stops the event, expiration watcher and the cleanup worker.
func (w *Watcher) Stop() error {
	w.setupMux.Lock()
	if !w.isSetup {
		w.setupMux.Unlock()
		return errors.New("Cannot teardown before calling Setup()")
	}
	w.setupMux.Unlock()

	// Stop event subscription
	w.blockSubscription.Unsubscribe()

	// Stop expiration watcher

	// Stop cleanup worker
	w.stopCleanupWorker()
	return nil
}

// Subscribe allows one to subscribe to the order events emitted by the OrderWatcher.
// To unsubscribe, simply call `Unsubscribe` on the returned subscription.
// The sink channel should have ample buffer space to avoid blocking other subscribers.
// Slow subscribers are not dropped.
func (w *Watcher) Subscribe(sink chan<- []*structs.Event) event.Subscription {
	return w.blockScope.Track(w.blockFeed.Subscribe(sink))
}
func (w *Watcher) stopCleanupWorker() {
	w.cleanupCancelFunc()
}


func (w *Watcher) handleDecodeErr(err error, eventType string) {
	switch err.(type) {
	case UnsupportedEventError:
		logger.WithFields(logger.Fields{
			"eventType":       eventType,
			"topics":          err.(UnsupportedEventError).Topics,
			"contractAddress": err.(UnsupportedEventError).ContractAddress,
		}).Warn("unsupported event found")

	default:
		logger.WithFields(logger.Fields{
			"error": err.Error(),
		}).Panic("unexpected event decoder error encountered")
	}
}
func (w *Watcher) AddPlug() {

	//w.blockWatcher.Start()
	blockEvents := make(chan []*structs.Event, 1)
	w.blockSubscription = w.blockWatcher.Subscribe(blockEvents)

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
				logger.WithFields(logger.Fields{
					"error": err.Error(),
				}).Error("subscription error encountered")
				return

			case events := <-blockEvents:

				//hashToOrderWithTxHashes := map[tool.Hash]*OrderWithTxHashes{}
				for _, event := range events {

					logger.WithFields(logger.Fields{
						"Hash":   event.BlockHeader.Hash.String(),
						"Number": event.BlockHeader.Number,
					}).Info("current block " )

					//fmt.Printf("current block number,%s \n", event.BlockHeader.Number)
					for _, log := range event.BlockHeader.Logs {
						//fmt.Printf("log11,txhash %s \n", log.BlockNumber.Hex())
						//fmt.Printf("log11 contract,%s \n", log.Address.Hex())
						////fmt.Printf("log11,%s \n",log.)

						eventType, err := w.eventDecoder.FindEventType(log)

						 //fmt.Printf("eventType:%s\n", eventType)
						///fmt.Printf("eventType err:%s\n", err)
						if err != nil {
							switch err.(type) {
							case UntrackedTokenError:
								continue
							case UnsupportedEventError:
								logger.WithFields(logger.Fields{
									"topics":          err.(UnsupportedEventError).Topics,
									"contractAddress": err.(UnsupportedEventError).ContractAddress,
								}).Info("unsupported event found while trying to find its event type")
								continue
							default:
								logger.WithFields(logger.Fields{
									"error": err.Error(),
								}).Panic("unexpected event decoder error encountered")
							}
						}

						//var orders []*meshdb.Order
						switch eventType {
						case "ERC20TransferEvent":
							var transferEvent ERC20TransferEvent

							// fmt.Printf("log 22  %v \n",log)

							err = w.eventDecoder.Decode(log, &transferEvent)
							if err != nil {
								w.handleDecodeErr(err, eventType)
								continue
							}

							to := strings.ToLower(transferEvent.To.String())
							is := w.bowdb.Cache().Has(to)

							 //fmt.Printf("log 22  %s %t\n", to, is)

							if is {

								order := &VgwOrder{
									From:        transferEvent.From.String(),
									To:          to,
									Amount:      transferEvent.Value,
									BlockNumber: log.BlockNumber,
									TxHash:      log.TxHash.Hex(),
									Token:       strings.ToLower(log.Address.String()),
								}
								err1 := w.bowdb.Insert(order.TxHash, order)
								if err1 != nil {
									logger.WithFields(logger.Fields{
										"error": err1.Error(),
									}).Warn("insert order warn")
								} else {
									logger.WithFields(logger.Fields{
										"to":order.To,
										"amount":order.Amount,
										"token":order.Token,
										"orderHash": "db order.insert",
									}).Info("Onegw Watched Deposite order")
								}

								// Publish the message
								if err := w.nat.Publish("vgw.order.insert", &transferEvent); err != nil {

									//logger.WithFields(logger.Fields{
									//	"error":     err.Error(),
									//	"orderHash": "nats order.insert",
									//}).Info("Deposite order: to %s,token:%s,amount:%d", order.To, order.Token, order.Amount)

								}
							}

							//
							//	//

							//}
							//orders = w.findOrdersAndGenerateOrderEvents(transferEvent.From, log.Address, nil)

						default:
							logger.WithFields(logger.Fields{
								"eventType": eventType,
								"log":       log,
							}).Info("unknown eventType encountered")
						}

					}
				}


			}
		}
	}()

}

func HexToDecimal(hex string) (decimal.Decimal, bool) {
	if hex[0:2] == "0x" || hex[0:2] == "0X" {
		hex = hex[2:]
	}

	b := new(big.Int)
	b, ok := b.SetString(hex, 16)
	if !ok {
		return decimal.Zero, false
	}

	return decimal.NewFromBigInt(b, 0), true
}
