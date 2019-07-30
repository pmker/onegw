package plugin

import (
 	"github.com/pmker/onegw/oneplus/watch/structs"
)

type IEventPlugin interface {
	AcceptEvent(event *structs.Event)
}

type EventHashPlugin struct {
	callback func(event *structs.Event)
}

func (p EventHashPlugin) AcceptEvent(event *structs.Event) {
	if p.callback != nil {
		p.callback(event)
	}
}

func NewEventHashPlugin(callback func(event *structs.Event)) *EventHashPlugin {
	return &EventHashPlugin{
		callback: callback,
	}
}
//
//type TxPlugin struct {
//	callback func(tx structs.RemovableTx)
//}
//
//func (p TxPlugin) AcceptTx(transaction structs.RemovableTx) {
//	if p.callback != nil {
//		p.callback(transaction)
//	}
//}
//
//func NewTxPlugin(callback func(tx structs.RemovableTx)) TxPlugin {
//	return TxPlugin{
//		callback: callback,
//	}
//}
