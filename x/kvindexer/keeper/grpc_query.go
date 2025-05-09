package keeper

import (
	"context"

	"github.com/initia-labs/kvindexer/x/kvindexer/types"
)

var _ types.QueryServer = (*Querier)(nil)

type Querier struct {
	*Keeper
}

// VMType implements types.QueryServer.
func (q Querier) VMType(context.Context, *types.QueryVMTypeRequest) (*types.QueryVMTypeResponse, error) {
	return &types.QueryVMTypeResponse{Vmtype: q.vmType}, nil
}

// Versions implements types.QueryServer.
func (q Querier) Versions(context.Context, *types.QueryVersionRequest) (*types.QueryVersionResponse, error) {
	res := []*types.SubmoduleVersion{}
	for _, sm := range q.submodules {
		res = append(res, &types.SubmoduleVersion{
			Submodule: sm.Name(),
			Version:   sm.Version(),
		})
	}
	return &types.QueryVersionResponse{Versions: res}, nil
}

// NewQuerier return new Querier instance
func NewQuerier(k *Keeper) Querier {
	return Querier{k}
}
