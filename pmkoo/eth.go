package pmkoo

import (
	"fmt"
	"github.com/pmker/onegw/oneplus/backend/sdk/ethereum"
	"os"

)

func EthCheckHash(hash string) bool {
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
	} else {
		fmt.Printf("hash %s %t", hash, tx.GetResult())

	}
	status := false
	if tx.GetResult() {
		status = true
		//executedAt := time.Now() //      .now(time.Unix(int64(time.), 0)
		//transaction := models.TransactionDao.FindTransactionByHash(hash)
		//transaction.Status = status
		//transaction.ExecutedAt = executedAt
		//_ = models.TransactionDao.UpdateTransaction(transaction)
		//_ = models.LaunchLogDao.UpdateLaunchLogsStatusByItemID(status, transaction.ID)

	}
	return status

}