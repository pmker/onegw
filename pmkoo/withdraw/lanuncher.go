package withdraw

import (
	"context"
	"fmt"
	"github.com/pmker/onegw/oneplus/common/bgdb"
	"github.com/pmker/onegw/pmkoo"

	//"github.com/pmker/onegw/oneplus/backend/cli"
	"github.com/pmker/onegw/pmkoo/models"
	"github.com/pmker/onegw/oneplus/backend/launcher"
	"github.com/pmker/onegw/oneplus/backend/sdk/ethereum"
	"github.com/pmker/onegw/oneplus/backend/utils"
	"github.com/shopspring/decimal"
	"os"
	"time"
)
func (v *Watcher) getkey()(*pmkoo.Wallet,error) {
	wallet:=  []pmkoo.Wallet{}
	err := v.bowdb.Find(&wallet, bgdb.Where("Coin").Eq("ETH").SortBy("Index"))
	if err != nil {
		fmt.Printf(err.Error())
		return nil,err
	}

	return &wallet[1],nil
}
func (v *Watcher) Lanuncher()  {
	 ctx, _ := context.WithCancel(context.Background())
	////go cli.WaitExitSignal(stop)
	 fmt.Print(os.Getenv("ONE_DATABASE_URL"))
	 models.Connect(os.Getenv("ONE_DATABASE_URL"))

	// blockchain
	hydro := ethereum.NewEthereumHydro(os.Getenv("ETHEREUM_RPC_URL"), os.Getenv("HSK_HYBRID_EXCHANGE_ADDRESS"))
	if os.Getenv("HSK_LOG_LEVEL") == "DEBUG" {
		hydro.EnableDebug(true)
	}
	wallet,_:=v.getkey()

	fmt.Printf("wallet %+v",wallet)
	signService := launcher.NewDefaultSignService(wallet.Pk, hydro.GetTransactionCount)

	fallbackGasPrice := decimal.New(3, 9) // 3Gwei
	priceDecider := launcher.NewGasStationGasPriceDecider(fallbackGasPrice)

	launcher := launcher.NewLauncher(ctx, signService, hydro, priceDecider)

	go v.LanuncerLoop(launcher, utils.StartMetrics)


}

const pollingIntervalSeconds = 5

func (v *Watcher) LanuncerLoop(l *launcher.Launcher, startMetrics func()) {
	defer func(){ // 必须要先声明defer，否则不能捕获到panic异常
		//fmt.Println("c")
		if err:=recover();err!=nil{
			fmt.Println(err) // 这里的err其实就是panic传入的内容，55
		}
		utils.Infof("launcher stop!")
		//fmt.Println("d")
	}()


	utils.Infof("launcher start!")
	//defer utils.Infof("launcher stop!")
	//go startMetrics()

	for {
		launchLogs := models.LaunchLogDao.FindAllCreated()

		if len(launchLogs) == 0 {
			select {
			case <-l.Ctx.Done():
				utils.Infof("main loop Exit")
				return
			default:
				//utils.Infof("no logs need to be sent. sleep %ds", pollingIntervalSeconds)

				time.Sleep(pollingIntervalSeconds * time.Second)
				continue
			}
		}

		for _, modelLaunchLog := range launchLogs {
			modelLaunchLog.GasPrice = decimal.NullDecimal{
				Decimal: l.GasPriceDecider.GasPriceInWei(),
				Valid:   true,
			}

			log := launcher.LaunchLog{
				ID:          modelLaunchLog.ID,
				ItemType:    modelLaunchLog.ItemType,
				ItemID:      modelLaunchLog.ItemID,
				Status:      modelLaunchLog.Status,
				Hash:        modelLaunchLog.Hash,
				BlockNumber: modelLaunchLog.BlockNumber,
				From:        modelLaunchLog.From,
				To:          modelLaunchLog.To,
				Value:       modelLaunchLog.Value,
				GasLimit:    modelLaunchLog.GasLimit,
				GasUsed:     modelLaunchLog.GasUsed,
				GasPrice:    modelLaunchLog.GasPrice,
				Nonce:       modelLaunchLog.Nonce,
				Data:        modelLaunchLog.Data,
				ExecutedAt:  modelLaunchLog.ExecutedAt,
				CreatedAt:   modelLaunchLog.CreatedAt,
				UpdatedAt:   modelLaunchLog.UpdatedAt,
			}
			//payload, _ := json.Marshal(launchLog)
			//json.Unmarshal(payload, &log)

			signedRawTransaction := l.SignService.Sign(&log)
			transactionHash, err := l.BlockChain.SendRawTransaction(signedRawTransaction)

			if err != nil {
				utils.Debugf("%+v", modelLaunchLog)
				utils.Infof("Send Tx failed, launchLog ID: %d, err: %+v", modelLaunchLog.ID, err)
				//panic(err)
			}

			utils.Infof("Send Tx, launchLog ID: %d, hash: %s", modelLaunchLog.ID, transactionHash)

			// todo any other fields?
			modelLaunchLog.Hash = log.Hash

			models.UpdateLaunchLogToPending(modelLaunchLog)

			if err != nil {
				utils.Infof("Update Launch Log Failed, ID: %d, err: %s", modelLaunchLog.ID, err)
				//panic(err)
			}

			l.SignService.AfterSign()
		}
	}
}
