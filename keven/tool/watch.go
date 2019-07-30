package tool

import (
	"database/sql"
	"fmt"
	_ "github.com/joho/godotenv/autoload"
	"github.com/pmker/onegw/keven/launcher"
	"github.com/pmker/onegw/oneplus/backend/common"
	"github.com/pmker/onegw/oneplus/backend/sdk/ethereum"
	"github.com/pmker/onegw/oneplus/common/bgdb"
	"os"
	//log "github.com/sirupsen/logrus"

)

func (p *Pmker) CheckOrderStatus() {

	repo := []launcher.LaunchLog{}
	err := p.Db.Find(&repo, bgdb.Where("Status").Eq(common.STATUS_PENDING))
 	if err!=nil{

	}
	fmt.Printf("Repo size:%d\n",len(repo))
	for _,log:=range repo{
		go p.checkHash(log.Hash)
	}
}
func  (p *Pmker) checkHash(hash sql.NullString) {
	defer func(){ // 必须要先声明defer，否则不能捕获到panic异常
		//fmt.Println("c")
		if err:=recover();err!=nil{
			fmt.Println(err) // 这里的err其实就是panic传入的内容，55
		}
		//fmt.Println("d")
	}()

	hydro := ethereum.NewEthereumHydro(os.Getenv("ETHEREUM_RPC_URL"), os.Getenv("HSK_HYBRID_EXCHANGE_ADDRESS"))
	//leven:=rpc.NewEthRPC(os.Getenv("ETHEREUM_RPC_URL"))
	tx, err := hydro.GetTransactionReceipt(hash.String)
	if err != nil {
		fmt.Printf(err.Error())
	} else {
		fmt.Printf("hash %s %t \n", hash, tx.GetResult())

	}

	if tx.GetResult() {

		//executedAt := time.Now() //      .now(time.Unix(int64(time.), 0)

		err = p.Db.UpdateMatching(&launcher.LaunchLog{}, bgdb.Where("Hash").Eq(hash), func(record interface{}) error {

			update, ok := record.(*launcher.LaunchLog) // record will always be a pointer
			if !ok {
				return fmt.Errorf("Record isn't the correct type!  Wanted Person, got %T", record)
			}

			update.Status=common.STATUS_SUCCESSFUL
			fmt.Printf("Update %v \n", update)

			return nil
		})
	}

}