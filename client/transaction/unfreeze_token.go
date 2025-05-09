package transaction

import (
	"fmt"

	"gitlab.com/thorchain/binance-sdk/types/msg"
	"gitlab.com/thorchain/binance-sdk/types/tx"
)

type UnfreezeTokenResult struct {
	tx.TxCommitResult
}

func (c *client) UnfreezeToken(symbol string, amount int64, sync bool, options ...Option) (*UnfreezeTokenResult, error) {
	if symbol == "" {
		return nil, fmt.Errorf("Freeze token symbol can'c be empty ")
	}
	fromAddr := c.keyManager.GetAddr()

	unfreezeMsg := msg.NewUnfreezeMsg(
		fromAddr,
		symbol,
		amount,
	)
	commit, err := c.broadcastMsg(unfreezeMsg, sync, options...)
	if err != nil {
		return nil, err
	}

	return &UnfreezeTokenResult{*commit}, nil

}
