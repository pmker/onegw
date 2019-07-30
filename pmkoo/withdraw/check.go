package withdraw

import (
	"fmt"
	"github.com/pmker/onegw/oneplus/backend/sdk/ethereum"
	"github.com/pmker/onegw/pmkoo/models"
	"os"
	"time"
	logger "github.com/sirupsen/logrus"

)

func (v *Watcher) checkOrderStatus() {


	launchLogs := models.LaunchLogDao.FindAllPending()
	//fmt.Printf("%+v",launchLogs)

	if launchLogs!=nil {
		///fmt.Printf(launchLogs.Hash.String)
		go v.checkHash(launchLogs.Hash.String)
	}
}
func  (v *Watcher)checkHash(hash string) {
	defer func(){ // 必须要先声明defer，否则不能捕获到panic异常
		//fmt.Println("c")
		if err:=recover();err!=nil{
			fmt.Println(err) // 这里的err其实就是panic传入的内容，55
		}
		//fmt.Println("d")
	}()

	hydro := ethereum.NewEthereumHydro(os.Getenv("ETHEREUM_RPC_URL"), os.Getenv("HSK_HYBRID_EXCHANGE_ADDRESS"))
	//leven:=rpc.NewEthRPC(os.Getenv("ETHEREUM_RPC_URL"))
	tx, err := hydro.GetTransactionReceipt(hash)
	if err != nil {
		fmt.Printf(err.Error())

		logger.WithFields(logger.Fields{
			"error": err.Error(),
		}).Error("checkHash error")
	} else {

		logger.WithFields(logger.Fields{
			"hash": hash,
			 "result":tx.GetTxHash(),
		}).Info("checkHash ")
		 fmt.Printf("hash %s %t", hash, tx.GetTxHash())

	}

	if (hash==tx.GetTxHash()){
		status := "false"
		if tx.GetResult() {
			status = "success"
			executedAt := time.Now() //      .now(time.Unix(int64(time.), 0)
			transaction := models.TransactionDao.FindTransactionByHash(hash)
			transaction.Status = status
			transaction.ExecutedAt = executedAt
			_ = models.TransactionDao.UpdateTransaction(transaction)
			_ = models.LaunchLogDao.UpdateLaunchLogsStatusByItemID(status, transaction.ID)


			log:= models.LaunchLogDao.FindLaunchLogByItemID(int(transaction.ID))
			fmt.Printf("log id %d,%d,%v\n",int(transaction.ID),(transaction.ID),log)

			if err := v.nat.Publish("vgw.order.withdraw", log); err != nil {

				logger.WithFields(logger.Fields{
					"error": err.Error(),
				}).Error("Nat vgw.order.update")
			}
		}
	}


}