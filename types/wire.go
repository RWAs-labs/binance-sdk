package types

import (
	"github.com/tendermint/go-amino"

	ntypes "gitlab.com/thorchain/binance-sdk/common/types"
	"gitlab.com/thorchain/binance-sdk/types/tx"
)

func NewCodec() *amino.Codec {
	cdc := amino.NewCodec()

	ntypes.RegisterWire(cdc)
	tx.RegisterCodec(cdc)
	return cdc
}
