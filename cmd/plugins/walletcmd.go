package plugins

import (
	"encoding/json"
	"fmt"
	"github.com/pmker/onegw/oneplus/common/bgdb"
	"github.com/pmker/onegw/oneplus/hdwallet"
	"github.com/pmker/onegw/pmkoo/types"
	"io/ioutil"
	"strings"
	"time"

	"github.com/pmker/onegw/pmkoo"

	"github.com/koinotice/vite/cmd/utils"
	"gopkg.in/urfave/cli.v1"
)

var (
	walletFlags = utils.MergeFlags(configFlags, generalFlags)

	//demo,please add this `demoCommand` to one.go
	/**
	app.Commands = []cli.Command{
		versionCommand,
		licenseCommand,
		consoleCommand,
		attachCommand,
		demoCommand,
	}
	*/
	walletCommand = cli.Command{
		Action:      utils.MigrateFlags(walletAction),
		Name:        "wallet",
		Usage:       "eth hd wallet create",
		Flags:       walletFlags,
		Category:    "WALLET COMMANDS",
		Description: `wallet create`,
	}
)
func config() bgdb.Options {
	opt := bgdb.DefaultOptions
	opt.Dir = "./bgdb"
	opt.ValueDir = opt.Dir
	return opt
}
// localConsole starts chain new gvite node, attaching chain JavaScript console to it at the same time.
func walletAction(ctx *cli.Context) error {

	//// Create and start the node based on the CLI flags
	//nodeManager, err := nodemanager.NewSubCmdNodeManager(ctx, nodemanager.FullNodeMaker{})
	//if err != nil {
	//	fmt.Println("demo error", err)
	//	return err
	//}
	//nodeManager.Start()
	//defer nodeManager.Stop()

	//msg := &pmkoo.WalletMessage{
	//	Coin:   "ETH",
	//	Number: 100,
	//}
	//nats.Publish("vgw.wallet.create", msg)
	//
	//msg1 := &pmkoo.WalletMessage{
	//	Coin:   "VITE",
	//	Number: 100,
	//}
	//nats.Publish("vgw.wallet.create", msg1)
	////Tips: add your code here
	//fmt.Println("msg send")
	//
	go createUserEthWallet()
	//go ViteWallet(cfg)
	select {

	}
	return nil
}

func createUserEthWallet() {

	nats := Nats()

	opt := config()
	bowdb, err := bgdb.Open(opt)
	if err != nil {
		fmt.Printf("err %s", err.Error())
	}


	config := &types.Config{}
	if jsonConf, err := ioutil.ReadFile(defaultNodeConfigFileName); err == nil {

		err = json.Unmarshal(jsonConf, &config)
		if err != nil {
			log.Error("Cannot unmarshal the default config file content", "error", err)

		}

	}

	repo := []pmkoo.Wallet{}
	Coin:="USERETH"
	err = bowdb.Find(&repo, bgdb.Where("Coin").Eq(Coin).SortBy("Index").Reverse())
	if err != nil {
		fmt.Printf(err.Error())
	}

	if err != nil {
		fmt.Printf(err.Error())

	}
	fmt.Printf("repo size  %d", uint32(len(repo)))
	wallet, err := hdwallet.NewFromMnemonic(config.Mnemonic)
	if err != nil {
		fmt.Printf(err.Error())
	}
	fmt.Print("\n----------------------EthWallet2----------------------\n")

	for i := uint32(len(repo)); i <uint32(len(repo))+ 1; i++ {

		path := hdwallet.MustParseDerivationPath(fmt.Sprintf(config.Epath, i))

		account, err := wallet.Derive(path, false)
		if err != nil {
			log.Error("create account", "error", err)

		}

		privateKey, err := wallet.PrivateKeyHex(account)
		if err != nil {
			log.Error("create account privateKey", "error", err)
		}

		 fmt.Printf("Private key in hex: %s ,Address:%s\n", privateKey, account.Address.Hex())

		qb := &pmkoo.Wallet{}
		qb.Created = time.Now()
		qb.Address = strings.ToLower(account.Address.Hex())
		qb.Index = i
		qb.Coin = Coin

		err1 := bowdb.Insert(qb.Address, qb)
		if err1 != nil {
			fmt.Printf(err.Error())
		} else {


			if err := nats.Publish("vgw.wallet.create", qb); err != nil {
				fmt.Printf("asdfasdf err %s", err.Error())
			}
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

	for i,o:=range repo {
		fmt.Printf("index %d,%+v\n",i,o)


	}
	fmt.Print("\n----------------------EthWallet2----------------------\n")

}
func ethWallet(cfg *types.Config) {

	nats := Nats()

	//mnemonic := "tag volcano eight thank tide danger coast health above argue embrace heavy"
	wallet, err := hdwallet.NewFromMnemonic(cfg.Mnemonic)
	if err != nil {
		log.Error("create wallet", "error", err)

	}
	fmt.Print("----------------------EthWallet2----------------------\n")

	for i := uint32(0); i < 10; i++ {

		path := hdwallet.MustParseDerivationPath(fmt.Sprintf(cfg.Epath, i))

		account, err := wallet.Derive(path, false)
		if err != nil {
			log.Error("create account", "error", err)

		}

		privateKey, err := wallet.PrivateKeyHex(account)
		if err != nil {
			log.Error("create account privateKey", "error", err)
		}

		fmt.Printf("Private key in hex: %s ,Address:%s\n", privateKey, account.Address.Hex())

		qb := &pmkoo.Wallet{

		}
		qb.Created = time.Now()
		qb.Address = strings.ToLower(account.Address.Hex())
		qb.Index = i
		qb.Coin = "USERETH"
		if err := nats.Publish("vgw.token.create", qb); err != nil {
			fmt.Printf("asdfasdf err %s", err.Error())
		}
		//publicKey, _ := wallet.PublicKeyHex(account)
		//if err != nil {
		//	log.Error("create account publicKey", "error", err)
		//}
		//
		//fmt.Println(account.Address.Hex()) // 0xC49926C4124cEe1cbA0Ea94Ea31a6c12318df947
		//fmt.Println(publicKey)             // 0xC49926C4124cEe1cbA0Ea94Ea31a6c12318df947
	}
	fmt.Print("----------------------EthWallet2----------------------\n")

	//	fmt.Printf("config %+v \n", cfg)
	//
	//path := hdwallet.MustParseDerivationPath(cfg.Epath)

}

//
//func ViteWallet(cfg *types.Config) (*entropystore.Manager, error) {
//
//	nats := Nats()
//	w := wallet.New(&wallet.Config{
//		DataDir:        pmkoo.VedexWalletDir(),
//		MaxSearchIndex: 100000,
//	})
//	w.Start()
//
//	//fmt.Printf("config %+v \n", cfg)
//
//	w2, err := w.RecoverEntropyStoreFromMnemonic(cfg.Mnemonic, cfg.OneKey)
//	if err != nil {
//		fmt.Errorf("wallet error, %+v", err)
//		//return nil, err
//	}
//	err = w2.Unlock(cfg.OneKey)
//	if err != nil {
//
//		fmt.Errorf("wallet error, %+v", err)
//		//return nil, err
//	}
//
//	fmt.Print("----------------------ViteWallet2----------------------\n")
//	for i := uint32(0); i < 10; i++ {
//		// _, key, err := w2.DeriveForIndexPath(i)
//		_, key, err := w2.DeriveForFullPath(fmt.Sprintf(cfg.Vpath, i))
//		if err != nil {
//
//		}
//		addr, err := key.Address()
//		if err != nil {
//			fmt.Printf(err.Error())
//		}
//		keys, err := key.PrivateKey()
//		if err != nil {
//			fmt.Print(err)
//		}
//
//		fmt.Printf("address:%s,private:%s\n", addr, keys.Hex())
//	}
//	fmt.Print("----------------------ViteWallet2----------------------")
//	return nil, nil
//}
