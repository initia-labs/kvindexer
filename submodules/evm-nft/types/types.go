package types

import (
	abci "github.com/cometbft/cometbft/abci/types"
	nfttypes "github.com/initia-labs/kvindexer/nft/types"
)

type EventWithAttributeMap struct {
	*abci.Event
	AttributesMap map[string]string
}

// internal use only: struct from move resource
type CollectionResource struct {
	Type       string              `json:"type,omitempty"`
	Collection nfttypes.Collection `json:"data"`
	// from here is additional fields, not original collection data
	//ObjectAddr string `json:"object_addr,omitempty"`
}

// internal use only: struct from move resource
type NftResource struct {
	TokenUri  string      `json:"token_uri"`
	Extension interface{} `json:"extension"`
}

type ContractInfo struct {
	Name   string `json:"name"`
	Symbol string `json:"symbol"`
}

// evm-specifics

// NOTE: remove this once getCollectionMinter() removed
type Minter struct {
	Minter string `json:"minter"`
}

// NOTE: remove this once getCollectionNumTokens() removed
type NumTokens struct {
	Count int64 `json:"count"`
}
