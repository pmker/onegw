package pmkoo

import (
	"fmt"
	"github.com/koinotice/vite/wallet"
	"github.com/koinotice/vite/wallet/entropystore"
	"github.com/koinotice/vite/wallet/hd-bip/derivation"
	"github.com/pmker/onegw/pmkoo/types"
	"os"
	"os/user"
	"path/filepath"
)
func HomeDir() string {
	if home := os.Getenv("HOME"); home != "" {
		return home
	}
	if usr, err := user.Current(); err == nil {
		return usr.HomeDir
	}
	return ""
}
func VedexWalletDir() string {
	home := HomeDir()
	if home != "" {
		return filepath.Join(home, "onegw", "wallet")
	}
	return ""
}


//生成钱包
func  WalletCreate(password string) {
	manager := wallet.New(&wallet.Config{
		DataDir: VedexWalletDir(),
	})
	mnemonic, em, err := manager.NewMnemonicAndEntropyStore(password)
	if err != nil {
		fmt.Print(err)
	}
	fmt.Println(mnemonic, ",", em.GetEntropyStoreFile())
	//
	//manager.Start()
	//files := manager.ListAllEntropyFiles()
	//for _, v := range files {
	//	storeManager, err := manager.GetEntropyStoreManager(v)
	//	if err != nil {
	//		fmt.Print(err)
	//	}
	//	storeManager.Unlock("123456")
	//	_, key, err := storeManager.DeriveForIndexPath(0)
	//
	//	fmt.Printf("key %v \n",key)
	//	if err != nil {
	//		fmt.Print(err)
	//	}
	//	keys, err := key.PrivateKey()
	//	if err != nil {
	//		fmt.Print(err)
	//	}
	//	addr, err := key.Address()
	//	if err != nil {
	//		fmt.PrintWatchOrder(err)
	//	}
	//	fmt.Printf("%s,0:%s,%s\n", addr, addr, keys.Hex())
	//}
}

func  UnlockViteWallet(config *types.Config)  (*entropystore.Manager ,error) {
	w := wallet.New(&wallet.Config{
		DataDir:        VedexWalletDir(),
		MaxSearchIndex: 100000,
	})
	w.Start()

	w2, err := w.RecoverEntropyStoreFromMnemonic(config.Mnemonic, config.OneKey)
	if err != nil {
		fmt.Errorf("wallet error, %+v", err)
		return nil,err
	}
	err = w2.Unlock(config.OneKey)
	if err != nil {

		fmt.Errorf("wallet error, %+v", err)
		return  nil ,err
	}

	//fmt.Print("----------------------Wallet2----------------------\n")
	//for i := uint32(0); i < 10; i++ {
	//	_, key, err := w2.DeriveForIndexPath(i)
	//	if err != nil {
	//
	//	}
	//	addr, _ := key.Address()
	//	fmt.Println(addr)
	//}
	//fmt.Print("----------------------Wallet2----------------------")

	return w2,nil
}

func GetKey(walletAddress string,index int) (key *derivation.Key) {
	manager := wallet.New(&wallet.Config{
		DataDir: VedexWalletDir(),
	})

	manager.Start()

	  //mnemonic, em, err := manager.NewMnemonicAndEntropyStore(os.Getenv("VITE_MNEMONIC_PASSWORD"))
	//
	//fmt.Println(mnemonic)
	//fmt.Println(em.GetPrimaryAddr())
	//fmt.Println(em.GetEntropyStoreFile())

	storeManager, err := manager.GetEntropyStoreManager(walletAddress)
	if err != nil {
		fmt.Print(err)
	}
	storeManager.Unlock(os.Getenv("VITE_MNEMONIC_PASSWORD"))

	_, key, err = storeManager.DeriveForIndexPath(uint32(index))
	if err != nil {
		panic(err)
	}
	return key
	//keys, err := key.PrivateKey()
	//	//if err != nil {
	//	//	panic(err)
	//	//}
	//	//addr, err := key.Address()
	//	//if err != nil {
	//	//	panic(err)
	//	//}
	//	//fmt.Printf("%s,0:%s,%s %d,\n", addr, keys, keys.Hex(), len(keys))

}