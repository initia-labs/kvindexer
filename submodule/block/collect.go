package block

import (
	"context"
	"encoding/json"
	"fmt"

	"cosmossdk.io/errors"
	abci "github.com/cometbft/cometbft/abci/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/initia-labs/kvindexer/config"
	"github.com/initia-labs/kvindexer/module/keeper"
	"github.com/initia-labs/kvindexer/submodule/block/types"
)

func collectBlock(k *keeper.Keeper, ctx context.Context, req abci.RequestFinalizeBlock, res abci.ResponseFinalizeBlock, cfg config.SubmoduleConfig) error {
	var b types.Block

	b.ChainId = k.GetChainId()
	b.Height = req.Height
	b.Hash = fmt.Sprintf("%X", req.Hash)
	b.Timestamp = req.Time

	validator, found := k.OPChildKeeper.GetValidatorByConsAddr(ctx, req.ProposerAddress)
	if !found {
		return fmt.Errorf("cannot find valoper address by consensus address:%s", string(req.ProposerAddress))
	}
	b.Proposer = validator.Moniker

	var feeCoins sdk.Coins
	for _, txBytes := range req.Txs {
		tx, err := parseTx(k, txBytes)
		if err != nil {
			return err
		}
		f := tx.GetFee()
		feeCoins = feeCoins.Add(f...)
	}
	b.TotalFee = feeCoins

	b.TxCount = int64(len(req.Txs))

	b.GasUsed = 0
	b.GasWanted = 0
	for tx := range req.Txs {
		res := toExecTxResultJSON(res.TxResults[tx])
		b.GasUsed += res.GasUsed
		b.GasWanted += res.GasWanted
	}

	if req.Height > 1 {
		pb, err := blockByHeight.Get(ctx, uint64(req.Height-1))
		if err != nil {
			return err
		}

		var prevBlock types.Block
		err = json.Unmarshal(pb, &prevBlock)
		if err != nil {
			return err
		}
		b.BlockTime = req.Time.Sub(prevBlock.Timestamp).Milliseconds()
	}

	block, err := json.Marshal(b)
	if err != nil {
		return fmt.Errorf("connot marshal block")
	}

	if err := blockByHeight.Set(ctx, uint64(req.Height), block); err != nil {
		return errors.Wrap(err, "failed to set block by height")
	}

	return nil
}
