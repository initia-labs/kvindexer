package keeper

import (
	"context"

	"github.com/initia-labs/kvindexer/module/types"
)

var _ types.QueryServer = (*Querier)(nil)

type Querier struct {
	*Keeper
}

// Version implements types.QueryServer.
func (q Querier) Version(context.Context, *types.QueryVersionRequest) (*types.QueryVersionResponse, error) {
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
