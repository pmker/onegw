package models

//import "time"

type IWalletDao interface {
	GetAllWallets() []*Wallet
	InsertWallet(*Wallet) error
	//FindWalletBySymbol(string) *Wallet
}

type Wallet struct {
	EthAddress string `json:"eth_address"   db:"eth_address" gorm:"primary_key"`
	ViteAddress string `json:"vite_address"     db:"vite_address"`
	//UpdatedAt       time.Time       `json:"updatedAt"       db:"updatedAt"`
	//CreatedAt       time.Time       `json:"createdAt"       db:"createdAt"`
}

func (Wallet) TableName() string {
	return "wallets"
}

var WalletDao IWalletDao
var WalletDaoPG IWalletDao

func init() {
	WalletDao = &walletDaoPG{}
	WalletDaoPG = WalletDao
}

type walletDaoPG struct {
}

func (walletDaoPG) GetAllWallets() []*Wallet {
	var wallets []*Wallet
	DB.Find(&wallets)
	return wallets
}

func (walletDaoPG) InsertWallet(wallet *Wallet) error {
	return DB.Create(wallet).Error
}

//
//func (walletDaoPG) FindWalletBySymbol(symbol string) *Wallet {
//	var wallet Wallet
//
//	DB.Where("symbol = ?", symbol).Find(&wallet)
//	if wallet.Symbol == "" {
//		return nil
//	}
//
//	return &wallet
//}
//
//func GetBaseWalletSymbol(marketID string) string {
//	splits := strings.Split(marketID, "-")
//
//	if len(splits) != 2 {
//		return ""
//	} else {
//		return splits[0]
//	}
//}
//
//func GetBaseWalletDecimals(marketID string) int {
//	walletSymbol := GetBaseWalletSymbol(marketID)
//
//	wallet := WalletDao.FindWalletBySymbol(walletSymbol)
//	if wallet == nil {
//		panic("invalid base wallet, symbol:" + walletSymbol)
//	}
//
//	return wallet.Decimals
//}
