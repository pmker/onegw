package tool

import (
	"crypto/ecdsa"
	"fmt"
	common2 "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"strings"
)

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
func (v *Pmker) getPublic(privateKeyString string) common2.Address {
	privateKey, err := crypto.HexToECDSA(privateKeyString)
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
