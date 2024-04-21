package types

// internal use only: struct from move event
type NftMintAndBurnEventData struct {
	Collection string `json:"collection"`
	Index      string `json:"index"`
	Nft        string `json:"nft"`
}

// internal use only: struct from move event
type NftTransferEventData struct {
	Object string `json:"object"`
	From   string `json:"from"`
	To     string `json:"to"`
}

// internal use only: struct from move event
type MutationEventData struct {
	Nft              string `json:"nft,omitempty"`
	Collection       string `json:"collection,omitempty"`
	MutatedFieldName string `json:"mutated_field_name"`
	OldValue         string `json:"old_value,omitempty"`
	NewValue         string `json:"new_value,omitempty"`
}
