package plugin

import (
	"github.com/onrik/ethrpc"
	"github.com/pmker/onegw/oneplus/backend/sdk"
	"github.com/pmker/onegw/oneplus/backend/sdk/ethereum"
	"github.com/pmker/onegw/oneplus/watch/structs"
	"math/big"

	//"github.com/pmker/onegw/oneplus/watch/utils"
	"github.com/shopspring/decimal"
)

type ITxReceiptPlugin interface {
	Accept(tx *structs.RemovableTxAndReceipt)
}

type TxReceiptPluginWithFilter struct {
	ITxReceiptPlugin
	filterFunc func(transaction sdk.Transaction) bool
}

func (p TxReceiptPluginWithFilter) NeedReceipt(tx sdk.Transaction) bool {
	return p.filterFunc(tx)
}

func NewTxReceiptPluginWithFilter(
	callback func(tx *structs.RemovableTxAndReceipt),
	filterFunc func(transaction sdk.Transaction) bool) *TxReceiptPluginWithFilter {

	p := NewTxReceiptPlugin(callback)
	return &TxReceiptPluginWithFilter{p, filterFunc}
}

type TxReceiptPlugin struct {
	callback func(tx *structs.RemovableTxAndReceipt)
}

func NewTxReceiptPlugin(callback func(tx *structs.RemovableTxAndReceipt)) *TxReceiptPlugin {
	return &TxReceiptPlugin{callback}
}

func (p TxReceiptPlugin) Accept(tx *structs.RemovableTxAndReceipt) {
	if p.callback != nil {
		p.callback(tx)
	}
}

type ERC20TransferPlugin struct {
	callback func(tokenAddress, from, to string, amount decimal.Decimal, isRemoved bool,log ethrpc.Log)
}

func NewERC20TransferPlugin(callback func(tokenAddress, from, to string, amount decimal.Decimal, isRemoved bool,log ethrpc.Log)) *ERC20TransferPlugin {
	return &ERC20TransferPlugin{callback}
}

func (p *ERC20TransferPlugin) Accept(tx *structs.RemovableTxAndReceipt) {
	if p.callback != nil {
		events := extractERC20TransfersIfExist(tx)

		for _, e := range events {
			p.callback(e.token, e.from, e.to, e.value, tx.IsRemoved, e.log)
		}
	}
}

type TransferEvent struct {
	token string
	from  string
	to    string
	value decimal.Decimal
	log  ethrpc.Log
}

func extractERC20TransfersIfExist(r *structs.RemovableTxAndReceipt) (rst []TransferEvent) {
	transferEventSig := "0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef"

	// todo a little weird
	if receipt, ok := r.Receipt.(*ethereum.EthereumTransactionReceipt); ok {
		for _, log := range receipt.Logs {
			if len(log.Topics) == 3 && log.Topics[0] == transferEventSig {
				from := log.Topics[1]
				to := log.Topics[2]
				//fmt.Printf("is log %v",log)

				amount, ok := HexToDecimal(log.Data)

				if ok {
					rst = append(rst, TransferEvent{log.Address, from, to, amount,log})
				}
			}
		}
	}

	return
}

func HexToDecimal(hex string) (decimal.Decimal, bool) {
	if hex[0:2] == "0x" || hex[0:2] == "0X" {
		hex = hex[2:]
	}

	b := new(big.Int)
	b, ok := b.SetString(hex, 16)
	if !ok {
		return decimal.Zero, false
	}

	return decimal.NewFromBigInt(b, 0), true
}
