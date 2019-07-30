package tool

import (
	"database/sql"
	"encoding/json"
	"fmt"
	common2 "github.com/ethereum/go-ethereum/common"
	_ "github.com/joho/godotenv/autoload"
	"github.com/nats-io/nats.go"
	"github.com/pmker/onegw/keven/launcher"
	"github.com/pmker/onegw/keven/nonce"
	"github.com/pmker/onegw/oneplus/backend/common"
	"github.com/pmker/onegw/oneplus/backend/sdk/crypto"
	"github.com/pmker/onegw/oneplus/backend/sdk/ethereum"
	"github.com/pmker/onegw/oneplus/backend/utils"
	"github.com/pmker/onegw/oneplus/common/bgdb"
	"github.com/pmker/onegw/oneplus/common/blero"
	"github.com/shopspring/decimal"
	log "github.com/sirupsen/logrus"
	"math/big"
	"strings"
	"sync"
	"time"
)

type Wallet struct {
	Address     string `bgdb:"index" json:"Address"`
	Pk          string `json:"pk"`
	ViteAddress string `josn:"ViteAddress"`
	Created     time.Time
}
type Balance struct {
	Address string `bgdb:"index" json:"Address"`
	Token   string `bgdb:"index" json:"Token"`
	Dice    int    `  json:"dice"`
	Amount  int    `  json:"amount"`
}

type TxOrder struct {
	ID                int64
	ItemType          string
	ItemID            int64
	Status            string
	Hash              sql.NullString
	BlockNumber       sql.NullInt64
	From              string
	To                string
	Value             decimal.Decimal
	GasLimit          int64
	GasUsed           sql.NullInt64
	GasPrice          decimal.NullDecimal
	Nonce             sql.NullInt64
	Data              string
	Amount            string
	TokenAddress      string
	ViteHash          string
	ViteWalletAddress string
	ViteTokenID       string
	ExecutedAt        time.Time
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

type Pmker struct {
	Db           *bgdb.Store
	Nat          *nats.EncodedConn
	Bot          *launcher.Launcher
	Hydro        *ethereum.EthereumHydro
	NonceManager *nonce.NonceManager // nonce 管理者实例
	Queue        *blero.Blero
	Lock         sync.Mutex
	Lans         map[string]*launcher.ISignService
	EventQueue   common.IQueue
}

type Message struct {
	Type         string
	TokenAddress string
	Prinkey      string
	Value        string
	ToAddress    string
}

func (p *Pmker) SendEth(message *Message) error {

	p.Lock.Lock()         // 加锁
	defer p.Lock.Unlock() // 当该函数执行完毕，进行解锁

	privateKey, err := crypto.NewPrivateKeyByHex(message.Prinkey)

	address := strings.ToLower(crypto.PubKey2Address(privateKey.PublicKey))

	if _, ok := p.Lans[address]; ok {
		//存在
		//p.Bot.SignService=p.Lans[address]
		//fmt.Printf("\n我在:%s \n", address)
		p.Bot.SignService = *p.Lans[address]

	} else {
		//fmt.Printf("我不在:%s \n", address)
		sign := launcher.NewDefaultSignService(message.Prinkey, p.NonceManager)
		p.Lans[address] = &sign
		p.Bot.SignService = sign

	}

	//fmt.Printf("Message :%+v\n",message)
	//	signService := launcher.NewDefaultSignService(message.Prinkey, p.NonceManager)
	//p.Bot.SignService=&p.Lans[address]

	d2 := GetRealDecimalValue(message.Value, 12) //
	amount, err := decimal.NewFromString(d2)
	if err != nil {

	}
	from := address
	data := &launcher.LaunchLog{
		//Nonce:p.NonceManager.(from)
		ItemType:  "SendEth",
		Status:    "created",
		From:      from,
		To:        message.ToAddress,
		Value:     amount,
		GasLimit:  int64(1 * 210000),
		Data:      "0x",
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}
	///fmt.Printf("from address:%s,%+v",from,data)

	err = p.Bot.Add(data)
 	if err != nil {
		log.WithFields(log.Fields{
			"error": err.Error(),
			"to":    message.ToAddress,
			"value": amount,
		}).Info("send eth ")
	}
	return err
}

//发送erc20
func (p *Pmker) SendErc20(msg []byte) error {

	p.Lock.Lock()         // 加锁
	defer p.Lock.Unlock() // 当该函数执行完毕，进行解锁
	var message Message //无需实例化，Unmarshal内部实现了
	err := json.Unmarshal(msg, &message)
	if err != nil {
		log.WithFields(log.Fields{
			"error": err.Error(),

		}).Error("SendErc20 json error ")
		return err
	}


	privateKey, err := crypto.NewPrivateKeyByHex(message.Prinkey)
	if err!=nil{
		log.WithFields(log.Fields{
			"error": err.Error(),
		}).Error("SendErc20 Prinkey error ")
		return err

	}
	address := strings.ToLower(crypto.PubKey2Address(privateKey.PublicKey))

	if _, ok := p.Lans[address]; ok {
		//存在
		//p.Bot.SignService=p.Lans[address]
		//fmt.Printf("\n我在:%s \n", address)
		p.Bot.SignService = *p.Lans[address]

	} else {
		//fmt.Printf("我不在:%s \n", address)
		sign := launcher.NewDefaultSignService(message.Prinkey, p.NonceManager)
		p.Lans[address] = &sign
		p.Bot.SignService = sign

	}

	//fmt.Printf("Message :%+v\n",message)
	//	signService := launcher.NewDefaultSignService(message.Prinkey, p.NonceManager)
	//p.Bot.SignService=&p.Lans[address]

	d2 := GetRealDecimalValue(message.Value, 0) //

	//d2 := GetRealDecimalValue("100", 0) //

	dataBytes := BuildERC20TransferData(d2, message.ToAddress, 0)

	from := address
	data := &launcher.LaunchLog{
		//Nonce:p.NonceManager.(from)
		ItemType:  "SendErc20",
		Status:    "created",
		From:      from,
		To:        message.TokenAddress,
		Value:     decimal.Zero,
		GasLimit:  int64(1 * 210000),
		Data:      utils.Bytes2HexP(dataBytes),
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}
	err = p.Bot.Add(data)

	if err != nil {
		log.WithFields(log.Fields{
			"error": err.Error(),
			"to":    message.ToAddress,
			"token": message.TokenAddress,
			"value": d2,
		}).Error("send erc20 ")
		//p.SendErc20(message)
		p.EventQueue.Push(msg)

	}
	return err
}

func (p *Pmker) SendEthAll(m *nats.Msg) {
	message := &Message{}
	err := json.Unmarshal(m.Data, message)
	if err != nil {

	}
	repo := []Wallet{}
	err2 := p.Db.Find(&repo, &bgdb.Query{})
	//err2 := store.Find(&repo, bgdb.Where("Symbol").Eq("kev"))
	//err2 := store.Find(&repo, badgerhold.)
	if err2 != nil {
		log.WithFields(log.Fields{
			"error": err2.Error(),
		}).Info("wallet find")
	}

	for _, to := range repo {
		message.ToAddress = to.Address
		go p.SendEth(message)

	}

}


// 构建符合“ERC20”标准的“transfer”合约函数的“data”入参
func BuildERC20TransferData(value, receiver string, decimal int) []byte {

	toAddress := common2.HexToAddress(receiver)

	transferFnSignature := []byte("transfer(address,uint256)")
	methodID := crypto.Keccak256(transferFnSignature)[:4]
	//hash.Write(transferFnSignature)
	//methodID := hash.Sum(nil)[:4]
	//fmt.Printf("Method ID: %s\n", hexutil.Encode(methodID))
	//fmt.Printf("fromAddress ID: %s\n", fromAddress)

	paddedAddress := common2.LeftPadBytes(toAddress.Bytes(), 32)
	//fmt.Printf("To address: %s\n", hexutil.Encode(paddedAddress))

	amount := new(big.Int)
	amount.SetString(value, 10) // 1000 tokens
	paddedAmount := common2.LeftPadBytes(amount.Bytes(), 32)
	//fmt.Printf("Token amount: %s \n", hexutil.Encode(paddedAmount))

	var data []byte
	data = append(data, methodID...)
	data = append(data, paddedAddress...)
	data = append(data, paddedAmount...)
	return data
}

func (v *Pmker) createWallet(m *nats.Msg) {

	wallet := &Wallet{}
	err := json.Unmarshal(m.Data, wallet)
	wallet.Created = time.Now()
	if err != nil {
		log.WithFields(log.Fields{
			"error": err.Error(),
		}).Warn("create wallet json")
	}

	err1 := v.Db.Insert(wallet.Address, wallet)
	if err1 != nil {
		log.WithFields(log.Fields{
			"error": err1.Error(),
		}).Warn("insert wallet warn")
	} else {

	}

	repo := []Wallet{}
	err2 := v.Db.Find(&repo, &bgdb.Query{})
	//err2 := store.Find(&repo, bgdb.Where("Symbol").Eq("kev"))
	//err2 := store.Find(&repo, badgerhold.)
	if err2 != nil {
		log.WithFields(log.Fields{
			"error": err2.Error(),
		}).Info("wallet find")
	}
	for i, to := range repo {
		fmt.Printf("index %d,%v \n", i, to)
	}
}
func (p *Pmker) Start() {

	if _, err := p.Nat.Subscribe("pmker.>", func(m *nats.Msg) {
		//log.Printf("%s: %s", m.Subject, m.Data)
		switch m.Subject {
		case "pmker.wallet.create":

			p.createWallet(m)
		case "pmker.send.eth":
			p.EventQueue.Push(m.Data)
			//p.Queue.EnqueueJob("pmker.send.eth", m.Data)
			//message := &Message{}
			//err := json.Unmarshal(m.Data, message)
			//if err != nil {
			//
			//}
			//p.SendEth(message)

		case "pmker.send.ethall":

			p.SendEthAll(m)

		case "pmker.send.erc20":

			p.EventQueue.Push(m.Data)

		default:
			log.WithFields(log.Fields{
				"subject": m.Subject,
			}).Info("no match nats")
		}

	}); err != nil {
		log.WithFields(log.Fields{
			"error": err.Error(),
		}).Error("nats subscribe error")
	}
}
