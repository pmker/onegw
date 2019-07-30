package withdraw

import (
	"crypto/ecdsa"
	"database/sql"
	"errors"
	"fmt"
	types2 "github.com/pmker/onegw/pmkoo/types"

	common2 "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/event"
	client2 "github.com/koinotice/vite/client"
	"github.com/koinotice/vite/common/types"
	"github.com/koinotice/vite/rpcapi/api"
	"github.com/koinotice/vite/wallet/entropystore"
	"github.com/nats-io/nats.go"
	"github.com/pmker/onegw/oneplus/backend/common"
	"github.com/pmker/onegw/oneplus/backend/utils"
	"github.com/pmker/onegw/pmkoo"
	"github.com/shopspring/decimal"

	"github.com/pmker/onegw/oneplus/common/bgdb"
	"github.com/pmker/onegw/oneplus/common/blero"
	Queue "github.com/pmker/onegw/oneplus/common/queue"
	"github.com/pmker/onegw/oneplus/watch/rpc"
	"github.com/pmker/onegw/pmkoo/models"
	logger "github.com/sirupsen/logrus"
	"math"
	"math/big"
	"os"
	"strings"
	"time"
	"unicode"

	"strconv"
	"sync"
)

//const pollingIntervalSeconds = 5
const (
	BlockTypeSendCreate   = byte(1) // send
	BlockTypeSendCall     = byte(2) // send
	BlockTypeSendReward   = byte(3) // send
	BlockTypeReceive      = byte(4) // receive
	BlockTypeReceiveError = byte(5) // receive
	BlockTypeSendRefund   = byte(6) // send

	BlockTypeGenesisReceive = byte(7) // receive
)

type Watcher struct {
	setupMux          sync.RWMutex
	blockSubscription event.Subscription
	bowdb             *bgdb.Store
	nat               *nats.EncodedConn
	viteRpc           client2.RpcClient
	viteClient        client2.Client

	isSetup                bool
	pollingIntervalSeconds int64
	ethRpc                 *rpc.EthBlockChainRPCWithRetry
	mu                     sync.Mutex
	queue                  Queue.UniqueList
	bl                     *blero.Blero
	pageSize               int
	withdrawAddress        string
	Wallet2                *entropystore.Manager
	Config                 *types2.Config
}

func New(bowdb *bgdb.Store, nat *nats.EncodedConn, wallet *entropystore.Manager, config *types2.Config) (*Watcher, error) {

	pollingDur, err := strconv.Atoi(os.Getenv("ORDER_POLLING_DUR"))
	if err != nil {
	}

	ethRpc := rpc.NewEthRPCWithRetry(os.Getenv("ETHEREUM_RPC_URL"), 5)

	rpc, err := client2.NewRpcClient(os.Getenv("VITE_RPC_URL"))
	if err != nil {
		fmt.Print(err)

	}

	client, err := client2.NewClient(rpc)

	if err != nil {
		fmt.Print(err)
	}
	//wallet, err := pmkoo.UnlockViteWallet(config)

	if err != nil {
		fmt.Print(err)
	}

	_, key, err := wallet.DeriveForFullPath(fmt.Sprintf(config.Vpath, 0))
	if err != nil {
		fmt.Printf(err.Error())
	}
	withdrawAddress, err := key.Address()
	if err != nil {
		fmt.Printf(err.Error())
	}
	 fmt.Println("vite_withdra%s",withdrawAddress)

	w := &Watcher{
		bowdb: bowdb,
		nat:   nat,

		pollingIntervalSeconds: int64(pollingDur),
		ethRpc:                 ethRpc,
		withdrawAddress:        withdrawAddress.String(),
		viteRpc:                rpc,
		viteClient:             client,
		pageSize:               100,
		Wallet2:                wallet,
		Config:                 config,
	}
	return w, nil
}

func (v *Watcher) Start() error {

	v.setupMux.Lock()
	defer v.setupMux.Unlock()

	//go pmkoo.UnlockViteWallet()
	if v.isSetup {
		return errors.New("Setup can only be called once")
	}

	logger.WithFields(logger.Fields{

	}).Info("withdraw start")

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
			go v.ListenWithdraw()

			go v.checkOrderStatus()

			go v.GetOnroadInfoByAddress()
		}
	}(ticker)
	//fmt.Printf("withdraw start %s", 1)
	//v.ListenWithdraw()
	//v.WatchOrder()
	//
	go v.Lanuncher()
	v.isSetup = true

	//go v.CheckOrderSend()
	return nil
}

//监听收款地址的收款记录
func (v *Watcher) ListenWithdraw() {

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
	var resp *api.RpcAccountInfo
	if err := v.viteRpc.GetClient().Call(&resp, "ledger_getAccountByAccAddr", to); err != nil {
		fmt.Printf(err.Error())
	}

	totalNUmber, err := strconv.Atoi(resp.TotalNumber)

	pageCount := math.Ceil(float64(totalNUmber) / float64(v.pageSize))

	//fmt.Printf("pageCount %f ,%d ",pageCount,totalNUmber)
	for page := int64(0); page <= int64(pageCount); page++ {

		go v.batchGetBlock(to, page)
	}

}

//监听收款地址的收款记录
func (v *Watcher) batchGetBlock(to types.Address, page int64) {

	//logger.WithFields(logger.Fields{
	//	"to": to,
	//	"page":page,
	//}).Info("batchGetBlock %s", to)
	//
	//
	//to, err := types.HexToAddress("vite_159df2046e8ab0348fd3d25070bbeaec58abd2c03bb93dce0a")

	var blocks []*api.AccountBlock
	if err := v.viteRpc.GetClient().Call(&blocks, "ledger_getBlocksByAccAddr", to, page, v.pageSize); err != nil {
		fmt.Printf(err.Error())

		//go v.addHistory(height)
	}

	for _, block := range blocks {

		err1 := v.bowdb.Insert(block.Hash, block)
		if err1 != nil {

		} else {

			if block.BlockType == BlockTypeReceive {
				logger.WithFields(logger.Fields{
					"BlockType": block.BlockType,
					"Height":    block.Height,
					"Hash":      block.Hash,
				}).Info("Account block.type %s,%s", block.Hash, block.Height)

				go v.addSignErc20(block)

			}
		}

	}

}

type DbBlock struct {
	BlockType      byte              `json:"blockType"`
	Height         string            `json:"height"`
	Hash           types.Hash        `json:"hash"`
	FromAddress    types.Address     `json:"fromAddress"`
	ToAddress      types.Address     `json:"toAddress"`
	TokenId        types.TokenTypeId `json:"tokenId"`
	Amount         string            `json:"amount"`
	ToErc20Address string            `json:"data2"`
	Data           []byte            `json:"data"`
}

func TrimZero(s string) string {
	str := make([]rune, 0, len(s))
	for _, v := range []rune(s) {
		if !unicode.IsLetter(v) && !unicode.IsDigit(v) {
			continue
		}

		str = append(str, v)
	}
	return string(str)
}

//通过hash获取request Data
func (v *Watcher) addSignErc20(block *api.AccountBlock) {

	//hash := ("32068917cd97591a77955f2d7e6c4bfc49e6bc17bd4b6c21332347c0fb7005cb")
	resBlock := v.getViteHashData(block.FromBlockHash)
	address := TrimZero(string(resBlock.Data))

	dblock := &DbBlock{
		BlockType:      resBlock.BlockType,
		Height:         resBlock.Height,
		Hash:           resBlock.Hash,
		TokenId:        resBlock.TokenId,
		Amount:         *resBlock.Amount,
		ToErc20Address: address,
		Data:           (resBlock.Data),
	}

	//如果地址是有效以太地址，则打币
	//re := regexp.MustCompile("^0x[0-9a-fA-F]{40}$")re.MatchString(address)
	if common2.IsHexAddress(address) {
		if tokenAddress := v.getErc20ByViteTokenId(dblock.TokenId.String()); tokenAddress != "" {

			go v.sendErc20Sign(dblock, tokenAddress)
		}
	} else {
		fmt.Printf("no address %s", address)
	}
	/*
		fmt.Printf("is valid: %v\n", re.MatchString()) // is valid: true

		fmt.Printf("block  %s", dblock.Data)
	*/
	v.nat.Publish("vgw.withdraw.blocks", dblock)

}

//通过hash获取request Data
func (v *Watcher) getViteHashData(hash types.Hash) *api.AccountBlock {

	var block *api.AccountBlock

	if err := v.viteRpc.GetClient().Call(&block, "ledger_getBlockByHash", hash); err != nil {
		//fmt.Printf(err.Error())
	}

	return block

}
func (v *Watcher) getErc20ByViteTokenId(tokenid string) string {
	result := []pmkoo.Token{}

	err := v.bowdb.Find(&result, bgdb.Where("ViteTokenAddress").Eq(tokenid))
	if err != nil {
		fmt.Printf("getErc20ByViteTokenId %s \n", err.Error())
	}
	if len(result) > 0 {
		return result[0].Address
	}
	fmt.Printf("getErc20ByViteTokenId no token %s \n", tokenid)

	return ""

}

func (v *Watcher) getPublic() common2.Address {

	wallet, _ := v.getkey()

	fmt.Printf("wallet %+v", wallet)
	privateKey, err := crypto.HexToECDSA(wallet.Pk)
	if err != nil {
		fmt.Printf(err.Error())
	}

	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		fmt.Printf("error casting public key to ECDSA")

	}

	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)
	return fromAddress
}

//保存签名信息到数据库
func (v *Watcher) sendErc20Sign(block *DbBlock, tokenAddress string) {

	transaction := &models.Transaction{
		Status: common.STATUS_PENDING,
		TransactionHash: &sql.NullString{
			Valid:  false,
			String: "",
		},
		MarketID:   "vgw",
		Action:     "withdraw_deposit",
		ExecutedAt: time.Now().UTC(),
		CreatedAt:  time.Now().UTC(),
	}
	err := models.TransactionDao.InsertTransaction(transaction)

	if err != nil {
		fmt.Printf("insert trans err %s", err.Error())
	}

	//t1:=v.getErc20ByViteTokenId(block.TokenId.String())

	//if tokenAddress := v.getErc20ByViteTokenId(block.TokenId.String()); tokenAddress != "" {

	fmt.Printf("Sign send to: %s amount:%s,token:%s \n", block.ToErc20Address, block.Amount, tokenAddress)

	dataBytes := BuildERC20TransferData(block.Amount, block.ToErc20Address, 0)
	//dataBytes := []byte(data)

	launchLog := &models.LaunchLog{
		ItemType:          "vgwWithdraw",
		ItemID:            transaction.ID,
		Status:            "created",
		From:              v.getPublic().String(),
		To:                tokenAddress,
		Value:             decimal.Zero,
		GasLimit:          int64(1 * 250000),
		Data:              utils.Bytes2HexP(dataBytes),
		Amount:            block.Amount,
		TokenAddress:      tokenAddress,
		ViteHash:          block.Hash.String(),
		ViteWalletAddress: block.FromAddress.String(),
		ViteTokenID:       block.TokenId.String(),
		CreatedAt:         time.Now().UTC(),
		UpdatedAt:         time.Now().UTC(),
	}

	err = models.LaunchLogDao.InsertLaunchLog(launchLog)

	if err != nil {
		fmt.Printf(err.Error())
	}

	logger.WithFields(logger.Fields{
		"amount": block.Amount,
		"token":  tokenAddress,
		"to":     block.ToErc20Address,
	}).Info("Withdraw insert Sign order")
	//}

}

// 根据代币的 decimal 得出乘上 10^decimal 后的值
// value 是包含浮点数的。例如 0.5 个 ETH
func GetRealDecimalValue(value string, decimal int) string {
	if strings.Contains(value, ".") {
		// 小数
		arr := strings.Split(value, ".")
		if len(arr) != 2 {
			return ""
		}
		num := len(arr[1])
		left := decimal - num
		return arr[0] + arr[1] + strings.Repeat("0", left)
	} else {
		// 整数
		return value + strings.Repeat("0", decimal)
	}
}

// 构建符合“ERC20”标准的“transfer”合约函数的“data”入参
func BuildERC20TransferData(value, receiver string, decimal int) []byte {

	toAddress := common2.HexToAddress(receiver)

	transferFnSignature := []byte("transfer(address,uint256)")
	methodID := crypto.Keccak256(transferFnSignature)[:4]
	//hash.Write(transferFnSignature)
	//methodID := hash.Sum(nil)[:4]
	fmt.Printf("Method ID: %s\n", hexutil.Encode(methodID))
	//fmt.Printf("fromAddress ID: %s\n", fromAddress)

	paddedAddress := common2.LeftPadBytes(toAddress.Bytes(), 32)
	fmt.Printf("To address: %s\n", hexutil.Encode(paddedAddress))

	amount := new(big.Int)
	amount.SetString(value, 10) // 1000 tokens
	paddedAmount := common2.LeftPadBytes(amount.Bytes(), 32)
	fmt.Printf("Token amount: %s", hexutil.Encode(paddedAmount))

	var data []byte
	data = append(data, methodID...)
	data = append(data, paddedAddress...)
	data = append(data, paddedAmount...)
	return data
}
