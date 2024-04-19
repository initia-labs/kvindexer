package types

import (
	abci "github.com/cometbft/cometbft/abci/types"
	nfttypes "github.com/initia-labs/kvindexer/internal/nft/types"
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
	Type string         `json:"type"`
	Nft  nfttypes.Token `json:"data"`
	// from here is additional fields, not original collection data
	//CollectionAddr string `json:"collection_addr,omitempty"`
	//ObjectAddr     string `json:"object_addr,omitempty"`
}
