package block

import (
	"context"
	"fmt"

	"cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"

	abci "github.com/cometbft/cometbft/abci/types"

	"github.com/initia-labs/kvindexer/submodules/block/types"
)

func (bs BlockSubmodule) finalizeBlock(ctx context.Context, req abci.RequestFinalizeBlock, res abci.ResponseFinalizeBlock) error {
	bs.Logger(ctx).Debug("finalizeBlock", "submodule", types.SubmoduleName, "txs", len(req.Txs), "height", req.Height)

	if err := bs.collectBlock(ctx, req, res); err != nil {
		return err
	}

	return nil
}

func (bs BlockSubmodule) collectBlock(ctx context.Context, req abci.RequestFinalizeBlock, res abci.ResponseFinalizeBlock) error {
	var block types.Block

	block.ChainId = sdk.UnwrapSDKContext(ctx).ChainID()
	block.Height = req.Height
	block.Hash = fmt.Sprintf("%X", req.Hash)
	block.Timestamp = req.Time

	validator, found := bs.opChildKeeper.GetValidatorByConsAddr(ctx, req.ProposerAddress)
	if !found {
		return fmt.Errorf("cannot find valoper address by consensus address:%s", string(req.ProposerAddress))
	}
	block.Proposer = &types.Proposer{
		Moniker:         validator.Moniker,
		OperatorAddress: validator.OperatorAddress,
	}

	var feeCoins sdk.Coins
	for _, txBytes := range req.Txs {
		tx, err := parseTx(bs.cdc, txBytes)
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
		prevBlock, err := bs.blockByHeight.Get(ctx, req.Height-1)
		if err != nil {
			bs.Logger(ctx).Warn("failed to get previous block", "error", err, "height", req.Height-1)
			block.BlockTime = 0
		} else {
			block.BlockTime = req.Time.Sub(prevBlock.Timestamp).Milliseconds()
		}
	}
	if err := bs.blockByHeight.Set(ctx, req.Height, block); err != nil {
		return errors.Wrap(err, "failed to set block by height")
	}

	return nil
}
