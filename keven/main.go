package main

import (
	"context"
	"encoding/json"
	"fmt"
	_ "github.com/joho/godotenv/autoload"
	"github.com/nats-io/nats.go"
	"github.com/pmker/onegw/keven/launcher"
	"github.com/pmker/onegw/keven/nonce"
	"github.com/pmker/onegw/keven/tool"
	"github.com/pmker/onegw/oneplus/backend/common"
	"github.com/pmker/onegw/oneplus/backend/sdk/ethereum"
	"github.com/pmker/onegw/oneplus/common/bgdb"
	"github.com/pmker/onegw/oneplus/connection"
	"github.com/shopspring/decimal"
	log "github.com/sirupsen/logrus"
	"os"
	"sync"
	"time"
)

func config() bgdb.Options {
	opt := bgdb.DefaultOptions
	opt.Dir = "./bgdb"
	opt.ValueDir = opt.Dir
	return opt
}

func main() {
	opt := config()
	db, err := bgdb.Open(opt)
	if err != nil {
		fmt.Printf("err %s", err.Error())
	}
	nc, err := nats.Connect("nats://localhost:4222")
	if err != nil {
		log.WithFields(map[string]interface{}{

			"err": err.Error(),
		}).Error("started p2p node")
	}
	//defer nc.Close()

	ec, err := nats.NewEncodedConn(nc, nats.JSON_ENCODER)
	if err != nil {
		log.WithFields(map[string]interface{}{

			"err": err.Error(),
		}).Error("started p2p node")
	}

	ctx, _ := context.WithCancel(context.Background())

	hydro := ethereum.NewEthereumHydro(os.Getenv("ETHEREUM_RPC_URL"), os.Getenv("HSK_HYBRID_EXCHANGE_ADDRESS"))
	//if os.Getenv("HSK_LOG_LEVEL") == "DEBUG" {
	//	hydro.EnableDebug(true)
	//}

	fallbackGasPrice := decimal.New(3, 9) // 3Gwei
	priceDecider := launcher.NewGasStationGasPriceDecider(fallbackGasPrice)

	nonceManager := nonce.NewNonceManager(hydro)

	signService := launcher.NewDefaultSignService("0x4C7AD49703E3D10FC7D9042CCFCDD629648D89902CFF4E8CCC450E05BA8E2926", nonceManager)

	bot := launcher.NewLauncher(ctx, signService, hydro, priceDecider, db)
	lans := make(map[string]*launcher.ISignService)

	redis := connection.NewRedisClient(os.Getenv("ONEGW_REDIS_URL"))

	eventQueue, _ := common.InitQueue(
		&common.RedisQueueConfig{
			Name:   common.HYDRO_ENGINE_EVENTS_QUEUE_KEY,
			Client: redis,
			Ctx:    ctx,
		})

	pm := &tool.Pmker{
		Db:           db,
		Nat:          ec,
		Bot:          bot,
		Hydro:        hydro,
		NonceManager: nonceManager,
		Lock:         sync.Mutex{},
		Lans:         lans,

		EventQueue: eventQueue,
	}

	//pm.start()
	//pm.SendEth()
	go pm.Start()

	go func() {
		for {
			select {
			case <-ctx.Done():

				return
			default:
				data, err := eventQueue.Pop()
				if err != nil {
					panic(err)
				}

				var std tool.Message //无需实例化，Unmarshal内部实现了
				err = json.Unmarshal(data, &std)
				if err != nil {
					fmt.Println("反序列化失败", err)
				}
				switch std.Type {
				case "sendeth":
					pm.SendEth(&std)
				case "senderc20":
					_ = pm.SendErc20(data)
				default:
					fmt.Printf(std.Type)
				}



			}
		}
	}()
	//address := "0x9d5ba7AD0B8aC6C88B9e112A26b12ED39700724e"
	////nonce, _ := pm.NonceManager.SyncNonce(address)
	/////pm.NonceManager.SetNonce(address,nonce)
	//
	//nonce1, err := pm.NonceManager.GetNonceOk(address)
	//if err != nil {
	//	fmt.Printf("aa nonce:%s,%d", err.Error(), nonce1)
	//
	//}
	//fmt.Printf("bb nonce %d", nonce1)

	var wg sync.WaitGroup
	wg.Add(1)
	//NewTimer 创建一个 Timer，它会在最少过去时间段 d 后到期，向其自身的 C 字段发送当时的时间
	//timer1 := time.NewTimer(2 * time.Second)

	//NewTicker 返回一个新的 Ticker，该 Ticker 包含一个通道字段，并会每隔时间段 d 就向该通道发送当时的时间。它会调  //整时间间隔或者丢弃 tick 信息以适应反应慢的接收者。如果d <= 0会触发panic。关闭该 Ticker 可以释放相关资源。
	ticker := time.NewTicker(10 * time.Second)

	go func(t *time.Ticker) {
		defer wg.Done()
		for {
			<-t.C
			//fmt.Printf("ticket %s",time.Now())
			//go pm.CheckOrderStatus()
		}
	}(ticker)
	wg.Wait()
	//NICE
	// go pm.SendErc20("0xc246b4da903ae8a7de212a3ed6e3353cdb98602b")

	select {}
}
