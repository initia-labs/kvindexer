package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type CollectedTokenPair struct {
	L1   string `json:"l1_denom"`
	L2   string `json:"l2_denom"`
	Path string `json:"path"`
}

type CollectedNftPair struct {
	L1   string `json:"l1_collection"`
	L2   string `json:"l2_collection"`
	Path string `json:"path"`
}

type TokenPairsResponse struct {
	TokenPairs []struct {
		L1Denom string `json:"l1_denom"`
		L2Denom string `json:"l2_denom"`
	} `json:"token_pairs"`
	Pagination struct {
		NextKey string `json:"next_key"`
		Total   string `json:"total"`
	} `json:"pagination"`
}

type MetadataResource struct {
	Resource struct {
		Address     string `json:"address"`
		StructTag    string `json:"struct_tag"`
		MoveResource string `json:"move_resource"`
		RawBytes     []byte `json:"raw_bytes"`
	}
}

type MoveResource struct {
	Type string                 `json:"type"`
	Data map[string]interface{} `json:"data"`
}

type NftClassData struct {
	Description string `json:"description"`
	Name        string `json:"name"`
}

type PacketData struct {
	ClassData string         `json:"class_data"`
	ClassId   string         `json:"class_id"`
	ClassUri  string         `json:"class_uri"`
	Receiver  sdk.AccAddress `json:"receiver"`
	Sender    sdk.AccAddress `json:"sender"`
	//TokenData []string       `json:"tokenData"`
	//TokenIds  []string       `json:"tokenIds"`
	//TokenUris []string       `json:"tokenUris"`
}

type ClassTrace struct {
	TraceHash string `json:"trace_hash"`
	ClassId   string `json:"class_id"`
	//MsgIndex  int    `json:"msg_index"`
}
