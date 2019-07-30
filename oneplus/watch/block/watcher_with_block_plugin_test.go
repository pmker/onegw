package watch

import (
	"context"
	"fmt"
	"github.com/pmker/onegw/oneplus/watch/plugin"
	"github.com/pmker/onegw/oneplus/watch/structs"
	"testing"
)

func TestNewBlockNumPlugin(t *testing.T) {
	api := "https://mainnet.infura.io/v3/19d753b2600445e292d54b1ef58d4df4"
	w := NewHttpBasedEthWatcher(context.Background(), api)

	w.RegisterBlockPlugin(plugin.NewBlockNumPlugin(func(i uint64, b bool) {
		fmt.Println(">>", i, b)
	}))

	w.RunTillExit()
}

func TestSimpleBlockPlugin(t *testing.T) {
	api := "https://mainnet.infura.io/v3/19d753b2600445e292d54b1ef58d4df4"
	w := NewHttpBasedEthWatcher(context.Background(), api)

	w.RegisterBlockPlugin(plugin.NewSimpleBlockPlugin(func(block *structs.RemovableBlock) {
		fmt.Println(">>", block, block.IsRemoved)
	}))

	w.RunTillExit()
}
