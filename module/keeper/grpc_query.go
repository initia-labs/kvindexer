package keeper

import (
	"context"

	"github.com/initia-labs/kvindexer/module/types"
)

var _ types.QueryServer = (*Querier)(nil)

type Querier struct {
	*Keeper
}

// VMType implements types.QueryServer.
func (q Querier) VMType(context.Context, *types.QueryVMTypeRequest) (*types.QueryVMTypeResponse, error) {
	return &types.QueryVMTypeResponse{Vmtype: q.VMKeeper.GetVMType()}, nil
}

// Versions implements types.QueryServer.
func (q Querier) Versions(context.Context, *types.QueryVersionRequest) (*types.QueryVersionResponse, error) {
	res := []*types.SubmoduleVersion{}
	for name, sm := range q.submodules {
		if name == "" {
			name = "unknown"
		}
		res = append(res, &types.SubmoduleVersion{
			Submodule: name,
			Version:   sm.Version,
		})
	}
	return &types.QueryVersionResponse{Versions: res}, nil
}

// NewQuerier return new Querier instance
func NewQuerier(k *Keeper) Querier {
	return Querier{k}
}
