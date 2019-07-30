package structs

import (
	"github.com/pmker/onegw/oneplus/backend/sdk"
	"github.com/pmker/onegw/oneplus/meshdb"
)

type RemovableBlock struct {
	sdk.Block
	IsRemoved bool
}

func NewRemovableBlock(block sdk.Block, isRemoved bool) *RemovableBlock {
	return &RemovableBlock{
		block,
		isRemoved,
	}
}

type TxAndReceipt struct {
	Tx      sdk.Transaction
	Receipt sdk.TransactionReceipt
}

type RemovableTxAndReceipt struct {
	*TxAndReceipt
	IsRemoved bool
	TimeStamp uint64
}


type EventType int

const (
	Added EventType = iota
	Removed
)

// Event describes a gateway event emitted by a Watcher
type Event struct {
	Type        EventType
	BlockHeader *meshdb.MiniHeader
}

type RemovableReceiptLog struct {
	sdk.IReceiptLog
	IsRemoved bool
}

func NewRemovableTxAndReceipt(tx sdk.Transaction, receipt sdk.TransactionReceipt, removed bool, timeStamp uint64) *RemovableTxAndReceipt {
	return &RemovableTxAndReceipt{
		&TxAndReceipt{
			tx,
			receipt,
		},
		removed,
		timeStamp,
	}
}

type RemovableTx struct {
	sdk.Transaction
	IsRemoved bool
}

func NewRemovableTx(tx sdk.Transaction, removed bool) RemovableTx {
	return RemovableTx{
		tx,
		removed,
	}
}

//
//type RemovableReceipt struct {
//	sdk.TransactionReceipt
//	IsRemoved bool
//}
//
//func NewRemovableReceipt(receipt sdk.TransactionReceipt, removed bool) RemovableReceipt {
//	return RemovableReceipt{
//		receipt,
//		removed,
//	}
//}
