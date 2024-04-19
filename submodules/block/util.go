package block

import (
	"github.com/cometbft/cometbft/abci/types"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types/tx"
)

func parseTx(cdc codec.Codec, txBytes []byte) (*sdk.Tx, error) {
	tx := sdk.Tx{}
	err := cdc.Unmarshal(txBytes, &tx)
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
