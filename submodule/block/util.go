package block

import (
	"encoding/json"

	"github.com/cometbft/cometbft/abci/types"
	sdk "github.com/cosmos/cosmos-sdk/types/tx"
	"github.com/initia-labs/kvindexer/module/keeper"
	blocktypes "github.com/initia-labs/kvindexer/submodule/block/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func parseTx(k *keeper.Keeper, txBytes []byte) (*sdk.Tx, error) {
	tx := sdk.Tx{}
	err := k.GetCodec().Unmarshal(txBytes, &tx)
	if err != nil {
		return nil, err
	}
	return &tx, nil
}

func toExecTxResultJSON(responseDeliverTx *types.ExecTxResult) *types.ExecTxResult {
	result := &types.ExecTxResult{}
	result.GasUsed = responseDeliverTx.GasUsed
	result.GasWanted = responseDeliverTx.GasWanted

	return result
}

func makeBlock(b []byte) (blocktypes.Block, error) {
	var block blocktypes.Block
	err := json.Unmarshal(b, &block)
	if err != nil {
		return block, status.Error(codes.Internal, err.Error())
	}
	return block, nil
}
