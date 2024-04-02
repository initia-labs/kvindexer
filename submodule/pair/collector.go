package pair

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"cosmossdk.io/collections"
	cosmoserr "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/gofiber/fiber/v2"

	exportedibc "github.com/cosmos/ibc-go/v8/modules/core/exported"
	ibctm "github.com/cosmos/ibc-go/v8/modules/light-clients/07-tendermint"
	"github.com/initia-labs/kvindexer/module/keeper"
	"github.com/initia-labs/kvindexer/submodule/pair/types"
)

const (
	ibcTransferPort     = "transfer"
	ibcNftTransferPort  = "nft-transfer"
	collectionStructTag = "0x1::collection::Collection"

	queryOpTokenFmt    = "%s/opinit/ophost/v1/bridges/%s/token_pairs"
	queryCollectionFmt = "%s/initia/move/v1/accounts/%s/resources/by_struct_tag?struct_tag=%s"
	paginationFmt      = "%s?pagination.key=%s"
)

func collectIbcTokenPairs(k *keeper.Keeper, ctx context.Context) (err error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	val := croncfg.ibcChannels.Load()
	if val == nil {
		return nil
	}
	ibcChannels := val.([]string)
	if len(ibcChannels) == 0 {
		return nil
	}

	traces := k.TransferKeeper.GetAllDenomTraces(sdkCtx)

	for _, ibcChannel := range ibcChannels {
		for _, trace := range traces {

			if trace.Path != fmt.Sprintf("%s/%s", ibcTransferPort, ibcChannel) {
				continue
			}

			prevDenom, err := fungiblePairsMap.Get(ctx, trace.IBCDenom())
			if err != nil && !cosmoserr.IsOf(err, collections.ErrNotFound) {
				continue
			}
			// prevDenom should be empty string if not found, or already set
			if prevDenom == trace.BaseDenom {
				continue
			}

			err = fungiblePairsMap.Set(ctx, trace.IBCDenom(), trace.BaseDenom)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func collecOpTokenPairs(k *keeper.Keeper, ctx context.Context) (err error) {
	k.BankKeeper.IterateTotalSupply(ctx, func(supply sdk.Coin) bool {

		if !strings.HasPrefix(supply.Denom, "l2/") {
			return false
		}

		l1denomAny, ok := fungiblePairsFromL1.Load(supply.Denom)
		if !ok {
			// not found in L1, just ignore and continue: it'll be processed by next iteration
			return false
		}
		l1denom, ok := l1denomAny.(string)
		if !ok || l1denom == "" {
			return false
		}

		// prevDenom should be empty string if not found, or already set
		prevDenom, err := fungiblePairsMap.Get(ctx, supply.Denom)
		if err != nil && !cosmoserr.IsOf(err, collections.ErrNotFound) {
			return true
		}

		if prevDenom == l1denomAny {
			return false
		}
		err = fungiblePairsMap.Set(ctx, supply.Denom, l1denom)
		return err != nil
	})

	return nil
}

// get OPinit token pairs from L1 and insert them into the syncmap
// data in syncmap will be used by collecOpTokenPairs()
func collectOpTokenPairsFromL1(client *fiber.Client, cfg *cronConfig) (err error) {
	ctps := []types.CollectedTokenPair{}
	nextKey := ""

	bridgeId := fmt.Sprintf("%d", cfg.bridgeId)

	for {
		queryStr := fmt.Sprintf(queryOpTokenFmt, cfg.l1LcdUrl, bridgeId)
		if nextKey != "" {
			queryStr = fmt.Sprintf(paginationFmt, queryStr, nextKey)
		}
		code, body, errs := client.Get(queryStr).Bytes()
		if err = errors.Join(errs...); err != nil {
			return err
		}
		if fiber.StatusOK != code {
			return fmt.Errorf("http response: %d", code)
		}

		var response types.TokenPairsResponse
		if err = json.Unmarshal(body, &response); err != nil {
			return err
		}

		for _, pair := range response.TokenPairs {
			ctps = append(ctps, types.CollectedTokenPair{
				L1:   pair.L1Denom,
				L2:   pair.L2Denom,
				Path: bridgeId,
			})
		}

		if response.Pagination.NextKey == "null" || response.Pagination.NextKey == "" {
			break
		}
		nextKey = response.Pagination.NextKey
	}

	for _, ctp := range ctps {
		fungiblePairsFromL1.Store(ctp.L2, ctp.L1)
	}

	return nil
}

func collectNftTokensFromL2(k *keeper.Keeper, ctx context.Context) (err error) {
	ibcNftChannels := croncfg.ibcNftChannels.Load().([]string)
	if len(ibcNftChannels) == 0 {
		return nil
	}

	traces, err := k.NftTransferKeeper.GetAllClassTraces(ctx)
	if err != nil {
		return err
	}
	for _, ibcNftChannel := range ibcNftChannels {
		for _, trace := range traces {
			// only from allowed channel
			if trace.Path != fmt.Sprintf("%s/%s", ibcNftTransferPort, ibcNftChannel) {
				continue
			}

			// only gather move based class
			splitted := strings.Split(trace.BaseClassId, "/")
			if splitted[0] != "move" || len(splitted) < 2 {
				continue
			}

			classId := trace.IBCClassId()
			l2collAddr := "0x" + splitted[1]
			l1collName, ok := nonFungiblePairsFromL2.Load(classId)
			if !ok {
				// insert them into nft syncmap if not exists
				nonFungiblePairsFromL2.Store(classId, l2collAddr)
			} else {
				if l1collName == "" || l1collName == l2collAddr {
					continue
				}
				_, err := nonFungiblePairsMap.Get(ctx, classId)
				if !cosmoserr.IsOf(err, collections.ErrNotFound) || err == nil {
					continue
				}

				err = nonFungiblePairsMap.Set(ctx, classId, l1collName.(string))
				if err != nil {
					return err
				}

				nonFungiblePairsFromL2.Store(classId, "")
			}
		}
	}

	return nil
}

// get OPinit token pairs from L1 and insert them into the syncmap
// data in syncmap will be used by collecOpTokenPairs()
func collectNftTokenPairsFromL1(client *fiber.Client, cfg *cronConfig) (err error) {
	if cfg.l1LcdUrl == "" {
		return errors.New("l1LcdUrl is not set")
	}
	val := cfg.ibcNftChannels.Load()
	if val == nil {
		return nil
	}
	ibcNftChannels := val.([]string)
	if len(ibcNftChannels) == 0 {
		return nil
	}

	nonFungiblePairsFromL2.Range(func(key, value interface{}) bool {
		ibcClassId := key.(string)
		l2CollAddr := value.(string)

		// already set
		if l2CollAddr == "" {
			return false
		}

		// if it has value, it means it's already set
		collectionName, err := getCollectionNameFromL1(client, cfg, l2CollAddr)
		if err != nil {
			// just continue to next iteration. it'll be processed by next iteration
			return false
		}

		nonFungiblePairsFromL2.Store(ibcClassId, collectionName)
		return false
	})

	return nil
}

// query collection name from L1
func getCollectionNameFromL1(client *fiber.Client, cfg *cronConfig, addr string) (collectionName string, err error) {
	queryStr := fmt.Sprintf(queryCollectionFmt, cfg.l1LcdUrl, addr, collectionStructTag)
	code, body, errs := client.Get(queryStr).Bytes()
	if err = errors.Join(errs...); err != nil {
		return collectionName, err
	}
	if fiber.StatusOK != code {
		return collectionName, fmt.Errorf("http response: %d", code)
	}

	var response types.MetadataResource
	if err = json.Unmarshal(body, &response); err != nil {
		return collectionName, err
	}

	var moveResource types.MoveResource
	if err = json.Unmarshal([]byte(response.Resource.MoveResource), &moveResource); err != nil {
		return collectionName, err
	}

	name, ok := moveResource.Data["name"].(string)
	if !ok {
		return collectionName, fmt.Errorf("move resource: %+v", moveResource.Data)
	}

	return name, nil
}

func updateIBCChannels(k *keeper.Keeper, ctx context.Context) error {

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	chKeeper := k.IBCKeeper.ChannelKeeper

	// get all channels
	channels := chKeeper.GetAllChannels(sdkCtx)

	ibcChannels := []string{}
	ibcNftChannels := []string{}
	for _, channel := range channels {
		if channel.PortId != ibcTransferPort && channel.PortId != ibcNftTransferPort {
			continue
		}

		_ /*clientId*/, cs, err := chKeeper.GetChannelClientState(sdkCtx, channel.PortId, channel.ChannelId)
		if err != nil {
			k.Logger(ctx).Warn("GetChannelClientState", "error", err)
		}
		counterpartyChainId := getChainIdFromClientState(cs)
		if counterpartyChainId == "" {
			k.Logger(ctx).Warn("channel id is nil")
			continue
		}
		if counterpartyChainId != croncfg.l1ChainId {
			continue
		}
		if channel.PortId == ibcTransferPort {
			ibcChannels = append(ibcChannels, channel.ChannelId)
		} else {
			ibcNftChannels = append(ibcNftChannels, channel.ChannelId)
		}

	}

	croncfg.ibcChannels.Store(ibcChannels)
	croncfg.ibcNftChannels.Store(ibcNftChannels)

	return nil
}

func getChainIdFromClientState(csi exportedibc.ClientState) string {
	if csi == nil {
		return ""
	}
	cs, ok := csi.(*ibctm.ClientState)
	if !ok {
		return ""
	}
	return cs.ChainId
}
