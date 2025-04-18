package util

import "github.com/cosmos/cosmos-sdk/types/query"

// ValidatePageRequest validates pagination
// - suppress limit to 100
// - force off to count_total
func ValidatePageRequest(pageReq *query.PageRequest) {
	if pageReq == nil {
		return
	}
	if pageReq.Limit > 100 {
		pageReq.Limit = 100
	}
	pageReq.CountTotal = false
}
