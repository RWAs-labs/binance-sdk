package transaction

import (
	"fmt"

	"gitlab.com/thorchain/binance-sdk/types/msg"
	"gitlab.com/thorchain/binance-sdk/types/tx"
)

type FreezeTokenResult struct {
	tx.TxCommitResult
}

func (c *client) FreezeToken(symbol string, amount int64, sync bool, options ...Option) (*FreezeTokenResult, error) {
	if symbol == "" {
		return nil, fmt.Errorf("Freeze token symbol can'c be empty ")
	}
	fromAddr := c.keyManager.GetAddr()

	freezeMsg := msg.NewFreezeMsg(
		fromAddr,
		symbol,
		amount,
	)
	commit, err := c.broadcastMsg(freezeMsg, sync, options...)
	if err != nil {
		return nil, err
	}

	return &FreezeTokenResult{*commit}, nil

}
