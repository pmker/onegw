package deposit

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/ethereum/go-ethereum/event"
	client2 "github.com/koinotice/vite/client"
	types2 "github.com/pmker/onegw/pmkoo/types"

	"github.com/koinotice/vite/common/types"
	"github.com/koinotice/vite/wallet/entropystore"
	"github.com/nats-io/nats.go"
	"github.com/pmker/onegw/oneplus/common/bgdb"
	"github.com/pmker/onegw/oneplus/common/blero"
	Queue "github.com/pmker/onegw/oneplus/common/queue"
	bw "github.com/pmker/onegw/oneplus/watch"
	"github.com/pmker/onegw/oneplus/watch/rpc"
	"github.com/pmker/onegw/pmkoo"
	logger "github.com/sirupsen/logrus"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

//const pollingIntervalSeconds = 5

type Action struct {
	setupMux               sync.RWMutex
	blockSubscription      event.Subscription
	bowdb                  *bgdb.Store
	nat                    *nats.EncodedConn
	blockWatcher           *bw.Watcher
	viteRpc                client2.RpcClient
	viteClient             client2.Client
	Wallet2                *entropystore.Manager
	depositAddress         string
	isSetup                bool
	pollingIntervalSeconds int64
	ethRpc                 *rpc.EthBlockChainRPCWithRetry
	mu                     sync.Mutex
	queue                  Queue.UniqueList
	bl                     *blero.Blero
	config                 *types2.Config
}

func New(bowdb *bgdb.Store, nat *nats.EncodedConn, blockWatcher *bw.Watcher, wallet *entropystore.Manager, config *types2.Config) (*Action, error) {

	pollingDur, err := strconv.Atoi(os.Getenv("ORDER_POLLING_DUR"))
	if err != nil {
	}

	ethRpc := rpc.NewEthRPCWithRetry(os.Getenv("ETHEREUM_RPC_URL"), 5)
	queue := Queue.New()
	bl := blero.New("")

	// Start Blero
	bl.Start(bowdb.Badger())

	// defer Stopping Blero
	//defer bl.Stop()
	rpc, err := client2.NewRpcClient(os.Getenv("VITE_RPC_URL"))
	if err != nil {
		fmt.Print(err)

	}

	client, err := client2.NewClient(rpc)

	if err != nil {
		fmt.Print(err)

	}

	if err != nil {
		fmt.Print(err)
	}
	//depositAddress, _ := key.Address()
	_, key, err := wallet.DeriveForFullPath(fmt.Sprintf(config.Vpath, 0))


	if err != nil {
		fmt.Printf(err.Error())
	}
	depositAddress, err := key.Address()
	if err != nil {
		fmt.Printf(err.Error())
	}

	fmt.Println("depoo_withdra%s\n", depositAddress)

	w := &Action{
		bowdb:                  bowdb,
		nat:                    nat,
		blockWatcher:           blockWatcher,
		pollingIntervalSeconds: int64(pollingDur),
		ethRpc:                 ethRpc,
		queue:                  queue,
		bl:                     bl,
		viteRpc:                rpc,
		viteClient:             client,
		Wallet2:                wallet,
		config:                 config,

		depositAddress: depositAddress.String(),
	}
	return w, nil
}

func (v *Action) Start() error {

	v.setupMux.Lock()
	defer v.setupMux.Unlock()
	if v.isSetup {
		return errors.New("Setup can only be called once")
	}

	logger.WithFields(logger.Fields{

	}).Info("deposit start ok")

	v.setupEventWatcher()
	//v.WatchOrder()
	//
	v.isSetup = true

	//go v.CheckOrderSend()
	return nil
}

func (v *Action) setupEventWatcher() {
	var wg sync.WaitGroup
	wg.Add(1)
	//NewTimer 创建一个 Timer，它会在最少过去时间段 d 后到期，向其自身的 C 字段发送当时的时间
	//timer1 := time.NewTimer(2 * time.Second)

	//NewTicker 返回一个新的 Ticker，该 Ticker 包含一个通道字段，并会每隔时间段 d 就向该通道发送当时的时间。它会调  //整时间间隔或者丢弃 tick 信息以适应反应慢的接收者。如果d <= 0会触发panic。关闭该 Ticker 可以释放相关资源。

	ticker := time.NewTicker(15 * time.Second)

	go func(t *time.Ticker) {
		defer wg.Done()
		for {
			<-t.C
			//fmt.Printf("ticket %s",time.Now())
			go v.WatchOrder()
		}
	}(ticker)

	v.bl.RegisterProcessorFunc(func(j *blero.Job) error {
		//data:=j.Data.(*gateway.VgwOrder)

		///fmt.Printf("job name%s \n",j.Name)

		var std VgwOrder //无需实例化，Unmarshal内部实现了
		err := json.Unmarshal(j.Data, &std)
		if err != nil {
			fmt.Println("反序列化失败", err)
		}
		fmt.Printf("job name%s, %v \n", j.Name, std)
		err = v.SendToken(std)

		if err != nil {
			keys := []string{std.TxHash}
			v.bowdb.Cache().Del(keys)

			fmt.Printf("send token err:%s", err.Error())
		}
		return err
		// Do some processing, access job name with j.Name, job data with j.Data
	})
}

//检查确认数

func (v *Action) WatchOrder() {

	currentNumber, err := v.ethRpc.GetCurrentBlockNum()
	if err != nil {
		fmt.Printf("error", err.Error())
	}

	confrimNum, err := strconv.Atoi(os.Getenv("BLOCK_CONFIRM_NUM"))

	cc := (currentNumber) - uint64(confrimNum)

	order := []VgwOrder{}

	v.bowdb.Find(&order, bgdb.Where("Status").Eq(0).And("BlockNumber").Lt(cc))
	//fmt.Printf("order size %d \n", len(order))

	for _, o := range order {

		if exists := v.bowdb.Cache().Has(o.TxHash); !exists {
			bytes, err := json.Marshal(o)
			if err != nil {
				fmt.Println(err)
			}

			success := pmkoo.EthCheckHash(o.TxHash)
			if success {

				logger.WithFields(logger.Fields{

					"hash": o.TxHash,
				}).Info("Job Add  send vite token !!")

				v.bowdb.Cache().Set(o.TxHash, "1", 0)

				v.bl.EnqueueJob(o.TxHash, bytes)
			} else {
				fmt.Printf("tx nothingness %v \n", o)
			}

		} else {

			fmt.Printf("exists %v \n", o)
		}

	}

}

func (v *Action) getInfo(order *VgwOrder) (string, string) {
	walletAddress, err := v.bowdb.Cache().Get(strings.ToLower(order.To));
	if err != nil {
		fmt.Printf("err %s", err.Error())
	}
	ViteTokenAddress, err := v.bowdb.Cache().Get(strings.ToLower(order.Token));
	if err != nil {
		fmt.Printf("err %s", err.Error())
	}
	return walletAddress, ViteTokenAddress

}

func (v *Action) SendToken(order VgwOrder) error {
	v.mu.Lock()
	defer v.mu.Unlock()

	viteAddress, viteTokenAddress := v.getInfo(&order)
	to, err := types.HexToAddress(viteAddress)

	self, err := types.HexToAddress(v.depositAddress)
	if err != nil {
		fmt.Print(err)
		return err
	}

	viteToken, err := types.HexToTokenTypeId(viteTokenAddress)
	if err != nil {
		fmt.Print(err)
	}
	viteAmount := order.Amount

	block, err := v.viteClient.BuildNormalRequestBlock(client2.RequestTxParams{
		ToAddr:   to,
		SelfAddr: self,
		Amount:   viteAmount,
		TokenId:  viteToken,
		Data:     []byte("XS vite gateway send"),
	}, nil)

	if err != nil {
		fmt.Print(err.Error())
	}

	err = v.viteClient.SignData(v.Wallet2, block)
	if err != nil {
		fmt.Print(err.Error())
	}

	err = v.viteRpc.SendRawTx(block)

	if err != nil {

		fmt.Print("send deposit chain token err:%s", err.Error())
		return err
	}

	err = v.bowdb.UpdateMatching(&VgwOrder{}, bgdb.Where("TxHash").Eq(order.TxHash), func(record interface{}) error {
		update, ok := record.(*VgwOrder) // record will always be a pointer
		if !ok {
			return fmt.Errorf("Record isn't the correct type!  Wanted Person, got %T", record)
		}
		update.Status = 2
		update.ViteStaus = 1
		update.ViteHash = block.Hash.Hex()
		update.ViteToken = viteTokenAddress
		update.ViteWalletAddress = viteAddress
		update.ViteAmout = viteAmount

		logger.WithFields(logger.Fields{
			"token":  viteTokenAddress,
			"to":     viteAddress,
			"amount": viteAmount,
		}).Info("Vite send token success")

		if err := v.nat.Publish("vgw.order.create", update); err != nil {

			logger.WithFields(logger.Fields{
				"error": err.Error(),
			}).Error("Nat vgw.order.update")
		}

		return nil
	})
	return err

}
