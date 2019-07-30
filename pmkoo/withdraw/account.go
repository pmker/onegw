package withdraw

import (
	"fmt"
	client2 "github.com/koinotice/vite/client"
	"github.com/koinotice/vite/common/types"
	"github.com/koinotice/vite/rpcapi/api"
	logger "github.com/sirupsen/logrus"
	"math"
	"strconv"
)

const PageSize = 256

//分页获取在途交易记录
func (v *Watcher) GetOnroadInfoByAddress() {

	defer func() { // 必须要先声明defer，否则不能捕获到panic异常
		//fmt.Println("c")
		if err := recover(); err != nil {
			fmt.Println(err) // 这里的err其实就是panic传入的内容，55
		}
		//fmt.Println("d")
	}()

	to, err := types.HexToAddress(v.withdrawAddress)
	if err != nil {
		fmt.Print("%s", err.Error())

	}
	resp := &api.RpcAccountInfo{}
	if err := v.viteRpc.GetClient().Call(&resp, "onroad_getOnroadInfoByAddress", to); err != nil {
		fmt.Printf(err.Error())
	}

	totalNUmber, err := strconv.Atoi(resp.TotalNumber)

	pageCount := math.Ceil(float64(totalNUmber) / float64(PageSize))
	logger.WithFields(logger.Fields{
		"pageCount":   pageCount,
		"totalNUmber": totalNUmber,
	}).Info("getOnroadInfoByAddress")
	for page := int64(0); page <= int64(pageCount); page++ {

		go v.GetOnroadBlocksByAddress(to, page)
	}

}

//获取在途交易信息
func (v *Watcher) GetOnroadBlocksByAddress(to types.Address, page int64) {

	var blocks []*api.AccountBlock
	if err := v.viteRpc.GetClient().Call(&blocks, "onroad_getOnroadBlocksByAddress", to, page, PageSize); err != nil {
		fmt.Printf(err.Error())

	}

	for _, block := range blocks {

		logger.WithFields(logger.Fields{
			"block.Hash": block.Hash,
		}).Info("onroad_getOnroadBlocksByAddress ")
		//fmt.Printf("hash %s \n ",block.Hash)

		go v.SubmitResponseTx(block.Hash)

	}

}

func (v *Watcher) SubmitResponseTx(hash types.Hash) error {
	//wallet,err:=pmkoo.UnlockViteWallet();
	//if err!=nil{
	//
	//}
	to, err := types.HexToAddress(v.withdrawAddress)
	if err != nil {

		fmt.Printf(err.Error())
	}
	requestHash := types.HexToHashPanic(hash.String())

	block, err := v.viteClient.BuildResponseBlock(client2.ResponseTxParams{
		SelfAddr:    to,
		RequestHash: requestHash,
	}, nil)

	if err != nil {
		logger.WithFields(logger.Fields{

			"to":          to,
			"requestHash": requestHash,
		}).Error("vite Account BuildResponseBlock error")
	}

 //fmt.Printf("viteClient.SignData wallet:%s, block:%v",v.Wallet2,block)

	err = v.viteClient.SignData(v.Wallet2, block)
	if err != nil {
		logger.WithFields(logger.Fields{
		"to":to,
		"block":block,


		}).Error("SignData error")
	}
	err = v.viteRpc.SendRawTx(block)

	if err != nil {
		logger.WithFields(logger.Fields{
			"block.Hash": block.Hash,
		}).Error("vite Account  SendRawTx ")


	}
	return err
}
