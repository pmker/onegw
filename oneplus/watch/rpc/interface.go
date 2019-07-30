package rpc

import "github.com/pmker/onegw/oneplus/backend/sdk"

type IBlockChainRPC interface {
	GetCurrentBlockNum() (uint64, error)

	GetBlockByNum(uint64) (sdk.Block, error)
	GetTransactionReceipt(txHash string) (sdk.TransactionReceipt, error)

	GetLogs(from, to uint64, address string, topics []string) ([]sdk.IReceiptLog, error)
}
