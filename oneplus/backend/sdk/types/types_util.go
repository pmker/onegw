package types

import (
	"math/big"

	"github.com/pmker/onegw/oneplus/backend/utils"
)

func HexToAddress(s string) Address { return BytesToAddress(utils.Hex2Bytes(s)) }
func HexToHash(s string) Hash       { return BytesToHash(utils.Hex2Bytes(s)) }

func BytesToAddress(b []byte) Address {
	var a Address
	a.SetBytes(b)
	return a
}

func BytesToHash(b []byte) Hash {
	var h Hash
	h.SetBytes(b)
	return h
}

func BigToHash(b *big.Int) Hash { return BytesToHash(b.Bytes()) }
