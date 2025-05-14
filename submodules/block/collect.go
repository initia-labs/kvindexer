package block

import (
	"context"
	"fmt"

	"cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"

	abci "github.com/cometbft/cometbft/abci/types"

	"github.com/initia-labs/kvindexer/submodules/block/types"
)

func (sub BlockSubmodule) finalizeBlock(ctx context.Context, req abci.RequestFinalizeBlock, res abci.ResponseFinalizeBlock) error {
	sub.Logger(ctx).Debug("finalizeBlock", "submodule", types.SubmoduleName, "txs", len(req.Txs), "height", req.Height)

	if err := sub.collectBlock(ctx, req, res); err != nil {
		return err
	}

	return nil
}

func (sub BlockSubmodule) collectBlock(ctx context.Context, req abci.RequestFinalizeBlock, res abci.ResponseFinalizeBlock) error {
	var block types.Block

	block.ChainId = sdk.UnwrapSDKContext(ctx).ChainID()
	block.Height = req.Height
	block.Hash = fmt.Sprintf("%X", req.Hash)
	block.Timestamp = req.Time

	validator, found := sub.opChildKeeper.GetValidatorByConsAddr(ctx, req.ProposerAddress)
	if !found {
		return fmt.Errorf("cannot find valoper address by consensus address:%s", string(req.ProposerAddress))
	}
	block.Proposer = &types.Proposer{
		Moniker:         validator.Moniker,
		OperatorAddress: validator.OperatorAddress,
	}

	var feeCoins sdk.Coins
	for _, txBytes := range req.Txs {
		tx, err := sub.parseTx(txBytes)
		if err != nil {
			return err
		}
		feeTx, ok := tx.(sdk.FeeTx)
		if !ok {
			sub.Logger(ctx).Debug("not a fee tx", "tx", tx)
		}
		f := feeTx.GetFee()
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
		prevBlock, err := sub.blockByHeight.Get(ctx, req.Height-1)
		if err != nil {
			sub.Logger(ctx).Warn("failed to get previous block", "error", err, "height", req.Height-1)
			block.BlockTime = 0
		} else {
			block.BlockTime = req.Time.Sub(prevBlock.Timestamp).Milliseconds()
		}
	}
	if err := sub.blockByHeight.Set(ctx, req.Height, block); err != nil {
		return errors.Wrap(err, "failed to set block by height")
	}

	return nil
}

func (sub BlockSubmodule) parseTx(txBytes []byte) (sdk.Tx, error) {
	tx, err := sub.encodingConfig.TxConfig.TxDecoder()(txBytes)
	if err != nil {
		return nil, err
	}
	return tx, nil
}
