package block

import (
	"context"
	"fmt"

	"cosmossdk.io/errors"
	abci "github.com/cometbft/cometbft/abci/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/initia-labs/kvindexer/module/keeper"
	"github.com/initia-labs/kvindexer/submodule/block/types"
)

func collectBlock(k *keeper.Keeper, ctx context.Context, req abci.RequestFinalizeBlock, res abci.ResponseFinalizeBlock) error {
	var block types.Block

	block.ChainId = k.GetChainId()
	block.Height = req.Height
	block.Hash = fmt.Sprintf("%X", req.Hash)
	block.Timestamp = req.Time

	validator, found := k.OPChildKeeper.GetValidatorByConsAddr(ctx, req.ProposerAddress)
	if !found {
		return fmt.Errorf("cannot find valoper address by consensus address:%s", string(req.ProposerAddress))
	}
	block.Proposer = &types.Proposer{
		Moniker:         validator.Moniker,
		OperatorAddress: validator.OperatorAddress,
	}

	var feeCoins sdk.Coins
	for _, txBytes := range req.Txs {
		tx, err := parseTx(k, txBytes)
		if err != nil {
			return err
		}
		f := tx.GetFee()
		feeCoins = feeCoins.Add(f...)
	}
	block.TotalFee = feeCoins

	block.TxCount = int64(len(req.Txs))

	block.GasUsed = 0
	block.GasWanted = 0
	for tx := range req.Txs {
		res := toExecTxResultJSON(res.TxResults[tx])
		block.GasUsed += res.GasUsed
		block.GasWanted += res.GasWanted
	}

	if req.Height > 1 {
		prevBlock, err := blockByHeight.Get(ctx, req.Height-1)
		if err != nil {
			return err
		}
		block.BlockTime = req.Time.Sub(prevBlock.Timestamp).Milliseconds()
	}

	if err := blockByHeight.Set(ctx, req.Height, block); err != nil {
		return errors.Wrap(err, "failed to set block by height")
	}

	return nil
}
