package pmkoo

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/koinotice/vite/wallet"
	"github.com/nats-io/nats.go"
	"github.com/pmker/onegw/oneplus/common/bgdb"
	"github.com/pmker/onegw/oneplus/hdwallet"
	"github.com/pmker/onegw/oneplus/openssl"
	"github.com/pmker/onegw/pmkoo/models"
	"github.com/pmker/onegw/pmkoo/types"
	log "github.com/sirupsen/logrus"
	"strings"
	"time"
)

type Vgw struct {
	Bowdb  *bgdb.Store
	Nat    *nats.EncodedConn
	Config *types.Config
}

type Token struct {
	//ID       uint64 `bgdb:"key"`
	Address          string `bgdb:"index"`
	Symbol           string
	Name             string
	Decimals         int
	ViteTokenAddress string `bgdb:"unique"`
	Created          time.Time
}

type Wallet struct {
	Address     string `bgdb:"index" json:"Address"`
	Index       uint32 `json:"index"`
	Coin        string `json:"coin"`
	Pk          string `json:"pk"`
	ViteAddress string `josn:"ViteAddress"`
	Created     time.Time
}

func Nats(bowdb *bgdb.Store, nat *nats.EncodedConn, config *types.Config) (*Vgw, error) {
	v := &Vgw{

		Bowdb: bowdb,

		Nat:    nat,
		Config: config,
	}

	return v, nil
}

func (v *Vgw) Start() error {
	log.WithFields(log.Fields{
		"action": "Withdraw",
	}).Info("nats subscribe start")

	if _, err := v.Nat.Subscribe("vgw.>", func(m *nats.Msg) {
		//log.Printf("%s: %s", m.Subject, m.Data)
		switch m.Subject {
		case "vgw.wallet.create":
			v.createUserWallet(m)
			//msg := &WalletMessage{}
			//err := json.Unmarshal(m.Data, msg)
			//
			////mnemonic := "tag volcano eight thank tide danger coast health above argue embrace heavy"
			//if err != nil {
			//	log.Error("create wallet", "error", err)
			//
			//}
			//if msg.Coin == "ETH" {
			//	v.createETHWallet(msg)
			//}
			//if msg.Coin == "VITE" {
			//	v.createVITEWallet(msg)
			//}

		case "vgw.wallet.update":
			v.updateWallet(m)

		case "vgw.token.create":
			v.createToken(m)

		case "vgw.onegw.init":
			v.SysInit(m)

		default:
			log.WithFields(log.Fields{
				"subject": m.Subject,
			}).Info("no match nats")
		}

	});
		err != nil {
		log.WithFields(log.Fields{
			"error": err.Error(),
		}).Error("nats subscribe error")
	}

	return nil
}

//type sysConfig struct {
//	key string
//	value string
//}

//func (v *Vgw) LoadConfigFromDb() {
//
//	if exists := v.bowdb.Cache().Has(types.OPEN_SYS_CONFIG); exists {
//		value, err := v.bowdb.Cache().Get(types.OPEN_SYS_CONFIG)
//
//		data, err := base64.StdEncoding.DecodeString(value)
//		if err != nil {
//			log.WithFields(log.Fields{
//				"error": err.Error(),
//			}).Warn("json.Unmarshal")
//		}
//		key, err := v.bowdb.Cache().Get(types.OPEN_SYS_CONFIG_KEY)
//
//		dst, err := openssl.AesECBDecrypt(data, []byte(key), openssl.PKCS7_PADDING)
//		fmt.Print(string(dst)) // 123456
//		if err != nil {
//			fmt.Printf(err.Error())
//		}
//		config := &types.Config{}
//		err = json.Unmarshal(dst, config)
//		if err != nil {
//			log.WithFields(log.Fields{
//				"error": err.Error(),
//			}).Warn("json.Unmarshal")
//		}
//		fmt.Printf("adfadf config%+v", config)
//	}
//
//}

func (v *Vgw) SysInit(m *nats.Msg) {
	config := &types.Config{}
	err := json.Unmarshal(m.Data, config)
	if err != nil {
		log.WithFields(log.Fields{
			"error": err.Error(),
		}).Warn("Unmarshal json")
	}
	data := m.Data
	key := []byte(config.OneKey)

	//fmt.Printf("pre fron config%+v",config)

	//
	dst, err := openssl.AesECBEncrypt(data, key, openssl.PKCS7_PADDING)
	if err != nil {
		log.WithFields(log.Fields{
			"error": err.Error(),
		}).Warn("openssl.AesECBEncrypt")
	}

	fmt.Printf("types%s, config%s\n", types.OPEN_SYS_CONFIG_KEY, config.OneKey)
	err = v.Bowdb.Cache().Set(types.OPEN_SYS_CONFIG_KEY, config.OneKey, 0)
	if err != nil {
		log.WithFields(log.Fields{
			"error": err.Error(),
		}).Warn("set key")
	}
	err = v.Bowdb.Cache().Set(types.OPEN_SYS_CONFIG, base64.StdEncoding.EncodeToString(dst), 0)
	if err != nil {
		log.WithFields(log.Fields{
			"error": err.Error(),
		}).Warn("set config")
	}
	msg := &WalletMessage{
		Coin:   "ETH",
		Number: 2,
	}
	v.createETHWallet(msg)
	msg = &WalletMessage{
		Coin:   "VITE",
		Number: 2,
	}
	v.createVITEWallet(msg)

	fmt.Printf("init ok")
	//
	// fmt.Printf("%s\n",base64.StdEncoding.EncodeToString(dst))  // yXVUkR45PFz0UfpbDB8/ew==
	// fmt.Printf("%s\n",(dst))  // yXVUkR45PFz0UfpbDB8/ew==

	///v.LoadConfigFromDb()

	//
	//if err!=nil{
	//	fmt.Printf(err.Error())
	//}
	//dst , err = openssl.AesECBDecrypt(dst, key, openssl.PKCS7_PADDING)
	//fmt.Print(string(dst)) // 123456
	//if err!=nil{
	//	fmt.Printf(err.Error())
	//}
	//
	//cfg:=&types.Config{}
	//err = json.Unmarshal(dst, cfg)
	//if err != nil {
	//	log.WithFields(log.Fields{
	//		"error": err.Error(),
	//	}).Warn("json.Unmarshal")
	//}
	//fmt.Printf("adfadf config%+v",cfg)

	//err1 := v.bowdb.Insert(wallet.Address, wallet)
	//if err1 != nil {
	//	log.WithFields(log.Fields{
	//		"error": err1.Error(),
	//	}).Warn("insert wallet warn")
	//} else {
	//	walletData := &models.Wallet{
	//		EthAddress: wallet.Address,
	//		//UpdatedAt:  time.Now().UTC(),
	//		//CreatedAt:  time.Now().UTC(),
	//	}
	//	err2 := models.WalletDaoPG.InsertWallet(walletData)
	//
	//	if err2 != nil {
	//		fmt.Printf("insert wallet err %s", err2.Error())
	//	}
	//	v.bowdb.Cache().Set(wallet.Address, wallet.ViteAddress, 0)
	//}

}

type WalletMessage struct {
	Coin   string
	Number uint32
}

func (v *Vgw) createUserWallet(m *nats.Msg) {

	wallet := &Wallet{}
	err := json.Unmarshal(m.Data, wallet)
	wallet.Created = time.Now()
	if err != nil {
		log.WithFields(log.Fields{
			"error": err.Error(),
		}).Warn("create wallet json")
	}

	err1 := v.Bowdb.Insert(wallet.Address, wallet)
	if err1 != nil {
		log.WithFields(log.Fields{
			"error": err1.Error(),
		}).Warn("insert wallet warn")
	} else {
		walletData := &models.Wallet{
			EthAddress: wallet.Address,
		}
		err2 := models.WalletDaoPG.InsertWallet(walletData)

		if err2 != nil {
			fmt.Printf("insert wallet err %s", err2.Error())
		}
		v.Bowdb.Cache().Set(wallet.Address, "1", 0)
	}

	repo := []Wallet{}
	err2 := v.Bowdb.Find(&repo, &bgdb.Query{})
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

func (v *Vgw) createETHWallet(msg *WalletMessage) {
	//fmt.Printf("fucksss config %+v \n",v.config)

	//for i, to := range repo {
	//	fmt.Printf("index %d,%+v \n", i, to)
	//}

	//wallet := &Wallet{}
	wallet, err := hdwallet.NewFromMnemonic(v.Config.Mnemonic)
	if err != nil {
		fmt.Printf(err.Error())
	}
	fmt.Print("\n----------------------EthWallet2----------------------\n")

	for i := uint32(0); i < msg.Number; i++ {

		path := hdwallet.MustParseDerivationPath(fmt.Sprintf(v.Config.Epath, i))

		account, err := wallet.Derive(path, false)
		if err != nil {
			log.Error("create account", "error", err)

		}

		privateKey, err := wallet.PrivateKeyHex(account)
		if err != nil {
			log.Error("create account privateKey", "error", err)
		}

		fmt.Printf("Private key in hex: %s ,Address:%s\n", privateKey, account.Address.Hex())

		qb := &Wallet{}
		qb.Created = time.Now()
		qb.Address = strings.ToLower(account.Address.Hex())
		qb.Index = i
		qb.Coin = "ETH"
		qb.Pk=privateKey

		err1 := v.Bowdb.Insert(qb.Address, qb)
		if err1 != nil {
			log.WithFields(log.Fields{
				"error": err1.Error(),
			}).Error("create eth wallet error")
		} else {
			//walletData := &models.Wallet{
			//	EthAddress: qb.Address,
			//	//UpdatedAt:  time.Now().UTC(),
			//	//CreatedAt:  time.Now().UTC(),
			//}
			//err2 := models.WalletDaoPG.InsertWallet(walletData)
			//
			//if err2 != nil {
			//	fmt.Printf("insert wallet err %s", err2.Error())
			//}
			//v.bowdb.Cache().Set(qb.Address, wallet.ViteAddress, 0)
		}

	}
	repo := []Wallet{}
	err = v.Bowdb.Find(&repo, bgdb.Where("Coin").Eq(msg.Coin).SortBy("Index").Reverse())
	if err != nil {
		fmt.Printf(err.Error())
	}

	if err != nil {
		log.WithFields(log.Fields{
			"error": err.Error(),
		}).Info("wallet find")
	}
	fmt.Printf("repo size  %d", uint32(len(repo)))
	fmt.Print("\n----------------------EthWallet2----------------------\n")

}

func (v *Vgw) createVITEWallet(msg *WalletMessage) {

	w := wallet.New(&wallet.Config{
		DataDir:        VedexWalletDir(),
		MaxSearchIndex: 100000,
	})
	w.Start()

	//fmt.Printf("config %+v \n", cfg)

	w2, err := w.RecoverEntropyStoreFromMnemonic(v.Config.Mnemonic, v.Config.OneKey)
	if err != nil {
		fmt.Errorf("wallet error, %+v", err)
		//return nil, err
	}
	err = w2.Unlock(v.Config.OneKey)
	if err != nil {

		fmt.Errorf("wallet error, %+v", err)
		//return nil, err
	}

	fmt.Print("\n----------------------ViteWallet2----------------------\n")
	for i := uint32(0); i < msg.Number; i++ {
		// _, key, err := w2.DeriveForIndexPath(i)
		_, key, err := w2.DeriveForFullPath(fmt.Sprintf(v.Config.Vpath, i))
		if err != nil {

		}
		addr, err := key.Address()
		if err != nil {
			fmt.Printf(err.Error())
		}
		keys, err := key.PrivateKey()
		if err != nil {
			fmt.Print(err)
		}

		fmt.Printf("address:%s,private:%s\n", addr, keys.Hex())
	}
	fmt.Print("\n----------------------ViteWallet2----------------------\n")

}

func (v *Vgw) updateWallet(m *nats.Msg) {

	wallet := &Wallet{}
	err := json.Unmarshal(m.Data, wallet)
	wallet.Created = time.Now()
	if err != nil {
		log.WithFields(log.Fields{
			"error": err.Error(),
		}).Warn("create wallet json")
	}
	err = v.Bowdb.UpdateMatching(&Wallet{}, bgdb.Where("Address").Eq(wallet.Address), func(record interface{}) error {

		update, ok := record.(*Wallet) // record will always be a pointer
		if !ok {
			return fmt.Errorf("Record isn't the correct type!  Wanted Person, got %T", record)
		}

		update.ViteAddress = wallet.ViteAddress
		update.Created = wallet.Created

		log.WithFields(log.Fields{
			"ethAddress":  wallet.Address,
			"viteAddress": wallet.ViteAddress,
		}).Info("wallet update success")

		v.Bowdb.Cache().Set(wallet.Address, wallet.ViteAddress, 0)

		return nil
	})
	if err != nil {
		fmt.Print("update wallet %s", err.Error())
	}

	repo := []Wallet{}
	err2 := v.Bowdb.Find(&repo, &bgdb.Query{})
	//err2 := store.Find(&repo, bgdb.Where("Symbol").Eq("kev"))
	//err2 := store.Find(&repo, badgerhold.)
	if err2 != nil {
		log.WithFields(log.Fields{
			"error": err2.Error(),
		}).Info("wallet find")
	}
	for i, to := range repo {
		fmt.Printf("index %d,ethAddress:%s,viteAddress:%s \n", i, to.Address, to.ViteAddress)
	}
}

func (v *Vgw) createToken(m *nats.Msg) {

	token := &Token{}
	err := json.Unmarshal(m.Data, token)
	token.Created = time.Now()
	if err != nil {
		log.WithFields(log.Fields{
			"error": err.Error(),
		}).Warn("create token json")
	}

	err1 := v.Bowdb.Insert(token.Address, token)
	if err1 != nil {
		log.WithFields(log.Fields{
			"error": err1.Error(),
		}).Warn("insert token warn")
	} else {
		//
		//tokenData := &models.Token{
		//	TokenId:     token.ViteTokenAddress,
		//	TokenSymbol: token.Symbol,
		//	Type:        1,
		//	//UpdatedAt:  time.Now().UTC(),
		//	//CreatedAt:  time.Now().UTC(),
		//}
		//err2 := models.TokenDaoPG.InsertToken(tokenData)

		//if err2 != nil {
		//	fmt.Printf("insert wallet err %s", err2.Error())
		//}
		v.Bowdb.Cache().Set(token.Address, token.ViteTokenAddress, 0)
	}

	repo := []Token{}
	err2 := v.Bowdb.Find(&repo, &bgdb.Query{})
	//err2 := store.Find(&repo, bgdb.Where("Symbol").Eq("kev"))
	//err2 := store.Find(&repo, badgerhold.)
	if err2 != nil {
		log.WithFields(log.Fields{
			"error": err2.Error(),
		}).Info("token find")
	}
	for i, to := range repo {
		fmt.Printf("Token list index %d,%v \n", i, to)
	}
}
