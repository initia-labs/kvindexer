package block

import (
	"github.com/cometbft/cometbft/abci/types"
)

func toExecTxResultJSON(responseDeliverTx *types.ExecTxResult) *types.ExecTxResult {
	result := &types.ExecTxResult{}
	result.GasUsed = responseDeliverTx.GasUsed
	result.GasWanted = responseDeliverTx.GasWanted

	return result
}
