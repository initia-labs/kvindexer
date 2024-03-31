package types

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
		Adddress     string `json:"address"`
		StructTag    string `json:"struct_tag"`
		MoveResource string `json:"move_resource"`
		RawBytes     []byte `json:"raw_bytes"`
	}
}

type MoveResource struct {
	Type string                 `json:"type"`
	Data map[string]interface{} `json:"data"`
}
