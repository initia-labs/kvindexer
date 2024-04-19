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
	Extention interface{} `json:"extension"`
}

type ContractInfo struct {
	Name   string `json:"name"`
	Symbol string `json:"symbol"`
}

// wasm-specifics

type Minter struct {
	Minter string `json:"minter"`
}

type NumTokens struct {
	Count int64 `json:"count"`
}

type NftInfo struct {
	TokenUri  string `json:"token_uri"`
	Extension string `json:"extension"`
}

type OwnerOf struct {
	Owner     string     `json:"owner"`
	Approvals []Approval `json:"approvals"`
}

type Approval struct {
	Spender    string `json:"spender"`
	Expiration uint64 `json:"expiration"` // FIXME: height? timestamp?
}
