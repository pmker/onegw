package launcher

import (
	"crypto/ecdsa"
	"database/sql"
	"github.com/pmker/onegw/keven/nonce"
	"github.com/pmker/onegw/oneplus/backend/sdk/crypto"
	"github.com/pmker/onegw/oneplus/backend/sdk/signer"
	"github.com/pmker/onegw/oneplus/backend/sdk/types"
	"github.com/pmker/onegw/oneplus/backend/utils"
	"strings"
	"sync"
)

type ISignService interface {
	Sign(launchLog *LaunchLog) string
	AfterSign(address string) //what you want to do when signature has been used
}

type LocalSignService struct {
	privateKey *ecdsa.PrivateKey
	nonce      int64
	//cn         *big.Int
	mutex   sync.Mutex
	nm      *nonce.NonceManager
	address string
}

func (s *LocalSignService) AfterSign(address string) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	//s.nonce = s.nonce + 1
	s.nm.PlusNonce(address)
	//
	//no := s.nm.GetNonce(s.address)
	//
	//fmt.Printf("AfterSign no %d", no)

}

func (s *LocalSignService) Sign(launchLog *LaunchLog) string {
	nonce := s.nm.GetNonce(s.address)

	transaction := types.NewTransaction(
		nonce.Uint64(),
		launchLog.To,
		utils.DecimalToBigInt(launchLog.Value),
		uint64(launchLog.GasLimit),
		utils.DecimalToBigInt(launchLog.GasPrice.Decimal),
		utils.Hex2Bytes(launchLog.Data[2:]),
	)

	signedTransaction, err := signer.SignTx(transaction, s.privateKey)

	//fmt.Printf("private:%s ,%s\n", s.privateKey, s.address)
	if err != nil {
		utils.Errorf("sign transaction error: %v", err)
		panic(err)
	}
	//s.nm.PlusNonce(s.address)
	//no:=s.nm.GetNonce(s.address)
	//
	//fmt.Printf("nnnn no %d",no)
	launchLog.Nonce = nonce

	launchLog.Hash = sql.NullString{
		String: utils.Bytes2HexP(signer.Hash(signedTransaction)),
		Valid:  true,
	}

	return utils.Bytes2HexP(signer.EncodeRlp(signedTransaction))
}

func NewDefaultSignService(privateKeyStr string, NM *nonce.NonceManager) ISignService {
	utils.Infof(privateKeyStr)
	privateKey, err := crypto.NewPrivateKeyByHex(privateKeyStr)
	if err != nil {
		panic(err)
	}
	address := strings.ToLower(crypto.PubKey2Address(privateKey.PublicKey))
	chainNonce := NM.GetNonce(address)
	if chainNonce == nil {

		//fmt.Printf("更新 %d", chainNonce)

		chainNonce, _ = NM.GetNonceOk(address)
	}
	///chainNonce,err := NM.GetNonceOk()

	//fmt.Printf("Chain nonce %d \n", chainNonce)
	if err != nil {
		panic(err)
	}

	//if int64(chainNonce) > nonce {
	//	nonce = int64(chainNonce)
	//}

	return &LocalSignService{
		privateKey: privateKey,
		mutex:      sync.Mutex{},
		address:    address,
		//nonce:      int64(chainNonce),
		//cn: chainNonce,
		nm: NM,
	}
}
