package launcher

import (
	"crypto/ecdsa"
	"database/sql"
	"github.com/pmker/onegw/oneplus/backend/sdk/crypto"
	"github.com/pmker/onegw/oneplus/backend/sdk/signer"
	"github.com/pmker/onegw/oneplus/backend/sdk/types"
	"github.com/pmker/onegw/oneplus/backend/utils"
	"sync"
)

type ISignService interface {
	Sign(launchLog *LaunchLog) string
	AfterSign() //what you want to do when signature has been used
}

type localSignService struct {
	privateKey *ecdsa.PrivateKey
	nonce      int64
	mutex      sync.Mutex
}

func (s *localSignService) AfterSign() {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.nonce = s.nonce + 1
}

func (s *localSignService) Sign(launchLog *LaunchLog) string {
	transaction := types.NewTransaction(
		uint64(s.nonce),
		launchLog.To,
		utils.DecimalToBigInt(launchLog.Value),
		uint64(launchLog.GasLimit),
		utils.DecimalToBigInt(launchLog.GasPrice.Decimal),
		utils.Hex2Bytes(launchLog.Data[2:]),
	)

	signedTransaction, err := signer.SignTx(transaction, s.privateKey)

	if err != nil {
		utils.Errorf("sign transaction error: %v", err)
		panic(err)
	}

	launchLog.Nonce = sql.NullInt64{
		Int64: s.nonce,
		Valid: true,
	}

	launchLog.Hash = sql.NullString{
		String: utils.Bytes2HexP(signer.Hash(signedTransaction)),
		Valid:  true,
	}

	return utils.Bytes2HexP(signer.EncodeRlp(signedTransaction))
}

func NewDefaultSignService(privateKeyStr string, getNonce func(string) (int, error)) ISignService {
	utils.Infof(privateKeyStr)
	privateKey, err := crypto.NewPrivateKeyByHex(privateKeyStr)
	if err != nil {
		panic(err)
	}

	//nonce := LaunchLogDao.FindPendingLogWithMaxNonce() + 1
	chainNonce, err := getNonce(crypto.PubKey2Address(privateKey.PublicKey))

	if err != nil {
		panic(err)
	}

	//if int64(chainNonce) > nonce {
	//	nonce = int64(chainNonce)
	//}

	return &localSignService{
		privateKey: privateKey,
		mutex:      sync.Mutex{},
		nonce:      int64(chainNonce),
	}
}
