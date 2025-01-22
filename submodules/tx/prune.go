package tx

import (
	"context"

	"cosmossdk.io/collections"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (sub TxSubmodule) prune(ctx context.Context, minHeight int64) error {
	// clear txhashesBySequenceMap using sequenceByHeightMap
	var sequence uint64
	rnHeight := new(collections.Range[int64]).StartInclusive(1).EndInclusive(minHeight)
	err := sub.sequenceByHeightMap.Walk(ctx, rnHeight, func(key int64, seq uint64) (bool, error) {
		sequence = seq
		return false, nil
	})
	if err != nil {
		return err
	}

	if err = sub.sequenceByHeightMap.Clear(ctx, rnHeight); err != nil {
		return err
	}

	rnSeq := new(collections.Range[uint64]).EndInclusive(sequence)
	if err = sub.txhashesBySequenceMap.Clear(ctx, rnSeq); err != nil {
		return err
	}

	// clear txhashesByAccountMap using accountSequenceByHeightMap
	accountSequenceMap := make(map[string]uint64)
	rnTriple := collections.NewPrefixUntilTripleRange[int64, sdk.AccAddress, uint64](minHeight)
	err = sub.accountSequenceByHeightMap.Walk(ctx, rnTriple, func(key collections.Triple[int64, sdk.AccAddress, uint64], value bool) (bool, error) {
		accountSequenceMap[key.K2().String()] = key.K3()
		return false, nil
	})
	if err != nil {
		return err
	}

	if err = sub.accountSequenceByHeightMap.Clear(ctx, rnTriple); err != nil {
		return err
	}

	for addr, seq := range accountSequenceMap {
		acc, _ := sdk.AccAddressFromBech32(addr)
		rnPair := collections.NewPrefixedPairRange[sdk.AccAddress, uint64](acc).EndInclusive(seq)
		if err = sub.txhashesByAccountMap.Clear(ctx, rnPair); err != nil {
			return err
		}
	}

	// clear everything else
	// do not clear sequence and accountSequenceMap as they merely track sequence numbers
	var txHashes []string
	rnPair := collections.NewPrefixUntilPairRange[int64, uint64](minHeight)
	err = sub.txhashesByHeightMap.Walk(ctx, rnPair, func(key collections.Pair[int64, uint64], txHash string) (bool, error) {
		txHashes = append(txHashes, txHash)
		return false, nil
	})
	if err != nil {
		return err
	}

	if err = sub.txhashesByHeightMap.Clear(ctx, rnPair); err != nil {
		return err
	}

	for _, txHash := range txHashes {
		if err := sub.txMap.Remove(ctx, txHash); err != nil {
			return err
		}
	}

	return nil
}
