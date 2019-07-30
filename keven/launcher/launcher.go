package launcher

import (
	"context"
	"github.com/pmker/onegw/oneplus/backend/common"
	"github.com/pmker/onegw/oneplus/backend/sdk"
	"github.com/pmker/onegw/oneplus/backend/utils"
	"github.com/pmker/onegw/oneplus/common/bgdb"
	"github.com/shopspring/decimal"
)

type Launcher struct {
	Ctx             context.Context `json:"ctx"`
	GasPriceDecider GasPriceDecider
	SignService     ISignService
	BlockChain      sdk.BlockChain
	Db            *bgdb.Store
}

func NewLauncher(ctx context.Context, sign ISignService, hydro sdk.Hydro, gasPriceDecider GasPriceDecider,db *bgdb.Store) *Launcher {
	return &Launcher{
		Ctx:             ctx,
		SignService:     sign,
		BlockChain:      hydro,
		GasPriceDecider: gasPriceDecider,
		Db:db,
	}
}

func (l *Launcher) Add(launchLog *LaunchLog)error {

	//defer func(){ // 必须要先声明defer，否则不能捕获到panic异常
	//	//fmt.Println("c")
	//	if err:=recover();err!=nil{
	//		fmt.Println(err) // 这里的err其实就是panic传入的内容，55
	//	}
	//	//fmt.Println("d")
	//}()


	launchLog.GasPrice = decimal.NullDecimal{
		Decimal: l.GasPriceDecider.GasPriceInWei(),
		Valid:   true,
	}

	signedRawTransaction := l.SignService.Sign(launchLog)


	transactionHash, err := l.BlockChain.SendRawTransaction(signedRawTransaction)

	if err != nil {
		//utils.Debugf("%+v", launchLog)
		utils.Errorf("\n Send Tx failed,  %+v err: %+v \n",  launchLog, err)
		return err
	}


	launchLog.Status = common.STATUS_PENDING

	if err != nil {
		utils.Errorf("\n Update Launch Log Failed,   err: %s",   err)
		return err
	}
	//launchLog.Hash=transactionHash
	//utils.Infof("\n Send Tx,   hash: %s log: %+v\n ",   transactionHash,launchLog)

	err=l.Db.Insert(transactionHash, launchLog)
	if err!=nil{
		utils.Errorf("\n launchLog insert Db error %+v",err)
	}

	l.SignService.AfterSign(launchLog.From)
	return err
}
